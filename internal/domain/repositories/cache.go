package repositories

import (
	"shop-microservice/internal/domain/model"
)

type Cache interface {
	Set(uid string, order *model.Order)
	Get(uid string) (*model.Order, bool)
	GetAll() []*model.Order
	Delete(uid string)
	Size() int
	Clear()
	WarmUp(repo OrderRepository) error
}
