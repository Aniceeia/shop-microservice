package kafka

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProducer_Produce(t *testing.T) {
	manager := NewKafkaManager([]string{"localhost:9093"})
	err := manager.WaitForKafka(30 * time.Second)
	if err != nil {
		t.Skipf("Kafka not available, skipping test: %v", err)
	}

	testTopic := "test-producer-topic"
	err = manager.CreateTopicIfNotExists(testTopic, 1, 1)
	require.NoError(t, err)

	// ДАЕМ ВРЕМЯ НА РЕПЛИКАЦИЮ ТОПИКА
	time.Sleep(5 * time.Second) // ← ДОБАВЬТЕ ЭТУ СТРОЧКУ

	producer := NewProducer(ProducerConfig{
		Brokers: []string{"localhost:9093"},
		Topic:   testTopic,
	})
	defer producer.Close()

	// Тестовое сообщение
	testMessage := map[string]string{
		"order_uid": "test-order-123",
		"action":    "created",
	}

	// Отправляем сообщение
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = producer.Produce(ctx, "test-key", testMessage)
	require.NoError(t, err)
}
func TestProducer_Produce_WithInvalidBrokers(t *testing.T) {
	// Producer с неверными брокерами
	producer := NewProducer(ProducerConfig{
		Brokers: []string{"invalid-host:9092"},
		Topic:   "test-topic",
	})
	defer producer.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := producer.Produce(ctx, "test-key", "test-value")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to write message")
}
