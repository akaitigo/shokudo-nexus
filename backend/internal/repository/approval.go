package repository

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/akaitigo/shokudo-nexus/backend/internal/domain"
)

// ApprovalTx は融通リクエスト承認トランザクション内での読み書きを抽象化する。
// Firestore トランザクションの制約上、すべての読み取り（Get系）は書き込み（Set系）より前に行う。
type ApprovalTx interface {
	GetFusionRequest(id string) (*domain.FusionRequest, error)
	GetFoodItem(id string) (*domain.FoodItem, error)
	SetFusionRequest(req *domain.FusionRequest) error
	SetFoodItem(item *domain.FoodItem) error
}

// ApprovalTxFunc はトランザクション内で実行する処理。
type ApprovalTxFunc func(tx ApprovalTx) error

// ApprovalRunner は承認処理を単一のアトミックなトランザクションとして実行する。
type ApprovalRunner interface {
	RunApproval(ctx context.Context, fn ApprovalTxFunc) error
}

// FirestoreApprovalRunner は Firestore トランザクションで ApprovalRunner を実装する。
// 楽観的並行制御により、複数の Cloud Run インスタンス間でも同一在庫の二重予約を防止する。
type FirestoreApprovalRunner struct {
	client *firestore.Client
}

// NewFirestoreApprovalRunner は新しい FirestoreApprovalRunner を生成する。
func NewFirestoreApprovalRunner(client *firestore.Client) *FirestoreApprovalRunner {
	return &FirestoreApprovalRunner{client: client}
}

// RunApproval は fn を Firestore トランザクション内で実行する。
// 読み取ったドキュメントがコミットまでに他インスタンスによって変更された場合、
// トランザクションは競合を検知して自動的に再試行される。
func (r *FirestoreApprovalRunner) RunApproval(ctx context.Context, fn ApprovalTxFunc) error {
	return r.client.RunTransaction(ctx, func(_ context.Context, tx *firestore.Transaction) error {
		return fn(&firestoreApprovalTx{
			tx:          tx,
			fusionCol:   r.client.Collection(fusionRequestsCollection),
			foodItemCol: r.client.Collection(foodItemsCollection),
		})
	})
}

// firestoreApprovalTx は Firestore トランザクションを ApprovalTx として提供する。
type firestoreApprovalTx struct {
	tx          *firestore.Transaction
	fusionCol   *firestore.CollectionRef
	foodItemCol *firestore.CollectionRef
}

func (a *firestoreApprovalTx) GetFusionRequest(id string) (*domain.FusionRequest, error) {
	doc, err := a.tx.Get(a.fusionCol.Doc(id))
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, status.Errorf(codes.NotFound, "fusion request %q not found", id)
		}
		return nil, fmt.Errorf("failed to get fusion request: %w", err)
	}
	var req domain.FusionRequest
	if err := doc.DataTo(&req); err != nil {
		return nil, fmt.Errorf("failed to decode fusion request: %w", err)
	}
	return &req, nil
}

func (a *firestoreApprovalTx) GetFoodItem(id string) (*domain.FoodItem, error) {
	doc, err := a.tx.Get(a.foodItemCol.Doc(id))
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

func (a *firestoreApprovalTx) SetFusionRequest(req *domain.FusionRequest) error {
	return a.tx.Set(a.fusionCol.Doc(req.ID), req)
}

func (a *firestoreApprovalTx) SetFoodItem(item *domain.FoodItem) error {
	return a.tx.Set(a.foodItemCol.Doc(item.ID), item)
}
