package repository

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/akaitigo/shokudo-nexus/backend/internal/domain"
)

func newFusionRequest(category string, status domain.FusionRequestStatus, createdAt time.Time) *domain.FusionRequest {
	return &domain.FusionRequest{
		RequesterShokudoID: "shokudo-1",
		DesiredCategory:    category,
		DesiredQuantity:    5,
		Unit:               "kg",
		Message:            "テストメッセージ",
		Status:             status,
		CreatedAt:          createdAt,
		UpdatedAt:          createdAt,
	}
}

func TestFusionRequestRepository_CreateAndGet(t *testing.T) {
	ctx := context.Background()
	repo := NewFusionRequestRepository(newTestClient(t))

	created, err := repo.Create(ctx, newFusionRequest("野菜", domain.FusionRequestStatusPending, time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)))
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if created.ID == "" {
		t.Fatal("expected generated ID, got empty string")
	}

	got, err := repo.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("ID mismatch: got %q, want %q", got.ID, created.ID)
	}
	if got.RequesterShokudoID != "shokudo-1" {
		t.Errorf("RequesterShokudoID mismatch: got %q", got.RequesterShokudoID)
	}
	if got.DesiredCategory != "野菜" {
		t.Errorf("DesiredCategory mismatch: got %q", got.DesiredCategory)
	}
	if got.DesiredQuantity != 5 {
		t.Errorf("DesiredQuantity mismatch: got %d", got.DesiredQuantity)
	}
	if got.Status != domain.FusionRequestStatusPending {
		t.Errorf("Status mismatch: got %q", got.Status)
	}
}

func TestFusionRequestRepository_Get_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := NewFusionRequestRepository(newTestClient(t))

	_, err := repo.Get(ctx, "does-not-exist")
	if err == nil {
		t.Fatal("expected error for missing request, got nil")
	}
	if status.Code(err) != codes.NotFound {
		t.Errorf("expected NotFound, got %v", status.Code(err))
	}
}

func TestFusionRequestRepository_Update(t *testing.T) {
	ctx := context.Background()
	repo := NewFusionRequestRepository(newTestClient(t))

	created, err := repo.Create(ctx, newFusionRequest("野菜", domain.FusionRequestStatusPending, time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)))
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	created.Status = domain.FusionRequestStatusApproved
	created.ResponderShokudoID = "shokudo-2"
	created.FoodItemID = "food-1"
	if updateErr := repo.Update(ctx, created); updateErr != nil {
		t.Fatalf("Update failed: %v", updateErr)
	}

	got, err := repo.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.Status != domain.FusionRequestStatusApproved {
		t.Errorf("Status not updated: got %q", got.Status)
	}
	if got.ResponderShokudoID != "shokudo-2" {
		t.Errorf("ResponderShokudoID not updated: got %q", got.ResponderShokudoID)
	}
	if got.FoodItemID != "food-1" {
		t.Errorf("FoodItemID not updated: got %q", got.FoodItemID)
	}
}

func TestFusionRequestRepository_List_All(t *testing.T) {
	ctx := context.Background()
	repo := NewFusionRequestRepository(newTestClient(t))

	base := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	for i := range 3 {
		req := newFusionRequest("野菜", domain.FusionRequestStatusPending, base.Add(time.Duration(i)*time.Hour))
		if _, err := repo.Create(ctx, req); err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	result, err := repo.List(ctx, FusionListParams{PageSize: 10})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(result.Requests) != 3 {
		t.Fatalf("expected 3 requests, got %d", len(result.Requests))
	}
	if result.TotalCount != 3 {
		t.Errorf("expected TotalCount 3, got %d", result.TotalCount)
	}

	// created_at 降順で返ることを確認する。
	for i := 1; i < len(result.Requests); i++ {
		if result.Requests[i-1].CreatedAt.Before(result.Requests[i].CreatedAt) {
			t.Errorf("results not ordered by created_at desc at position %d", i)
		}
	}
}

func TestFusionRequestRepository_List_StatusFilter(t *testing.T) {
	ctx := context.Background()
	repo := NewFusionRequestRepository(newTestClient(t))

	base := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	if _, err := repo.Create(ctx, newFusionRequest("野菜", domain.FusionRequestStatusPending, base)); err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if _, err := repo.Create(ctx, newFusionRequest("肉", domain.FusionRequestStatusApproved, base.Add(time.Hour))); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	result, err := repo.List(ctx, FusionListParams{PageSize: 10, StatusFilter: string(domain.FusionRequestStatusApproved)})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(result.Requests) != 1 {
		t.Fatalf("expected 1 request for status filter, got %d", len(result.Requests))
	}
	if result.Requests[0].Status != domain.FusionRequestStatusApproved {
		t.Errorf("expected approved status, got %q", result.Requests[0].Status)
	}
	if result.TotalCount != 1 {
		t.Errorf("expected TotalCount 1, got %d", result.TotalCount)
	}
}

func TestFusionRequestRepository_List_Pagination(t *testing.T) {
	ctx := context.Background()
	repo := NewFusionRequestRepository(newTestClient(t))

	base := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	const total = 5
	for i := range total {
		req := newFusionRequest("野菜", domain.FusionRequestStatusPending, base.Add(time.Duration(i)*time.Minute))
		if _, err := repo.Create(ctx, req); err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	seen := make(map[string]bool)
	pageToken := ""
	pages := 0
	for {
		result, err := repo.List(ctx, FusionListParams{PageSize: 2, PageToken: pageToken})
		if err != nil {
			t.Fatalf("List page failed: %v", err)
		}
		if result.TotalCount != total {
			t.Errorf("expected TotalCount %d, got %d", total, result.TotalCount)
		}
		for _, req := range result.Requests {
			if seen[req.ID] {
				t.Errorf("duplicate request across pages: %q", req.ID)
			}
			seen[req.ID] = true
		}
		pages++
		if pages > total+2 {
			t.Fatal("pagination did not terminate")
		}
		if result.NextPageToken == "" {
			break
		}
		pageToken = result.NextPageToken
	}

	if len(seen) != total {
		t.Errorf("expected %d unique requests across pages, got %d", total, len(seen))
	}
}

func TestFusionRequestRepository_List_InvalidPageToken(t *testing.T) {
	ctx := context.Background()
	repo := NewFusionRequestRepository(newTestClient(t))

	_, err := repo.List(ctx, FusionListParams{PageSize: 10, PageToken: "nonexistent-doc-id"})
	if err == nil {
		t.Fatal("expected error for invalid page token, got nil")
	}
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", status.Code(err))
	}
}
