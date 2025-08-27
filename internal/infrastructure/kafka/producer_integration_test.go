package kafka

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestProducer_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	manager := NewKafkaManager([]string{"localhost:9093"})
	err := manager.WaitForKafka(30 * time.Second)
	require.NoError(t, err)

	testTopic := "test-orders-integration"
	err = manager.CreateTopicIfNotExists(testTopic, 1, 1)
	require.NoError(t, err)

	// ДАЕМ ВРЕМЯ НА РЕПЛИКАЦИЮ ТОПИКА
	time.Sleep(5 * time.Second) // ← ДОБАВЬТЕ ЭТУ СТРОЧКУ

	producer := NewProducer(ProducerConfig{
		Brokers: []string{"localhost:9093"},
		Topic:   testTopic,
	})
	defer producer.Close()

	// Тестовые данные
	testOrder := map[string]interface{}{
		"order_uid":    "test-order-integration-123",
		"track_number": "WBILMTESTTRACK",
		"entry":        "WBIL",
		"customer_id":  "test-customer",
	}

	// Отправляем сообщение
	ctxProduce, cancelProduce := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelProduce()

	err = producer.Produce(ctxProduce, testOrder["order_uid"].(string), testOrder)
	require.NoError(t, err)

	t.Logf("Successfully produced message to topic %s", testTopic)
}
