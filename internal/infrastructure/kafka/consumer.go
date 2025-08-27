package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
	topic  string
}

type ConsumerConfig struct {
	Brokers     []string
	Topic       string
	GroupID     string
	StartOffset int64
}

type MessageHandler func(key string, value []byte) error

func NewConsumer(cfg ConsumerConfig) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.Brokers,
		Topic:          cfg.Topic,
		GroupID:        cfg.GroupID,
		MinBytes:       10e3, // 10KB
		MaxBytes:       10e6, // 10MB
		CommitInterval: time.Second,
		StartOffset:    cfg.StartOffset,
		Logger:         kafka.LoggerFunc(log.Printf),
		ErrorLogger:    kafka.LoggerFunc(log.Printf),
	})

	return &Consumer{
		reader: reader,
		topic:  cfg.Topic,
	}
}

func (c *Consumer) Consume(ctx context.Context, handler MessageHandler) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				return fmt.Errorf("failed to read message: %w", err)
			}

			if err := handler(string(msg.Key), msg.Value); err != nil {
				log.Printf("Error handling message: %v", err)
				continue
			}

			log.Printf("Consumed message: topic=%s key=%s", c.topic, string(msg.Key))
		}
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}

func (c *Consumer) ConsumeJSON(ctx context.Context, handler func(key string, value interface{}) error, target interface{}) error {
	return c.Consume(ctx, func(key string, value []byte) error {
		if err := json.Unmarshal(value, target); err != nil {
			return fmt.Errorf("failed to unmarshal JSON: %w", err)
		}
		return handler(key, target)
	})
}
