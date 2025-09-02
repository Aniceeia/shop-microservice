package unit

import (
	"context"
	"testing"
	"time"

	"shop-microservice/internal/domain/model"
	"shop-microservice/internal/domain/repositories"
	"shop-microservice/internal/infrastructure/cache"

	"github.com/stretchr/testify/assert"
)

func TestCache_SetGetDelete(t *testing.T) {
	c := cache.NewCash()
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
	c := cache.NewCash()
	c.Set("a", &model.Order{OrderUID: "a"})
	c.Set("b", &model.Order{OrderUID: "b"})
	all := c.GetAll()
	assert.Len(t, all, 2)
	assert.Equal(t, 2, c.Size())
	c.Clear()
	assert.Equal(t, 0, c.Size())
}

func TestCache_WarmUp(t *testing.T) {
	c := cache.NewCash()
	mr := &mockOrderRepo{orders: []*model.Order{{OrderUID: "x"}}}
	assert.NoError(t, c.WarmUp(mr))
	assert.Equal(t, 1, c.Size())
}

func TestCache_WarmUp_Error(t *testing.T) {
	c := cache.NewCash()
	mr := &mockOrderRepoErr{}
	err := c.WarmUp(mr)
	assert.Error(t, err)
}

type mockOrderRepo struct{ orders []*model.Order }

func (m *mockOrderRepo) Save(ctx context.Context, order *model.Order) error { return nil }
func (m *mockOrderRepo) FindByID(ctx context.Context, uid string) (*model.Order, error) {
	return nil, repositories.ErrOrderNotFound
}
func (m *mockOrderRepo) FindAll(ctx context.Context) ([]*model.Order, error) {
	time.Sleep(5 * time.Millisecond)
	return m.orders, nil
}

type mockOrderRepoErr struct{}

func (m *mockOrderRepoErr) Save(ctx context.Context, order *model.Order) error { return nil }
func (m *mockOrderRepoErr) FindByID(ctx context.Context, uid string) (*model.Order, error) {
	return nil, repositories.ErrOrderNotFound
}
func (m *mockOrderRepoErr) FindAll(ctx context.Context) ([]*model.Order, error) {
	return nil, assert.AnError
}
