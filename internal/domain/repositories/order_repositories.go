package repositories

import (
	"context"
	"errors"
	"shop-microservice/internal/domain/model"
)

var (
	ErrOrderNotFound = errors.New("not found")
)

type OrderRepository interface {
	Save(ctx context.Context, order *model.Order) error
	FindByID(ctx context.Context, uid string) (*model.Order, error)
	FindAll(ctx context.Context) ([]*model.Order, error)
}
