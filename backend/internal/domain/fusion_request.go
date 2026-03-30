package domain

import "time"

// FusionRequestStatus は融通リクエストのステータスを表す。
type FusionRequestStatus string

const (
	// FusionRequestStatusPending は未対応状態。
	FusionRequestStatusPending FusionRequestStatus = "pending"
	// FusionRequestStatusApproved は承認済み状態。
	FusionRequestStatusApproved FusionRequestStatus = "approved"
	// FusionRequestStatusRejected は拒否済み状態。
	FusionRequestStatusRejected FusionRequestStatus = "rejected"
	// FusionRequestStatusCompleted は完了状態。
	FusionRequestStatusCompleted FusionRequestStatus = "completed"
)

// ValidFusionRequestStatuses は許可されたステータス一覧。
var ValidFusionRequestStatuses = map[FusionRequestStatus]bool{
	FusionRequestStatusPending:   true,
	FusionRequestStatusApproved:  true,
	FusionRequestStatusRejected:  true,
	FusionRequestStatusCompleted: true,
}

// FusionRequest は拠点間食材融通リクエストを表すドメインモデル。
type FusionRequest struct {
	ID                 string              `firestore:"id"`
	RequesterShokudoID string              `firestore:"requester_shokudo_id"`
	DesiredCategory    string              `firestore:"desired_category"`
	DesiredQuantity    int32               `firestore:"desired_quantity"`
	Unit               string              `firestore:"unit"`
	Message            string              `firestore:"message"`
	Status             FusionRequestStatus `firestore:"status"`
	ResponderShokudoID string              `firestore:"responder_shokudo_id"`
	FoodItemID         string              `firestore:"food_item_id"`
	CreatedAt          time.Time           `firestore:"created_at"`
	UpdatedAt          time.Time           `firestore:"updated_at"`
}
