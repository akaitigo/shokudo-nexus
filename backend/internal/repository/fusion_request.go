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

const fusionRequestsCollection = "fusion_requests"

// FusionRequestRepository はFusionRequestのFirestoreリポジトリ。
type FusionRequestRepository struct {
	client *firestore.Client
}

// NewFusionRequestRepository は新しいFusionRequestRepositoryを生成する。
func NewFusionRequestRepository(client *firestore.Client) *FusionRequestRepository {
	return &FusionRequestRepository{client: client}
}

// Create は融通リクエストを作成する。
func (r *FusionRequestRepository) Create(ctx context.Context, req *domain.FusionRequest) (*domain.FusionRequest, error) {
	req.ID = uuid.New().String()
	_, err := r.client.Collection(fusionRequestsCollection).Doc(req.ID).Set(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create fusion request: %w", err)
	}
	return req, nil
}

// Get はIDを指定して融通リクエストを取得する。
func (r *FusionRequestRepository) Get(ctx context.Context, id string) (*domain.FusionRequest, error) {
	doc, err := r.client.Collection(fusionRequestsCollection).Doc(id).Get(ctx)
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

// FusionListParams はList操作のパラメータ。
type FusionListParams struct {
	PageSize     int
	PageToken    string
	StatusFilter string
}

// FusionListResult はList操作の結果。
type FusionListResult struct {
	Requests      []*domain.FusionRequest
	NextPageToken string
	TotalCount    int32
}

// List は融通リクエストの一覧を取得する。
func (r *FusionRequestRepository) List(ctx context.Context, params FusionListParams) (*FusionListResult, error) {
	col := r.client.Collection(fusionRequestsCollection)

	var q firestore.Query
	if params.StatusFilter != "" {
		q = col.Where("status", "==", params.StatusFilter)
	} else {
		q = col.Query
	}

	// 全件数を取得するための別クエリ（Select で転送量を最小化）
	totalCount, err := countDocuments(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("failed to count fusion requests: %w", err)
	}

	q = q.OrderBy("created_at", firestore.Desc)

	if params.PageToken != "" {
		tokenDoc, err := col.Doc(params.PageToken).Get(ctx)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid page token: %v", err)
		}
		q = q.StartAfter(tokenDoc)
	}

	limit := params.PageSize + 1
	q = q.Limit(limit)

	iter := q.Documents(ctx)
	defer iter.Stop()

	var requests []*domain.FusionRequest
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate fusion requests: %w", err)
		}

		var req domain.FusionRequest
		if err := doc.DataTo(&req); err != nil {
			return nil, fmt.Errorf("failed to decode fusion request: %w", err)
		}
		requests = append(requests, &req)
	}

	result := &FusionListResult{}

	if len(requests) > params.PageSize {
		result.NextPageToken = requests[params.PageSize-1].ID
		requests = requests[:params.PageSize]
	}

	result.Requests = requests
	result.TotalCount = int32(totalCount)

	return result, nil
}

// Update は融通リクエストを更新する。
func (r *FusionRequestRepository) Update(ctx context.Context, req *domain.FusionRequest) error {
	_, err := r.client.Collection(fusionRequestsCollection).Doc(req.ID).Set(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to update fusion request: %w", err)
	}
	return nil
}
