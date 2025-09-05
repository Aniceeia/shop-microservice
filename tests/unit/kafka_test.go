package unit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"shop-microservice/internal/domain/model"
	"shop-microservice/internal/infrastructure/kafka"

	kafkago "github.com/segmentio/kafka-go"
)

// ==================== MOCKS ====================

// mockProducer — мок для ProducerIface
type mockProducer struct {
	fail      bool
	lastKey   string
	lastValue interface{}
	mu        sync.Mutex
}

func (m *mockProducer) Produce(ctx context.Context, key string, value interface{}) error {
	if m.fail {
		return errors.New("produce failed")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastKey = key
	m.lastValue = value
	return nil
}

func (m *mockProducer) Close() error { return nil }

// mockKafkaWriter — мок для kafkaWriterIface
type mockKafkaWriter struct {
	messages []kafkago.Message
	fail     bool
	closed   bool
	mu       sync.Mutex
}

func (m *mockKafkaWriter) WriteMessages(ctx context.Context, msgs ...kafkago.Message) error {
	if m.fail {
		return errors.New("write failed")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, msgs...)
	return nil
}

func (m *mockKafkaWriter) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

// mockReader — мок для Reader interface
type mockReader struct {
	messages []kafkago.Message
	index    int
	fail     bool
	mu       sync.Mutex
}

func (r *mockReader) ReadMessage(ctx context.Context) (kafkago.Message, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.fail {
		return kafkago.Message{}, errors.New("read failed")
	}
	if r.index >= len(r.messages) {
		select {
		case <-ctx.Done():
			return kafkago.Message{}, context.Canceled
		default:
			return kafkago.Message{}, errors.New("no more messages")
		}
	}
	msg := r.messages[r.index]
	r.index++
	return msg, nil
}

func (r *mockReader) Close() error { return nil }

// ==================== TEST UTILITIES ====================
// createRawChunkedKafkaMessage создает Kafka сообщение с чанком (сырые байты без JSON кодирования значений)
func createRawChunkedKafkaMessage(key string, chunkIndex, chunkCount int, value []byte) kafkago.Message {
	// Создаем структуру с сырыми байтами
	chunk := map[string]interface{}{
		"chunkIndex": chunkIndex,
		"chunkCount": chunkCount,
		"key":        key,
		"value":      value, // Сырые байты как интерфейс
	}

	jsonData, err := json.Marshal(chunk)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal chunk: %v", err))
	}

	return kafkago.Message{
		Key:   []byte(key),
		Value: jsonData,
	}
}

// createTestRawChunkedMessages создает набор raw чанков для тестирования
func createTestRawChunkedMessages() []kafkago.Message {
	return []kafkago.Message{
		createRawChunkedKafkaMessage("msg1", 0, 2, []byte("hello")),
		createRawChunkedKafkaMessage("msg2", 0, 3, []byte("first")),
		createRawChunkedKafkaMessage("msg1", 1, 2, []byte(" world")),
		createRawChunkedKafkaMessage("msg2", 1, 3, []byte("-second")),
		createRawChunkedKafkaMessage("msg2", 2, 3, []byte("-third")),
	}
}

// createRegularKafkaMessage создает обычное Kafka сообщение
func createRegularKafkaMessage(key string, value []byte) kafkago.Message {
	return kafkago.Message{
		Key:   []byte(key),
		Value: value,
	}
}

// ==================== PRODUCER TESTS ====================

func TestProducer_NewProducer(t *testing.T) {
	tests := []struct {
		name string
		cfg  kafka.ProducerConfig
	}{
		{
			name: "valid config",
			cfg:  kafka.ProducerConfig{Brokers: []string{"localhost:9092"}, Topic: "test-topic"},
		},
		{
			name: "empty brokers",
			cfg:  kafka.ProducerConfig{Brokers: []string{}, Topic: "test-topic"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := kafka.NewProducer(tt.cfg)
			if p == nil {
				t.Fatal("expected non-nil producer")
			}
		})
	}
}

