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

// --- CreateFoodItem テスト ---

func TestCreateFoodItem_Success(t *testing.T) {
	mock := newMockFoodItemStore()
	svc := NewFoodInventoryService(mock)

	resp, err := svc.CreateFoodItem(context.Background(), &pb.CreateFoodItemRequest{
		Name:       "にんじん",
		Category:   "野菜",
		ExpiryDate: "2026-04-15T00:00:00Z",
		Quantity:   10,
		Unit:       "kg",
		DonorId:    "donor-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	item := resp.GetFoodItem()
	if item.GetId() == "" {
		t.Error("expected non-empty ID")
	}
	if item.GetName() != "にんじん" {
		t.Errorf("expected name 'にんじん', got %q", item.GetName())
	}
	if item.GetCategory() != "野菜" {
		t.Errorf("expected category '野菜', got %q", item.GetCategory())
	}
	if item.GetQuantity() != 10 {
		t.Errorf("expected quantity 10, got %d", item.GetQuantity())
	}
	if item.GetUnit() != "kg" {
		t.Errorf("expected unit 'kg', got %q", item.GetUnit())
	}
	if item.GetDonorId() != "donor-1" {
		t.Errorf("expected donor_id 'donor-1', got %q", item.GetDonorId())
	}
	if item.GetStatus() != "available" {
		t.Errorf("expected status 'available', got %q", item.GetStatus())
	}
	if item.GetCreatedAt() == "" {
		t.Error("expected non-empty created_at")
	}
	if item.GetUpdatedAt() == "" {
		t.Error("expected non-empty updated_at")
	}
}

func TestCreateFoodItem_InvalidExpiryDateFormat(t *testing.T) {
	mock := newMockFoodItemStore()
	svc := NewFoodInventoryService(mock)

	_, err := svc.CreateFoodItem(context.Background(), &pb.CreateFoodItemRequest{
		Name:       "にんじん",
		Category:   "野菜",
		ExpiryDate: "2026/04/15",
		Quantity:   10,
		Unit:       "kg",
	})
	if err == nil {
		t.Fatal("expected error for invalid expiry_date format")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got: %v", err)
	}
	if st.Code() != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", st.Code())
	}
	if !strings.Contains(st.Message(), "expiry_date") {
		t.Errorf("expected message to mention 'expiry_date', got %q", st.Message())
	}
}

func TestCreateFoodItem_RepoError(t *testing.T) {
	mock := newMockFoodItemStore()
	mock.createFunc = func(_ context.Context, _ *domain.FoodItem) (*domain.FoodItem, error) {
		return nil, errors.New("firestore unavailable")
	}
	svc := NewFoodInventoryService(mock)

	_, err := svc.CreateFoodItem(context.Background(), &pb.CreateFoodItemRequest{
		Name:       "にんじん",
		Category:   "野菜",
		ExpiryDate: "2026-04-15T00:00:00Z",
		Quantity:   10,
		Unit:       "kg",
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

func TestCreateFoodItem_AllCategories(t *testing.T) {
	for category := range domain.ValidCategories {
		t.Run(category, func(t *testing.T) {
			mock := newMockFoodItemStore()
			svc := NewFoodInventoryService(mock)

			resp, err := svc.CreateFoodItem(context.Background(), &pb.CreateFoodItemRequest{
				Name:       "テスト食品",
				Category:   category,
				ExpiryDate: "2026-04-15T00:00:00Z",
				Quantity:   1,
				Unit:       "個",
			})
			if err != nil {
				t.Fatalf("unexpected error for category %q: %v", category, err)
			}
			if resp.GetFoodItem().GetCategory() != category {
				t.Errorf("expected category %q, got %q", category, resp.GetFoodItem().GetCategory())
			}
		})
	}
}

func TestCreateFoodItem_AllUnits(t *testing.T) {
	for unit := range domain.ValidUnits {
		t.Run(unit, func(t *testing.T) {
			mock := newMockFoodItemStore()
			svc := NewFoodInventoryService(mock)

			resp, err := svc.CreateFoodItem(context.Background(), &pb.CreateFoodItemRequest{
				Name:       "テスト食品",
				Category:   "野菜",
				ExpiryDate: "2026-04-15T00:00:00Z",
				Quantity:   1,
				Unit:       unit,
			})
			if err != nil {
				t.Fatalf("unexpected error for unit %q: %v", unit, err)
			}
			if resp.GetFoodItem().GetUnit() != unit {
				t.Errorf("expected unit %q, got %q", unit, resp.GetFoodItem().GetUnit())
			}
		})
	}
}

func TestCreateFoodItem_BoundaryQuantities(t *testing.T) {
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
			mock := newMockFoodItemStore()
			svc := NewFoodInventoryService(mock)

			_, err := svc.CreateFoodItem(context.Background(), &pb.CreateFoodItemRequest{
				Name:       "テスト",
				Category:   "野菜",
				ExpiryDate: "2026-04-15T00:00:00Z",
				Quantity:   tt.quantity,
				Unit:       "kg",
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

// --- GetFoodItem テスト ---

func TestGetFoodItem_Success(t *testing.T) {
	mock := newMockFoodItemStore()
	svc := NewFoodInventoryService(mock)
	// 未来の消費期限を持つ食品を設定
	fixedNow := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	svc.nowFunc = func() time.Time { return fixedNow }

	mock.items["item-1"] = &domain.FoodItem{
		ID:         "item-1",
		Name:       "トマト",
		Category:   "野菜",
		ExpiryDate: time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC),
		Quantity:   5,
		Unit:       "個",
		DonorID:    "donor-1",
		Status:     domain.FoodItemStatusAvailable,
		CreatedAt:  time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:  time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
	}

	resp, err := svc.GetFoodItem(context.Background(), &pb.GetFoodItemRequest{Id: "item-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	item := resp.GetFoodItem()
	if item.GetId() != "item-1" {
		t.Errorf("expected id 'item-1', got %q", item.GetId())
	}
	if item.GetName() != "トマト" {
		t.Errorf("expected name 'トマト', got %q", item.GetName())
	}
	if item.GetStatus() != "available" {
		t.Errorf("expected status 'available', got %q", item.GetStatus())
	}
}

func TestGetFoodItem_NotFound(t *testing.T) {
	mock := newMockFoodItemStore()
	svc := NewFoodInventoryService(mock)

	_, err := svc.GetFoodItem(context.Background(), &pb.GetFoodItemRequest{Id: "nonexistent"})
	if err == nil {
		t.Fatal("expected error for nonexistent item")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got: %v", err)
	}
	if st.Code() != codes.NotFound {
		t.Errorf("expected NotFound, got %v", st.Code())
	}
}

func TestGetFoodItem_ExpiryCheck_MarksExpired(t *testing.T) {
	mock := newMockFoodItemStore()
	svc := NewFoodInventoryService(mock)

	// 消費期限を過去に設定
	expiredDate := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	fixedNow := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	svc.nowFunc = func() time.Time { return fixedNow }

	mock.items["item-expired"] = &domain.FoodItem{
		ID:         "item-expired",
		Name:       "古い牛乳",
		Category:   "乳製品",
		ExpiryDate: expiredDate,
		Quantity:   1,
		Unit:       "本",
		Status:     domain.FoodItemStatusAvailable,
		CreatedAt:  time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:  time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
	}

	resp, err := svc.GetFoodItem(context.Background(), &pb.GetFoodItemRequest{Id: "item-expired"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// レスポンスのステータスが expired になっていること
	if resp.GetFoodItem().GetStatus() != "expired" {
		t.Errorf("expected status 'expired', got %q", resp.GetFoodItem().GetStatus())
	}

	// ストア内のアイテムも更新されていること
	if mock.items["item-expired"].Status != domain.FoodItemStatusExpired {
		t.Errorf("expected stored item status to be 'expired', got %q", mock.items["item-expired"].Status)
	}

	// UpdatedAt が更新されていること
	if !mock.items["item-expired"].UpdatedAt.Equal(fixedNow) {
		t.Errorf("expected updated_at to be %v, got %v", fixedNow, mock.items["item-expired"].UpdatedAt)
	}
}

func TestGetFoodItem_ExpiryCheck_SkipsNonAvailable(t *testing.T) {
	mock := newMockFoodItemStore()
	svc := NewFoodInventoryService(mock)

	expiredDate := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	fixedNow := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	svc.nowFunc = func() time.Time { return fixedNow }

	// reserved ステータスの期限切れアイテム -- ステータスが変わらないことを確認
	mock.items["item-reserved"] = &domain.FoodItem{
		ID:         "item-reserved",
		Name:       "予約済み牛乳",
		Category:   "乳製品",
		ExpiryDate: expiredDate,
		Quantity:   1,
		Unit:       "本",
		Status:     domain.FoodItemStatusReserved,
		CreatedAt:  time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:  time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
	}

	resp, err := svc.GetFoodItem(context.Background(), &pb.GetFoodItemRequest{Id: "item-reserved"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// reserved のままであること
	if resp.GetFoodItem().GetStatus() != "reserved" {
		t.Errorf("expected status 'reserved', got %q", resp.GetFoodItem().GetStatus())
	}
}

func TestGetFoodItem_ExpiryCheck_UpdateError(t *testing.T) {
	mock := newMockFoodItemStore()
	mock.updateFunc = func(_ context.Context, _ *domain.FoodItem) error {
		return errors.New("firestore unavailable")
	}
	svc := NewFoodInventoryService(mock)

	expiredDate := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	fixedNow := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	svc.nowFunc = func() time.Time { return fixedNow }

	mock.items["item-expired"] = &domain.FoodItem{
		ID:         "item-expired",
		Name:       "古い牛乳",
		Category:   "乳製品",
		ExpiryDate: expiredDate,
		Quantity:   1,
		Unit:       "本",
		Status:     domain.FoodItemStatusAvailable,
		CreatedAt:  time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:  time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
	}

	_, err := svc.GetFoodItem(context.Background(), &pb.GetFoodItemRequest{Id: "item-expired"})
	if err == nil {
		t.Fatal("expected error when update fails")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got: %v", err)
	}
	if st.Code() != codes.Internal {
		t.Errorf("expected Internal, got %v", st.Code())
	}
}

// --- ListFoodItems テスト ---

func TestListFoodItems_Success(t *testing.T) {
	mock := newMockFoodItemStore()
	svc := NewFoodInventoryService(mock)
	fixedNow := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	svc.nowFunc = func() time.Time { return fixedNow }

	mock.items["item-1"] = &domain.FoodItem{
		ID:         "item-1",
		Name:       "にんじん",
		Category:   "野菜",
		ExpiryDate: time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC),
		Quantity:   10,
		Unit:       "kg",
		Status:     domain.FoodItemStatusAvailable,
		CreatedAt:  time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:  time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
	}
	mock.items["item-2"] = &domain.FoodItem{
		ID:         "item-2",
		Name:       "鶏肉",
		Category:   "肉",
		ExpiryDate: time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC),
		Quantity:   5,
		Unit:       "パック",
		Status:     domain.FoodItemStatusAvailable,
		CreatedAt:  time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC),
		UpdatedAt:  time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC),
	}

	resp, err := svc.ListFoodItems(context.Background(), &pb.ListFoodItemsRequest{
		PageSize: 20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.GetFoodItems()) != 2 {
		t.Errorf("expected 2 items, got %d", len(resp.GetFoodItems()))
	}
}

func TestListFoodItems_DefaultPageSize(t *testing.T) {
	mock := newMockFoodItemStore()
	mock.listFunc = func(_ context.Context, params repository.ListParams) (*repository.ListResult, error) {
		if params.PageSize != defaultPageSize {
			t.Errorf("expected default page size %d, got %d", defaultPageSize, params.PageSize)
		}
		return &repository.ListResult{Items: nil, TotalCount: 0}, nil
	}
	svc := NewFoodInventoryService(mock)

	_, err := svc.ListFoodItems(context.Background(), &pb.ListFoodItemsRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListFoodItems_CategoryFilter(t *testing.T) {
	mock := newMockFoodItemStore()
	mock.listFunc = func(_ context.Context, params repository.ListParams) (*repository.ListResult, error) {
		if params.CategoryFilter != "野菜" {
			t.Errorf("expected category filter '野菜', got %q", params.CategoryFilter)
		}
		return &repository.ListResult{
			Items: []*domain.FoodItem{
				{
					ID:         "item-1",
					Name:       "にんじん",
					Category:   "野菜",
					ExpiryDate: time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC),
					Quantity:   10,
					Unit:       "kg",
					Status:     domain.FoodItemStatusAvailable,
					CreatedAt:  time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
					UpdatedAt:  time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
				},
			},
			TotalCount: 1,
		}, nil
	}
	svc := NewFoodInventoryService(mock)
	fixedNow := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	svc.nowFunc = func() time.Time { return fixedNow }

	resp, err := svc.ListFoodItems(context.Background(), &pb.ListFoodItemsRequest{
		CategoryFilter: "野菜",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.GetFoodItems()) != 1 {
		t.Errorf("expected 1 item, got %d", len(resp.GetFoodItems()))
	}
}

func TestListFoodItems_ExpiryCheck_MarksExpiredOnList(t *testing.T) {
	expiredDate := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	futureDate := time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)
	fixedNow := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)

	expiredItem := &domain.FoodItem{
		ID:         "item-expired",
		Name:       "古い牛乳",
		Category:   "乳製品",
		ExpiryDate: expiredDate,
		Quantity:   1,
		Unit:       "本",
		Status:     domain.FoodItemStatusAvailable,
		CreatedAt:  time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:  time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
	}
	freshItem := &domain.FoodItem{
		ID:         "item-fresh",
		Name:       "新鮮な牛乳",
		Category:   "乳製品",
		ExpiryDate: futureDate,
		Quantity:   2,
		Unit:       "本",
		Status:     domain.FoodItemStatusAvailable,
		CreatedAt:  time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC),
		UpdatedAt:  time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC),
	}

	updateCalled := false
	mock := newMockFoodItemStore()
	mock.listFunc = func(_ context.Context, _ repository.ListParams) (*repository.ListResult, error) {
		return &repository.ListResult{
			Items:      []*domain.FoodItem{expiredItem, freshItem},
			TotalCount: 2,
		}, nil
	}
	mock.updateFunc = func(_ context.Context, item *domain.FoodItem) error {
		if item.ID == "item-expired" {
			updateCalled = true
		}
		return nil
	}

	svc := NewFoodInventoryService(mock)
	svc.nowFunc = func() time.Time { return fixedNow }

	resp, err := svc.ListFoodItems(context.Background(), &pb.ListFoodItemsRequest{PageSize: 20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.GetFoodItems()) != 2 {
		t.Fatalf("expected 2 items, got %d", len(resp.GetFoodItems()))
	}

	// 期限切れアイテムが expired ステータスになっていること
	for _, item := range resp.GetFoodItems() {
		switch item.GetId() {
		case "item-expired":
			if item.GetStatus() != "expired" {
				t.Errorf("expected expired item to have status 'expired', got %q", item.GetStatus())
			}
		case "item-fresh":
			if item.GetStatus() != "available" {
				t.Errorf("expected fresh item to have status 'available', got %q", item.GetStatus())
			}
		}
	}

	if !updateCalled {
		t.Error("expected update to be called for expired item")
	}
}

func TestListFoodItems_RepoError(t *testing.T) {
	mock := newMockFoodItemStore()
	mock.listFunc = func(_ context.Context, _ repository.ListParams) (*repository.ListResult, error) {
		return nil, errors.New("firestore unavailable")
	}
	svc := NewFoodInventoryService(mock)

	_, err := svc.ListFoodItems(context.Background(), &pb.ListFoodItemsRequest{PageSize: 20})
	if err == nil {
		t.Fatal("expected error when repo fails")
	}
}

func TestListFoodItems_Pagination(t *testing.T) {
	mock := newMockFoodItemStore()
	mock.listFunc = func(_ context.Context, params repository.ListParams) (*repository.ListResult, error) {
		if params.PageToken != "cursor-abc" {
			t.Errorf("expected page token 'cursor-abc', got %q", params.PageToken)
		}
		return &repository.ListResult{
			Items:         nil,
			NextPageToken: "cursor-def",
			TotalCount:    0,
		}, nil
	}
	svc := NewFoodInventoryService(mock)

	resp, err := svc.ListFoodItems(context.Background(), &pb.ListFoodItemsRequest{
		PageSize:  10,
		PageToken: "cursor-abc",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.GetNextPageToken() != "cursor-def" {
		t.Errorf("expected next_page_token 'cursor-def', got %q", resp.GetNextPageToken())
	}
}

// --- DeleteFoodItem テスト ---

func TestDeleteFoodItem_Success(t *testing.T) {
	mock := newMockFoodItemStore()
	mock.items["item-1"] = &domain.FoodItem{
		ID:     "item-1",
		Name:   "にんじん",
		Status: domain.FoodItemStatusAvailable,
	}
	svc := NewFoodInventoryService(mock)

	_, err := svc.DeleteFoodItem(context.Background(), &pb.DeleteFoodItemRequest{Id: "item-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// ストア内のアイテムが deleted ステータスになっていること
	if mock.items["item-1"].Status != domain.FoodItemStatusDeleted {
		t.Errorf("expected status 'deleted', got %q", mock.items["item-1"].Status)
	}
}

func TestDeleteFoodItem_NotFound(t *testing.T) {
	mock := newMockFoodItemStore()
	svc := NewFoodInventoryService(mock)

	_, err := svc.DeleteFoodItem(context.Background(), &pb.DeleteFoodItemRequest{Id: "nonexistent"})
	if err == nil {
		t.Fatal("expected error for nonexistent item")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got: %v", err)
	}
	if st.Code() != codes.NotFound {
		t.Errorf("expected NotFound, got %v", st.Code())
	}
}

// --- domainFoodItemToProto テスト ---

func TestDomainFoodItemToProto(t *testing.T) {
	now := time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC)
	expiry := time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)

	item := &domain.FoodItem{
		ID:         "test-id",
		Name:       "テスト食品",
		Category:   "野菜",
		ExpiryDate: expiry,
		Quantity:   10,
		Unit:       "kg",
		DonorID:    "donor-1",
		Status:     domain.FoodItemStatusAvailable,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	pbItem := domainFoodItemToProto(item)

	if pbItem.GetId() != "test-id" {
		t.Errorf("expected id 'test-id', got %q", pbItem.GetId())
	}
	if pbItem.GetName() != "テスト食品" {
		t.Errorf("expected name 'テスト食品', got %q", pbItem.GetName())
	}
	if pbItem.GetCategory() != "野菜" {
		t.Errorf("expected category '野菜', got %q", pbItem.GetCategory())
	}
	if pbItem.GetExpiryDate() != "2026-04-15T00:00:00Z" {
		t.Errorf("expected expiry_date '2026-04-15T00:00:00Z', got %q", pbItem.GetExpiryDate())
	}
	if pbItem.GetQuantity() != 10 {
		t.Errorf("expected quantity 10, got %d", pbItem.GetQuantity())
	}
	if pbItem.GetUnit() != "kg" {
		t.Errorf("expected unit 'kg', got %q", pbItem.GetUnit())
	}
	if pbItem.GetDonorId() != "donor-1" {
		t.Errorf("expected donor_id 'donor-1', got %q", pbItem.GetDonorId())
	}
	if pbItem.GetStatus() != "available" {
		t.Errorf("expected status 'available', got %q", pbItem.GetStatus())
	}
	if pbItem.GetCreatedAt() != "2026-04-01T12:00:00Z" {
		t.Errorf("expected created_at '2026-04-01T12:00:00Z', got %q", pbItem.GetCreatedAt())
	}
	if pbItem.GetUpdatedAt() != "2026-04-01T12:00:00Z" {
		t.Errorf("expected updated_at '2026-04-01T12:00:00Z', got %q", pbItem.GetUpdatedAt())
	}
}

// --- 網羅的バリデーションテスト (追加) ---

func TestCreateFoodItem_DonorIDOptional(t *testing.T) {
	mock := newMockFoodItemStore()
	svc := NewFoodInventoryService(mock)

	resp, err := svc.CreateFoodItem(context.Background(), &pb.CreateFoodItemRequest{
		Name:       "テスト",
		Category:   "野菜",
		ExpiryDate: "2026-04-15T00:00:00Z",
		Quantity:   1,
		Unit:       "kg",
		DonorId:    "", // 任意フィールド
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.GetFoodItem().GetDonorId() != "" {
		t.Errorf("expected empty donor_id, got %q", resp.GetFoodItem().GetDonorId())
	}
}
