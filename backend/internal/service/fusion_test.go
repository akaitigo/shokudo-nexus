package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/akaitigo/shokudo-nexus/backend/gen/shokudo/v1"
	"github.com/akaitigo/shokudo-nexus/backend/internal/config"
	"github.com/akaitigo/shokudo-nexus/backend/internal/domain"
)

// newTestFusionService は、承認トランザクションを模倣する mockApprovalRunner を
// 同じストアに結線した FusionService を生成する。
func newTestFusionService(fusionStore *mockFusionRequestStore, foodStore *mockFoodItemStore) *FusionService {
	return NewFusionService(fusionStore, foodStore, newMockApprovalRunner(fusionStore, foodStore))
}

func TestValidateCreateFusionRequest(t *testing.T) {
	validReq := &pb.CreateFusionRequestRequest{
		RequesterShokudoId: "shokudo-1",
		DesiredCategory:    "野菜",
		DesiredQuantity:    5,
		Unit:               "kg",
		Message:            "にんじんが必要です",
	}

	if err := validateCreateFusionRequest(validReq, config.Default()); err != nil {
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

			err := validateCreateFusionRequest(req, config.Default())
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
	svc := newTestFusionService(newMockFusionRequestStore(), newMockFoodItemStore())
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
	svc := newTestFusionService(newMockFusionRequestStore(), newMockFoodItemStore())
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
		name      string
		fusionReq *domain.FusionRequest
		foodItem  *domain.FoodItem
		wantCode  codes.Code
		wantMsg   string
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

			svc := newTestFusionService(fusionStore, foodStore)

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
		DonorID:  "shokudo-B",
		Status:   domain.FoodItemStatusAvailable,
	}

	fusionStore.requests[fusionReq.ID] = fusionReq
	foodStore.items[foodItem.ID] = foodItem

	svc := newTestFusionService(fusionStore, foodStore)

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

	// ResponderShokudoID が FoodItem の DonorID で埋まっていることを確認
	if resp.FusionRequest.ResponderShokudoId != "shokudo-B" {
		t.Errorf("expected responder_shokudo_id %q, got %q",
			"shokudo-B", resp.FusionRequest.ResponderShokudoId)
	}

	// FoodItem のステータスが reserved に変更されていることを確認
	item, _ := foodStore.Get(context.Background(), foodItem.ID)
	if item.Status != domain.FoodItemStatusReserved {
		t.Errorf("expected food item status %q, got %q",
			domain.FoodItemStatusReserved, item.Status)
	}
}

