package repository

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/akaitigo/shokudo-nexus/backend/internal/domain"
)

func newFoodItem(name, category string, createdAt time.Time) *domain.FoodItem {
	return &domain.FoodItem{
		Name:       name,
		Category:   category,
		ExpiryDate: createdAt.Add(30 * 24 * time.Hour),
		Quantity:   10,
		Unit:       "kg",
		DonorID:    "donor-1",
		Status:     domain.FoodItemStatusAvailable,
		CreatedAt:  createdAt,
		UpdatedAt:  createdAt,
	}
}

func TestFoodItemRepository_CreateAndGet(t *testing.T) {
	ctx := context.Background()
	repo := NewFoodItemRepository(newTestClient(t))

	created, err := repo.Create(ctx, newFoodItem("にんじん", "野菜", time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)))
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

	// データ変換のラウンドトリップを検証する。
	if got.ID != created.ID {
		t.Errorf("ID mismatch: got %q, want %q", got.ID, created.ID)
	}
	if got.Name != "にんじん" {
		t.Errorf("Name mismatch: got %q, want %q", got.Name, "にんじん")
	}
	if got.Category != "野菜" {
		t.Errorf("Category mismatch: got %q, want %q", got.Category, "野菜")
	}
	if got.Quantity != 10 {
		t.Errorf("Quantity mismatch: got %d, want %d", got.Quantity, 10)
	}
	if got.Unit != "kg" {
		t.Errorf("Unit mismatch: got %q, want %q", got.Unit, "kg")
	}
	if got.Status != domain.FoodItemStatusAvailable {
		t.Errorf("Status mismatch: got %q, want %q", got.Status, domain.FoodItemStatusAvailable)
	}
	if !got.CreatedAt.Equal(created.CreatedAt) {
		t.Errorf("CreatedAt mismatch: got %v, want %v", got.CreatedAt, created.CreatedAt)
	}
}

func TestFoodItemRepository_Get_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := NewFoodItemRepository(newTestClient(t))

	_, err := repo.Get(ctx, "does-not-exist")
	if err == nil {
		t.Fatal("expected error for missing item, got nil")
	}
	if status.Code(err) != codes.NotFound {
		t.Errorf("expected NotFound, got %v", status.Code(err))
	}
}

func TestFoodItemRepository_Update(t *testing.T) {
	ctx := context.Background()
	repo := NewFoodItemRepository(newTestClient(t))

	created, err := repo.Create(ctx, newFoodItem("牛乳", "乳製品", time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)))
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	created.Quantity = 3
	created.Status = domain.FoodItemStatusReserved
	if updateErr := repo.Update(ctx, created); updateErr != nil {
		t.Fatalf("Update failed: %v", updateErr)
	}

	got, err := repo.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.Quantity != 3 {
		t.Errorf("Quantity not updated: got %d, want 3", got.Quantity)
	}
	if got.Status != domain.FoodItemStatusReserved {
		t.Errorf("Status not updated: got %q, want %q", got.Status, domain.FoodItemStatusReserved)
	}
}

func TestFoodItemRepository_Delete_LogicalDeletion(t *testing.T) {
	ctx := context.Background()
	repo := NewFoodItemRepository(newTestClient(t))

	created, err := repo.Create(ctx, newFoodItem("鶏肉", "肉", time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)))
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if delErr := repo.Delete(ctx, created.ID); delErr != nil {
		t.Fatalf("Delete failed: %v", delErr)
	}

	// 論理削除なのでドキュメントは残り、ステータスが deleted になる。
	got, err := repo.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("Get after delete failed: %v", err)
	}
	if got.Status != domain.FoodItemStatusDeleted {
		t.Errorf("expected status %q after delete, got %q", domain.FoodItemStatusDeleted, got.Status)
	}
}

func TestFoodItemRepository_Delete_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := NewFoodItemRepository(newTestClient(t))

	err := repo.Delete(ctx, "does-not-exist")
	if err == nil {
		t.Fatal("expected error deleting missing item, got nil")
	}
	if status.Code(err) != codes.NotFound {
		t.Errorf("expected NotFound, got %v", status.Code(err))
	}
}

func TestFoodItemRepository_List_OrdersByCreatedAtDesc(t *testing.T) {
	ctx := context.Background()
	repo := NewFoodItemRepository(newTestClient(t))

	base := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	// 作成順とは逆に登録し、created_at 降順で返ることを確認する。
	items := []*domain.FoodItem{
		newFoodItem("古い", "野菜", base),
		newFoodItem("新しい", "野菜", base.Add(2*time.Hour)),
		newFoodItem("中間", "野菜", base.Add(1*time.Hour)),
	}
	for _, it := range items {
		if _, err := repo.Create(ctx, it); err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	result, err := repo.List(ctx, ListParams{PageSize: 10})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(result.Items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(result.Items))
	}
	if result.TotalCount != 3 {
		t.Errorf("expected TotalCount 3, got %d", result.TotalCount)
	}

	wantOrder := []string{"新しい", "中間", "古い"}
	for i, want := range wantOrder {
		if result.Items[i].Name != want {
			t.Errorf("position %d: got %q, want %q", i, result.Items[i].Name, want)
		}
	}
}

