package unit

import (
	"errors"
	"testing"
	"time"

	"shop-microservice/internal/domain/model"
	"shop-microservice/internal/infrastructure/cache"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// cache

func TestCache_SetGetDelete(t *testing.T) {
	c := cache.NewCache(100, 1*time.Hour)
	defer c.Stop()

	ord := &model.Order{OrderUID: "u1"}
	c.Set("u1", ord)

	got, ok := c.Get("u1")
	assert.True(t, ok)
	assert.Equal(t, "u1", got.OrderUID)

	c.Delete("u1")
	_, ok = c.Get("u1")
	assert.False(t, ok)
}

func TestCache_GetAll_Size_Clear(t *testing.T) {
	c := cache.NewCache(100, 1*time.Hour)
	defer c.Stop()

	c.Set("a", &model.Order{OrderUID: "a"})
	c.Set("b", &model.Order{OrderUID: "b"})

	all := c.GetAll()
	assert.Len(t, all, 2)
	assert.Equal(t, 2, c.Size())

	c.Clear()
	assert.Equal(t, 0, c.Size())
}

func TestCache_TTL_Expiration(t *testing.T) {
	c := cache.NewCache(100, 10*time.Millisecond)
	defer c.Stop()

	ord := &model.Order{OrderUID: "expire"}
	c.Set("expire", ord)

	_, ok := c.Get("expire")
	assert.True(t, ok)

	time.Sleep(20 * time.Millisecond)
	_, ok = c.Get("expire")
	assert.False(t, ok)
}

func TestCache_MaxSize_Eviction(t *testing.T) {
	c := cache.NewCache(2, 1*time.Hour)
	defer c.Stop()

	c.Set("a", &model.Order{OrderUID: "a"})
	c.Set("b", &model.Order{OrderUID: "b"})
	c.Set("c", &model.Order{OrderUID: "c"})

	assert.Equal(t, 2, c.Size())
	all := c.GetAll()
	assert.Len(t, all, 2)
}

func TestCache_NewCash(t *testing.T) {
	c := cache.NewCache(100, 10*time.Minute)
	defer c.Stop()

	assert.Equal(t, 0, c.Size())
	c.Set("test", &model.Order{OrderUID: "test"})
	assert.Equal(t, 1, c.Size())
}

func TestCache_Preload_Error(t *testing.T) {
	c := cache.NewCache(100, 1*time.Hour)
	defer c.Stop()

	mockRepo := new(mockOrderRepository)
	mockRepo.On("FindAll", mock.Anything).Return(nil, errors.New("database error"))

	err := c.PreloadCacheFromDB(mockRepo)
	assert.Error(t, err)
	assert.Equal(t, 0, c.Size())

	mockRepo.AssertExpectations(t)
}

func TestCache_CleanupExpired(t *testing.T) {
	c := cache.NewCache(100, 10*time.Millisecond)
	defer c.Stop()

	c.CleanupTickerForTest(20 * time.Millisecond)

	c.Set("toExpire", &model.Order{OrderUID: "toExpire"})

	_, ok := c.Get("toExpire")
	assert.True(t, ok)

	time.Sleep(50 * time.Millisecond)

	_, ok = c.Get("toExpire")
	assert.False(t, ok)
}

// adapter

func TestCacheAdapter_BasicOps(t *testing.T) {
	c := cache.NewCache(100, 1*time.Hour)
	defer c.Stop()

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
	assert.Equal(t, 0, ad.Size())
}

func TestCacheAdapter_GetAll(t *testing.T) {
	c := cache.NewCache(100, 1*time.Hour)
	defer c.Stop()

	ad := cache.NewCacheAdapter(c)
	ad.Set("a", &model.Order{OrderUID: "a"})
	ad.Set("b", &model.Order{OrderUID: "b"})

	all := ad.GetAll()
	assert.Len(t, all, 2)
}

func TestCacheAdapter_Preload(t *testing.T) {
	c := cache.NewCache(100, 1*time.Hour)
	defer c.Stop()

	ad := cache.NewCacheAdapter(c)
	r := &repoForPreload{}

	testOrders := []*model.Order{
		{OrderUID: "x"},
		{OrderUID: "y"},
	}

	r.On("FindAll", mock.AnythingOfType("*context.timerCtx")).Return(testOrders, nil)

	assert.NoError(t, ad.PreloadCacheFromDB(r))
	assert.Equal(t, 2, ad.Size())

	r.AssertCalled(t, "FindAll", mock.AnythingOfType("*context.timerCtx"))

	orderX, exists := ad.Get("x")
	assert.True(t, exists)
	assert.Equal(t, "x", orderX.OrderUID)

	orderY, exists := ad.Get("y")
	assert.True(t, exists)
	assert.Equal(t, "y", orderY.OrderUID)
}
