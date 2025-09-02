package metrics

import (
	"sync"
	"time"
)

type Metrics struct {
	mu            sync.RWMutex
	requestCount  map[string]int64
	errorCount    map[string]int64
	responseTime  map[string][]time.Duration
	cacheHits     int64
	cacheMisses   int64
	dbConnections int64
	kafkaMessages int64
}

func NewMetrics() *Metrics {
	return &Metrics{
		requestCount: make(map[string]int64),
		errorCount:   make(map[string]int64),
		responseTime: make(map[string][]time.Duration),
	}
}

func (m *Metrics) IncrementRequest(endpoint string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requestCount[endpoint]++
}

func (m *Metrics) IncrementError(endpoint string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errorCount[endpoint]++
}

func (m *Metrics) RecordResponseTime(endpoint string, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.responseTime[endpoint]) >= 100 {
		m.responseTime[endpoint] = m.responseTime[endpoint][1:]
	}
	m.responseTime[endpoint] = append(m.responseTime[endpoint], duration)
}

func (m *Metrics) IncrementCacheHit() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cacheHits++
}

func (m *Metrics) IncrementCacheMiss() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cacheMisses++
}

func (m *Metrics) SetDBConnections(count int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.dbConnections = count
}

func (m *Metrics) IncrementKafkaMessages() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.kafkaMessages++
}

func (m *Metrics) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make(map[string]interface{})

	requests := make(map[string]interface{})
	for endpoint, count := range m.requestCount {
		requests[endpoint] = float64(count)
	}
	stats["requests"] = requests

	errors := make(map[string]interface{})
	for endpoint, count := range m.errorCount {
		errors[endpoint] = float64(count)
	}
	stats["errors"] = errors

	stats["cache_hits"] = float64(m.cacheHits)
	stats["cache_misses"] = float64(m.cacheMisses)
	stats["db_connections"] = float64(m.dbConnections)
	stats["kafka_messages"] = float64(m.kafkaMessages)

	avgResponseTimes := make(map[string]interface{})
	for endpoint, times := range m.responseTime {
		if len(times) > 0 {
			var total time.Duration
			for _, t := range times {
				total += t
			}
			// в миллисекунды
			avgMs := float64(total) / float64(len(times)) / float64(time.Millisecond)
			avgResponseTimes[endpoint] = avgMs
		}
	}
	stats["avg_response_times"] = avgResponseTimes

	return stats
}

func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.requestCount = make(map[string]int64)
	m.errorCount = make(map[string]int64)
	m.responseTime = make(map[string][]time.Duration)
	m.cacheHits = 0
	m.cacheMisses = 0
	m.dbConnections = 0
	m.kafkaMessages = 0
}
