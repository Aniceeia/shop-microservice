package middleware

import (
	"context"
	"time"

	"shop-microservice/internal/infrastructure/logger"
	"shop-microservice/internal/infrastructure/metrics"

	"github.com/gin-gonic/gin"
)

func LoggingMiddleware(log *logger.Logger, metrics *metrics.Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		requestID := generateRequestID()

		ctx := context.WithValue(c.Request.Context(), "request_id", requestID)
		c.Request = c.Request.WithContext(ctx)

		log.Info(ctx, "Request started",
			logger.String("method", c.Request.Method),
			logger.String("path", c.Request.URL.Path),
			logger.String("user_agent", c.Request.UserAgent()),
		)

		metrics.IncrementRequest(c.Request.URL.Path)

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()

		if status >= 400 {
			log.Error(ctx, "Request failed",
				logger.Int("status", status),
				logger.Duration("duration", duration),
			)
			metrics.IncrementError(c.Request.URL.Path)
		} else {
			log.Info(ctx, "Request completed",
				logger.Int("status", status),
				logger.Duration("duration", duration),
			)
		}

		metrics.RecordResponseTime(c.Request.URL.Path, duration)
	}
}

func generateRequestID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(6)
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
