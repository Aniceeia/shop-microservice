package metrics

import (
	"shop-microservice/internal/api/middleware"
	"sync"
)

type Metrics struct {
	mu sync.RWMutex
}

func NewMetrics() *Metrics {
	return &Metrics{}
}

func (m *Metrics) IncrementCacheHit() {
	middleware.CacheHits.Inc()
}

func (m *Metrics) IncrementCacheMiss() {
	middleware.CacheMisses.Inc()
}

func (m *Metrics) SetDBConnections(count int64) {
	middleware.DBConnections.Set(float64(count))
}

func (m *Metrics) IncrementKafkaMessages() {
	middleware.KafkaMessagesProduced.Inc()
}
