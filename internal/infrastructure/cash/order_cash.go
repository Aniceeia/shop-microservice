package cash

import (
	"shop-microservice/internal/domain/model"
	"sync"
)

type Cash struct {
	mu     sync.RWMutex
	memory map[string]*model.Order
}

func (cash *Cash) Set(uid string, order *model.Order) {
	// wright to map
}

func (cash *Cash) Get(uid string) {
	// get from map
}
