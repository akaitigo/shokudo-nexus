package domain

import "time"

// DonationRecord は寄付履歴を表すドメインモデル。レポート機能の基盤データとなる。
type DonationRecord struct {
	ID              string    `firestore:"id"`
	FoodItemID      string    `firestore:"food_item_id"`
	DonorID         string    `firestore:"donor_id"`
	RecipientID     string    `firestore:"recipient_id"`
	FusionRequestID string    `firestore:"fusion_request_id"`
	Category        string    `firestore:"category"`
	Quantity        int32     `firestore:"quantity"`
	Unit            string    `firestore:"unit"`
	CreatedAt       time.Time `firestore:"created_at"`
}
