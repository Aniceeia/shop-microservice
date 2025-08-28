package repositories

import (
	"context"
	"shop-microservice/internal/domain/model"
)

type OrderRepository interface {
	Save(ctx context.Context, order *model.Order) error
	FindByID(ctx context.Context, uid string) (*model.Order, error)
	FindAll(ctx context.Context) ([]*model.Order, error)
}
