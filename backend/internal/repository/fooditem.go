// Package repository はFirestoreを用いたデータアクセス層を提供する。
package repository

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/akaitigo/shokudo-nexus/backend/internal/domain"
)

const foodItemsCollection = "food_items"

// FoodItemRepository はFoodItemのFirestoreリポジトリ。
type FoodItemRepository struct {
	client *firestore.Client
}

// NewFoodItemRepository は新しいFoodItemRepositoryを生成する。
func NewFoodItemRepository(client *firestore.Client) *FoodItemRepository {
	return &FoodItemRepository{client: client}
}

// Create は食品アイテムを作成する。IDは自動生成される。
func (r *FoodItemRepository) Create(ctx context.Context, item *domain.FoodItem) (*domain.FoodItem, error) {
	item.ID = uuid.New().String()
	_, err := r.client.Collection(foodItemsCollection).Doc(item.ID).Set(ctx, item)
	if err != nil {
		return nil, fmt.Errorf("failed to create food item: %w", err)
	}
	return item, nil
}

// Get はIDを指定して食品アイテムを取得する。
func (r *FoodItemRepository) Get(ctx context.Context, id string) (*domain.FoodItem, error) {
	doc, err := r.client.Collection(foodItemsCollection).Doc(id).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, status.Errorf(codes.NotFound, "food item %q not found", id)
		}
		return nil, fmt.Errorf("failed to get food item: %w", err)
	}

	var item domain.FoodItem
	if err := doc.DataTo(&item); err != nil {
		return nil, fmt.Errorf("failed to decode food item: %w", err)
	}
	return &item, nil
}

// ListParams はList操作のパラメータ。
type ListParams struct {
	PageSize       int
	PageToken      string
	CategoryFilter string
}

// ListResult はList操作の結果。
type ListResult struct {
	Items         []*domain.FoodItem
	NextPageToken string
	TotalCount    int32
}

// List は食品アイテムの一覧を取得する。カテゴリフィルタとページネーションに対応。
func (r *FoodItemRepository) List(ctx context.Context, params ListParams) (*ListResult, error) {
	col := r.client.Collection(foodItemsCollection)

	// 削除済みを除外するクエリ
	q := col.Where("status", "!=", string(domain.FoodItemStatusDeleted))

	if params.CategoryFilter != "" {
		q = col.Where("category", "==", params.CategoryFilter).
			Where("status", "!=", string(domain.FoodItemStatusDeleted))
	}

	q = q.OrderBy("created_at", firestore.Desc)

	// ページトークンによるオフセット
	if params.PageToken != "" {
		tokenDoc, err := col.Doc(params.PageToken).Get(ctx)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid page token: %v", err)
		}
		q = q.StartAfter(tokenDoc)
	}

	// +1 件多く取得して next_page_token の判定に使う
	limit := params.PageSize + 1
	q = q.Limit(limit)

	iter := q.Documents(ctx)
	defer iter.Stop()

	var items []*domain.FoodItem
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate food items: %w", err)
		}

		var item domain.FoodItem
		if err := doc.DataTo(&item); err != nil {
			return nil, fmt.Errorf("failed to decode food item: %w", err)
		}
		items = append(items, &item)
	}

	result := &ListResult{}

	// 次ページがあるか判定
	if len(items) > params.PageSize {
		result.NextPageToken = items[params.PageSize-1].ID
		items = items[:params.PageSize]
	}

	result.Items = items
	result.TotalCount = int32(len(items))

	return result, nil
}

// Update は食品アイテムを更新する。
func (r *FoodItemRepository) Update(ctx context.Context, item *domain.FoodItem) error {
	_, err := r.client.Collection(foodItemsCollection).Doc(item.ID).Set(ctx, item)
	if err != nil {
		return fmt.Errorf("failed to update food item: %w", err)
	}
	return nil
}

// Delete は食品アイテムを論理削除する（ステータスをdeletedに変更）。
func (r *FoodItemRepository) Delete(ctx context.Context, id string) error {
	item, err := r.Get(ctx, id)
	if err != nil {
		return err
	}

	item.Status = domain.FoodItemStatusDeleted
	return r.Update(ctx, item)
}
