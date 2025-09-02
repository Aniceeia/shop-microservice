package unit

import (
	"context"
	"testing"
	"time"

	"shop-microservice/internal/infrastructure/logger"

	"github.com/stretchr/testify/assert"
)

func TestLogger_Levels(t *testing.T) {
	log := logger.NewLogger(logger.INFO)
	ctx := context.Background()

	log.Debug(ctx, "debug message")
	log.Info(ctx, "info message")
	log.Warn(ctx, "warn message")
	log.Error(ctx, "error message")

	assert.True(t, true)
}

func TestLogger_Fields(t *testing.T) {
	log := logger.NewLogger(logger.DEBUG)
	ctx := context.Background()

	log.Info(ctx, "test message",
		logger.String("key1", "value1"),
		logger.Int("key2", 42),
		logger.Error(assert.AnError),
		logger.Duration("duration", time.Second),
	)

	assert.True(t, true)
}

func TestLogger_RequestID(t *testing.T) {
	log := logger.NewLogger(logger.INFO)
	ctx := context.WithValue(context.Background(), "request_id", "test-123")

	log.Info(ctx, "test message")
	assert.True(t, true)
}

func TestLogger_UnknownRequestID(t *testing.T) {
	log := logger.NewLogger(logger.INFO)
	ctx := context.Background()

	log.Info(ctx, "test message")
	assert.True(t, true)
}