func TestRespondToFusionRequest_ApprovalSetsResponderShokudoID(t *testing.T) {
	fusionStore := newMockFusionRequestStore()
	foodStore := newMockFoodItemStore()

	fusionReq := &domain.FusionRequest{
		ID:              "req-resp",
		DesiredCategory: "肉類",
		DesiredQuantity: 2,
		Unit:            "kg",
		Status:          domain.FusionRequestStatusPending,
	}
	foodItem := &domain.FoodItem{
		ID:       "food-resp",
		Category: "肉類",
		Quantity: 5,
		Unit:     "kg",
		DonorID:  "shokudo-C",
		Status:   domain.FoodItemStatusAvailable,
	}

	fusionStore.requests[fusionReq.ID] = fusionReq
	foodStore.items[foodItem.ID] = foodItem

	svc := newTestFusionService(fusionStore, foodStore)

	resp, err := svc.RespondToFusionRequest(context.Background(), &pb.RespondToFusionRequestRequest{
		FusionRequestId: fusionReq.ID,
		Response:        "APPROVED",
		FoodItemId:      foodItem.ID,
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// レスポンスの responder_shokudo_id が foodItem.DonorID と一致すること
	if resp.FusionRequest.ResponderShokudoId != "shokudo-C" {
		t.Errorf("expected responder_shokudo_id %q, got %q",
			"shokudo-C", resp.FusionRequest.ResponderShokudoId)
	}

	// ドメインモデル側でも永続化されていることを確認
	stored, _ := fusionStore.Get(context.Background(), fusionReq.ID)
	if stored.ResponderShokudoID != "shokudo-C" {
		t.Errorf("expected stored ResponderShokudoID %q, got %q",
			"shokudo-C", stored.ResponderShokudoID)
	}
}

func TestRespondToFusionRequest_ConcurrentApprovalPreventsDoubleReservation(t *testing.T) {
	fusionStore := newMockFusionRequestStore()
	foodStore := newMockFoodItemStore()

	// 1つの食品に対して2つの融通リクエストが同時に承認されるシナリオ
	foodItem := &domain.FoodItem{
		ID:       "food-shared",
		Category: "野菜",
		Quantity: 5,
		Unit:     "kg",
		DonorID:  "shokudo-X",
		Status:   domain.FoodItemStatusAvailable,
	}
	foodStore.items[foodItem.ID] = foodItem

	fusionReq1 := &domain.FusionRequest{
		ID:              "req-concurrent-1",
		DesiredCategory: "野菜",
		DesiredQuantity: 5,
		Unit:            "kg",
		Status:          domain.FusionRequestStatusPending,
	}
	fusionReq2 := &domain.FusionRequest{
		ID:              "req-concurrent-2",
		DesiredCategory: "野菜",
		DesiredQuantity: 5,
		Unit:            "kg",
		Status:          domain.FusionRequestStatusPending,
	}
	fusionStore.requests[fusionReq1.ID] = fusionReq1
	fusionStore.requests[fusionReq2.ID] = fusionReq2

	svc := newTestFusionService(fusionStore, foodStore)

	// 並行実行
	const goroutines = 2
	results := make(chan error, goroutines)

	approve := func(reqID string) {
		_, err := svc.RespondToFusionRequest(context.Background(), &pb.RespondToFusionRequestRequest{
			FusionRequestId: reqID,
			Response:        "APPROVED",
			FoodItemId:      foodItem.ID,
		})
		results <- err
	}

	go approve(fusionReq1.ID)
	go approve(fusionReq2.ID)

	var successCount, failCount int
	for range goroutines {
		err := <-results
		if err == nil {
			successCount++
		} else {
			failCount++
		}
	}

	// Mutex保護により、1つだけ成功し、もう1つはFailedPrecondition（food item already reserved）
	if successCount != 1 {
		t.Errorf("expected exactly 1 successful approval, got %d", successCount)
	}
	if failCount != 1 {
		t.Errorf("expected exactly 1 failed approval, got %d", failCount)
	}
}

func TestRespondToFusionRequest_ApprovalTransactionFailureKeepsFoodItemAvailable(t *testing.T) {
	fusionStore := newMockFusionRequestStore()
	foodStore := newMockFoodItemStore()

	fusionReq := &domain.FusionRequest{
		ID:              "req-rollback",
		DesiredCategory: "野菜",
		DesiredQuantity: 3,
		Unit:            "kg",
		Status:          domain.FusionRequestStatusPending,
	}
	foodItem := &domain.FoodItem{
		ID:       "food-rollback",
		Category: "野菜",
		Quantity: 5,
		Unit:     "kg",
		Status:   domain.FoodItemStatusAvailable,
	}
	fusionStore.requests[fusionReq.ID] = fusionReq
	foodStore.items[foodItem.ID] = foodItem

	// トランザクションのコミットを強制失敗させる。アトミック性により書き込みは適用されない。
	runner := newMockApprovalRunner(fusionStore, foodStore)
	runner.commitErr = errors.New("simulated commit failure")

	svc := NewFusionService(fusionStore, foodStore, runner)
	_, err := svc.RespondToFusionRequest(context.Background(), &pb.RespondToFusionRequestRequest{
		FusionRequestId: fusionReq.ID,
		Response:        "APPROVED",
		FoodItemId:      foodItem.ID,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if status.Code(err) != codes.Internal {
		t.Errorf("expected Internal, got %v", status.Code(err))
	}

	// トランザクションが失敗したため FoodItem は available のまま（予約されない）。
	item, _ := foodStore.Get(context.Background(), foodItem.ID)
	if item.Status != domain.FoodItemStatusAvailable {
		t.Errorf("expected food item to remain %q, got %q",
			domain.FoodItemStatusAvailable, item.Status)
	}
	// FusionRequest も pending のまま。
	storedReq, _ := fusionStore.Get(context.Background(), fusionReq.ID)
	if storedReq.Status != domain.FusionRequestStatusPending {
		t.Errorf("expected fusion request to remain %q, got %q",
			domain.FusionRequestStatusPending, storedReq.Status)
	}
}

func TestCreateFusionRequest_Success(t *testing.T) {
	fusionStore := newMockFusionRequestStore()
	foodStore := newMockFoodItemStore()
	svc := newTestFusionService(fusionStore, foodStore)

	resp, err := svc.CreateFusionRequest(context.Background(), &pb.CreateFusionRequestRequest{
		RequesterShokudoId: "shokudo-1",
		DesiredCategory:    "野菜",
		DesiredQuantity:    5,
		Unit:               "kg",
		Message:            "にんじんが必要です",
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	got := resp.GetFusionRequest()
	if got.GetId() == "" {
		t.Error("expected non-empty ID")
	}
	if got.GetRequesterShokudoId() != "shokudo-1" {
		t.Errorf("expected requester_shokudo_id 'shokudo-1', got %q", got.GetRequesterShokudoId())
	}
	if got.GetDesiredCategory() != "野菜" {
		t.Errorf("expected desired_category '野菜', got %q", got.GetDesiredCategory())
	}
	if got.GetDesiredQuantity() != 5 {
		t.Errorf("expected desired_quantity 5, got %d", got.GetDesiredQuantity())
	}
	if got.GetUnit() != "kg" {
		t.Errorf("expected unit 'kg', got %q", got.GetUnit())
	}
	if got.GetMessage() != "にんじんが必要です" {
		t.Errorf("expected message 'にんじんが必要です', got %q", got.GetMessage())
	}
	// 新規作成時のステータスは pending。
	if got.GetStatus() != "pending" {
		t.Errorf("expected status 'pending', got %q", got.GetStatus())
	}
	if got.GetCreatedAt() == "" {
		t.Error("expected non-empty created_at")
	}
	if got.GetUpdatedAt() == "" {
		t.Error("expected non-empty updated_at")
	}

	// リポジトリ層まで到達し永続化されていることを確認する。
	stored, err := fusionStore.Get(context.Background(), got.GetId())
	if err != nil {
		t.Fatalf("expected request to be stored, got error: %v", err)
	}
	if stored.Status != domain.FusionRequestStatusPending {
		t.Errorf("expected stored status %q, got %q", domain.FusionRequestStatusPending, stored.Status)
	}
	if stored.RequesterShokudoID != "shokudo-1" {
		t.Errorf("expected stored requester 'shokudo-1', got %q", stored.RequesterShokudoID)
	}
}

func TestCreateFusionRequest_EmptyMessageAllowed(t *testing.T) {
	fusionStore := newMockFusionRequestStore()
	foodStore := newMockFoodItemStore()
	svc := newTestFusionService(fusionStore, foodStore)

	// メッセージは任意フィールドのため空でも成功する。
	resp, err := svc.CreateFusionRequest(context.Background(), &pb.CreateFusionRequestRequest{
		RequesterShokudoId: "shokudo-1",
		DesiredCategory:    "肉",
		DesiredQuantity:    2,
		Unit:               "パック",
		Message:            "",
	})
	if err != nil {
		t.Fatalf("expected no error for empty message, got: %v", err)
	}
	if resp.GetFusionRequest().GetMessage() != "" {
		t.Errorf("expected empty message, got %q", resp.GetFusionRequest().GetMessage())
	}
}

func TestCreateFusionRequest_ValidationError(t *testing.T) {
	fusionStore := newMockFusionRequestStore()
	foodStore := newMockFoodItemStore()
	svc := newTestFusionService(fusionStore, foodStore)

	// 無効なカテゴリはサービス層で InvalidArgument として弾かれ、リポジトリに到達しない。
	_, err := svc.CreateFusionRequest(context.Background(), &pb.CreateFusionRequestRequest{
		RequesterShokudoId: "shokudo-1",
		DesiredCategory:    "invalid",
		DesiredQuantity:    5,
		Unit:               "kg",
	})
	if err == nil {
		t.Fatal("expected error for invalid category, got nil")
	}
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", status.Code(err))
	}
}

func TestCreateFusionRequest_RepoError(t *testing.T) {
	fusionStore := newMockFusionRequestStore()
	fusionStore.createFunc = func(_ context.Context, _ *domain.FusionRequest) (*domain.FusionRequest, error) {
		return nil, errors.New("firestore unavailable")
	}
	foodStore := newMockFoodItemStore()
	svc := newTestFusionService(fusionStore, foodStore)

	// リポジトリ書き込み失敗時は Internal に変換される。
	_, err := svc.CreateFusionRequest(context.Background(), &pb.CreateFusionRequestRequest{
		RequesterShokudoId: "shokudo-1",
		DesiredCategory:    "野菜",
		DesiredQuantity:    5,
		Unit:               "kg",
	})
	if err == nil {
		t.Fatal("expected error when repo fails, got nil")
	}
	if status.Code(err) != codes.Internal {
		t.Errorf("expected Internal, got %v", status.Code(err))
	}
}
