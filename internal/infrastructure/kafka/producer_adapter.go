package kafka

import (
	"context"
	"shop-microservice/internal/domain/model"
	"shop-microservice/internal/domain/repositories"
)

type KafkaMessageProducer struct {
	producer *Producer
}

func NewKafkaMessageProducer(producer *Producer) repositories.MessageProducer {
	return &KafkaMessageProducer{producer: producer}
}

func (kp *KafkaMessageProducer) ProduceOrder(ctx context.Context, order *model.Order) error {
	return kp.producer.Produce(ctx, order.OrderUID, order)
}
