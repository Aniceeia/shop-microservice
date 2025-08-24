package repositories

import "shop-microservice/domain/entities"

type OrderRepository interface {
	Save(order *entities.Order) error
	FindByID(uid string) (*entities.Order, error)
	// GetAll() ([]*entities.Order, error) // Для восстановления кеша
}
