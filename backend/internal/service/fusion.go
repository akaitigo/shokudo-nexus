package service

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/akaitigo/shokudo-nexus/backend/internal/domain"
	"github.com/akaitigo/shokudo-nexus/backend/internal/repository"

	pb "github.com/akaitigo/shokudo-nexus/backend/gen/shokudo/v1"
)

// FusionService はFusionServiceのgRPC実装。
type FusionService struct {
	pb.UnimplementedFusionServiceServer
	fusionRepo   repository.FusionRequestStore
	foodItemRepo repository.FoodItemStore
}

// NewFusionService は新しいFusionServiceを生成する。
func NewFusionService(fusionRepo repository.FusionRequestStore, foodItemRepo repository.FoodItemStore) *FusionService {
	return &FusionService{
		fusionRepo:   fusionRepo,
		foodItemRepo: foodItemRepo,
	}
}

// CreateFusionRequest は食材融通リクエストを作成する。
func (s *FusionService) CreateFusionRequest(ctx context.Context, req *pb.CreateFusionRequestRequest) (*pb.CreateFusionRequestResponse, error) {
	if err := validateCreateFusionRequest(req); err != nil {
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
		return nil, status.Errorf(codes.Internal, "failed to create fusion request: %v", err)
	}

	return &pb.CreateFusionRequestResponse{
		FusionRequest: domainFusionRequestToProto(created),
	}, nil
}

// ListFusionRequests は融通リクエストの一覧を取得する。
func (s *FusionService) ListFusionRequests(ctx context.Context, req *pb.ListFusionRequestsRequest) (*pb.ListFusionRequestsResponse, error) {
	pageSize := int(req.GetPageSize())
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		return nil, status.Errorf(codes.InvalidArgument, "page_size must be between 1 and %d", maxPageSize)
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
func (s *FusionService) RespondToFusionRequest(ctx context.Context, req *pb.RespondToFusionRequestRequest) (*pb.RespondToFusionRequestResponse, error) {
	if err := validateRespondToFusionRequest(req); err != nil {
		return nil, err
	}

	fusionReq, err := s.fusionRepo.Get(ctx, req.GetFusionRequestId())
	if err != nil {
		return nil, err
	}

	// ステータス遷移バリデーション: pending からのみ応答可能
	if fusionReq.Status != domain.FusionRequestStatusPending {
		return nil, status.Errorf(codes.FailedPrecondition,
			"fusion request status is %q, only %q can be responded to",
			fusionReq.Status, domain.FusionRequestStatusPending)
	}

	now := time.Now().UTC()
	fusionReq.UpdatedAt = now

	var foodItem *domain.FoodItem // ロールバック用にスコープを関数レベルに引き上げ

	switch req.GetResponse() {
	case "APPROVED":
		// 承認時: FoodItem のステータスを reserved に変更
		var foodErr error
		foodItem, foodErr = s.foodItemRepo.Get(ctx, req.GetFoodItemId())
		if foodErr != nil {
			return nil, foodErr
		}
		if foodItem.Status != domain.FoodItemStatusAvailable {
			return nil, status.Errorf(codes.FailedPrecondition,
				"food item status is %q, must be %q to approve",
				foodItem.Status, domain.FoodItemStatusAvailable)
		}

		// カテゴリ整合性チェック
		if foodItem.Category != fusionReq.DesiredCategory {
			return nil, status.Errorf(codes.InvalidArgument,
				"food item category %q does not match desired category %q",
				foodItem.Category, fusionReq.DesiredCategory)
		}

		// 単位整合性チェック
		if foodItem.Unit != fusionReq.Unit {
			return nil, status.Errorf(codes.InvalidArgument,
				"food item unit %q does not match desired unit %q",
				foodItem.Unit, fusionReq.Unit)
		}

		// 数量充足チェック
		if foodItem.Quantity < fusionReq.DesiredQuantity {
			return nil, status.Errorf(codes.InvalidArgument,
				"food item quantity %d is less than desired quantity %d",
				foodItem.Quantity, fusionReq.DesiredQuantity)
		}

		foodItem.Status = domain.FoodItemStatusReserved
		foodItem.UpdatedAt = now
		if updateErr := s.foodItemRepo.Update(ctx, foodItem); updateErr != nil {
			return nil, status.Errorf(codes.Internal, "failed to update food item status: %v", updateErr)
		}

		fusionReq.Status = domain.FusionRequestStatusApproved
		fusionReq.FoodItemID = req.GetFoodItemId()
	case "REJECTED":
		fusionReq.Status = domain.FusionRequestStatusRejected
	}

	if updateErr := s.fusionRepo.Update(ctx, fusionReq); updateErr != nil {
		// Rollback: food item が reserved に更新済みの場合、available に戻す
		if req.GetResponse() == "APPROVED" && foodItem != nil {
			foodItem.Status = domain.FoodItemStatusAvailable
			foodItem.UpdatedAt = now
			_ = s.foodItemRepo.Update(ctx, foodItem) // best-effort rollback
		}
		return nil, status.Errorf(codes.Internal, "failed to update fusion request: %v", updateErr)
	}

	return &pb.RespondToFusionRequestResponse{
		FusionRequest: domainFusionRequestToProto(fusionReq),
	}, nil
}

func validateCreateFusionRequest(req *pb.CreateFusionRequestRequest) error {
	if req.GetRequesterShokudoId() == "" {
		return status.Error(codes.InvalidArgument, "requester_shokudo_id is required")
	}
	if !domain.IsValidCategory(req.GetDesiredCategory()) {
		return status.Errorf(codes.InvalidArgument, "invalid desired_category: %q", req.GetDesiredCategory())
	}
	if req.GetDesiredQuantity() < minQuantity || req.GetDesiredQuantity() > maxQuantity {
		return status.Errorf(codes.InvalidArgument, "desired_quantity must be between %d and %d", minQuantity, maxQuantity)
	}
	if !domain.IsValidUnit(req.GetUnit()) {
		return status.Errorf(codes.InvalidArgument, "invalid unit: %q", req.GetUnit())
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
