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

type mockRepoForGoroutines struct{ mock.Mock }

func (m *mockRepoForGoroutines) Save(ctx context.Context, order *model.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *mockRepoForGoroutines) FindByID(ctx context.Context, uid string) (*model.Order, error) {
	args := m.Called(ctx, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Order), args.Error(1)
}

func (m *mockRepoForGoroutines) FindAll(ctx context.Context) ([]*model.Order, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Order), args.Error(1)
}

type mockProducerForGoroutines struct{ mock.Mock }

func (m *mockProducerForGoroutines) ProduceOrder(ctx context.Context, order *model.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

type mockCacheForGoroutines struct{ mock.Mock }

func (m *mockCacheForGoroutines) Set(uid string, order *model.Order) {
	m.Called(uid, order)
}

func (m *mockCacheForGoroutines) Get(uid string) (*model.Order, bool) {
	args := m.Called(uid)
	if args.Get(0) == nil {
		return nil, args.Bool(1)
	}
	return args.Get(0).(*model.Order), args.Bool(1)
}

func (m *mockCacheForGoroutines) GetAll() []*model.Order {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*model.Order)
}

func (m *mockCacheForGoroutines) Delete(uid string) {
	m.Called(uid)
}

func (m *mockCacheForGoroutines) Size() int {
	args := m.Called()
	return args.Int(0)
}

func (m *mockCacheForGoroutines) Clear() {
	m.Called()
}

func (m *mockCacheForGoroutines) WarmUp(repo repositories.OrderRepository) error {
	args := m.Called(repo)
	return args.Error(0)
}
func TestGoroutinesGracefulShutdown(t *testing.T) {
	repo := new(mockRepoForGoroutines)
	prod := new(mockProducerForGoroutines)
	cache := new(mockCacheForGoroutines)
	metrics := new(mockMetrics)

	uc := usecases.NewOrderUseCase(repo, prod, cache, metrics, 2, 10)

	dateCreated, err := time.Parse(time.RFC3339, "2021-11-26T06:22:19Z")
	if err != nil {
		t.Fatalf("Failed to parse date: %v", err)
	}

	order := &model.Order{
		OrderUID:    "test123",
		TrackNumber: "WBILMTESTTRACK",
		Entry:       "WBIL",
		Delivery: model.Delivery{
			Name:    "Test Testov",
			Phone:   "+9720000000",
			Zip:     "2639809",
			City:    "Kiryat Mozkin",
			Address: "Ploshad Mira 15",
			Region:  "Kraiot",
			Email:   "test@gmail.com",
		},
		Payment: model.Payment{
			Transaction:  "test123",
			RequestID:    "",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       1817,
			PaymentDt:    1637907727,
			Bank:         "alpha",
			DeliveryCost: 1500,
			GoodsTotal:   317,
			CustomFee:    0,
		},
		Items: []model.Item{
			{
				ChrtID:      9934930,
				TrackNumber: "WBILMTESTTRACK",
				Price:       453,
				Rid:         "ab4219087a764ae0btest",
				Name:        "Mascaras",
				Sale:        30,
				Size:        "0",
				TotalPrice:  317,
				NmID:        2389212,
				Brand:       "Vivienne Sabo",
				Status:      202,
			},
		},
		Locale:            "en",
		InternalSignature: "",
		CustomerID:        "test",
		DeliveryService:   "meest",
		Shardkey:          "9",
		SmID:              99,
		DateCreated:       dateCreated,
		OofShard:          "1",
	}

	repo.On("Save", mock.Anything, mock.Anything).Return(nil)
	prod.On("ProduceOrder", mock.Anything, mock.Anything).Return(nil)
	cache.On("Set", mock.Anything, mock.Anything)

	done := make(chan bool)

	go func() {
		err := uc.CreateOrder(context.Background(), order)
		assert.NoError(t, err)
		done <- true
	}()

	time.Sleep(50 * time.Millisecond)

	start := time.Now()
	uc.Shutdown()
	shutdownTime := time.Since(start)

	assert.Less(t, shutdownTime, 2*time.Second, "Shutdown should complete within 2 seconds")

	<-done

	repo.AssertExpectations(t)
	prod.AssertExpectations(t)
	cache.AssertExpectations(t)
}
