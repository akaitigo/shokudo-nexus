package service

import (
	"context"
	"strings"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/akaitigo/shokudo-nexus/backend/gen/shokudo/v1"
	"github.com/akaitigo/shokudo-nexus/backend/internal/domain"
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

func TestRespondToFusionRequest_ApprovalValidation(t *testing.T) {
	tests := []struct {
		name        string
		fusionReq   *domain.FusionRequest
		foodItem    *domain.FoodItem
		wantCode    codes.Code
		wantMsg     string
	}{
		{
			name: "category mismatch rejects approval",
			fusionReq: &domain.FusionRequest{
				ID:              "req-1",
				DesiredCategory: "野菜",
				DesiredQuantity: 5,
				Unit:            "kg",
				Status:          domain.FusionRequestStatusPending,
			},
			foodItem: &domain.FoodItem{
				ID:       "food-1",
				Category: "肉類",
				Quantity: 10,
				Unit:     "kg",
				Status:   domain.FoodItemStatusAvailable,
			},
			wantCode: codes.InvalidArgument,
			wantMsg:  "does not match desired category",
		},
		{
			name: "unit mismatch rejects approval",
			fusionReq: &domain.FusionRequest{
				ID:              "req-2",
				DesiredCategory: "野菜",
				DesiredQuantity: 5,
				Unit:            "kg",
				Status:          domain.FusionRequestStatusPending,
			},
			foodItem: &domain.FoodItem{
				ID:       "food-2",
				Category: "野菜",
				Quantity: 10,
				Unit:     "個",
				Status:   domain.FoodItemStatusAvailable,
			},
			wantCode: codes.InvalidArgument,
			wantMsg:  "does not match desired unit",
		},
		{
			name: "insufficient quantity rejects approval",
			fusionReq: &domain.FusionRequest{
				ID:              "req-3",
				DesiredCategory: "野菜",
				DesiredQuantity: 10,
				Unit:            "kg",
				Status:          domain.FusionRequestStatusPending,
			},
			foodItem: &domain.FoodItem{
				ID:       "food-3",
				Category: "野菜",
				Quantity: 3,
				Unit:     "kg",
				Status:   domain.FoodItemStatusAvailable,
			},
			wantCode: codes.InvalidArgument,
			wantMsg:  "is less than desired quantity",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fusionStore := newMockFusionRequestStore()
			foodStore := newMockFoodItemStore()

			fusionStore.requests[tt.fusionReq.ID] = tt.fusionReq
			foodStore.items[tt.foodItem.ID] = tt.foodItem

			svc := NewFusionService(fusionStore, foodStore)

			_, err := svc.RespondToFusionRequest(context.Background(), &pb.RespondToFusionRequestRequest{
				FusionRequestId: tt.fusionReq.ID,
				Response:        "APPROVED",
				FoodItemId:      tt.foodItem.ID,
			})

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

			// FoodItem のステータスが変更されていないことを確認
			item, _ := foodStore.Get(context.Background(), tt.foodItem.ID)
			if item.Status != domain.FoodItemStatusAvailable {
				t.Errorf("expected food item status to remain %q, got %q",
					domain.FoodItemStatusAvailable, item.Status)
			}
		})
	}
}

func TestRespondToFusionRequest_ApprovalSuccess(t *testing.T) {
	fusionStore := newMockFusionRequestStore()
	foodStore := newMockFoodItemStore()

	fusionReq := &domain.FusionRequest{
		ID:              "req-ok",
		DesiredCategory: "野菜",
		DesiredQuantity: 5,
		Unit:            "kg",
		Status:          domain.FusionRequestStatusPending,
	}
	foodItem := &domain.FoodItem{
		ID:       "food-ok",
		Category: "野菜",
		Quantity: 10,
		Unit:     "kg",
		Status:   domain.FoodItemStatusAvailable,
	}

	fusionStore.requests[fusionReq.ID] = fusionReq
	foodStore.items[foodItem.ID] = foodItem

	svc := NewFusionService(fusionStore, foodStore)

	resp, err := svc.RespondToFusionRequest(context.Background(), &pb.RespondToFusionRequestRequest{
		FusionRequestId: fusionReq.ID,
		Response:        "APPROVED",
		FoodItemId:      foodItem.ID,
	})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if resp.FusionRequest.Status != "approved" {
		t.Errorf("expected status %q, got %q", "approved", resp.FusionRequest.Status)
	}

	// FoodItem のステータスが reserved に変更されていることを確認
	item, _ := foodStore.Get(context.Background(), foodItem.ID)
	if item.Status != domain.FoodItemStatusReserved {
		t.Errorf("expected food item status %q, got %q",
			domain.FoodItemStatusReserved, item.Status)
	}
}
