package kafka

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

// KafkaManager управляет топиками и проверкой состояния Kafka
type KafkaManager struct {
	brokers []string
}

// NewKafkaManager создает новый менеджер Kafka
func NewKafkaManager(brokers []string) *KafkaManager {
	return &KafkaManager{brokers: brokers}
}

// CreateTopicIfNotExists создает топик, если он не существует
//
// swagger:operation POST /kafka/topic createKafkaTopic
// ---
// summary: Создает топик в Kafka
// parameters:
//   - name: topic
//     in: query
//     type: string
//     required: true
//   - name: partitions
//     in: query
//     type: integer
//     required: true
//   - name: replicationFactor
//     in: query
//     type: integer
//     required: true
//
// responses:
//
//	"200":
//	  description: Топик создан или уже существует
//	"500":
//	  description: Ошибка создания топика
func (m *KafkaManager) CreateTopicIfNotExists(topic string, partitions, replicationFactor int) error {
	return m.withController(func(conn *kafka.Conn) error {
		topicConfigs := []kafka.TopicConfig{
			{
				Topic:             topic,
				NumPartitions:     partitions,
				ReplicationFactor: replicationFactor,
			},
		}

		if err := conn.CreateTopics(topicConfigs...); err != nil {
			if err.Error() == "topic already exists" {
				log.Printf("topic %s already exists", topic)
				return nil
			}
			return fmt.Errorf("failed to create topic: %w", err)
		}

		log.Printf("topic %s created successfully", topic)
		return nil
	})
}

// HealthCheck проверяет доступность Kafka
//
// swagger:operation GET /kafka/health kafkaHealthCheck
// ---
// summary: Проверка доступности Kafka
// responses:
//
//	"200":
//	  description: Kafka доступна
//	"500":
//	  description: Kafka недоступна
func (m *KafkaManager) HealthCheck() error {
	conn, err := kafka.Dial("tcp", m.brokers[0])
	if err != nil {
		return fmt.Errorf("failed to connect to kafka: %w", err)
	}
	defer conn.Close()

	if _, err := conn.Brokers(); err != nil {
		return fmt.Errorf("failed to get brokers: %w", err)
	}

	return nil
}

// TryToConnectKafkaFor пытается подключиться к Kafka в течение заданного таймаута
//
// swagger:operation GET /kafka/connect kafkaTryConnect
// ---
// summary: Попытка подключения к Kafka с таймаутом
// parameters:
//   - name: timeout
//     in: query
//     type: string
//     required: true
//     description: Время ожидания (например, 10s, 1m)
//
// responses:
//
//	"200":
//	  description: Подключение успешно
//	"500":
//	  description: Kafka недоступна
func (m *KafkaManager) TryToConnectKafkaFor(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("kafka not available after %v", timeout)
		case <-ticker.C:
			if err := m.HealthCheck(); err == nil {
				return nil
			}
		}
	}
}

func (m *KafkaManager) withController(fn func(conn *kafka.Conn) error) error {
	conn, err := kafka.Dial("tcp", m.brokers[0])
	if err != nil {
		return fmt.Errorf("failed to dial kafka: %w", err)
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("failed to get controller: %w", err)
	}

	controllerConn, err := kafka.Dial("tcp", fmt.Sprintf("%s:%d", controller.Host, controller.Port))
	if err != nil {
		return fmt.Errorf("failed to dial controller: %w", err)
	}
	defer controllerConn.Close()

	return fn(controllerConn)
}
