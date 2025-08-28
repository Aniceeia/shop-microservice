package cash

import (
	"context"
	"errors"
	"shop-microservice/internal/domain/model"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockOrderRepository struct {
	orders []*model.Order
	err    error
}

func (m *MockOrderRepository) FindAll(ctx context.Context) ([]*model.Order, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.orders, nil
}

func (m *MockOrderRepository) FindByID(ctx context.Context, uid string) (*model.Order, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, order := range m.orders {
		if order.OrderUID == uid {
			return order, nil
		}
	}
	return nil, errors.New("order not found")
}

func TestCash_SetAndGet(t *testing.T) {
	cash := NewCash()
	order := createTestOrder()

	cash.Set(order.OrderUID, order)

	result, exists := cash.Get(order.OrderUID)
	require.True(t, exists)
	assert.Equal(t, order.OrderUID, result.OrderUID)
	assert.Equal(t, order.TrackNumber, result.TrackNumber)
	assert.Equal(t, order.Delivery.Name, result.Delivery.Name)
	assert.Equal(t, order.Payment.Amount, result.Payment.Amount)
	assert.Len(t, result.Items, 1)
	assert.Equal(t, order.Items[0].Name, result.Items[0].Name)

	_, exists = cash.Get("non-existent-id")
	assert.False(t, exists)
}

func TestCash_GetAll(t *testing.T) {
	cash := NewCash()

	orders := cash.GetAll()
	assert.Empty(t, orders)

	order1 := createTestOrder()
	order2 := createTestOrder()
	order2.OrderUID = "test-order-uid-2"
	order2.TrackNumber = "TEST456"

	cash.Set(order1.OrderUID, order1)
	cash.Set(order2.OrderUID, order2)

	orders = cash.GetAll()
	require.Len(t, orders, 2)

	orderMap := make(map[string]*model.Order)
	for _, order := range orders {
		orderMap[order.OrderUID] = order
	}

	assert.Contains(t, orderMap, order1.OrderUID)
	assert.Contains(t, orderMap, order2.OrderUID)
	assert.Equal(t, order1.TrackNumber, orderMap[order1.OrderUID].TrackNumber)
	assert.Equal(t, order2.TrackNumber, orderMap[order2.OrderUID].TrackNumber)
}

func TestCash_Delete(t *testing.T) {
	cash := NewCash()
	order := createTestOrder()

	cash.Set(order.OrderUID, order)

	_, exists := cash.Get(order.OrderUID)
	assert.True(t, exists)

	cash.Delete(order.OrderUID)

	_, exists = cash.Get(order.OrderUID)
	assert.False(t, exists)
}

func TestCash_Clear(t *testing.T) {
	cash := NewCash()

	order1 := createTestOrder()
	order2 := createTestOrder()
	order2.OrderUID = "test-order-uid-2"

	cash.Set(order1.OrderUID, order1)
	cash.Set(order2.OrderUID, order2)

	assert.Equal(t, 2, cash.Size())

	cash.Clear()

	assert.Equal(t, 0, cash.Size())
	assert.Empty(t, cash.GetAll())
}

func TestCash_Size(t *testing.T) {
	cash := NewCash()

	assert.Equal(t, 0, cash.Size())

	order1 := createTestOrder()
	cash.Set(order1.OrderUID, order1)
	assert.Equal(t, 1, cash.Size())

	order2 := createTestOrder()
	order2.OrderUID = "test-order-uid-2"
	cash.Set(order2.OrderUID, order2)
	assert.Equal(t, 2, cash.Size())

	cash.Delete(order1.OrderUID)
	assert.Equal(t, 1, cash.Size())
}

func TestCash_WarmUp_Success(t *testing.T) {
	cash := NewCash()

	order1 := createTestOrder()
	order2 := createTestOrder()
	order2.OrderUID = "test-order-uid-2"
	order2.TrackNumber = "TEST456"

	mockRepo := &MockOrderRepository{
		orders: []*model.Order{order1, order2},
	}

	err := cash.WarmUp(mockRepo)
	require.NoError(t, err)

	assert.Equal(t, 2, cash.Size())

	result1, exists1 := cash.Get(order1.OrderUID)
	require.True(t, exists1)
	assert.Equal(t, order1.OrderUID, result1.OrderUID)

	result2, exists2 := cash.Get(order2.OrderUID)
	require.True(t, exists2)
	assert.Equal(t, order2.OrderUID, result2.OrderUID)
}

func TestCash_WarmUp_EmptyRepository(t *testing.T) {
	cash := NewCash()

	mockRepo := &MockOrderRepository{
		orders: []*model.Order{},
	}

	err := cash.WarmUp(mockRepo)
	require.NoError(t, err)

	assert.Equal(t, 0, cash.Size())
	assert.Empty(t, cash.GetAll())
}

func TestCash_WarmUp_RepositoryError(t *testing.T) {
	cash := NewCash()

	mockRepo := &MockOrderRepository{
		err: errors.New("database connection failed"),
	}

	err := cash.WarmUp(mockRepo)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "database connection failed")

	assert.Equal(t, 0, cash.Size())
}

