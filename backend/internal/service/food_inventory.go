// Package service はgRPCサービスの実装を提供する。
package service

import (
	"context"
	"log/slog"
	"time"
	"unicode/utf8"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/akaitigo/shokudo-nexus/backend/internal/config"
	"github.com/akaitigo/shokudo-nexus/backend/internal/domain"
	"github.com/akaitigo/shokudo-nexus/backend/internal/repository"

	pb "github.com/akaitigo/shokudo-nexus/backend/gen/shokudo/v1"
)

// FoodInventoryService はFoodInventoryServiceのgRPC実装。
type FoodInventoryService struct {
	pb.UnimplementedFoodInventoryServiceServer
	repo    repository.FoodItemStore
	cfg     config.ServiceConfig
	nowFunc func() time.Time
}

// NewFoodInventoryService は新しいFoodInventoryServiceを生成する。
// 調整可能なパラメータは環境変数から読み込む。
func NewFoodInventoryService(repo repository.FoodItemStore) *FoodInventoryService {
	return &FoodInventoryService{
		repo:    repo,
		cfg:     config.LoadServiceConfig(),
		nowFunc: func() time.Time { return time.Now().UTC() },
	}
}

// CreateFoodItem は余剰食品を登録する。
func (s *FoodInventoryService) CreateFoodItem(ctx context.Context, req *pb.CreateFoodItemRequest) (*pb.CreateFoodItemResponse, error) {
	if err := validateCreateFoodItemRequest(req, s.cfg); err != nil {
		return nil, err
	}

	expiryDate, err := time.Parse(time.RFC3339, req.GetExpiryDate())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid expiry_date format, expected RFC3339: %v", err)
	}

	now := s.nowFunc()
	item := &domain.FoodItem{
		Name:       req.GetName(),
		Category:   req.GetCategory(),
		ExpiryDate: expiryDate,
		Quantity:   req.GetQuantity(),
		Unit:       req.GetUnit(),
		DonorID:    req.GetDonorId(),
		Status:     domain.FoodItemStatusAvailable,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	created, err := s.repo.Create(ctx, item)
	if err != nil {
		slog.Error("failed to create food item", "error", err)
		return nil, status.Error(codes.Internal, "failed to create food item")
	}

	return &pb.CreateFoodItemResponse{
		FoodItem: domainFoodItemToProto(created),
	}, nil
}

