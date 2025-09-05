package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	kafkaGo "github.com/segmentio/kafka-go"
)

// ReaderInterface — интерфейс для чтения сообщений
type ReaderInterface interface {
	ReadMessage(ctx context.Context) (kafkaGo.Message, error)
	Close() error
}

// MessageHandler — функция обработки сообщений
type MessageHandler func(key string, value []byte) error

// Consumer — основной консьюмер
type Consumer struct {
	Reader ReaderInterface
	topic  string
}

// messageBuffer для хранения чанков
type messageBuffer struct {
	mu        sync.Mutex
	chunks    [][]byte
	count     int
	createdAt time.Time
}

// NewConsumer создаёт Consumer с реальным kafka.Reader
func NewConsumer(brokers []string, topic, groupID string) *Consumer {
	reader := kafkaGo.NewReader(kafkaGo.ReaderConfig{
		Brokers:  brokers,
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 10e3,
		MaxBytes: 200e6,
	})
	return &Consumer{Reader: reader, topic: topic}
}

// Consume — цикл чтения сообщений
func (c *Consumer) Consume(ctx context.Context, handler MessageHandler) error {
	buffer := make(map[string]*messageBuffer)
	var bufferMu sync.Mutex

	// Горутина для очистки старых буферов
	go c.cleanupOldBuffers(ctx, &bufferMu, buffer)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			kafkaMsg, err := c.Reader.ReadMessage(ctx)
			if err != nil {
				if err == context.Canceled {
					return nil
				}
				return fmt.Errorf("read message failed: %w", err)
			}

			if err := c.processMessage(kafkaMsg, handler, &bufferMu, buffer); err != nil {
				log.Printf("process message failed: %v", err)
			}
		}
	}
}

// processMessage обрабатывает одно сообщение
func (c *Consumer) processMessage(
	kafkaMsg kafkaGo.Message,
	handler MessageHandler,
	bufferMu *sync.Mutex,
	buffer map[string]*messageBuffer,
) error {
	// Пытаемся разобрать как ChunkedMessage
	var chunk ChunkedMessage
	if err := json.Unmarshal(kafkaMsg.Value, &chunk); err != nil {
		// Если не JSON, обрабатываем как обычное сообщение
		return handler(string(kafkaMsg.Key), kafkaMsg.Value)
	}

	// Если это не чанкованное сообщение
	if chunk.ChunkCount <= 1 {
		return handler(chunk.Key, chunk.Value)
	}

	// Обработка чанкованного сообщения
	bufferMu.Lock()
	if _, exists := buffer[chunk.Key]; !exists {
		buffer[chunk.Key] = &messageBuffer{
			chunks:    make([][]byte, chunk.ChunkCount),
			count:     chunk.ChunkCount,
			createdAt: time.Now(),
		}
	}
	buf := buffer[chunk.Key]
	bufferMu.Unlock()

	buf.mu.Lock()
	defer buf.mu.Unlock()

	// Сохраняем чанк
	if chunk.ChunkIndex < len(buf.chunks) {
		buf.chunks[chunk.ChunkIndex] = chunk.Value
	}

	// Проверяем, все ли чанки получены
	if AllChunksReceived(buf.chunks) {
		fullMessage := MergeChunks(buf.chunks)

		bufferMu.Lock()
		delete(buffer, chunk.Key)
		bufferMu.Unlock()

		return handler(chunk.Key, fullMessage)
	}

	return nil
}

// cleanupOldBuffers очищает старые буферы
func (c *Consumer) cleanupOldBuffers(ctx context.Context, bufferMu *sync.Mutex, buffer map[string]*messageBuffer) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			bufferMu.Lock()
			for key, buf := range buffer {
				if time.Since(buf.createdAt) > 5*time.Minute {
					delete(buffer, key)
				}
			}
			bufferMu.Unlock()
		}
	}
}

// Close закрывает reader
func (c *Consumer) Close() error {
	return c.Reader.Close()
}

// Topic возвращает топик консьюмера
func (c *Consumer) Topic() string {
	return c.topic
}
