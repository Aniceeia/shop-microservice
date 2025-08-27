package kafka

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConsumer_Consume(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Используем localhost:9093 для внешнего доступа
	manager := NewKafkaManager([]string{"localhost:9093"})

	// Даем Kafka больше времени на запуск
	err := manager.WaitForKafka(45 * time.Second)
	require.NoError(t, err)

	testTopic := "test-consumer-topic"
	err = manager.CreateTopicIfNotExists(testTopic, 1, 1)
	require.NoError(t, err)

	// Создаем producer с портом 9093
	producer := NewProducer(ProducerConfig{
		Brokers: []string{"localhost:9093"},
		Topic:   testTopic,
	})
	defer producer.Close()

	// Создаем consumer с портом 9093
	consumer := NewConsumer(ConsumerConfig{
		Brokers:     []string{"localhost:9093"}, // Используем порт 9093
		Topic:       testTopic,
		GroupID:     "test-consumer-group",
		StartOffset: -2, // С начала
	})
	defer consumer.Close()

	// Даем время на создание топика
	time.Sleep(5 * time.Second)

	// Отправляем тестовое сообщение
	testMessage := map[string]string{
		"test": "message",
		"time": time.Now().String(),
	}

	ctxProduce, cancelProduce := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelProduce()

	err = producer.Produce(ctxProduce, "test-key", testMessage)
	require.NoError(t, err)

	// Даем время на доставку сообщения
	time.Sleep(3 * time.Second)

	// Запускаем consumer
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	messagesReceived := 0
	err = consumer.Consume(ctx, func(key string, value []byte) error {
		t.Logf("Received message: key=%s, value=%s", key, string(value))
		messagesReceived++
		// Завершаем после первого сообщения
		cancel()
		return nil
	})

	// Ожидаем контекст cancelled, это нормально
	if err != nil && err != context.Canceled {
		assert.NoError(t, err)
	}

	assert.Greater(t, messagesReceived, 0, "Should receive at least one message")
}
