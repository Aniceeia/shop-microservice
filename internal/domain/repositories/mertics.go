package repositories

type Metrics interface {
	IncrementCacheHit()
	IncrementCacheMiss()
	SetDBConnections(count int64)
	IncrementKafkaMessages()
}
