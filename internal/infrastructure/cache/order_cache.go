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

func NewCash(maxSize int, ttl time.Duration) *Cache {
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

func (c *Cache) Set(uid string, order *model.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.orders) >= c.maxSize {
		c.evictOldest()
	}

	c.orders[uid] = &cacheEntry{
		order:     order,
		expiresAt: time.Now().Add(c.ttl),
	}
	c.stats[uid] = cacheStats{
		lastAccess:  time.Now(),
		accessCount: 0,
	}
}

func (c *Cache) Get(uid string) (*model.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.orders[uid]
	if !exists {
		return nil, false
	}

	if time.Now().After(entry.expiresAt) {
		return nil, false
	}

	stats := c.stats[uid]
	stats.lastAccess = time.Now()
	stats.accessCount++
	c.stats[uid] = stats

	return entry.order, true
}

func (c *Cache) GetAll() []*model.Order {
	c.mu.RLock()
	defer c.mu.RUnlock()

	orders := make([]*model.Order, 0, len(c.orders))
	now := time.Now()
	for _, entry := range c.orders {
		if now.Before(entry.expiresAt) {
			orders = append(orders, entry.order)
		}
	}
	return orders
}

func (c *Cache) Delete(uid string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.orders, uid)
	delete(c.stats, uid)
}

func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.orders)
}

func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.orders = make(map[string]*cacheEntry)
	c.stats = make(map[string]cacheStats)
}

func (c *Cache) WarmUp(repo repositories.OrderRepository) error {
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orders, err := repo.FindAll(ctx)
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.orders = make(map[string]*cacheEntry)
	now := time.Now()

	for _, order := range orders {
		if order != nil {
			c.orders[order.OrderUID] = &cacheEntry{
				order:     order,
				expiresAt: now.Add(c.ttl),
			}
		}
	}

	log.Printf("Cache warm-up completed. Loaded %d orders in %v", len(orders), time.Since(start))
	return nil
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

func (c *Cache) Stop() {
	close(c.stopChan)
}
