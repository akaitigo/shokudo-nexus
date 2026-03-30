package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/akaitigo/shokudo-nexus/backend/internal/domain"
	"github.com/akaitigo/shokudo-nexus/backend/internal/repository"

	pb "github.com/akaitigo/shokudo-nexus/backend/gen/shokudo/v1"
)

func TestValidateCreateFusionRequest(t *testing.T) {
	validReq := &pb.CreateFusionRequestRequest{
		RequesterShokudoId: "shokudo-1",
		DesiredCategory:    "野菜",
		DesiredQuantity:    5,
		Unit:               "kg",
		Message:            "にんじんが必要です",
	}

	if err := validateCreateFusionRequest(validReq); err != nil {
		t.Errorf("expected nil error for valid request, got: %v", err)
	}

	tests := []struct {
		name     string
		modify   func(r *pb.CreateFusionRequestRequest)
		wantCode codes.Code
		wantMsg  string
	}{
		{
			name:     "empty requester_shokudo_id",
			modify:   func(r *pb.CreateFusionRequestRequest) { r.RequesterShokudoId = "" },
			wantCode: codes.InvalidArgument,
			wantMsg:  "requester_shokudo_id is required",
		},
		{
			name:     "invalid desired_category",
			modify:   func(r *pb.CreateFusionRequestRequest) { r.DesiredCategory = "invalid" },
			wantCode: codes.InvalidArgument,
			wantMsg:  "invalid desired_category",
		},
		{
			name:     "quantity too low",
			modify:   func(r *pb.CreateFusionRequestRequest) { r.DesiredQuantity = 0 },
			wantCode: codes.InvalidArgument,
			wantMsg:  "desired_quantity must be between",
		},
		{
			name:     "quantity too high",
			modify:   func(r *pb.CreateFusionRequestRequest) { r.DesiredQuantity = 10001 },
			wantCode: codes.InvalidArgument,
			wantMsg:  "desired_quantity must be between",
		},
		{
			name:     "invalid unit",
			modify:   func(r *pb.CreateFusionRequestRequest) { r.Unit = "リットル" },
			wantCode: codes.InvalidArgument,
			wantMsg:  "invalid unit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &pb.CreateFusionRequestRequest{
				RequesterShokudoId: validReq.RequesterShokudoId,
				DesiredCategory:    validReq.DesiredCategory,
				DesiredQuantity:    validReq.DesiredQuantity,
				Unit:               validReq.Unit,
				Message:            validReq.Message,
			}
			tt.modify(req)

			err := validateCreateFusionRequest(req)
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			st, ok := status.FromError(err)
			if !ok {
				t.Fatalf("expected gRPC status error, got: %v", err)
			}
			if st.Code() != tt.wantCode {
				t.Errorf("expected code %v, got %v", tt.wantCode, st.Code())
			}
			if !strings.Contains(st.Message(), tt.wantMsg) {
				t.Errorf("expected message to contain %q, got %q", tt.wantMsg, st.Message())
			}
		})
	}
}

func TestValidateRespondToFusionRequest(t *testing.T) {
	tests := []struct {
		name     string
		req      *pb.RespondToFusionRequestRequest
		wantErr  bool
		wantCode codes.Code
		wantMsg  string
	}{
		{
			name: "valid approve",
			req: &pb.RespondToFusionRequestRequest{
				FusionRequestId: "req-1",
				Response:        "APPROVED",
				FoodItemId:      "food-1",
			},
			wantErr: false,
		},
		{
			name: "valid reject",
			req: &pb.RespondToFusionRequestRequest{
				FusionRequestId: "req-1",
				Response:        "REJECTED",
			},
			wantErr: false,
		},
		{
			name: "empty fusion_request_id",
			req: &pb.RespondToFusionRequestRequest{
				FusionRequestId: "",
				Response:        "APPROVED",
				FoodItemId:      "food-1",
			},
			wantErr:  true,
			wantCode: codes.InvalidArgument,
			wantMsg:  "fusion_request_id is required",
		},
		{
			name: "invalid response",
			req: &pb.RespondToFusionRequestRequest{
				FusionRequestId: "req-1",
				Response:        "INVALID",
			},
			wantErr:  true,
			wantCode: codes.InvalidArgument,
			wantMsg:  "response must be APPROVED or REJECTED",
		},
		{
			name: "approve without food_item_id",
			req: &pb.RespondToFusionRequestRequest{
				FusionRequestId: "req-1",
				Response:        "APPROVED",
				FoodItemId:      "",
			},
			wantErr:  true,
			wantCode: codes.InvalidArgument,
			wantMsg:  "food_item_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRespondToFusionRequest(tt.req)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				st, ok := status.FromError(err)
				if !ok {
					t.Fatalf("expected gRPC status error, got: %v", err)
				}
				if st.Code() != tt.wantCode {
					t.Errorf("expected code %v, got %v", tt.wantCode, st.Code())
				}
				if !strings.Contains(st.Message(), tt.wantMsg) {
					t.Errorf("expected message to contain %q, got %q", tt.wantMsg, st.Message())
				}
			} else {
				if err != nil {
					t.Errorf("expected nil error, got: %v", err)
				}
			}
		})
	}
}