// GetFoodItem は指定されたIDの食品を取得する。
func (s *FoodInventoryService) GetFoodItem(ctx context.Context, req *pb.GetFoodItemRequest) (*pb.GetFoodItemResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	item, err := s.repo.Get(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	// 消費期限チェック（オンザフライ）
	now := s.nowFunc()
	if item.Status == domain.FoodItemStatusAvailable && item.IsExpired(now) {
		item.Status = domain.FoodItemStatusExpired
		item.UpdatedAt = now
		if updateErr := s.repo.Update(ctx, item); updateErr != nil {
			slog.Error("failed to update expired status", "item_id", item.ID, "error", updateErr)
			return nil, status.Error(codes.Internal, "failed to update food item status")
		}
	}

	return &pb.GetFoodItemResponse{
		FoodItem: domainFoodItemToProto(item),
	}, nil
}

// ListFoodItems は食品の一覧を取得する。
func (s *FoodInventoryService) ListFoodItems(ctx context.Context, req *pb.ListFoodItemsRequest) (*pb.ListFoodItemsResponse, error) {
	pageSize := int(req.GetPageSize())
	if pageSize <= 0 {
		pageSize = s.cfg.DefaultPageSize
	}
	if pageSize > s.cfg.MaxPageSize {
		return nil, status.Errorf(codes.InvalidArgument, "page_size must be between 1 and %d", s.cfg.MaxPageSize)
	}

	if req.GetCategoryFilter() != "" && !domain.IsValidCategory(req.GetCategoryFilter()) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid category_filter: %q", req.GetCategoryFilter())
	}

	result, err := s.repo.List(ctx, repository.ListParams{
		PageSize:       pageSize,
		PageToken:      req.GetPageToken(),
		CategoryFilter: req.GetCategoryFilter(),
	})
	if err != nil {
		return nil, err
	}

	// 消費期限チェック（オンザフライ）。
	//
	// 一覧取得時に available かつ期限切れのアイテムを検出したら、その場で expired へ
	// 永続化する。この書き込みは意図的にベストエフォートとしている:
	//   - 目的: 期限切れの反映は「次に誰かが読んだとき」で十分であり、一覧応答の
	//     レイテンシを Firestore 書き込みでブロックしたくない。
	//   - 失敗時の挙動: 書き込みに失敗しても警告ログのみでリクエストは成功させ、
	//     レスポンス上のステータスは expired として返す（表示は常に正しい）。
	//   - 許容リスク: 永続化が失敗したアイテムは、次回の Get/List で再度更新が試行される。
	//     期限切れ判定はレスポンス生成時に毎回オンザフライで行うため、DB 上のステータスが
	//     一時的に古くても表示・マッチング（RespondToFusionRequest は available のみ承認）に
	//     影響しない。リトライ機構・管理者通知は MVP では不要と判断した。
	// 設計判断の詳細は docs/design-notes.md を参照。
	now := s.nowFunc()
	pbItems := make([]*pb.FoodItem, 0, len(result.Items))
	for _, item := range result.Items {
		if item.Status == domain.FoodItemStatusAvailable && item.IsExpired(now) {
			item.Status = domain.FoodItemStatusExpired
			item.UpdatedAt = now
			// ベストエフォートで更新（リスト表示を遅延・失敗させない）。
			if updateErr := s.repo.Update(ctx, item); updateErr != nil {
				slog.Warn("failed to update expired status on list", "item_id", item.ID, "error", updateErr)
			}
		}
		pbItems = append(pbItems, domainFoodItemToProto(item))
	}

	return &pb.ListFoodItemsResponse{
		FoodItems:     pbItems,
		NextPageToken: result.NextPageToken,
		TotalCount:    result.TotalCount,
	}, nil
}

// DeleteFoodItem は食品を論理削除する。
func (s *FoodInventoryService) DeleteFoodItem(ctx context.Context, req *pb.DeleteFoodItemRequest) (*pb.DeleteFoodItemResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	if err := s.repo.Delete(ctx, req.GetId()); err != nil {
		return nil, err
	}

	return &pb.DeleteFoodItemResponse{}, nil
}

func validateCreateFoodItemRequest(req *pb.CreateFoodItemRequest, cfg config.ServiceConfig) error {
	if req.GetName() == "" {
		return status.Error(codes.InvalidArgument, "name is required")
	}
	if utf8.RuneCountInString(req.GetName()) > cfg.MaxNameLength {
		return status.Errorf(codes.InvalidArgument, "name must be at most %d characters", cfg.MaxNameLength)
	}
	if !domain.IsValidCategory(req.GetCategory()) {
		return status.Errorf(codes.InvalidArgument, "invalid category: %q", req.GetCategory())
	}
	if req.GetExpiryDate() == "" {
		return status.Error(codes.InvalidArgument, "expiry_date is required")
	}
	if req.GetQuantity() < cfg.MinQuantity || req.GetQuantity() > cfg.MaxQuantity {
		return status.Errorf(codes.InvalidArgument, "quantity must be between %d and %d", cfg.MinQuantity, cfg.MaxQuantity)
	}
	if !domain.IsValidUnit(req.GetUnit()) {
		return status.Errorf(codes.InvalidArgument, "invalid unit: %q", req.GetUnit())
	}
	return nil
}

func domainFoodItemToProto(item *domain.FoodItem) *pb.FoodItem {
	return &pb.FoodItem{
		Id:         item.ID,
		Name:       item.Name,
		Category:   item.Category,
		ExpiryDate: item.ExpiryDate.Format(time.RFC3339),
		Quantity:   item.Quantity,
		Unit:       item.Unit,
		DonorId:    item.DonorID,
		Status:     string(item.Status),
		CreatedAt:  item.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  item.UpdatedAt.Format(time.RFC3339),
	}
}
