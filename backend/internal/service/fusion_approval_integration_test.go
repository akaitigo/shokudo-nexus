package service

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/akaitigo/shokudo-nexus/backend/gen/shokudo/v1"
	"github.com/akaitigo/shokudo-nexus/backend/internal/domain"
	"github.com/akaitigo/shokudo-nexus/backend/internal/repository"
)

// newEmulatorClient は Firestore エミュレータに接続したクライアントを返す。
// FIRESTORE_EMULATOR_HOST 未設定時はテストをスキップする。
func newEmulatorClient(t *testing.T) *firestore.Client {
	t.Helper()
	if os.Getenv("FIRESTORE_EMULATOR_HOST") == "" {
		t.Skip("FIRESTORE_EMULATOR_HOST is not set; skipping Firestore integration test")
	}
	client, err := firestore.NewClient(context.Background(), "test-"+uuid.NewString(), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("failed to create Firestore client: %v", err)
	}
	t.Cleanup(func() {
		if closeErr := client.Close(); closeErr != nil {
			t.Logf("failed to close Firestore client: %v", closeErr)
		}
	})
	return client
}

// TestRespondToFusionRequest_ConcurrentApprovalAcrossProcesses は #32 の統合テスト。
// 実際の Firestore トランザクション（エミュレータ）を用い、同一在庫への並行承認が
// ちょうど1件だけ成功することを検証する。単一プロセスの Mutex ではなく Firestore の
// 楽観的並行制御によって二重予約が防止されることを確認する。
func TestRespondToFusionRequest_ConcurrentApprovalAcrossProcesses(t *testing.T) {
	client := newEmulatorClient(t)
	ctx := context.Background()

	foodRepo := repository.NewFoodItemRepository(client)
	fusionRepo := repository.NewFusionRequestRepository(client)
	runner := repository.NewFirestoreApprovalRunner(client)

	now := time.Now().UTC()
	food, err := foodRepo.Create(ctx, &domain.FoodItem{
		Name:       "共有にんじん",
		Category:   "野菜",
		ExpiryDate: now.Add(720 * time.Hour),
		Quantity:   5,
		Unit:       "kg",
		DonorID:    "shokudo-donor",
		Status:     domain.FoodItemStatusAvailable,
		CreatedAt:  now,
		UpdatedAt:  now,
	})
	if err != nil {
		t.Fatalf("seed food item: %v", err)
	}

	newPendingRequest := func() *domain.FusionRequest {
		req, createErr := fusionRepo.Create(ctx, &domain.FusionRequest{
			RequesterShokudoID: "shokudo-requester",
			DesiredCategory:    "野菜",
			DesiredQuantity:    5,
			Unit:               "kg",
			Status:             domain.FusionRequestStatusPending,
			CreatedAt:          now,
			UpdatedAt:          now,
		})
		if createErr != nil {
			t.Fatalf("seed fusion request: %v", createErr)
		}
		return req
	}
	req1 := newPendingRequest()
	req2 := newPendingRequest()

	// 別プロセス相当の2つのサービスが、同一在庫を同時に承認する。
	svcA := NewFusionService(fusionRepo, foodRepo, runner)
	svcB := NewFusionService(fusionRepo, foodRepo, runner)

	var wg sync.WaitGroup
	results := make(chan error, 2)
	approve := func(svc *FusionService, reqID string) {
		defer wg.Done()
		_, respErr := svc.RespondToFusionRequest(ctx, &pb.RespondToFusionRequestRequest{
			FusionRequestId: reqID,
			Response:        "APPROVED",
			FoodItemId:      food.ID,
		})
		results <- respErr
	}
	wg.Add(2)
	go approve(svcA, req1.ID)
	go approve(svcB, req2.ID)
	wg.Wait()
	close(results)

	var successCount, failCount int
	for respErr := range results {
		if respErr == nil {
			successCount++
			continue
		}
		failCount++
		// 失敗側は在庫が予約済みになったことによる FailedPrecondition。
		if status.Code(respErr) != codes.FailedPrecondition {
			t.Errorf("expected FailedPrecondition for the losing approval, got %v (%v)", status.Code(respErr), respErr)
		}
	}
	if successCount != 1 {
		t.Errorf("expected exactly 1 successful approval, got %d", successCount)
	}
	if failCount != 1 {
		t.Errorf("expected exactly 1 failed approval, got %d", failCount)
	}

	// 在庫はちょうど1回だけ reserved になっている。
	gotFood, err := foodRepo.Get(ctx, food.ID)
	if err != nil {
		t.Fatalf("get food item: %v", err)
	}
	if gotFood.Status != domain.FoodItemStatusReserved {
		t.Errorf("expected food item status %q, got %q", domain.FoodItemStatusReserved, gotFood.Status)
	}
}
