package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	HttpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests",
	}, []string{"method", "endpoint", "status_code"})

	HttpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request duration in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "endpoint"})

	CacheHits = promauto.NewCounter(prometheus.CounterOpts{
		Name: "cache_hits_total",
		Help: "Total number of cache hits",
	})

	CacheMisses = promauto.NewCounter(prometheus.CounterOpts{
		Name: "cache_misses_total",
		Help: "Total number of cache misses",
	})

	DBConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "db_connections_current",
		Help: "Current number of database connections",
	})

	KafkaMessagesProduced = promauto.NewCounter(prometheus.CounterOpts{
		Name: "kafka_messages_produced_total",
		Help: "Total number of Kafka messages produced",
	})

	OrdersCreated = promauto.NewCounter(prometheus.CounterOpts{
		Name: "orders_created_total",
		Help: "Total number of orders created",
	})

	OrdersRetrieved = promauto.NewCounter(prometheus.CounterOpts{
		Name: "orders_retrieved_total",
		Help: "Total number of orders retrieved",
	})

	ValidationErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "validation_errors_total",
		Help: "Total number of validation errors",
	})

	QueueSize = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "producer_queue_size_current",
		Help: "Current size of producer queue",
	})

	QueueCapacity = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "producer_queue_capacity",
		Help: "Capacity of producer queue",
	})

	WorkerErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "worker_errors_total",
		Help: "Total number of worker errors",
	}, []string{"worker_number"})
)

func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		c.Next()

		duration := time.Since(start).Seconds()
		statusCode := strconv.Itoa(c.Writer.Status())

		HttpRequestDuration.WithLabelValues(
			c.Request.Method,
			path,
		).Observe(duration)

		HttpRequestsTotal.WithLabelValues(
			c.Request.Method,
			path,
			statusCode,
		).Inc()
	}
}

func GetPrometheusMetrics(c *gin.Context) {
	promhttp.Handler().ServeHTTP(c.Writer, c.Request)
}