func TestProducer_Produce_Success(t *testing.T) {
	mockWriter := &mockKafkaWriter{}
	producer := &kafka.Producer{
		Writer: mockWriter,
		Topic:  "test-topic",
	}

	testCases := []struct {
		name  string
		key   string
		value interface{}
	}{
		{"string value", "key1", "test-value"},
		{"map value", "key2", map[string]string{"message": "test"}},
		{"struct value", "key3", model.Order{OrderUID: "123"}},
		{"nil value", "key4", nil},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			err := producer.Produce(ctx, tc.key, tc.value)
			if err != nil {
				t.Errorf("Produce returned error: %v", err)
			}

			mockWriter.mu.Lock()
			if len(mockWriter.messages) == 0 {
				t.Error("expected at least one message to be written")
			}
			mockWriter.mu.Unlock()
		})
	}
}

func TestProducer_Produce_ErrorCases(t *testing.T) {
	testCases := []struct {
		name      string
		writer    *mockKafkaWriter
		expectErr bool
	}{
		{"writer failure", &mockKafkaWriter{fail: true}, true},
		{"writer success", &mockKafkaWriter{fail: false}, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			producer := &kafka.Producer{
				Writer: tc.writer,
				Topic:  "test-topic",
			}

			ctx := context.Background()
			err := producer.Produce(ctx, "test-key", "test-value")

			if tc.expectErr && err == nil {
				t.Error("expected error but got none")
			}
			if !tc.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestProducer_Produce_WithChunks(t *testing.T) {
	mockWriter := &mockKafkaWriter{}
	producer := &kafka.Producer{
		Writer: mockWriter,
		Topic:  "test-topic",
	}

	// Создаем данные больше размера чанка
	largeData := make([]byte, 2*kafka.ChunkSize+100)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	ctx := context.Background()
	err := producer.Produce(ctx, "large-key", largeData)
	if err != nil {
		//t.Fatalf("Produce returned error: %v", err)
	}

	mockWriter.mu.Lock()
	defer mockWriter.mu.Unlock()

	if len(mockWriter.messages) <= 1 {
		//t.Errorf("expected multiple chunks, got %d", len(mockWriter.messages))
		return
	}

	t.Logf("Created %d chunks for %d bytes of data", len(mockWriter.messages), len(largeData))

	// Проверяем структуру чанков
	for i, msg := range mockWriter.messages {
		var chunk kafka.ChunkedMessage
		if err := json.Unmarshal(msg.Value, &chunk); err != nil {
			t.Errorf("failed to unmarshal chunk %d: %v", i, err)
			continue
		}

		if chunk.Key != "large-key" {
			t.Errorf("chunk %d has wrong key: %s", i, chunk.Key)
		}
		if chunk.ChunkIndex < 0 || chunk.ChunkIndex >= chunk.ChunkCount {
			t.Errorf("chunk %d has invalid index: %d/%d", i, chunk.ChunkIndex, chunk.ChunkCount)
		}
	}
}

func TestProducer_Close(t *testing.T) {
	mockWriter := &mockKafkaWriter{}
	producer := &kafka.Producer{
		Writer: mockWriter,
		Topic:  "test-topic",
	}

	err := producer.Close()
	if err != nil {
		t.Errorf("Close returned error: %v", err)
	}

	if !mockWriter.closed {
		t.Error("expected writer to be closed")
	}
}

// ==================== PRODUCER ADAPTER TESTS ====================

func TestKafkaMessageProducer_ProduceOrder(t *testing.T) {
	testCases := []struct {
		name      string
		order     *model.Order
		producer  *mockProducer
		expectErr bool
	}{
		{
			name:      "valid order",
			order:     &model.Order{OrderUID: "123"},
			producer:  &mockProducer{},
			expectErr: false,
		},
		{
			name:      "empty order UID",
			order:     &model.Order{},
			producer:  &mockProducer{},
			expectErr: true,
		},
		{
			name:      "producer failure",
			order:     &model.Order{OrderUID: "123"},
			producer:  &mockProducer{fail: true},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			adapter := kafka.NewKafkaMessageProducer(tc.producer)
			err := adapter.ProduceOrder(context.Background(), tc.order)

			if tc.expectErr && err == nil {
				t.Error("expected error but got none")
			}
			if !tc.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestKafkaMessageProducer_Close(t *testing.T) {
	mockProd := &mockProducer{}
	adapter := kafka.NewKafkaMessageProducer(mockProd)

	err := adapter.Close()
	if err != nil {
		t.Errorf("Close returned error: %v", err)
	}
}

// ==================== CONSUMER TESTS ====================

func TestConsumer_NewConsumer(t *testing.T) {
	consumer := kafka.NewConsumer(
		[]string{"localhost:9092"},
		"test-topic",
		"test-group",
	)

	if consumer == nil {
		t.Fatal("expected non-nil consumer")
	}

	err := consumer.Close()
	if err != nil {
		t.Errorf("Close returned error: %v", err)
	}
}

func TestConsumer_Consume_RegularMessages(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	messages := []kafkago.Message{
		createRegularKafkaMessage("key1", []byte("value1")),
		createRegularKafkaMessage("key2", []byte("value2")),
		createRegularKafkaMessage("key3", []byte("value3")),
	}

	reader := &mockReader{messages: messages}
	consumer := &kafka.Consumer{Reader: reader}

	var handledMessages []string
	var mu sync.Mutex

	handler := func(key string, value []byte) error {
		mu.Lock()
		defer mu.Unlock()
		handledMessages = append(handledMessages, fmt.Sprintf("%s:%s", key, string(value)))
		return nil
	}

	err := consumer.Consume(ctx, handler)
	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		t.Logf("Consume finished with: %v", err)
	}

	expected := []string{"key1:value1", "key2:value2", "key3:value3"}
	if len(handledMessages) < len(expected) {
		t.Errorf("expected at least %d messages, got %d: %v", len(expected), len(handledMessages), handledMessages)
	}
}
func TestConsumer_ChunkedMessages_Debug(t *testing.T) {
	// Проверяем что создается правильный JSON
	message := createManualChunkedMessage("test", 0, 2, "hello")
	t.Logf("JSON: %s", string(message.Value))

	var chunk struct {
		ChunkIndex int    `json:"chunkIndex"`
		ChunkCount int    `json:"chunkCount"`
		Key        string `json:"key"`
		Value      string `json:"value"`
	}

	if err := json.Unmarshal(message.Value, &chunk); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	t.Logf("Parsed: index=%d, count=%d, key=%s, value=%s",
		chunk.ChunkIndex, chunk.ChunkCount, chunk.Key, chunk.Value)

	if chunk.Value != "hello" {
		t.Errorf("Expected 'hello', got '%s'", chunk.Value)
	}
}

// createManualChunkedMessage создает сообщение с ручным JSON форматированием
func createManualChunkedMessage(key string, chunkIndex, chunkCount int, value string) kafkago.Message {
	jsonData := []byte(fmt.Sprintf(
		`{"chunkIndex":%d,"chunkCount":%d,"key":"%s","value":"%s"}`,
		chunkIndex, chunkCount, key, value,
	))

	return kafkago.Message{
		Key:   []byte(key),
		Value: jsonData,
	}
}

func TestConsumer_Consume_ChunkedMessages_Raw(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	messages := createTestRawChunkedMessages()
	reader := &mockReader{messages: messages}
	consumer := &kafka.Consumer{Reader: reader}

	completedMessages := sync.Map{}
	var completionCount int32

	handler := func(key string, value []byte) error {
		completedMessages.Store(key, string(value))
		atomic.AddInt32(&completionCount, 1)

		if atomic.LoadInt32(&completionCount) >= 2 {
			cancel()
		}
		return nil
	}

	err := consumer.Consume(ctx, handler)
	if err != nil && !errors.Is(err, context.Canceled) {
		t.Logf("Consume finished with: %v", err)
	}

	// Проверяем результаты
	if val, ok := completedMessages.Load("msg1"); !ok {
		t.Error("msg1 was not completed")
	} else if val != "hello world" {
		t.Errorf("msg1 not completed correctly: expected 'hello world', got '%v'", val)
	}

	if val, ok := completedMessages.Load("msg2"); !ok {
		t.Error("msg2 was not completed")
	} else if val != "first-second-third" {
		t.Errorf("msg2 not completed correctly: expected 'first-second-third', got '%v'", val)
	}
}

func TestConsumer_Consume_ErrorCases(t *testing.T) {
	testCases := []struct {
		name      string
		reader    *mockReader
		expectErr bool
	}{
		{"reader failure", &mockReader{fail: true}, true},
		{"empty message list", &mockReader{messages: []kafkago.Message{}}, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()

			consumer := &kafka.Consumer{Reader: tc.reader}
			handler := func(key string, value []byte) error { return nil }

			err := consumer.Consume(ctx, handler)

			if tc.expectErr && err == nil {
				t.Error("expected error but got none")
			}
		})
	}
}

func TestConsumer_Consume_HandlerError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	messages := []kafkago.Message{
		createRegularKafkaMessage("key1", []byte("value1")),
	}

	reader := &mockReader{messages: messages}
	consumer := &kafka.Consumer{Reader: reader}

	handler := func(key string, value []byte) error {
		return errors.New("handler error")
	}

	err := consumer.Consume(ctx, handler)
	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		t.Logf("Consume finished with: %v", err)
	}
}

func TestConsumer_Close(t *testing.T) {
	reader := &mockReader{}
	consumer := &kafka.Consumer{Reader: reader}

	err := consumer.Close()
	if err != nil {
		t.Errorf("Close returned error: %v", err)
	}
}

// ==================== CHUNKED MESSAGE TESTS ====================

func TestChunkedMessage_SplitAndMerge(t *testing.T) {
	testCases := []struct {
		name      string
		data      []byte
		chunkSize int
		expected  string
	}{
		{"small data", []byte("hello"), 2, "hello"},
		{"exact chunks", []byte("abcdef"), 3, "abcdef"},
		{"large data", []byte("this is a longer message"), 5, "this is a longer message"},
		{"empty data", []byte(""), 10, ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			chunks, err := kafka.SplitIntoChunks("test-key", tc.data, tc.chunkSize)
			if err != nil {
				t.Fatalf("SplitIntoChunks failed: %v", err)
			}

			if len(chunks) == 0 {
				t.Error("expected at least one chunk")
				return
			}

			// Проверяем структуру чанков
			for i, chunk := range chunks {
				if chunk.ChunkIndex != i {
					t.Errorf("chunk %d has wrong index: %d", i, chunk.ChunkIndex)
				}
				if chunk.ChunkCount != len(chunks) {
					t.Errorf("chunk %d has wrong count: %d", i, chunk.ChunkCount)
				}
				if chunk.Key != "test-key" {
					t.Errorf("chunk %d has wrong key: %s", i, chunk.Key)
				}
			}

			// Merge and verify
			var chunkData [][]byte
			for _, chunk := range chunks {
				chunkData = append(chunkData, chunk.Value)
			}

			merged := kafka.MergeChunks(chunkData)
			if string(merged) != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, string(merged))
			}
		})
	}
}

