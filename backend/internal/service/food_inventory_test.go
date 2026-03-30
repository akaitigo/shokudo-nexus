package service

import (
	"context"
	"strings"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/akaitigo/shokudo-nexus/backend/gen/shokudo/v1"
)

func TestValidateCreateFoodItemRequest(t *testing.T) {
	validReq := &pb.CreateFoodItemRequest{
		Name:       "にんじん",
		Category:   "野菜",
		ExpiryDate: "2026-04-15T00:00:00Z",
		Quantity:   10,
		Unit:       "kg",
		DonorId:    "donor-1",
	}

	// 正常系
	if err := validateCreateFoodItemRequest(validReq); err != nil {
		t.Errorf("expected nil error for valid request, got: %v", err)
	}

	tests := []struct {
		name     string
		modify   func(r *pb.CreateFoodItemRequest)
		wantCode codes.Code
		wantMsg  string
	}{
		{
			name:     "empty name",
			modify:   func(r *pb.CreateFoodItemRequest) { r.Name = "" },
			wantCode: codes.InvalidArgument,
			wantMsg:  "name is required",
		},
		{
			name: "name too long",
			modify: func(r *pb.CreateFoodItemRequest) {
				r.Name = strings.Repeat("あ", 201)
			},
			wantCode: codes.InvalidArgument,
			wantMsg:  "at most 200 characters",
		},
		{
			name:     "invalid category",
			modify:   func(r *pb.CreateFoodItemRequest) { r.Category = "invalid" },
			wantCode: codes.InvalidArgument,
			wantMsg:  "invalid category",
		},
		{
			name:     "empty expiry_date",
			modify:   func(r *pb.CreateFoodItemRequest) { r.ExpiryDate = "" },
			wantCode: codes.InvalidArgument,
			wantMsg:  "expiry_date is required",
		},
		{
			name:     "quantity too low",
			modify:   func(r *pb.CreateFoodItemRequest) { r.Quantity = 0 },
			wantCode: codes.InvalidArgument,
			wantMsg:  "quantity must be between",
		},
		{
			name:     "quantity too high",
			modify:   func(r *pb.CreateFoodItemRequest) { r.Quantity = 10001 },
			wantCode: codes.InvalidArgument,
			wantMsg:  "quantity must be between",
		},
		{
			name:     "invalid unit",
			modify:   func(r *pb.CreateFoodItemRequest) { r.Unit = "リットル" },
			wantCode: codes.InvalidArgument,
			wantMsg:  "invalid unit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := cloneCreateFoodItemRequest(validReq)
			tt.modify(req)

			err := validateCreateFoodItemRequest(req)
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

func TestValidateCreateFoodItemRequest_NameExact200Chars(t *testing.T) {
	req := &pb.CreateFoodItemRequest{
		Name:       strings.Repeat("a", 200),
		Category:   "野菜",
		ExpiryDate: "2026-04-15T00:00:00Z",
		Quantity:   1,
		Unit:       "kg",
	}
	if err := validateCreateFoodItemRequest(req); err != nil {
		t.Errorf("expected nil error for 200-char name, got: %v", err)
	}
}

func TestGetFoodItem_EmptyID(t *testing.T) {
	svc := &FoodInventoryService{}
	_, err := svc.GetFoodItem(context.Background(), &pb.GetFoodItemRequest{Id: ""})
	if err == nil {
		t.Fatal("expected error for empty ID")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got: %v", err)
	}
	if st.Code() != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", st.Code())
	}
}

func TestDeleteFoodItem_EmptyID(t *testing.T) {
	svc := &FoodInventoryService{}
	_, err := svc.DeleteFoodItem(context.Background(), &pb.DeleteFoodItemRequest{Id: ""})
	if err == nil {
		t.Fatal("expected error for empty ID")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got: %v", err)
	}
	if st.Code() != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", st.Code())
	}
}

func TestListFoodItems_PageSizeTooLarge(t *testing.T) {
	svc := &FoodInventoryService{}
	_, err := svc.ListFoodItems(context.Background(), &pb.ListFoodItemsRequest{PageSize: 101})
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

func TestListFoodItems_InvalidCategoryFilter(t *testing.T) {
	svc := &FoodInventoryService{}
	_, err := svc.ListFoodItems(context.Background(), &pb.ListFoodItemsRequest{CategoryFilter: "invalid"})
	if err == nil {
		t.Fatal("expected error for invalid category filter")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got: %v", err)
	}
	if st.Code() != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", st.Code())
	}
}

func cloneCreateFoodItemRequest(r *pb.CreateFoodItemRequest) *pb.CreateFoodItemRequest {
	return &pb.CreateFoodItemRequest{
		Name:       r.Name,
		Category:   r.Category,
		ExpiryDate: r.ExpiryDate,
		Quantity:   r.Quantity,
		Unit:       r.Unit,
		DonorId:    r.DonorId,
	}
}
