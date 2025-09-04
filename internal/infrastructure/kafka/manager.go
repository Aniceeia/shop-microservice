package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaManager struct {
	brokers []string
}

func NewKafkaManager(brokers []string) *KafkaManager {
	return &KafkaManager{brokers: brokers}
}

func (m *KafkaManager) CreateTopicIfNotExists(topic string, partitions int, replicationFactor int) error {
	conn, err := kafka.Dial("tcp", m.brokers[0])
	if err != nil {
		return errFail("failed to dial kafka: %w", err)
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return errFail("failed to get controller: %w", err)
	}

	controllerConn, err := kafka.Dial("tcp", fmt.Sprintf("%s:%d", controller.Host, controller.Port))
	if err != nil {
		return errFail("failed to dial controller: %w", err)
	}
	defer controllerConn.Close()

	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             topic,
			NumPartitions:     partitions,
			ReplicationFactor: replicationFactor,
		},
	}

	err = controllerConn.CreateTopics(topicConfigs...)
	if err != nil {
		if err.Error() == "topic already exists" {
			kafkaLog("topic %s already exists", topic)
			return nil
		}
		return errFail("failed to create topic: %w", err)
	}

	kafkaLog("topic %s created successfully", topic)
	return nil
}

func (m *KafkaManager) HealthCheck() error {
	conn, err := kafka.Dial("tcp", m.brokers[0])
	if err != nil {
		return errFail("failed to connect to kafka: %w", err)
	}
	defer conn.Close()

	_, err = conn.Brokers()
	if err != nil {
		return errFail("failed to get brokers: %w", err)
	}

	return nil
}

func (m *KafkaManager) TryToConnectKafkaFor(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return errFail("kafka not available after %v", timeout)
		default:
			if err := m.HealthCheck(); err == nil {
				return nil
			}
			time.Sleep(1 * time.Second)
		}
	}
}