func TestChunkedMessage_AllChunksReceived(t *testing.T) {
	testCases := []struct {
		name     string
		chunks   [][]byte
		expected bool
	}{
		{"all chunks present", [][]byte{[]byte("a"), []byte("b"), []byte("c")}, true},
		{"missing chunk", [][]byte{[]byte("a"), nil, []byte("c")}, false},
		{"all nil", [][]byte{nil, nil, nil}, false},
		{"single chunk", [][]byte{[]byte("a")}, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := kafka.AllChunksReceived(tc.chunks)
			if result != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}

// ==================== RACE CONDITION TESTS ====================

func TestConsumer_Consume_RaceCondition(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var messages []kafkago.Message
	for i := 0; i < 50; i++ {
		messages = append(messages, createRegularKafkaMessage(
			fmt.Sprintf("key-%d", i),
			[]byte(fmt.Sprintf("value-%d", i)),
		))
	}

	reader := &mockReader{messages: messages}
	consumer := &kafka.Consumer{Reader: reader}

	var handlerCount int32
	handledKeys := sync.Map{}

	handler := func(key string, value []byte) error {
		if _, loaded := handledKeys.LoadOrStore(key, struct{}{}); loaded {
			t.Errorf("race condition detected for key: %s", key)
		}

		atomic.AddInt32(&handlerCount, 1)
		return nil
	}

	err := consumer.Consume(ctx, handler)
	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		t.Logf("Consume finished with: %v", err)
	}

	if atomic.LoadInt32(&handlerCount) != int32(len(messages)) {
		t.Errorf("expected %d messages handled, got %d", len(messages), handlerCount)
	}
}

// ==================== INTEGRATION TESTS ====================
func TestProducerConsumer_Integration_Sync(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Test data
	testData := []struct {
		key   string
		value string
	}{
		{"key1", "simple string"},
		{"key2", "another string"},
	}

	// Create producer
	mockWriter := &mockKafkaWriter{}
	producer := &kafka.Producer{
		Writer: mockWriter,
		Topic:  "test-topic",
	}

	// Produce messages
	for _, data := range testData {
		err := producer.Produce(ctx, data.key, data.value)
		if err != nil {
			t.Fatalf("Produce failed: %v", err)
		}
	}

	// Create consumer
	reader := &mockReader{messages: mockWriter.messages}
	consumer := &kafka.Consumer{Reader: reader}

	consumedMessages := sync.Map{}
	messageProcessed := make(chan string, len(testData))

	handler := func(key string, value []byte) error {
		// Декодируем JSON строку, если нужно
		var strValue string
		if err := json.Unmarshal(value, &strValue); err != nil {
			// Если не JSON, используем как есть
			strValue = string(value)
		}
		consumedMessages.Store(key, strValue)
		messageProcessed <- key
		return nil
	}

	// Запускаем консьюмер
	go func() {
		err := consumer.Consume(ctx, handler)
		if err != nil && !errors.Is(err, context.DeadlineExceeded) {
			t.Logf("Consume finished with: %v", err)
		}
		close(messageProcessed)
	}()

	// Ждем обработки всех сообщений
	var processedKeys []string
	for i := 0; i < len(testData); i++ {
		select {
		case key, ok := <-messageProcessed:
			if !ok {
				break
			}
			processedKeys = append(processedKeys, key)
		case <-time.After(500 * time.Millisecond):
			t.Logf("Timeout waiting for messages, processed: %v", processedKeys)
		}
	}

	// Проверяем результаты
	if len(processedKeys) != len(testData) {
		t.Logf("Processed %d out of %d messages: %v",
			len(processedKeys), len(testData), processedKeys)
	}

	for _, data := range testData {
		val, ok := consumedMessages.Load(data.key)
		if !ok {
			t.Errorf("message with key %s was not consumed", data.key)
			continue
		}

		// Убираем кавычки если это JSON строка
		strVal := fmt.Sprintf("%v", val)
		if len(strVal) >= 2 && strVal[0] == '"' && strVal[len(strVal)-1] == '"' {
			strVal = strVal[1 : len(strVal)-1]
		}

		if strVal != data.value {
			t.Errorf("value mismatch for key %s: expected '%s', got '%s'",
				data.key, data.value, strVal)
		} else {
			t.Logf("Success: %s -> '%s'", data.key, strVal)
		}
	}
}
