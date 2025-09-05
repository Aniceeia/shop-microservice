package cache

import (
	"context"
	"log"
	"shop-microservice/internal/api/middleware"
	"shop-microservice/internal/domain/model"
	"shop-microservice/internal/domain/repositories"
	"sync"
	"time"
)

// Cache хранит заказы в памяти с TTL и ограничением по размеру
type Cache struct {
	mu       sync.RWMutex
	orders   map[string]*cacheEntry
	stats    map[string]cacheStats
	maxSize  int
	ttl      time.Duration
	cleanup  *time.Ticker
	stopChan chan struct{}
}

type cacheEntry struct {
	order     *model.Order
	expiresAt time.Time
}

type cacheStats struct {
	lastAccess  time.Time
	accessCount int
}

// NewCache создает новый кэш
func NewCache(maxSize int, ttl time.Duration) *Cache {
	c := &Cache{
		orders:   make(map[string]*cacheEntry),
		stats:    make(map[string]cacheStats),
		maxSize:  maxSize,
		ttl:      ttl,
		cleanup:  time.NewTicker(5 * time.Minute),
		stopChan: make(chan struct{}),
	}
	go c.cleanupExpired()
	return c
}

// Set сохраняет заказ в кэш
func (c *Cache) Set(uid string, order *model.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ensureCapacity()

	c.setEntry(uid, order)
}

// Get возвращает заказ из кэша
func (c *Cache) Get(uid string) (*model.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.orders[uid]
	if !exists || time.Now().After(entry.expiresAt) {
		middleware.CacheMisses.Inc()
		return nil, false
	}

	c.updateStats(uid)
	middleware.CacheHits.Inc()
	return entry.order, true
}

// GetAll возвращает все актуальные заказы
func (c *Cache) GetAll() []*model.Order {
	c.mu.RLock()
	defer c.mu.RUnlock()

	now := time.Now()
	result := make([]*model.Order, 0, len(c.orders))
	for _, entry := range c.orders {
		if now.Before(entry.expiresAt) {
			result = append(result, entry.order)
		}
	}
	return result
}

// Delete удаляет заказ из кэша
func (c *Cache) Delete(uid string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.orders, uid)
	delete(c.stats, uid)
}

// Size возвращает количество заказов в кэше
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.orders)
}

// Clear очищает кэш
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.orders = make(map[string]*cacheEntry)
	c.stats = make(map[string]cacheStats)
}

// PreloadCacheFromDB загружает все заказы из БД
func (c *Cache) PreloadCacheFromDB(repo repositories.OrderRepository) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orders, err := repo.FindAll(ctx)
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.orders = make(map[string]*cacheEntry)
	for _, order := range orders {
		if order != nil {
			c.setEntry(order.OrderUID, order)
		}
	}

	log.Printf("Loaded %d orders into cache", len(orders))
	return nil
}

// Stop останавливает тикер очистки
func (c *Cache) Stop() {
	close(c.stopChan)
}

// CleanupTickerForTest тикер только для тестов
func (c *Cache) CleanupTickerForTest(d time.Duration) {
	c.cleanup.Stop()
	c.cleanup = time.NewTicker(d)
	go c.cleanupExpired()
}

func (c *Cache) setEntry(uid string, order *model.Order) {
	c.orders[uid] = &cacheEntry{
		order:     order,
		expiresAt: time.Now().Add(c.ttl),
	}
	c.stats[uid] = cacheStats{
		lastAccess:  time.Now(),
		accessCount: 0,
	}
	middleware.QueueSize.Set(float64(len(c.orders)))
}

func (c *Cache) updateStats(uid string) {
	stats := c.stats[uid]
	stats.lastAccess = time.Now()
	stats.accessCount++
	c.stats[uid] = stats
}

func (c *Cache) ensureCapacity() {
	if len(c.orders) < c.maxSize {
		return
	}
	c.evictOldest()
}

func (c *Cache) evictOldest() {
	var oldestUID string
	var oldestTime time.Time

	for uid, entry := range c.orders {
		if oldestUID == "" || entry.expiresAt.Before(oldestTime) {
			oldestUID = uid
			oldestTime = entry.expiresAt
		}
	}

	if oldestUID != "" {
		delete(c.orders, oldestUID)
		delete(c.stats, oldestUID)
	}
}

func (c *Cache) cleanupExpired() {
	for {
		select {
		case <-c.cleanup.C:
			c.mu.Lock()
			now := time.Now()
			for uid, entry := range c.orders {
				if now.After(entry.expiresAt) {
					delete(c.orders, uid)
					delete(c.stats, uid)
				}
			}
			c.mu.Unlock()
		case <-c.stopChan:
			c.cleanup.Stop()
			return
		}
	}
}
