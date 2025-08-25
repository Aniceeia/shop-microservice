package repositories

import (
	"errors"
	"shop-microservice/internal/domain/model"
	"sync"
)

type MockOrderRepository struct {
	mu     sync.RWMutex
	orders map[string]*model.Order
}

func NewMockOrderRepository() *MockOrderRepository {
	return &MockOrderRepository{
		orders: make(map[string]*model.Order),
	}
}

func (m *MockOrderRepository) Save(order *model.Order) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.orders[order.OrderUID] = order
	return nil
}

func (m *MockOrderRepository) FindByID(uid string) (*model.Order, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	order, exists := m.orders[uid]
	if !exists {
		return nil, ErrOrderNotFound
	}
	return order, nil
}

func (m *MockOrderRepository) FindAll() ([]*model.Order, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	orders := make([]*model.Order, 0, len(m.orders))
	for _, order := range m.orders {
		orders = append(orders, order)
	}
	return orders, nil
}

var ErrOrderNotFound = errors.New("order not found")
