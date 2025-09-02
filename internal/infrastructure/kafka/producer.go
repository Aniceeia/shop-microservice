package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/compress"
)

type Producer struct {
	writer *kafka.Writer
	topic  string
}

type ProducerConfig struct {
	Brokers []string
	Topic   string
}

func NewProducer(cfg ProducerConfig) *Producer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Topic:        cfg.Topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireOne,
		Compression:  compress.Snappy,
		BatchTimeout: 10 * time.Millisecond,
		BatchSize:    100,
		Async:        true,
		Logger:       kafka.LoggerFunc(log.Printf),
		ErrorLogger:  kafka.LoggerFunc(log.Printf),
	}

	return &Producer{
		writer: writer,
		topic:  cfg.Topic,
	}
}

func (p *Producer) Produce(ctx context.Context, key string, value interface{}) error {
	return p.produceWithRetry(ctx, key, value)
}

func (p *Producer) produceWithRetry(ctx context.Context, key string, value interface{}) error {
	maxRetries := 3
	baseDelay := 100 * time.Millisecond

	for attempt := 0; attempt < maxRetries; attempt++ {
		err := p.produceOnce(ctx, key, value)
		if err == nil {
			return nil
		}

		if attempt == maxRetries-1 {
			return fmt.Errorf("failed to produce message after %d attempts: %w", maxRetries, err)
		}

		delay := time.Duration(float64(baseDelay) * math.Pow(2, float64(attempt)))
		log.Printf("Kafka produce failed, retrying in %v: %v", delay, err)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			continue
		}
	}
	return nil
}

func (p *Producer) produceOnce(ctx context.Context, key string, value interface{}) error {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(key),
		Value: jsonValue,
		Time:  time.Now(),
	}

	err = p.writer.WriteMessages(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
