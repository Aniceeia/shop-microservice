package kafka

import (
	"context"
	"fmt"
	"log"
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
			log.Printf("Topic %s already exists", topic)
			return nil
		}
		return fmt.Errorf("failed to create topic: %w", err)
	}

	log.Printf("Topic %s created successfully", topic)
	return nil
}

func (m *KafkaManager) HealthCheck() error {
	conn, err := kafka.Dial("tcp", m.brokers[0])
	if err != nil {
		return fmt.Errorf("failed to connect to kafka: %w", err)
	}
	defer conn.Close()

	_, err = conn.Brokers()
	if err != nil {
		return fmt.Errorf("failed to get brokers: %w", err)
	}

	return nil
}

func (m *KafkaManager) WaitForKafka(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("kafka not available after %v", timeout)
		default:
			if err := m.HealthCheck(); err == nil {
				return nil
			}
			time.Sleep(1 * time.Second)
		}
	}
}
