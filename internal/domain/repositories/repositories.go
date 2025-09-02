package repositories

import "time"

type Metrics interface {
	IncrementRequest(endpoint string)
	IncrementError(endpoint string)
	RecordResponseTime(endpoint string, duration time.Duration)
	IncrementCacheHit()
	IncrementCacheMiss()
	SetDBConnections(count int64)
	IncrementKafkaMessages()
	GetStats() map[string]any
	Reset()
}