func TestFoodItemRepository_List_CategoryFilter(t *testing.T) {
	ctx := context.Background()
	repo := NewFoodItemRepository(newTestClient(t))

	base := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	if _, err := repo.Create(ctx, newFoodItem("にんじん", "野菜", base)); err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if _, err := repo.Create(ctx, newFoodItem("鶏肉", "肉", base.Add(time.Hour))); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	result, err := repo.List(ctx, ListParams{PageSize: 10, CategoryFilter: "野菜"})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("expected 1 item for category filter, got %d", len(result.Items))
	}
	if result.Items[0].Category != "野菜" {
		t.Errorf("expected category 野菜, got %q", result.Items[0].Category)
	}
	if result.TotalCount != 1 {
		t.Errorf("expected TotalCount 1, got %d", result.TotalCount)
	}
}

func TestFoodItemRepository_List_ExcludesDeleted(t *testing.T) {
	ctx := context.Background()
	repo := NewFoodItemRepository(newTestClient(t))

	base := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	active, err := repo.Create(ctx, newFoodItem("有効", "野菜", base))
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	deleted, err := repo.Create(ctx, newFoodItem("削除対象", "野菜", base.Add(time.Hour)))
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if delErr := repo.Delete(ctx, deleted.ID); delErr != nil {
		t.Fatalf("Delete failed: %v", delErr)
	}

	result, err := repo.List(ctx, ListParams{PageSize: 10})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("expected 1 item (deleted excluded), got %d", len(result.Items))
	}
	if result.Items[0].ID != active.ID {
		t.Errorf("expected active item %q, got %q", active.ID, result.Items[0].ID)
	}
}

func TestFoodItemRepository_List_IncludesExpiredStatus(t *testing.T) {
	ctx := context.Background()
	repo := NewFoodItemRepository(newTestClient(t))

	base := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	item := newFoodItem("期限切れ", "野菜", base)
	item.Status = domain.FoodItemStatusExpired
	if _, err := repo.Create(ctx, item); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	result, err := repo.List(ctx, ListParams{PageSize: 10})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	// expired はアクティブステータスに含まれるため一覧に表示される。
	if len(result.Items) != 1 {
		t.Fatalf("expected expired item to be listed, got %d items", len(result.Items))
	}
}

func TestFoodItemRepository_List_Pagination(t *testing.T) {
	ctx := context.Background()
	repo := NewFoodItemRepository(newTestClient(t))

	base := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	const total = 5
	for i := range total {
		// created_at を1分ずつずらして全順序を一意にする。
		it := newFoodItem("item", "野菜", base.Add(time.Duration(i)*time.Minute))
		if _, err := repo.Create(ctx, it); err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	// ページサイズ2で全件を走査し、重複・欠落がないことを確認する。
	seen := make(map[string]bool)
	pageToken := ""
	pages := 0
	for {
		result, err := repo.List(ctx, ListParams{PageSize: 2, PageToken: pageToken})
		if err != nil {
			t.Fatalf("List page failed: %v", err)
		}
		if result.TotalCount != total {
			t.Errorf("expected TotalCount %d, got %d", total, result.TotalCount)
		}
		for _, it := range result.Items {
			if seen[it.ID] {
				t.Errorf("duplicate item across pages: %q", it.ID)
			}
			seen[it.ID] = true
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
		t.Errorf("expected to see %d unique items across pages, got %d", total, len(seen))
	}
}

func TestFoodItemRepository_List_InvalidPageToken(t *testing.T) {
	ctx := context.Background()
	repo := NewFoodItemRepository(newTestClient(t))

	_, err := repo.List(ctx, ListParams{PageSize: 10, PageToken: "nonexistent-doc-id"})
	if err == nil {
		t.Fatal("expected error for invalid page token, got nil")
	}
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", status.Code(err))
	}
}

func TestFoodItemRepository_List_Empty(t *testing.T) {
	ctx := context.Background()
	repo := NewFoodItemRepository(newTestClient(t))

	result, err := repo.List(ctx, ListParams{PageSize: 10})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(result.Items) != 0 {
		t.Errorf("expected empty list, got %d items", len(result.Items))
	}
	if result.NextPageToken != "" {
		t.Errorf("expected empty NextPageToken, got %q", result.NextPageToken)
	}
	if result.TotalCount != 0 {
		t.Errorf("expected TotalCount 0, got %d", result.TotalCount)
	}
}
