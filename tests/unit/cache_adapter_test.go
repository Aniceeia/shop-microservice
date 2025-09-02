package unit

import (
	"context"
	"testing"

	"shop-microservice/internal/domain/model"
	"shop-microservice/internal/infrastructure/cache"

	"github.com/stretchr/testify/assert"
)

func TestCacheAdapter_BasicOps(t *testing.T) {
	c := cache.NewCash()
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
	c := cache.NewCash()
	ad := cache.NewCacheAdapter(c)
	ad.Set("a", &model.Order{OrderUID: "a"})
	ad.Set("b", &model.Order{OrderUID: "b"})
	all := ad.GetAll()
	assert.Len(t, all, 2)
}

type repoForWarmup struct{ orders []*model.Order }

func (r *repoForWarmup) Save(ctx context.Context, order *model.Order) error { return nil }
func (r *repoForWarmup) FindByID(ctx context.Context, uid string) (*model.Order, error) {
	return nil, nil
}
func (r *repoForWarmup) FindAll(ctx context.Context) ([]*model.Order, error) { return r.orders, nil }

func TestCacheAdapter_WarmUp(t *testing.T) {
	c := cache.NewCash()
	ad := cache.NewCacheAdapter(c)
	r := &repoForWarmup{orders: []*model.Order{{OrderUID: "x"}, {OrderUID: "y"}}}
	assert.NoError(t, ad.WarmUp(r))
	assert.Equal(t, 2, ad.Size())
}
