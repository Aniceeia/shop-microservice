package cache

import (
	"context"
	"log"
	"shop-microservice/internal/domain/model"
	"shop-microservice/internal/domain/repositories"

	"sync"
	"time"
)

type Cache struct {
	mu     sync.RWMutex
	orders map[string]*model.Order
	stats  map[string]cacheStats
}

// исправить потом
func (c *Cache) NewCacheAdapter(cache *Cache) repositories.Cache {
	return c
}

type cacheStats struct {
	lastAccess  time.Time
	accessCount int
}

func NewCash() *Cache {
	return &Cache{
		orders: make(map[string]*model.Order),
		stats:  make(map[string]cacheStats),
	}
}

func (c *Cache) Set(uid string, order *model.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.orders[uid] = order
	c.stats[uid] = cacheStats{
		lastAccess:  time.Now(),
		accessCount: 0,
	}
}

func (c *Cache) Get(uid string) (*model.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	order, exists := c.orders[uid]
	if exists {
		stats := c.stats[uid]
		stats.lastAccess = time.Now()
		stats.accessCount++
		c.stats[uid] = stats
	}
	return order, exists
}

func (cache *Cache) GetAll() []*model.Order {
	cache.mu.RLock()
	defer cache.mu.RUnlock()

	orders := make([]*model.Order, 0, len(cache.orders))
	for _, order := range cache.orders {
		orders = append(orders, order)
	}
	return orders
}

func (cache *Cache) Delete(uid string) {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	delete(cache.orders, uid)
}

func (cache *Cache) Size() int {
	cache.mu.RLock()
	defer cache.mu.RUnlock()
	return len(cache.orders)
}

func (cache *Cache) Clear() {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	cache.orders = make(map[string]*model.Order)
}

// WarmUp заполняет кэш данными из репозитория при старте сервиса
func (cache *Cache) WarmUp(repo repositories.OrderRepository) error {
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orders, err := repo.FindAll(ctx)
	if err != nil {
		return err
	}

	cache.mu.Lock()
	defer cache.mu.Unlock()

	// Очищаем текущий кэш перед заполнением
	cache.orders = make(map[string]*model.Order)

	for _, order := range orders {
		if order != nil {
			cache.orders[order.OrderUID] = order
		}
	}

	log.Printf("Cache warm-up completed. Loaded %d orders in %v", len(orders), time.Since(start))
	return nil
}
