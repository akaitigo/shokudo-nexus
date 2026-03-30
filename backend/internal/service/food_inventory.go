// Package service はgRPCサービスの実装を提供する。
package service

import (
	"context"
	"time"
	"unicode/utf8"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/akaitigo/shokudo-nexus/backend/internal/domain"
	"github.com/akaitigo/shokudo-nexus/backend/internal/repository"

	pb "github.com/akaitigo/shokudo-nexus/backend/gen/shokudo/v1"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
	maxNameLength   = 200
	minQuantity     = 1
	maxQuantity     = 10000
)

// FoodInventoryService はFoodInventoryServiceのgRPC実装。
type FoodInventoryService struct {
	pb.UnimplementedFoodInventoryServiceServer
	repo *repository.FoodItemRepository
}

// NewFoodInventoryService は新しいFoodInventoryServiceを生成する。
func NewFoodInventoryService(repo *repository.FoodItemRepository) *FoodInventoryService {
	return &FoodInventoryService{repo: repo}
}

// CreateFoodItem は余剰食品を登録する。
func (s *FoodInventoryService) CreateFoodItem(ctx context.Context, req *pb.CreateFoodItemRequest) (*pb.CreateFoodItemResponse, error) {
	if err := validateCreateFoodItemRequest(req); err != nil {
		return nil, err
	}

	expiryDate, err := time.Parse(time.RFC3339, req.GetExpiryDate())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid expiry_date format, expected RFC3339: %v", err)
	}

	now := time.Now().UTC()
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
		return nil, status.Errorf(codes.Internal, "failed to create food item: %v", err)
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
	now := time.Now().UTC()
	if item.Status == domain.FoodItemStatusAvailable && item.IsExpired(now) {
		item.Status = domain.FoodItemStatusExpired
		item.UpdatedAt = now
		if updateErr := s.repo.Update(ctx, item); updateErr != nil {
			return nil, status.Errorf(codes.Internal, "failed to update expired status: %v", updateErr)
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
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		return nil, status.Errorf(codes.InvalidArgument, "page_size must be between 1 and %d", maxPageSize)
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

	// 消費期限チェック（オンザフライ）
	now := time.Now().UTC()
	pbItems := make([]*pb.FoodItem, 0, len(result.Items))
	for _, item := range result.Items {
		if item.Status == domain.FoodItemStatusAvailable && item.IsExpired(now) {
			item.Status = domain.FoodItemStatusExpired
			item.UpdatedAt = now
			// ベストエフォートで更新（リスト表示を遅延させない）
			_ = s.repo.Update(ctx, item)
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

func validateCreateFoodItemRequest(req *pb.CreateFoodItemRequest) error {
	if req.GetName() == "" {
		return status.Error(codes.InvalidArgument, "name is required")
	}
	if utf8.RuneCountInString(req.GetName()) > maxNameLength {
		return status.Errorf(codes.InvalidArgument, "name must be at most %d characters", maxNameLength)
	}
	if !domain.IsValidCategory(req.GetCategory()) {
		return status.Errorf(codes.InvalidArgument, "invalid category: %q", req.GetCategory())
	}
	if req.GetExpiryDate() == "" {
		return status.Error(codes.InvalidArgument, "expiry_date is required")
	}
	if req.GetQuantity() < minQuantity || req.GetQuantity() > maxQuantity {
		return status.Errorf(codes.InvalidArgument, "quantity must be between %d and %d", minQuantity, maxQuantity)
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
