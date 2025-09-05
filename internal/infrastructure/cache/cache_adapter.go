package cache

import (
	"shop-microservice/internal/domain/model"
	"shop-microservice/internal/domain/repositories"
)

// CacheAdapter делает Cache совместимым с интерфейсом repositories.Cache
type CacheAdapter struct {
	cache *Cache
}

func NewCacheAdapter(cache *Cache) repositories.Cache {
	return &CacheAdapter{cache: cache}
}

func (ca *CacheAdapter) Set(uid string, order *model.Order) {
	ca.cache.Set(uid, order)
}
func (ca *CacheAdapter) Get(uid string) (*model.Order, bool) {
	return ca.cache.Get(uid)
}
func (ca *CacheAdapter) GetAll() []*model.Order {
	return ca.cache.GetAll()
}
func (ca *CacheAdapter) Delete(uid string) {
	ca.cache.Delete(uid)
}
func (ca *CacheAdapter) Size() int {
	return ca.cache.Size()
}
func (ca *CacheAdapter) Clear() {
	ca.cache.Clear()
}
func (ca *CacheAdapter) PreloadCacheFromDB(repo repositories.OrderRepository) error {
	return ca.cache.PreloadCacheFromDB(repo)
}
