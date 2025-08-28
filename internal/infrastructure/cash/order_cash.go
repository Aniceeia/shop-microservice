package cash

import (
	"context"
	"log"
	"shop-microservice/internal/domain/model"
	"sync"
	"time"
)

type Cash struct {
	mu     sync.RWMutex
	memory map[string]*model.Order
}

func NewCash() *Cash {
	return &Cash{
		memory: make(map[string]*model.Order),
	}
}

func (cash *Cash) Set(uid string, order *model.Order) {
	cash.mu.Lock()
	defer cash.mu.Unlock()
	cash.memory[uid] = order
}

func (cash *Cash) Get(uid string) (*model.Order, bool) {
	cash.mu.RLock()
	defer cash.mu.RUnlock()

	order, exists := cash.memory[uid]
	return order, exists
}

func (cash *Cash) GetAll() []*model.Order {
	cash.mu.RLock()
	defer cash.mu.RUnlock()

	orders := make([]*model.Order, 0, len(cash.memory))
	for _, order := range cash.memory {
		orders = append(orders, order)
	}
	return orders
}

func (cash *Cash) Delete(uid string) {
	cash.mu.Lock()
	defer cash.mu.Unlock()
	delete(cash.memory, uid)
}

func (cash *Cash) Size() int {
	cash.mu.RLock()
	defer cash.mu.RUnlock()
	return len(cash.memory)
}

func (cash *Cash) Clear() {
	cash.mu.Lock()
	defer cash.mu.Unlock()
	cash.memory = make(map[string]*model.Order)
}

// WarmUp заполняет кэш данными из репозитория при старте сервиса
func (cash *Cash) WarmUp(repo OrderRepository) error {
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orders, err := repo.FindAll(ctx)
	if err != nil {
		return err
	}

	cash.mu.Lock()
	defer cash.mu.Unlock()

	// Очищаем текущий кэш перед заполнением
	cash.memory = make(map[string]*model.Order)

	for _, order := range orders {
		if order != nil {
			cash.memory[order.OrderUID] = order
		}
	}

	log.Printf("Cache warm-up completed. Loaded %d orders in %v", len(orders), time.Since(start))
	return nil
}

// OrderRepository интерфейс для доступа к данным заказов
type OrderRepository interface {
	FindAll(ctx context.Context) ([]*model.Order, error)
	FindByID(ctx context.Context, uid string) (*model.Order, error)
}
