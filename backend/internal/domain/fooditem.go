// Package domain はshokudo-nexusのドメインモデルを定義する。
package domain

import "time"

// FoodItemStatus は食品のステータスを表す。
type FoodItemStatus string

const (
	// FoodItemStatusAvailable は利用可能な状態。
	FoodItemStatusAvailable FoodItemStatus = "available"
	// FoodItemStatusReserved は予約済み状態。
	FoodItemStatusReserved FoodItemStatus = "reserved"
	// FoodItemStatusConsumed は消費済み状態。
	FoodItemStatusConsumed FoodItemStatus = "consumed"
	// FoodItemStatusExpired は消費期限切れ状態。
	FoodItemStatusExpired FoodItemStatus = "expired"
	// FoodItemStatusDeleted は論理削除状態。
	FoodItemStatusDeleted FoodItemStatus = "deleted"
)

// ValidFoodItemStatuses は許可されたステータス一覧。
var ValidFoodItemStatuses = map[FoodItemStatus]bool{
	FoodItemStatusAvailable: true,
	FoodItemStatusReserved:  true,
	FoodItemStatusConsumed:  true,
	FoodItemStatusExpired:   true,
	FoodItemStatusDeleted:   true,
}

// FoodItem は余剰食品を表すドメインモデル。
type FoodItem struct {
	ID         string         `firestore:"id"`
	Name       string         `firestore:"name"`
	Category   string         `firestore:"category"`
	ExpiryDate time.Time      `firestore:"expiry_date"`
	Quantity   int32          `firestore:"quantity"`
	Unit       string         `firestore:"unit"`
	DonorID    string         `firestore:"donor_id"`
	Status     FoodItemStatus `firestore:"status"`
	CreatedAt  time.Time      `firestore:"created_at"`
	UpdatedAt  time.Time      `firestore:"updated_at"`
}

// IsExpired は消費期限を過ぎているかを判定する。
func (f *FoodItem) IsExpired(now time.Time) bool {
	return now.After(f.ExpiryDate)
}