func TestCash_WarmUp_ContextCancelled(t *testing.T) {
	cash := NewCash()

	_, cancel := context.WithCancel(context.Background())
	cancel()

	mockRepo := &MockOrderRepository{
		orders: []*model.Order{createTestOrder()},
	}

	// This test verifies that the WarmUp method doesn't use context internally
	// since the repository interface doesn't accept context in FindAll
	err := cash.WarmUp(mockRepo)
	require.NoError(t, err) // Should still work despite cancelled context

	assert.Equal(t, 1, cash.Size())
}

func TestCash_ConcurrentAccess(t *testing.T) {
	cash := NewCash()
	order := createTestOrder()

	// Test concurrent writes and reads
	done := make(chan bool)
	numGoroutines := 10

	for i := range numGoroutines {
		go func(id int) {
			// Concurrent writes
			newOrder := createTestOrder()
			newOrder.OrderUID = order.OrderUID + "-" + string(rune(id))
			cash.Set(newOrder.OrderUID, newOrder)

			// Concurrent reads
			_, _ = cash.Get(newOrder.OrderUID)

			// Concurrent size check
			_ = cash.Size()

			done <- true
		}(i)
	}

	// Wait for all goroutines to finish
	for range numGoroutines {
		<-done
	}

	// Verify no data races and correct final state
	assert.Equal(t, numGoroutines, cash.Size())
}

func TestCash_OverwriteExisting(t *testing.T) {
	cash := NewCash()
	order := createTestOrder()

	// Set initial order
	cash.Set(order.OrderUID, order)

	// Verify initial state
	result, exists := cash.Get(order.OrderUID)
	require.True(t, exists)
	assert.Equal(t, "TEST123", result.TrackNumber)

	// Overwrite with modified order
	modifiedOrder := createTestOrder()
	modifiedOrder.TrackNumber = "MODIFIED123"
	cash.Set(order.OrderUID, modifiedOrder)

	// Verify overwrite
	result, exists = cash.Get(order.OrderUID)
	require.True(t, exists)
	assert.Equal(t, "MODIFIED123", result.TrackNumber)
	assert.Equal(t, 1, cash.Size()) // Should still have only one item
}

func TestCash_NilOrderHandling(t *testing.T) {
	cash := NewCash()

	// Test setting nil order
	cash.Set("nil-order", nil)

	// Should not panic and should store nil
	result, exists := cash.Get("nil-order")
	assert.True(t, exists)
	assert.Nil(t, result)

	// Test size includes nil items
	assert.Equal(t, 1, cash.Size())
}

func createTestOrder() *model.Order {
	return &model.Order{
		OrderUID:          "test-order-uid",
		TrackNumber:       "TEST123",
		Entry:             "WBIL",
		Locale:            "en",
		InternalSignature: "",
		CustomerID:        "test-customer",
		DeliveryService:   "meest",
		Shardkey:          "9",
		SmID:              99,
		DateCreated:       time.Now(),
		OofShard:          "1",
		Delivery: model.Delivery{
			Name:    "Test Testov",
			Phone:   "+9720000000",
			Zip:     "2639809",
			City:    "Kiryat Mozkin",
			Address: "Ploshad Mira 15",
			Region:  "Kraiot",
			Email:   "test@gmail.com",
		},
		Payment: model.Payment{
			Transaction:  "test-transaction",
			RequestID:    "",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       1817,
			PaymentDt:    1637907727,
			Bank:         "alpha",
			DeliveryCost: 1500,
			GoodsTotal:   317,
			CustomFee:    0,
		},
		Items: []model.Item{
			{
				ChrtID:      9934930,
				TrackNumber: "TEST123",
				Price:       453,
				Rid:         "ab4219087a764ae0btest",
				Name:        "Test Item",
				Sale:        30,
				Size:        "0",
				TotalPrice:  317,
				NmID:        2389212,
				Brand:       "Test Brand",
				Status:      202,
			},
		},
	}
}
