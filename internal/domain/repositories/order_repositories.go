package repositories

import "shop-microservice/internal/domain/model"

type OrderRepository interface {
	Save(order *model.Order) error
	FindByID(uid string) (*model.Order, error)
	FindAll() ([]*model.Order, error) // Для восстановления кеша
}
