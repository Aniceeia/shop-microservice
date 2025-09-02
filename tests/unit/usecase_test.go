package unit

import (
	"context"
	"testing"
	"time"

	"shop-microservice/internal/application/usecases"
	"shop-microservice/internal/domain/model"
	"shop-microservice/internal/domain/repositories"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockOrderRepository struct{ mock.Mock }
type mockMessageProducer struct{ mock.Mock }
type mockCache struct{ mock.Mock }

type mockMetrics struct {
	mock.Mock
}

func (m *mockMetrics) IncrementRequest(endpoint string) {
	m.Called(endpoint)
}

func (m *mockMetrics) IncrementError(endpoint string) {
	m.Called(endpoint)
}

func (m *mockMetrics) RecordResponseTime(endpoint string, duration time.Duration) {
	m.Called(endpoint, duration)
}

func (m *mockMetrics) IncrementCacheHit() {
	m.Called()
}

func (m *mockMetrics) IncrementCacheMiss() {
	m.Called()
}

func (m *mockMetrics) SetDBConnections(count int64) {
	m.Called(count)
}

func (m *mockMetrics) IncrementKafkaMessages() {
	m.Called()
}

func (m *mockMetrics) Reset() {
	m.Called()
}

func (m *mockOrderRepository) Save(ctx context.Context, order *model.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *mockOrderRepository) FindByID(ctx context.Context, uid string) (*model.Order, error) {
	args := m.Called(ctx, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Order), args.Error(1)
}

func (m *mockOrderRepository) FindAll(ctx context.Context) ([]*model.Order, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Order), args.Error(1)
}

func (m *mockMessageProducer) ProduceOrder(ctx context.Context, order *model.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *mockCache) Set(uid string, order *model.Order) {
	m.Called(uid, order)
}

func (m *mockCache) Get(uid string) (*model.Order, bool) {
	args := m.Called(uid)
	if args.Get(0) == nil {
		return nil, args.Bool(1)
	}
	return args.Get(0).(*model.Order), args.Bool(1)
}

func (m *mockCache) GetAll() []*model.Order {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*model.Order)
}

func (m *mockCache) Delete(uid string) {
	m.Called(uid)
}

func (m *mockCache) Size() int {
	args := m.Called()
	return args.Int(0)
}

func (m *mockCache) Clear() {
	m.Called()
}

func (m *mockCache) WarmUp(repo repositories.OrderRepository) error {
	args := m.Called(repo)
	return args.Error(0)
}

func (m *mockMetrics) GetStats() map[string]interface{} {
	args := m.Called()
	return args.Get(0).(map[string]interface{})
}

func TestOrderUseCase_GetMetrics(t *testing.T) {
	mockRepo := new(mockOrderRepository)
	mockProducer := new(mockMessageProducer)
	mockCache := new(mockCache)
	mockMetrics := new(mockMetrics)

	expectedStats := map[string]interface{}{
		"requests": map[string]int64{"/test": 5},
		"errors":   map[string]int64{"/test": 1},
	}
	mockMetrics.On("GetStats").Return(expectedStats)

	uc := usecases.NewOrderUseCase(mockRepo, mockProducer, mockCache, mockMetrics, 3, 1000)

	stats := uc.GetMetrics()
	assert.Equal(t, expectedStats, stats)

	mockMetrics.AssertExpectations(t)
}

func TestOrderUseCase_Shutdown(t *testing.T) {
	mockRepo := new(mockOrderRepository)
	mockProducer := new(mockMessageProducer)
	mockCache := new(mockCache)
	mockMetrics := new(mockMetrics)

	uc := usecases.NewOrderUseCase(mockRepo, mockProducer, mockCache, mockMetrics, 1, 10)

	uc.Shutdown()

	mockProducer.AssertExpectations(t)
}
