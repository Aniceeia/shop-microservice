package repositories

import (
	"shop-microservice/internal/domain/model"
)

type Cache interface {
	PreloadCacheFromDB(repo OrderRepository) error
	Set(uid string, order *model.Order)
	Get(uid string) (*model.Order, bool)
	GetAll() []*model.Order
	Delete(uid string)
	Size() int
	Clear()
}
