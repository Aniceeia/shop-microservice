package unit

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"shop-microservice/internal/domain/model"
	"shop-microservice/internal/infrastructure/cache"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCache_SetGetDelete(t *testing.T) {
	c := cache.NewCash(100, 1*time.Hour)
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
	c := cache.NewCash(100, 1*time.Hour)
	c.Set("a", &model.Order{OrderUID: "a"})
	c.Set("b", &model.Order{OrderUID: "b"})
	all := c.GetAll()
	assert.Len(t, all, 2)
	assert.Equal(t, 2, c.Size())
	c.Clear()
	assert.Equal(t, 0, c.Size())
}

func TestCache_WarmUp(t *testing.T) {
	c := cache.NewCash(100, 1*time.Hour)
	mr := &mockOrderRepo{orders: []*model.Order{{OrderUID: "x"}}}
	assert.NoError(t, c.WarmUp(mr))
	assert.Equal(t, 1, c.Size())
}

func TestCache_TTL_Expiration(t *testing.T) {
	c := cache.NewCash(100, 10*time.Millisecond)
	ord := &model.Order{OrderUID: "expire"}
	c.Set("expire", ord)

	_, ok := c.Get("expire")
	assert.True(t, ok)

	time.Sleep(20 * time.Millisecond)
	_, ok = c.Get("expire")
	assert.False(t, ok)
}

func TestCache_MaxSize_Eviction(t *testing.T) {
	c := cache.NewCash(2, 1*time.Hour)
	c.Set("a", &model.Order{OrderUID: "a"})
	c.Set("b", &model.Order{OrderUID: "b"})
	c.Set("c", &model.Order{OrderUID: "c"})

	assert.Equal(t, 2, c.Size())
	all := c.GetAll()
	assert.Len(t, all, 2)
}

func TestCache_NewCash(t *testing.T) {
	cache := cache.NewCash(100, 10*time.Minute)
	assert.Equal(t, 0, cache.Size())

	cache.Set("test", &model.Order{OrderUID: "test"})
	assert.Equal(t, 1, cache.Size())
}

func TestCache_WarmUp_Error(t *testing.T) {
	c := cache.NewCash(100, 1*time.Hour)
	defer c.Stop()

	mockRepo := new(mockOrderRepository)
	mockRepo.On("FindAll", mock.Anything).Return(nil, errors.New("database error"))

	err := c.WarmUp(mockRepo)
	assert.Error(t, err)
	assert.Equal(t, 0, c.Size())

	mockRepo.AssertExpectations(t)
}

type mockOrderRepo struct{ orders []*model.Order }

func (m *mockOrderRepo) Save(ctx context.Context, order *model.Order) error { return nil }
func (m *mockOrderRepo) FindByID(ctx context.Context, uid string) (*model.Order, error) {
	return nil, fmt.Errorf("user not foud")
}
func (m *mockOrderRepo) FindAll(ctx context.Context) ([]*model.Order, error) {
	time.Sleep(5 * time.Millisecond)
	return m.orders, nil
}
