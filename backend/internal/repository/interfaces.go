// Package repository はFirestoreを用いたデータアクセス層を提供する。
package repository

import (
	"context"

	"github.com/akaitigo/shokudo-nexus/backend/internal/domain"
)

// FoodItemStore は食品アイテムのデータアクセスインターフェース。
type FoodItemStore interface {
	Create(ctx context.Context, item *domain.FoodItem) (*domain.FoodItem, error)
	Get(ctx context.Context, id string) (*domain.FoodItem, error)
	List(ctx context.Context, params ListParams) (*ListResult, error)
	Update(ctx context.Context, item *domain.FoodItem) error
	Delete(ctx context.Context, id string) error
}

// FusionRequestStore は融通リクエストのデータアクセスインターフェース。
type FusionRequestStore interface {
	Create(ctx context.Context, req *domain.FusionRequest) (*domain.FusionRequest, error)
	Get(ctx context.Context, id string) (*domain.FusionRequest, error)
	List(ctx context.Context, params FusionListParams) (*FusionListResult, error)
	Update(ctx context.Context, req *domain.FusionRequest) error
}

// DonationRecordStore は寄付履歴のデータアクセスインターフェース。
type DonationRecordStore interface {
	Create(ctx context.Context, record *domain.DonationRecord) (*domain.DonationRecord, error)
}
