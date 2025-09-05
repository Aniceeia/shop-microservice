package kafka

import (
	"context"
	"fmt"
	"shop-microservice/internal/domain/model"
)

type KafkaMessageProducer struct {
	producer ProducerIface
}

func NewKafkaMessageProducer(producer ProducerIface) KafkaMessageProducerIface {
	return &KafkaMessageProducer{producer: producer}
}

func (kp *KafkaMessageProducer) ProduceOrder(ctx context.Context, order *model.Order) error {
	if order.OrderUID == "" {
		return fmt.Errorf("empty order UID")
	}
	return kp.producer.Produce(ctx, order.OrderUID, order)
}

func (kp *KafkaMessageProducer) Close() error {
	return kp.producer.Close()
}

// KafkaMessageProducerIface — интерфейс для адаптера
type KafkaMessageProducerIface interface {
	ProduceOrder(ctx context.Context, order *model.Order) error
	Close() error
}
