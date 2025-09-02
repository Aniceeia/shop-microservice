package unit

import (
	"testing"
	"time"

	"shop-microservice/internal/infrastructure/metrics"

	"github.com/stretchr/testify/assert"
)

func TestMetrics_BasicOperations(t *testing.T) {
	metrics := metrics.NewMetrics()

	metrics.IncrementRequest("/test")
	metrics.IncrementRequest("/test")
	metrics.IncrementError("/test")
	metrics.IncrementCacheHit()
	metrics.IncrementCacheMiss()
	metrics.SetDBConnections(5)
	metrics.IncrementKafkaMessages()

	stats := metrics.GetStats()

	requests := stats["requests"].(map[string]interface{})
	assert.Equal(t, float64(2), requests["/test"])

	errors := stats["errors"].(map[string]interface{})
	assert.Equal(t, float64(1), errors["/test"])

	assert.Equal(t, float64(1), stats["cache_hits"])
	assert.Equal(t, float64(1), stats["cache_misses"])
	assert.Equal(t, float64(5), stats["db_connections"])
	assert.Equal(t, float64(1), stats["kafka_messages"])
}

func TestMetrics_ResponseTime(t *testing.T) {
	metrics := metrics.NewMetrics()

	metrics.RecordResponseTime("/test", 150*time.Millisecond)
	metrics.RecordResponseTime("/test", 200*time.Millisecond)

	stats := metrics.GetStats()
	responseTimes := stats["avg_response_times"].(map[string]interface{})
	avgTime := responseTimes["/test"].(float64)

	assert.InDelta(t, 175.0, avgTime, 1.0)
}

func TestMetrics_CacheOperations(t *testing.T) {
	metrics := metrics.NewMetrics()

	for i := 0; i < 3; i++ {
		metrics.IncrementCacheHit()
	}
	for i := 0; i < 2; i++ {
		metrics.IncrementCacheMiss()
	}

	stats := metrics.GetStats()
	assert.Equal(t, float64(3), stats["cache_hits"])
	assert.Equal(t, float64(2), stats["cache_misses"])
}

func TestMetrics_DBAndKafka(t *testing.T) {
	metrics := metrics.NewMetrics()

	metrics.SetDBConnections(10)
	for i := 0; i < 5; i++ {
		metrics.IncrementKafkaMessages()
	}

	stats := metrics.GetStats()
	assert.Equal(t, float64(10), stats["db_connections"])
	assert.Equal(t, float64(5), stats["kafka_messages"])
}
func TestMetrics_ResponseTimeLimit(t *testing.T) {
	metrics := metrics.NewMetrics()

	metrics.RecordResponseTime("/slow", 10*time.Second)

	stats := metrics.GetStats()
	responseTimes := stats["avg_response_times"].(map[string]interface{})

	avgTime, exists := responseTimes["/slow"]
	assert.True(t, exists, "Expected metrics for /slow endpoint")
	assert.InDelta(t, 10000.0, avgTime.(float64), 1.0)
}

func TestMetrics_Reset(t *testing.T) {
	m := metrics.NewMetrics()

	m.IncrementRequest("/test")
	m.IncrementCacheHit()
	m.SetDBConnections(5)

	m.Reset()

	stats := m.GetStats()
	assert.Equal(t, float64(0), stats["cache_hits"])
	assert.Equal(t, float64(0), stats["db_connections"])
}
