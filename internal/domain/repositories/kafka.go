package repositories

import (
	"context"
	"shop-microservice/internal/domain/model"
)

type MessageProducer interface {
	ProduceOrder(ctx context.Context, order *model.Order) error
	Close() error
}
