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

// FusionService はFusionServiceのgRPC実装。
type FusionService struct {
	pb.UnimplementedFusionServiceServer
	fusionRepo   repository.FusionRequestStore
	foodItemRepo repository.FoodItemStore
	// approvalRunner は承認処理を Firestore トランザクションで実行し、
	// 複数インスタンス間でも同一在庫の二重予約を防止する。
	approvalRunner repository.ApprovalRunner
	cfg            config.ServiceConfig
}

// NewFusionService は新しいFusionServiceを生成する。
// 調整可能なパラメータは環境変数から読み込む。
func NewFusionService(
	fusionRepo repository.FusionRequestStore,
	foodItemRepo repository.FoodItemStore,
	approvalRunner repository.ApprovalRunner,
) *FusionService {
	return &FusionService{
		fusionRepo:     fusionRepo,
		foodItemRepo:   foodItemRepo,
		approvalRunner: approvalRunner,
		cfg:            config.LoadServiceConfig(),
	}
}

// CreateFusionRequest は食材融通リクエストを作成する。
func (s *FusionService) CreateFusionRequest(ctx context.Context, req *pb.CreateFusionRequestRequest) (*pb.CreateFusionRequestResponse, error) {
	if err := validateCreateFusionRequest(req, s.cfg); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	fusionReq := &domain.FusionRequest{
		RequesterShokudoID: req.GetRequesterShokudoId(),
		DesiredCategory:    req.GetDesiredCategory(),
		DesiredQuantity:    req.GetDesiredQuantity(),
		Unit:               req.GetUnit(),
		Message:            req.GetMessage(),
		Status:             domain.FusionRequestStatusPending,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	created, err := s.fusionRepo.Create(ctx, fusionReq)
	if err != nil {
		slog.Error("failed to create fusion request", "error", err)
		return nil, status.Error(codes.Internal, "failed to create fusion request")
	}

	return &pb.CreateFusionRequestResponse{
		FusionRequest: domainFusionRequestToProto(created),
	}, nil
}

// ListFusionRequests は融通リクエストの一覧を取得する。
func (s *FusionService) ListFusionRequests(ctx context.Context, req *pb.ListFusionRequestsRequest) (*pb.ListFusionRequestsResponse, error) {
	pageSize := int(req.GetPageSize())
	if pageSize <= 0 {
		pageSize = s.cfg.DefaultPageSize
	}
	if pageSize > s.cfg.MaxPageSize {
		return nil, status.Errorf(codes.InvalidArgument, "page_size must be between 1 and %d", s.cfg.MaxPageSize)
	}

	if req.GetStatusFilter() != "" {
		validStatuses := map[string]bool{
			"pending": true, "approved": true, "rejected": true, "completed": true,
		}
		if !validStatuses[req.GetStatusFilter()] {
			return nil, status.Errorf(codes.InvalidArgument, "invalid status_filter: %q", req.GetStatusFilter())
		}
	}

	result, err := s.fusionRepo.List(ctx, repository.FusionListParams{
		PageSize:     pageSize,
		PageToken:    req.GetPageToken(),
		StatusFilter: req.GetStatusFilter(),
	})
	if err != nil {
		return nil, err
	}

	pbRequests := make([]*pb.FusionRequest, 0, len(result.Requests))
	for _, r := range result.Requests {
		pbRequests = append(pbRequests, domainFusionRequestToProto(r))
	}

	return &pb.ListFusionRequestsResponse{
		FusionRequests: pbRequests,
		NextPageToken:  result.NextPageToken,
		TotalCount:     result.TotalCount,
	}, nil
}

// RespondToFusionRequest は融通リクエストに応答する（承認/拒否）。
// 承認は Firestore トランザクションで在庫予約と状態遷移をアトミックに行い、
// 複数インスタンス間でも同一在庫の二重予約を防止する。
func (s *FusionService) RespondToFusionRequest(ctx context.Context, req *pb.RespondToFusionRequestRequest) (*pb.RespondToFusionRequestResponse, error) {
	if err := validateRespondToFusionRequest(req); err != nil {
		return nil, err
	}

	now := time.Now().UTC()

	var updated *domain.FusionRequest
	var err error
	if req.GetResponse() == "APPROVED" {
		updated, err = s.approveFusionRequest(ctx, req, now)
	} else {
		updated, err = s.rejectFusionRequest(ctx, req, now)
	}
	if err != nil {
		return nil, err
	}

	return &pb.RespondToFusionRequestResponse{
		FusionRequest: domainFusionRequestToProto(updated),
	}, nil
}

// approveFusionRequest は承認処理を単一の Firestore トランザクションで実行する。
// 融通リクエストと食品アイテムの読み取り・整合性チェック・書き込みをアトミックに行い、
// 途中で在庫状態が変化した場合はトランザクションが再試行される。
func (s *FusionService) approveFusionRequest(ctx context.Context, req *pb.RespondToFusionRequestRequest, now time.Time) (*domain.FusionRequest, error) {
	var updated *domain.FusionRequest
	err := s.approvalRunner.RunApproval(ctx, func(tx repository.ApprovalTx) error {
		// 読み取りは書き込みより前にまとめて行う（Firestore トランザクションの制約）。
		fusionReq, err := tx.GetFusionRequest(req.GetFusionRequestId())
		if err != nil {
			return err
		}
		if fusionReq.Status != domain.FusionRequestStatusPending {
			return status.Errorf(codes.FailedPrecondition,
				"fusion request status is %q, only %q can be responded to",
				fusionReq.Status, domain.FusionRequestStatusPending)
		}

		foodItem, err := tx.GetFoodItem(req.GetFoodItemId())
		if err != nil {
			return err
		}
		if foodItem.Status != domain.FoodItemStatusAvailable {
			return status.Errorf(codes.FailedPrecondition,
				"food item status is %q, must be %q to approve",
				foodItem.Status, domain.FoodItemStatusAvailable)
		}
		if foodItem.Category != fusionReq.DesiredCategory {
			return status.Errorf(codes.InvalidArgument,
				"food item category %q does not match desired category %q",
				foodItem.Category, fusionReq.DesiredCategory)
		}
		if foodItem.Unit != fusionReq.Unit {
			return status.Errorf(codes.InvalidArgument,
				"food item unit %q does not match desired unit %q",
				foodItem.Unit, fusionReq.Unit)
		}
		if foodItem.Quantity < fusionReq.DesiredQuantity {
			return status.Errorf(codes.InvalidArgument,
				"food item quantity %d is less than desired quantity %d",
				foodItem.Quantity, fusionReq.DesiredQuantity)
		}

		foodItem.Status = domain.FoodItemStatusReserved
		foodItem.UpdatedAt = now
		fusionReq.Status = domain.FusionRequestStatusApproved
		fusionReq.FoodItemID = req.GetFoodItemId()
		fusionReq.ResponderShokudoID = foodItem.DonorID
		fusionReq.UpdatedAt = now

		if err := tx.SetFoodItem(foodItem); err != nil {
			return err
		}
		if err := tx.SetFusionRequest(fusionReq); err != nil {
			return err
		}
		updated = fusionReq
		return nil
	})
	if err != nil {
		// バリデーション由来の gRPC status エラー（NotFound/FailedPrecondition/InvalidArgument）は
		// そのままクライアントへ返す。それ以外（トランザクションのコミット失敗等）は Internal に変換する。
		if _, ok := status.FromError(err); ok {
			return nil, err
		}
		slog.Error("failed to approve fusion request", "fusion_request_id", req.GetFusionRequestId(), "error", err)
		return nil, status.Error(codes.Internal, "failed to approve fusion request")
	}
	return updated, nil
}

// rejectFusionRequest は拒否処理を行う。食品在庫を伴わないため競合の余地はなく、
// 楽観ロックを伴わない単純な状態更新で足りる。
func (s *FusionService) rejectFusionRequest(ctx context.Context, req *pb.RespondToFusionRequestRequest, now time.Time) (*domain.FusionRequest, error) {
	fusionReq, err := s.fusionRepo.Get(ctx, req.GetFusionRequestId())
	if err != nil {
		return nil, err
	}
	if fusionReq.Status != domain.FusionRequestStatusPending {
		return nil, status.Errorf(codes.FailedPrecondition,
			"fusion request status is %q, only %q can be responded to",
			fusionReq.Status, domain.FusionRequestStatusPending)
	}

	fusionReq.Status = domain.FusionRequestStatusRejected
	fusionReq.UpdatedAt = now
	if updateErr := s.fusionRepo.Update(ctx, fusionReq); updateErr != nil {
		slog.Error("failed to update fusion request", "fusion_request_id", req.GetFusionRequestId(), "error", updateErr)
		return nil, status.Error(codes.Internal, "failed to update fusion request")
	}
	return fusionReq, nil
}

func validateCreateFusionRequest(req *pb.CreateFusionRequestRequest, cfg config.ServiceConfig) error {
	if req.GetRequesterShokudoId() == "" {
		return status.Error(codes.InvalidArgument, "requester_shokudo_id is required")
	}
	if !domain.IsValidCategory(req.GetDesiredCategory()) {
		return status.Errorf(codes.InvalidArgument, "invalid desired_category: %q", req.GetDesiredCategory())
	}
	if req.GetDesiredQuantity() < cfg.MinQuantity || req.GetDesiredQuantity() > cfg.MaxQuantity {
		return status.Errorf(codes.InvalidArgument, "desired_quantity must be between %d and %d", cfg.MinQuantity, cfg.MaxQuantity)
	}
	if !domain.IsValidUnit(req.GetUnit()) {
		return status.Errorf(codes.InvalidArgument, "invalid unit: %q", req.GetUnit())
	}
	if utf8.RuneCountInString(req.GetMessage()) > cfg.MaxMessageLength {
		return status.Errorf(codes.InvalidArgument, "message must be at most %d characters", cfg.MaxMessageLength)
	}
	return nil
}

func validateRespondToFusionRequest(req *pb.RespondToFusionRequestRequest) error {
	if req.GetFusionRequestId() == "" {
		return status.Error(codes.InvalidArgument, "fusion_request_id is required")
	}
	validResponses := map[string]bool{"APPROVED": true, "REJECTED": true}
	if !validResponses[req.GetResponse()] {
		return status.Errorf(codes.InvalidArgument, "response must be APPROVED or REJECTED, got %q", req.GetResponse())
	}
	if req.GetResponse() == "APPROVED" && req.GetFoodItemId() == "" {
		return status.Error(codes.InvalidArgument, "food_item_id is required when response is APPROVED")
	}
	return nil
}

func domainFusionRequestToProto(req *domain.FusionRequest) *pb.FusionRequest {
	return &pb.FusionRequest{
		Id:                 req.ID,
		RequesterShokudoId: req.RequesterShokudoID,
		DesiredCategory:    req.DesiredCategory,
		DesiredQuantity:    req.DesiredQuantity,
		Unit:               req.Unit,
		Message:            req.Message,
		Status:             string(req.Status),
		ResponderShokudoId: req.ResponderShokudoID,
		FoodItemId:         req.FoodItemID,
		CreatedAt:          req.CreatedAt.Format(time.RFC3339),
		UpdatedAt:          req.UpdatedAt.Format(time.RFC3339),
	}
}
