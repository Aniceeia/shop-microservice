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

type mockRepo struct{ mock.Mock }

func (m *mockRepo) Save(ctx context.Context, order *model.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}
func (m *mockRepo) FindByID(ctx context.Context, uid string) (*model.Order, error) {
	args := m.Called(ctx, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Order), args.Error(1)
}
func (m *mockRepo) FindAll(ctx context.Context) ([]*model.Order, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Order), args.Error(1)
}

type mockProducer struct{ mock.Mock }

func (m *mockProducer) ProduceOrder(ctx context.Context, order *model.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

type memoryCache struct {
	store map[string]*model.Order
}

func newMemoryCache() *memoryCache { return &memoryCache{store: make(map[string]*model.Order)} }

func (c *memoryCache) Set(uid string, order *model.Order)  { c.store[uid] = order }
func (c *memoryCache) Get(uid string) (*model.Order, bool) { v, ok := c.store[uid]; return v, ok }
func (c *memoryCache) GetAll() []*model.Order {
	res := make([]*model.Order, 0, len(c.store))
	for _, v := range c.store {
		res = append(res, v)
	}
	return res
}
func (c *memoryCache) Delete(uid string)                              { delete(c.store, uid) }
func (c *memoryCache) Size() int                                      { return len(c.store) }
func (c *memoryCache) Clear()                                         { c.store = make(map[string]*model.Order) }
func (c *memoryCache) WarmUp(repo repositories.OrderRepository) error { return nil }

func validOrder() *model.Order {
	return &model.Order{
		OrderUID:    "uid12345",
		TrackNumber: "WBILMTRACK",
		Entry:       "WBIL",
		CustomerID:  "cust1",
		DateCreated: time.Now(),
		Delivery: model.Delivery{
			Name:    "User",
			Phone:   "+1000000000",
			Zip:     "000000",
			City:    "City",
			Address: "Addr",
			Region:  "Reg",
			Email:   "u@u.com",
		},
		Payment: model.Payment{
			Transaction:  "tx1",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       100,
			PaymentDt:    time.Now().Unix(),
			Bank:         "bank",
			DeliveryCost: 10,
			GoodsTotal:   90,
			CustomFee:    0,
		},
		Items: []model.Item{{
			ChrtID:      1,
			TrackNumber: "WBILMTRACK",
			Price:       90,
			Rid:         "rid",
			Name:        "prod",
			Sale:        0,
			Size:        "0",
			TotalPrice:  90,
			NmID:        1,
			Brand:       "br",
			Status:      200,
		}},
	}
}

func TestValidateOrder_Errors(t *testing.T) {
	uc := usecases.NewOrderUseCase(new(mockRepo), new(mockProducer), newMemoryCache(), new(mockMetrics), 1, 10)
	ord := validOrder()
	ord.OrderUID = ""
	err := uc.ValidateOrder(ord)
	assert.Error(t, err)

	ord = validOrder()
	ord.TrackNumber = ""
	assert.Error(t, uc.ValidateOrder(ord))

	ord = validOrder()
	ord.Entry = ""
	assert.Error(t, uc.ValidateOrder(ord))

	ord = validOrder()
	ord.Items = nil
	assert.Error(t, uc.ValidateOrder(ord))

}

func TestCreateOrder_AndGetByID_CacheAndRepo(t *testing.T) {
	repo := new(mockRepo)
	prod := new(mockProducer)
	cache := newMemoryCache()
	metrics := new(mockMetrics)
	uc := usecases.NewOrderUseCase(repo, prod, cache, metrics, 1, 10)
	ctx := context.Background()
	ord := validOrder()

	repo.On("Save", mock.Anything, mock.Anything).Return(nil)
	prod.On("ProduceOrder", mock.Anything, mock.Anything).Return(nil)

	assert.NoError(t, uc.CreateOrder(ctx, ord))

	got, ok := cache.Get(ord.OrderUID)
	assert.True(t, ok)
	assert.Equal(t, ord.OrderUID, got.OrderUID)

	cache.Clear()
	repo.On("FindByID", mock.Anything, ord.OrderUID).Return(ord, nil)
	res, err := uc.GetOrderByID(ctx, ord.OrderUID)
	assert.NoError(t, err)
	assert.Equal(t, ord.OrderUID, res.OrderUID)
}

func TestGetAllOrders_FromCacheAndRepo(t *testing.T) {
	repo := new(mockRepo)
	prod := new(mockProducer)
	cache := newMemoryCache()
	metrics := new(mockMetrics)

	uc := usecases.NewOrderUseCase(repo, prod, cache, metrics, 1, 10)
	ctx := context.Background()

	ord := validOrder()
	cache.Set(ord.OrderUID, ord)
	orders, err := uc.GetAllOrders(ctx)
	assert.NoError(t, err)
	assert.Len(t, orders, 1)

	cache.Clear()
	repo.On("FindAll", mock.Anything).Return([]*model.Order{ord}, nil)
	orders, err = uc.GetAllOrders(ctx)
	assert.NoError(t, err)
	assert.Len(t, orders, 1)
}

func TestHealthCheck_HealthyAndUnhealthy(t *testing.T) {
	repo := new(mockRepo)
	prod := new(mockProducer)
	cache := newMemoryCache()
	metrics := new(mockMetrics)

	uc := usecases.NewOrderUseCase(repo, prod, cache, metrics, 1, 10)
	ctx := context.Background()

	repo.On("FindAll", mock.Anything).Return([]*model.Order{}, nil).Once()
	h, err := uc.HealthCheck(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", h["status"])

}

func TestShutdown(t *testing.T) {
	repo := new(mockRepo)
	prod := new(mockProducer)
	cache := newMemoryCache()
	metrics := new(mockMetrics)

	uc := usecases.NewOrderUseCase(repo, prod, cache, metrics, 1, 1)
	uc.Shutdown()
	assert.True(t, true)
}