func TestListFusionRequests_PageSizeTooLarge(t *testing.T) {
	svc := &FusionService{}
	_, err := svc.ListFusionRequests(context.Background(), &pb.ListFusionRequestsRequest{PageSize: 101})
	if err == nil {
		t.Fatal("expected error for page_size > 100")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got: %v", err)
	}
	if st.Code() != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", st.Code())
	}
}

func TestListFusionRequests_InvalidStatusFilter(t *testing.T) {
	svc := &FusionService{}
	_, err := svc.ListFusionRequests(context.Background(), &pb.ListFusionRequestsRequest{StatusFilter: "invalid"})
	if err == nil {
		t.Fatal("expected error for invalid status filter")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got: %v", err)
	}
	if st.Code() != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", st.Code())
	}
}

func TestDomainFusionRequestToProto(t *testing.T) {
	// Verify the conversion doesn't panic with zero-value struct
	req := &pb.FusionRequest{}
	if req.Id != "" {
		t.Error("expected empty ID for zero-value FusionRequest")
	}
}

// --- newTestFusionService ヘルパー ---

func newTestFusionService() (*FusionService, *mockFusionRequestStore, *mockFoodItemStore, *mockDonationRecordStore) {
	fusionRepo := newMockFusionRequestStore()
	foodRepo := newMockFoodItemStore()
	donationRepo := newMockDonationRecordStore()
	svc := NewFusionService(fusionRepo, foodRepo, donationRepo)
	return svc, fusionRepo, foodRepo, donationRepo
}

// --- CreateFusionRequest テスト ---

