package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/compress"
)

// ProducerIface — интерфейс для продюсера
type ProducerIface interface {
	Produce(ctx context.Context, key string, value interface{}) error
	Close() error
}

// kafkaWriterIface — внутренний интерфейс для writer
type kafkaWriterIface interface {
	WriteMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}

// Producer — реализация продюсера для Kafka
type Producer struct {
	Writer kafkaWriterIface
	Topic  string
}

type ProducerConfig struct {
	Brokers []string
	Topic   string
}

// NewProducer создаёт новый Kafka producer
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
		Writer: writer,
		Topic:  cfg.Topic,
	}
}

// Produce — отправка сообщения (с чанками, если нужно)
func (p *Producer) Produce(ctx context.Context, key string, value interface{}) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal message failed: %w", err)
	}

	// Проверяем, нужно ли чанковать
	if !ShouldChunk(jsonData) {
		// Отправляем как обычное сообщение
		return p.Writer.WriteMessages(ctx, kafka.Message{
			Key:   []byte(key),
			Value: jsonData,
			Time:  time.Now(),
		})
	}

	// Чанкуем данные
	chunks, err := SplitIntoChunks(key, jsonData, ChunkSize)
	if err != nil {
		return fmt.Errorf("split into chunks failed: %w", err)
	}

	// Отправляем каждый чанк
	for _, chunk := range chunks {
		chunkData, err := json.Marshal(chunk)
		if err != nil {
			return fmt.Errorf("marshal chunk failed: %w", err)
		}

		if err := p.Writer.WriteMessages(ctx, kafka.Message{
			Key:   []byte(key),
			Value: chunkData,
			Time:  time.Now(),
		}); err != nil {
			return fmt.Errorf("write chunk failed: %w", err)
		}
	}

	return nil
}

// Close — закрывает Kafka writer
func (p *Producer) Close() error {
	return p.Writer.Close()
}
