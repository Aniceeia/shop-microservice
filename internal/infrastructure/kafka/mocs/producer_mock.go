package mocks

import (
	"context"
)

type MockProducer struct {
	ProduceFunc func(ctx context.Context, key string, value interface{}) error
	CloseFunc   func() error
}

func (m *MockProducer) Produce(ctx context.Context, key string, value interface{}) error {
	if m.ProduceFunc != nil {
		return m.ProduceFunc(ctx, key, value)
	}
	return nil
}

func (m *MockProducer) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

// NewMockProducer создает mock producer для тестов
func NewMockProducer() *MockProducer {
	return &MockProducer{
		ProduceFunc: func(ctx context.Context, key string, value interface{}) error {
			return nil // По умолчанию успех
		},
		CloseFunc: func() error {
			return nil
		},
	}
}
