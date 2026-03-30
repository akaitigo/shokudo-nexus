package service

import (
	"context"
	"sync"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/akaitigo/shokudo-nexus/backend/internal/domain"
	"github.com/akaitigo/shokudo-nexus/backend/internal/repository"
)

// mockFoodItemStore はテスト用のFoodItemStoreモック。
type mockFoodItemStore struct {
	mu    sync.RWMutex
	items map[string]*domain.FoodItem

	createFunc func(ctx context.Context, item *domain.FoodItem) (*domain.FoodItem, error)
	getFunc    func(ctx context.Context, id string) (*domain.FoodItem, error)
	listFunc   func(ctx context.Context, params repository.ListParams) (*repository.ListResult, error)
	updateFunc func(ctx context.Context, item *domain.FoodItem) error
	deleteFunc func(ctx context.Context, id string) error

	// 呼び出し回数を追跡
	updateCalls int
}

func newMockFoodItemStore() *mockFoodItemStore {
	return &mockFoodItemStore{
		items: make(map[string]*domain.FoodItem),
	}
}

func (m *mockFoodItemStore) Create(ctx context.Context, item *domain.FoodItem) (*domain.FoodItem, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, item)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	item.ID = "generated-id"
	m.items[item.ID] = item
	return item, nil
}

func (m *mockFoodItemStore) Get(ctx context.Context, id string) (*domain.FoodItem, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, id)
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	item, ok := m.items[id]
	if !ok {
		return nil, errNotFound(id)
	}
	return item, nil
}

func (m *mockFoodItemStore) List(ctx context.Context, params repository.ListParams) (*repository.ListResult, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, params)
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	filtered := make([]*domain.FoodItem, 0, len(m.items))
	for _, item := range m.items {
		if item.Status == domain.FoodItemStatusDeleted {
			continue
		}
		if params.CategoryFilter != "" && item.Category != params.CategoryFilter {
			continue
		}
		filtered = append(filtered, item)
	}

	end := params.PageSize
	if end > len(filtered) {
		end = len(filtered)
	}

	return &repository.ListResult{
		Items:      filtered[:end],
		TotalCount: int32(end),
	}, nil
}

func (m *mockFoodItemStore) Update(ctx context.Context, item *domain.FoodItem) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, item)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items[item.ID] = item
	m.updateCalls++
	return nil
}

func (m *mockFoodItemStore) Delete(ctx context.Context, id string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	item, ok := m.items[id]
	if !ok {
		return errNotFound(id)
	}
	item.Status = domain.FoodItemStatusDeleted
	return nil
}

func errNotFound(id string) error {
	return status.Errorf(codes.NotFound, "item %q not found", id)
}

// mockFusionRequestStore はテスト用のFusionRequestStoreモック。
type mockFusionRequestStore struct {
	mu       sync.RWMutex
	requests map[string]*domain.FusionRequest

	createFunc func(ctx context.Context, req *domain.FusionRequest) (*domain.FusionRequest, error)
	getFunc    func(ctx context.Context, id string) (*domain.FusionRequest, error)
	listFunc   func(ctx context.Context, params repository.FusionListParams) (*repository.FusionListResult, error)
	updateFunc func(ctx context.Context, req *domain.FusionRequest) error
	updateErr  error // Set to force Update() to return an error

	updateCalls int
}

func newMockFusionRequestStore() *mockFusionRequestStore {
	return &mockFusionRequestStore{
		requests: make(map[string]*domain.FusionRequest),
	}
}

func (m *mockFusionRequestStore) Create(ctx context.Context, req *domain.FusionRequest) (*domain.FusionRequest, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, req)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	req.ID = "generated-fusion-id"
	m.requests[req.ID] = req
	return req, nil
}

func (m *mockFusionRequestStore) Get(ctx context.Context, id string) (*domain.FusionRequest, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, id)
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	req, ok := m.requests[id]
	if !ok {
		return nil, errNotFound(id)
	}
	return req, nil
}

func (m *mockFusionRequestStore) List(ctx context.Context, params repository.FusionListParams) (*repository.FusionListResult, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, params)
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*domain.FusionRequest, 0, len(m.requests))
	for _, req := range m.requests {
		result = append(result, req)
	}
	return &repository.FusionListResult{
		Requests:   result,
		TotalCount: int32(len(result)),
	}, nil
}

func (m *mockFusionRequestStore) Update(ctx context.Context, req *domain.FusionRequest) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if m.updateFunc != nil {
		return m.updateFunc(ctx, req)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requests[req.ID] = req
	m.updateCalls++
	return nil
}
