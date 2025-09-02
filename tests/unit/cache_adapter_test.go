package unit

import (
	"context"
	"testing"
	"time"

	"shop-microservice/internal/domain/model"
	"shop-microservice/internal/infrastructure/cache"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCacheAdapter_BasicOps(t *testing.T) {
	c := cache.NewCash(100, 1*time.Hour)
	ad := cache.NewCacheAdapter(c)
	ad.Set("u", &model.Order{OrderUID: "u"})
	got, ok := ad.Get("u")
	assert.True(t, ok)
	assert.Equal(t, "u", got.OrderUID)
	assert.Equal(t, 1, ad.Size())
	ad.Delete("u")
	_, ok = ad.Get("u")
	assert.False(t, ok)
	ad.Clear()
}

func TestCacheAdapter_GetAll(t *testing.T) {
	c := cache.NewCash(100, 1*time.Hour)
	ad := cache.NewCacheAdapter(c)
	ad.Set("a", &model.Order{OrderUID: "a"})
	ad.Set("b", &model.Order{OrderUID: "b"})
	all := ad.GetAll()
	assert.Len(t, all, 2)
}

type repoForWarmup struct {
	mock.Mock
	orders []*model.Order
}

func (r *repoForWarmup) Save(ctx context.Context, order *model.Order) error {
	args := r.Called(ctx, order)
	return args.Error(0)
}

func (r *repoForWarmup) FindByID(ctx context.Context, uid string) (*model.Order, error) {
	args := r.Called(ctx, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Order), args.Error(1)
}

func (r *repoForWarmup) FindAll(ctx context.Context) ([]*model.Order, error) {
	args := r.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Order), args.Error(1)
}

func TestCacheAdapter_WarmUp(t *testing.T) {
	c := cache.NewCash(100, 1*time.Hour)
	ad := cache.NewCacheAdapter(c)

	r := &repoForWarmup{}

	testOrders := []*model.Order{
		{OrderUID: "x"},
		{OrderUID: "y"},
	}

	r.On("FindAll", mock.AnythingOfType("*context.timerCtx")).Return(testOrders, nil)

	assert.NoError(t, ad.WarmUp(r))
	assert.Equal(t, 2, ad.Size())

	r.AssertCalled(t, "FindAll", mock.AnythingOfType("*context.timerCtx"))

	orderX, exists := ad.Get("x")
	assert.True(t, exists)
	assert.Equal(t, "x", orderX.OrderUID)

	orderY, exists := ad.Get("y")
	assert.True(t, exists)
	assert.Equal(t, "y", orderY.OrderUID)
}