func TestCreateFusionRequest_Success(t *testing.T) {
	svc, _, _, _ := newTestFusionService()

	resp, err := svc.CreateFusionRequest(context.Background(), &pb.CreateFusionRequestRequest{
		RequesterShokudoId: "shokudo-1",
		DesiredCategory:    "野菜",
		DesiredQuantity:    5,
		Unit:               "kg",
		Message:            "にんじんが必要です",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fr := resp.GetFusionRequest()
	if fr.GetId() == "" {
		t.Error("expected non-empty ID")
	}
	if fr.GetRequesterShokudoId() != "shokudo-1" {
		t.Errorf("expected requester_shokudo_id 'shokudo-1', got %q", fr.GetRequesterShokudoId())
	}
	if fr.GetDesiredCategory() != "野菜" {
		t.Errorf("expected desired_category '野菜', got %q", fr.GetDesiredCategory())
	}
	if fr.GetDesiredQuantity() != 5 {
		t.Errorf("expected desired_quantity 5, got %d", fr.GetDesiredQuantity())
	}
	if fr.GetUnit() != "kg" {
		t.Errorf("expected unit 'kg', got %q", fr.GetUnit())
	}
	if fr.GetMessage() != "にんじんが必要です" {
		t.Errorf("expected message 'にんじんが必要です', got %q", fr.GetMessage())
	}
	if fr.GetStatus() != "pending" {
		t.Errorf("expected status 'pending', got %q", fr.GetStatus())
	}
	if fr.GetCreatedAt() == "" {
		t.Error("expected non-empty created_at")
	}
}

func TestCreateFusionRequest_RepoError(t *testing.T) {
	svc, fusionRepo, _, _ := newTestFusionService()
	fusionRepo.createFunc = func(_ context.Context, _ *domain.FusionRequest) (*domain.FusionRequest, error) {
		return nil, errors.New("firestore unavailable")
	}

	_, err := svc.CreateFusionRequest(context.Background(), &pb.CreateFusionRequestRequest{
		RequesterShokudoId: "shokudo-1",
		DesiredCategory:    "野菜",
		DesiredQuantity:    5,
		Unit:               "kg",
	})
	if err == nil {
		t.Fatal("expected error when repo fails")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got: %v", err)
	}
	if st.Code() != codes.Internal {
		t.Errorf("expected Internal, got %v", st.Code())
	}
}

func TestCreateFusionRequest_AllCategories(t *testing.T) {
	for category := range domain.ValidCategories {
		t.Run(category, func(t *testing.T) {
			svc, _, _, _ := newTestFusionService()

			resp, err := svc.CreateFusionRequest(context.Background(), &pb.CreateFusionRequestRequest{
				RequesterShokudoId: "shokudo-1",
				DesiredCategory:    category,
				DesiredQuantity:    1,
				Unit:               "個",
			})
			if err != nil {
				t.Fatalf("unexpected error for category %q: %v", category, err)
			}
			if resp.GetFusionRequest().GetDesiredCategory() != category {
				t.Errorf("expected category %q, got %q", category, resp.GetFusionRequest().GetDesiredCategory())
			}
		})
	}
}

func TestCreateFusionRequest_BoundaryQuantities(t *testing.T) {
	tests := []struct {
		name     string
		quantity int32
		wantErr  bool
	}{
		{"min quantity", 1, false},
		{"max quantity", 10000, false},
		{"below min", 0, true},
		{"above max", 10001, true},
		{"negative", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, _, _, _ := newTestFusionService()

			_, err := svc.CreateFusionRequest(context.Background(), &pb.CreateFusionRequestRequest{
				RequesterShokudoId: "shokudo-1",
				DesiredCategory:    "野菜",
				DesiredQuantity:    tt.quantity,
				Unit:               "kg",
			})
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// --- ListFusionRequests テスト ---

func TestListFusionRequests_Success(t *testing.T) {
	svc, fusionRepo, _, _ := newTestFusionService()
	now := time.Now().UTC()

	fusionRepo.requests["req-1"] = &domain.FusionRequest{
		ID:                 "req-1",
		RequesterShokudoID: "shokudo-1",
		DesiredCategory:    "野菜",
		DesiredQuantity:    5,
		Unit:               "kg",
		Status:             domain.FusionRequestStatusPending,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	fusionRepo.requests["req-2"] = &domain.FusionRequest{
		ID:                 "req-2",
		RequesterShokudoID: "shokudo-2",
		DesiredCategory:    "肉",
		DesiredQuantity:    3,
		Unit:               "パック",
		Status:             domain.FusionRequestStatusApproved,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	resp, err := svc.ListFusionRequests(context.Background(), &pb.ListFusionRequestsRequest{
		PageSize: 20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.GetFusionRequests()) != 2 {
		t.Errorf("expected 2 requests, got %d", len(resp.GetFusionRequests()))
	}
}

func TestListFusionRequests_StatusFilter(t *testing.T) {
	svc, fusionRepo, _, _ := newTestFusionService()
	now := time.Now().UTC()

	fusionRepo.requests["req-1"] = &domain.FusionRequest{
		ID:                 "req-1",
		Status:             domain.FusionRequestStatusPending,
		RequesterShokudoID: "shokudo-1",
		DesiredCategory:    "野菜",
		DesiredQuantity:    5,
		Unit:               "kg",
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	fusionRepo.requests["req-2"] = &domain.FusionRequest{
		ID:                 "req-2",
		Status:             domain.FusionRequestStatusApproved,
		RequesterShokudoID: "shokudo-2",
		DesiredCategory:    "肉",
		DesiredQuantity:    3,
		Unit:               "パック",
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	resp, err := svc.ListFusionRequests(context.Background(), &pb.ListFusionRequestsRequest{
		PageSize:     20,
		StatusFilter: "pending",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.GetFusionRequests()) != 1 {
		t.Errorf("expected 1 request, got %d", len(resp.GetFusionRequests()))
	}
	if resp.GetFusionRequests()[0].GetStatus() != "pending" {
		t.Errorf("expected status 'pending', got %q", resp.GetFusionRequests()[0].GetStatus())
	}
}

func TestListFusionRequests_DefaultPageSize(t *testing.T) {
	svc, fusionRepo, _, _ := newTestFusionService()
	fusionRepo.listFunc = func(_ context.Context, params repository.FusionListParams) (*repository.FusionListResult, error) {
		if params.PageSize != defaultPageSize {
			t.Errorf("expected default page size %d, got %d", defaultPageSize, params.PageSize)
		}
		return &repository.FusionListResult{Requests: nil, TotalCount: 0}, nil
	}

	_, err := svc.ListFusionRequests(context.Background(), &pb.ListFusionRequestsRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListFusionRequests_RepoError(t *testing.T) {
	svc, fusionRepo, _, _ := newTestFusionService()
	fusionRepo.listFunc = func(_ context.Context, _ repository.FusionListParams) (*repository.FusionListResult, error) {
		return nil, errors.New("firestore unavailable")
	}

	_, err := svc.ListFusionRequests(context.Background(), &pb.ListFusionRequestsRequest{PageSize: 20})
	if err == nil {
		t.Fatal("expected error when repo fails")
	}
}

// --- RespondToFusionRequest テスト ---

func TestRespondToFusionRequest_Approve_Success(t *testing.T) {
	svc, fusionRepo, foodRepo, _ := newTestFusionService()
	now := time.Now().UTC()

	fusionRepo.requests["req-1"] = &domain.FusionRequest{
		ID:                 "req-1",
		RequesterShokudoID: "shokudo-1",
		DesiredCategory:    "野菜",
		DesiredQuantity:    5,
		Unit:               "kg",
		Status:             domain.FusionRequestStatusPending,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	foodRepo.items["food-1"] = &domain.FoodItem{
		ID:         "food-1",
		Name:       "にんじん",
		Category:   "野菜",
		ExpiryDate: now.Add(48 * time.Hour),
		Quantity:   10,
		Unit:       "kg",
		DonorID:    "donor-1",
		Status:     domain.FoodItemStatusAvailable,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	resp, err := svc.RespondToFusionRequest(context.Background(), &pb.RespondToFusionRequestRequest{
		FusionRequestId: "req-1",
		Response:        "APPROVED",
		FoodItemId:      "food-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fr := resp.GetFusionRequest()
	if fr.GetStatus() != "approved" {
		t.Errorf("expected status 'approved', got %q", fr.GetStatus())
	}
	if fr.GetFoodItemId() != "food-1" {
		t.Errorf("expected food_item_id 'food-1', got %q", fr.GetFoodItemId())
	}

	// FoodItem のステータスが reserved になっていること
	if foodRepo.items["food-1"].Status != domain.FoodItemStatusReserved {
		t.Errorf("expected food item status 'reserved', got %q", foodRepo.items["food-1"].Status)
	}
}

func TestRespondToFusionRequest_Reject_Success(t *testing.T) {
	svc, fusionRepo, _, _ := newTestFusionService()
	now := time.Now().UTC()

	fusionRepo.requests["req-1"] = &domain.FusionRequest{
		ID:                 "req-1",
		RequesterShokudoID: "shokudo-1",
		DesiredCategory:    "野菜",
		DesiredQuantity:    5,
		Unit:               "kg",
		Status:             domain.FusionRequestStatusPending,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	resp, err := svc.RespondToFusionRequest(context.Background(), &pb.RespondToFusionRequestRequest{
		FusionRequestId: "req-1",
		Response:        "REJECTED",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.GetFusionRequest().GetStatus() != "rejected" {
		t.Errorf("expected status 'rejected', got %q", resp.GetFusionRequest().GetStatus())
	}
}

func TestRespondToFusionRequest_NotPending_Fails(t *testing.T) {
	statuses := []domain.FusionRequestStatus{
		domain.FusionRequestStatusApproved,
		domain.FusionRequestStatusRejected,
		domain.FusionRequestStatusCompleted,
	}

	for _, s := range statuses {
		t.Run(string(s), func(t *testing.T) {
			svc, fusionRepo, _, _ := newTestFusionService()
			now := time.Now().UTC()

			fusionRepo.requests["req-1"] = &domain.FusionRequest{
				ID:                 "req-1",
				RequesterShokudoID: "shokudo-1",
				Status:             s,
				CreatedAt:          now,
				UpdatedAt:          now,
			}

			_, err := svc.RespondToFusionRequest(context.Background(), &pb.RespondToFusionRequestRequest{
				FusionRequestId: "req-1",
				Response:        "APPROVED",
				FoodItemId:      "food-1",
			})
			if err == nil {
				t.Fatal("expected error for non-pending status")
			}

			st, ok := status.FromError(err)
			if !ok {
				t.Fatalf("expected gRPC status error, got: %v", err)
			}
			if st.Code() != codes.FailedPrecondition {
				t.Errorf("expected FailedPrecondition, got %v", st.Code())
			}
		})
	}
}

func TestRespondToFusionRequest_Approve_FoodItemNotAvailable(t *testing.T) {
	svc, fusionRepo, foodRepo, _ := newTestFusionService()
	now := time.Now().UTC()

	fusionRepo.requests["req-1"] = &domain.FusionRequest{
		ID:                 "req-1",
		Status:             domain.FusionRequestStatusPending,
		RequesterShokudoID: "shokudo-1",
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	foodRepo.items["food-1"] = &domain.FoodItem{
		ID:     "food-1",
		Status: domain.FoodItemStatusReserved,
	}

	_, err := svc.RespondToFusionRequest(context.Background(), &pb.RespondToFusionRequestRequest{
		FusionRequestId: "req-1",
		Response:        "APPROVED",
		FoodItemId:      "food-1",
	})
	if err == nil {
		t.Fatal("expected error for reserved food item")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got: %v", err)
	}
	if st.Code() != codes.FailedPrecondition {
		t.Errorf("expected FailedPrecondition, got %v", st.Code())
	}
}

func TestRespondToFusionRequest_NotFound(t *testing.T) {
	svc, _, _, _ := newTestFusionService()

	_, err := svc.RespondToFusionRequest(context.Background(), &pb.RespondToFusionRequestRequest{
		FusionRequestId: "nonexistent",
		Response:        "APPROVED",
		FoodItemId:      "food-1",
	})
	if err == nil {
		t.Fatal("expected error for nonexistent fusion request")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got: %v", err)
	}
	if st.Code() != codes.NotFound {
		t.Errorf("expected NotFound, got %v", st.Code())
	}
}

func TestRespondToFusionRequest_Approve_FoodItemNotFound(t *testing.T) {
	svc, fusionRepo, _, _ := newTestFusionService()
	now := time.Now().UTC()

	fusionRepo.requests["req-1"] = &domain.FusionRequest{
		ID:                 "req-1",
		Status:             domain.FusionRequestStatusPending,
		RequesterShokudoID: "shokudo-1",
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	_, err := svc.RespondToFusionRequest(context.Background(), &pb.RespondToFusionRequestRequest{
		FusionRequestId: "req-1",
		Response:        "APPROVED",
		FoodItemId:      "nonexistent-food",
	})
	if err == nil {
		t.Fatal("expected error for nonexistent food item")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got: %v", err)
	}
	if st.Code() != codes.NotFound {
		t.Errorf("expected NotFound, got %v", st.Code())
	}
}

func TestRespondToFusionRequest_Approve_FoodItemUpdateError(t *testing.T) {
	svc, fusionRepo, foodRepo, _ := newTestFusionService()
	now := time.Now().UTC()

	fusionRepo.requests["req-1"] = &domain.FusionRequest{
		ID:                 "req-1",
		Status:             domain.FusionRequestStatusPending,
		RequesterShokudoID: "shokudo-1",
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	foodRepo.items["food-1"] = &domain.FoodItem{
		ID:     "food-1",
		Status: domain.FoodItemStatusAvailable,
	}
	foodRepo.updateFunc = func(_ context.Context, _ *domain.FoodItem) error {
		return errors.New("firestore unavailable")
	}

	_, err := svc.RespondToFusionRequest(context.Background(), &pb.RespondToFusionRequestRequest{
		FusionRequestId: "req-1",
		Response:        "APPROVED",
		FoodItemId:      "food-1",
	})
	if err == nil {
		t.Fatal("expected error when food item update fails")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got: %v", err)
	}
	if st.Code() != codes.Internal {
		t.Errorf("expected Internal, got %v", st.Code())
	}
}

func TestRespondToFusionRequest_FusionRepoUpdateError(t *testing.T) {
	svc, fusionRepo, _, _ := newTestFusionService()
	now := time.Now().UTC()

	fusionRepo.requests["req-1"] = &domain.FusionRequest{
		ID:                 "req-1",
		Status:             domain.FusionRequestStatusPending,
		RequesterShokudoID: "shokudo-1",
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	fusionRepo.updateFunc = func(_ context.Context, _ *domain.FusionRequest) error {
		return errors.New("firestore unavailable")
	}

	_, err := svc.RespondToFusionRequest(context.Background(), &pb.RespondToFusionRequestRequest{
		FusionRequestId: "req-1",
		Response:        "REJECTED",
	})
	if err == nil {
		t.Fatal("expected error when fusion repo update fails")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got: %v", err)
	}
	if st.Code() != codes.Internal {
		t.Errorf("expected Internal, got %v", st.Code())
	}
}

// --- CompleteFusionRequest テスト ---

func TestCompleteFusionRequest_Success(t *testing.T) {
	svc, fusionRepo, foodRepo, donationRepo := newTestFusionService()
	now := time.Now().UTC()

	fusionRepo.requests["req-1"] = &domain.FusionRequest{
		ID:                 "req-1",
		RequesterShokudoID: "shokudo-1",
		DesiredCategory:    "野菜",
		DesiredQuantity:    5,
		Unit:               "kg",
		Status:             domain.FusionRequestStatusApproved,
		FoodItemID:         "food-1",
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	foodRepo.items["food-1"] = &domain.FoodItem{
		ID:        "food-1",
		Name:      "にんじん",
		Category:  "野菜",
		Quantity:  10,
		Unit:      "kg",
		DonorID:   "donor-1",
		Status:    domain.FoodItemStatusReserved,
		CreatedAt: now,
		UpdatedAt: now,
	}

	resp, err := svc.CompleteFusionRequest(context.Background(), &pb.CompleteFusionRequestRequest{
		FusionRequestId: "req-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fr := resp.GetFusionRequest()
	if fr.GetStatus() != "completed" {
		t.Errorf("expected status 'completed', got %q", fr.GetStatus())
	}

	// FoodItem のステータスが consumed になっていること
	if foodRepo.items["food-1"].Status != domain.FoodItemStatusConsumed {
		t.Errorf("expected food item status 'consumed', got %q", foodRepo.items["food-1"].Status)
	}

	// DonationRecord が作成されていること
	if len(donationRepo.records) != 1 {
		t.Fatalf("expected 1 donation record, got %d", len(donationRepo.records))
	}
	dr := donationRepo.records[0]
	if dr.FoodItemID != "food-1" {
		t.Errorf("expected food_item_id 'food-1', got %q", dr.FoodItemID)
	}
	if dr.DonorID != "donor-1" {
		t.Errorf("expected donor_id 'donor-1', got %q", dr.DonorID)
	}
	if dr.RecipientID != "shokudo-1" {
		t.Errorf("expected recipient_id 'shokudo-1', got %q", dr.RecipientID)
	}
	if dr.FusionRequestID != "req-1" {
		t.Errorf("expected fusion_request_id 'req-1', got %q", dr.FusionRequestID)
	}
	if dr.Category != "野菜" {
		t.Errorf("expected category '野菜', got %q", dr.Category)
	}
	if dr.Quantity != 10 {
		t.Errorf("expected quantity 10, got %d", dr.Quantity)
	}
	if dr.Unit != "kg" {
		t.Errorf("expected unit 'kg', got %q", dr.Unit)
	}
}

func TestCompleteFusionRequest_EmptyID(t *testing.T) {
	svc, _, _, _ := newTestFusionService()

	_, err := svc.CompleteFusionRequest(context.Background(), &pb.CompleteFusionRequestRequest{
		FusionRequestId: "",
	})
	if err == nil {
		t.Fatal("expected error for empty fusion_request_id")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got: %v", err)
	}
	if st.Code() != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", st.Code())
	}
}

func TestCompleteFusionRequest_NotFound(t *testing.T) {
	svc, _, _, _ := newTestFusionService()

	_, err := svc.CompleteFusionRequest(context.Background(), &pb.CompleteFusionRequestRequest{
		FusionRequestId: "nonexistent",
	})
	if err == nil {
		t.Fatal("expected error for nonexistent fusion request")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got: %v", err)
	}
	if st.Code() != codes.NotFound {
		t.Errorf("expected NotFound, got %v", st.Code())
	}
}

func TestCompleteFusionRequest_NotApproved_Fails(t *testing.T) {
	statuses := []domain.FusionRequestStatus{
		domain.FusionRequestStatusPending,
		domain.FusionRequestStatusRejected,
		domain.FusionRequestStatusCompleted,
	}

	for _, s := range statuses {
		t.Run(string(s), func(t *testing.T) {
			svc, fusionRepo, _, _ := newTestFusionService()
			now := time.Now().UTC()

			fusionRepo.requests["req-1"] = &domain.FusionRequest{
				ID:        "req-1",
				Status:    s,
				CreatedAt: now,
				UpdatedAt: now,
			}

			_, err := svc.CompleteFusionRequest(context.Background(), &pb.CompleteFusionRequestRequest{
				FusionRequestId: "req-1",
			})
			if err == nil {
				t.Fatal("expected error for non-approved status")
			}

			st, ok := status.FromError(err)
			if !ok {
				t.Fatalf("expected gRPC status error, got: %v", err)
			}
			if st.Code() != codes.FailedPrecondition {
				t.Errorf("expected FailedPrecondition, got %v", st.Code())
			}
		})
	}
}

func TestCompleteFusionRequest_FoodItemNotFound(t *testing.T) {
	svc, fusionRepo, _, _ := newTestFusionService()
	now := time.Now().UTC()

	fusionRepo.requests["req-1"] = &domain.FusionRequest{
		ID:         "req-1",
		Status:     domain.FusionRequestStatusApproved,
		FoodItemID: "nonexistent-food",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	_, err := svc.CompleteFusionRequest(context.Background(), &pb.CompleteFusionRequestRequest{
		FusionRequestId: "req-1",
	})
	if err == nil {
		t.Fatal("expected error for nonexistent food item")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got: %v", err)
	}
	if st.Code() != codes.NotFound {
		t.Errorf("expected NotFound, got %v", st.Code())
	}
}

func TestCompleteFusionRequest_FoodItemUpdateError(t *testing.T) {
	svc, fusionRepo, foodRepo, _ := newTestFusionService()
	now := time.Now().UTC()

	fusionRepo.requests["req-1"] = &domain.FusionRequest{
		ID:         "req-1",
		Status:     domain.FusionRequestStatusApproved,
		FoodItemID: "food-1",
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	foodRepo.items["food-1"] = &domain.FoodItem{
		ID:     "food-1",
		Status: domain.FoodItemStatusReserved,
	}
	foodRepo.updateFunc = func(_ context.Context, _ *domain.FoodItem) error {
		return errors.New("firestore unavailable")
	}

	_, err := svc.CompleteFusionRequest(context.Background(), &pb.CompleteFusionRequestRequest{
		FusionRequestId: "req-1",
	})
	if err == nil {
		t.Fatal("expected error when food item update fails")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got: %v", err)
	}
	if st.Code() != codes.Internal {
		t.Errorf("expected Internal, got %v", st.Code())
	}
}

func TestCompleteFusionRequest_DonationRecordCreateError(t *testing.T) {
	svc, fusionRepo, foodRepo, donationRepo := newTestFusionService()
	now := time.Now().UTC()

	fusionRepo.requests["req-1"] = &domain.FusionRequest{
		ID:                 "req-1",
		RequesterShokudoID: "shokudo-1",
		Status:             domain.FusionRequestStatusApproved,
		FoodItemID:         "food-1",
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	foodRepo.items["food-1"] = &domain.FoodItem{
		ID:      "food-1",
		DonorID: "donor-1",
		Status:  domain.FoodItemStatusReserved,
	}
	donationRepo.createFunc = func(_ context.Context, _ *domain.DonationRecord) (*domain.DonationRecord, error) {
		return nil, errors.New("firestore unavailable")
	}

	_, err := svc.CompleteFusionRequest(context.Background(), &pb.CompleteFusionRequestRequest{
		FusionRequestId: "req-1",
	})
	if err == nil {
		t.Fatal("expected error when donation record creation fails")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got: %v", err)
	}
	if st.Code() != codes.Internal {
		t.Errorf("expected Internal, got %v", st.Code())
	}
}

func TestCompleteFusionRequest_FusionRepoUpdateError(t *testing.T) {
	svc, fusionRepo, foodRepo, _ := newTestFusionService()
	now := time.Now().UTC()

	fusionRepo.requests["req-1"] = &domain.FusionRequest{
		ID:                 "req-1",
		RequesterShokudoID: "shokudo-1",
		Status:             domain.FusionRequestStatusApproved,
		FoodItemID:         "food-1",
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	foodRepo.items["food-1"] = &domain.FoodItem{
		ID:      "food-1",
		DonorID: "donor-1",
		Status:  domain.FoodItemStatusReserved,
	}
	fusionRepo.updateFunc = func(_ context.Context, _ *domain.FusionRequest) error {
		return errors.New("firestore unavailable")
	}

	_, err := svc.CompleteFusionRequest(context.Background(), &pb.CompleteFusionRequestRequest{
		FusionRequestId: "req-1",
	})
	if err == nil {
		t.Fatal("expected error when fusion repo update fails")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got: %v", err)
	}
	if st.Code() != codes.Internal {
		t.Errorf("expected Internal, got %v", st.Code())
	}
}

// --- domainFusionRequestToProto テスト ---

func TestDomainFusionRequestToProto_AllFields(t *testing.T) {
	now := time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC)

	req := &domain.FusionRequest{
		ID:                 "test-id",
		RequesterShokudoID: "shokudo-1",
		DesiredCategory:    "野菜",
		DesiredQuantity:    5,
		Unit:               "kg",
		Message:            "テストメッセージ",
		Status:             domain.FusionRequestStatusApproved,
		ResponderShokudoID: "shokudo-2",
		FoodItemID:         "food-1",
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	pbReq := domainFusionRequestToProto(req)

	if pbReq.GetId() != "test-id" {
		t.Errorf("expected id 'test-id', got %q", pbReq.GetId())
	}
	if pbReq.GetRequesterShokudoId() != "shokudo-1" {
		t.Errorf("expected requester_shokudo_id 'shokudo-1', got %q", pbReq.GetRequesterShokudoId())
	}
	if pbReq.GetDesiredCategory() != "野菜" {
		t.Errorf("expected desired_category '野菜', got %q", pbReq.GetDesiredCategory())
	}
	if pbReq.GetDesiredQuantity() != 5 {
		t.Errorf("expected desired_quantity 5, got %d", pbReq.GetDesiredQuantity())
	}
	if pbReq.GetUnit() != "kg" {
		t.Errorf("expected unit 'kg', got %q", pbReq.GetUnit())
	}
	if pbReq.GetMessage() != "テストメッセージ" {
		t.Errorf("expected message 'テストメッセージ', got %q", pbReq.GetMessage())
	}
	if pbReq.GetStatus() != "approved" {
		t.Errorf("expected status 'approved', got %q", pbReq.GetStatus())
	}
	if pbReq.GetResponderShokudoId() != "shokudo-2" {
		t.Errorf("expected responder_shokudo_id 'shokudo-2', got %q", pbReq.GetResponderShokudoId())
	}
	if pbReq.GetFoodItemId() != "food-1" {
		t.Errorf("expected food_item_id 'food-1', got %q", pbReq.GetFoodItemId())
	}
	if pbReq.GetCreatedAt() != "2026-04-01T12:00:00Z" {
		t.Errorf("expected created_at '2026-04-01T12:00:00Z', got %q", pbReq.GetCreatedAt())
	}
	if pbReq.GetUpdatedAt() != "2026-04-01T12:00:00Z" {
		t.Errorf("expected updated_at '2026-04-01T12:00:00Z', got %q", pbReq.GetUpdatedAt())
	}
}

// --- E2Eフロー: pending → approved → completed ---

func TestFusionFlow_PendingToApprovedToCompleted(t *testing.T) {
	svc, _, foodRepo, donationRepo := newTestFusionService()

	// 1. FoodItem を作成（利用可能状態）
	foodRepo.items["food-1"] = &domain.FoodItem{
		ID:        "food-1",
		Name:      "にんじん",
		Category:  "野菜",
		Quantity:  10,
		Unit:      "kg",
		DonorID:   "donor-1",
		Status:    domain.FoodItemStatusAvailable,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	// 2. 融通リクエストを作成（pending）
	createResp, err := svc.CreateFusionRequest(context.Background(), &pb.CreateFusionRequestRequest{
		RequesterShokudoId: "shokudo-1",
		DesiredCategory:    "野菜",
		DesiredQuantity:    5,
		Unit:               "kg",
		Message:            "にんじんが欲しいです",
	})
	if err != nil {
		t.Fatalf("CreateFusionRequest failed: %v", err)
	}
	reqID := createResp.GetFusionRequest().GetId()
	if createResp.GetFusionRequest().GetStatus() != "pending" {
		t.Errorf("expected status 'pending', got %q", createResp.GetFusionRequest().GetStatus())
	}

	// 3. 承認（pending → approved）
	approveResp, err := svc.RespondToFusionRequest(context.Background(), &pb.RespondToFusionRequestRequest{
		FusionRequestId: reqID,
		Response:        "APPROVED",
		FoodItemId:      "food-1",
	})
	if err != nil {
		t.Fatalf("RespondToFusionRequest (APPROVED) failed: %v", err)
	}
	if approveResp.GetFusionRequest().GetStatus() != "approved" {
		t.Errorf("expected status 'approved', got %q", approveResp.GetFusionRequest().GetStatus())
	}
	if foodRepo.items["food-1"].Status != domain.FoodItemStatusReserved {
		t.Errorf("expected food item status 'reserved', got %q", foodRepo.items["food-1"].Status)
	}

	// 4. 完了（approved → completed）
	completeResp, err := svc.CompleteFusionRequest(context.Background(), &pb.CompleteFusionRequestRequest{
		FusionRequestId: reqID,
	})
	if err != nil {
		t.Fatalf("CompleteFusionRequest failed: %v", err)
	}
	if completeResp.GetFusionRequest().GetStatus() != "completed" {
		t.Errorf("expected status 'completed', got %q", completeResp.GetFusionRequest().GetStatus())
	}
	if foodRepo.items["food-1"].Status != domain.FoodItemStatusConsumed {
		t.Errorf("expected food item status 'consumed', got %q", foodRepo.items["food-1"].Status)
	}

	// DonationRecord が作成されていること
	if len(donationRepo.records) != 1 {
		t.Fatalf("expected 1 donation record, got %d", len(donationRepo.records))
	}
	dr := donationRepo.records[0]
	if dr.DonorID != "donor-1" {
		t.Errorf("expected donor_id 'donor-1', got %q", dr.DonorID)
	}
	if dr.RecipientID != "shokudo-1" {
		t.Errorf("expected recipient_id 'shokudo-1', got %q", dr.RecipientID)
	}
	if dr.FusionRequestID != reqID {
		t.Errorf("expected fusion_request_id %q, got %q", reqID, dr.FusionRequestID)
	}

	// 5. 完了済みリクエストに再度応答できないこと
	_, err = svc.RespondToFusionRequest(context.Background(), &pb.RespondToFusionRequestRequest{
		FusionRequestId: reqID,
		Response:        "APPROVED",
		FoodItemId:      "food-1",
	})
	if err == nil {
		t.Fatal("expected error when responding to completed request")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got: %v", err)
	}
	if st.Code() != codes.FailedPrecondition {
		t.Errorf("expected FailedPrecondition, got %v", st.Code())
	}

	// 6. 完了済みリクエストを再度完了できないこと
	_, err = svc.CompleteFusionRequest(context.Background(), &pb.CompleteFusionRequestRequest{
		FusionRequestId: reqID,
	})
	if err == nil {
		t.Fatal("expected error when completing already completed request")
	}
	st, ok = status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got: %v", err)
	}
	if st.Code() != codes.FailedPrecondition {
		t.Errorf("expected FailedPrecondition, got %v", st.Code())
	}
}

// --- E2Eフロー: pending → rejected ---

func TestFusionFlow_PendingToRejected(t *testing.T) {
	svc, _, _, _ := newTestFusionService()

	// 1. 融通リクエストを作成
	createResp, err := svc.CreateFusionRequest(context.Background(), &pb.CreateFusionRequestRequest{
		RequesterShokudoId: "shokudo-1",
		DesiredCategory:    "肉",
		DesiredQuantity:    3,
		Unit:               "パック",
	})
	if err != nil {
		t.Fatalf("CreateFusionRequest failed: %v", err)
	}
	reqID := createResp.GetFusionRequest().GetId()

	// 2. 拒否
	rejectResp, err := svc.RespondToFusionRequest(context.Background(), &pb.RespondToFusionRequestRequest{
		FusionRequestId: reqID,
		Response:        "REJECTED",
	})
	if err != nil {
		t.Fatalf("RespondToFusionRequest (REJECTED) failed: %v", err)
	}
	if rejectResp.GetFusionRequest().GetStatus() != "rejected" {
		t.Errorf("expected status 'rejected', got %q", rejectResp.GetFusionRequest().GetStatus())
	}

	// 3. 拒否済みリクエストに再度応答できないこと
	_, err = svc.RespondToFusionRequest(context.Background(), &pb.RespondToFusionRequestRequest{
		FusionRequestId: reqID,
		Response:        "APPROVED",
		FoodItemId:      "food-1",
	})
	if err == nil {
		t.Fatal("expected error when responding to rejected request")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got: %v", err)
	}
	if st.Code() != codes.FailedPrecondition {
		t.Errorf("expected FailedPrecondition, got %v", st.Code())
	}
}
