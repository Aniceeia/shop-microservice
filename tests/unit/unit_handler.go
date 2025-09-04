package unit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"shop-microservice/internal/api/handlers"
	"shop-microservice/internal/domain/model"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockOrderUseCase struct {
	mock.Mock
}

func (m *MockOrderUseCase) GetMetrics() map[string]interface{} {
	args := m.Called()
	return args.Get(0).(map[string]interface{})
}

func (m *MockOrderUseCase) CreateOrder(ctx context.Context, order *model.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderUseCase) GetOrderByID(ctx context.Context, id string) (*model.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Order), args.Error(1)
}

func (m *MockOrderUseCase) GetAllOrders(ctx context.Context) ([]*model.Order, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*model.Order), args.Error(1)
}

func (m *MockOrderUseCase) ValidateOrder(order *model.Order) error {
	args := m.Called(order)
	return args.Error(0)
}

func (m *MockOrderUseCase) HealthCheck(ctx context.Context) (map[string]interface{}, error) {
	args := m.Called(ctx)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockOrderUseCase) Shutdown() {
	m.Called()
}

func TestHandler_CreateOrder_Success(t *testing.T) {
	// Setup
	mockUseCase := new(MockOrderUseCase)
	handler := handlers.NewHandler(mockUseCase)

	// Test data
	order := model.Order{
		OrderUID:    "test123",
		TrackNumber: "WBILMTESTTRACK",
		Entry:       "WBIL",
		CustomerID:  "test",
		DateCreated: time.Now(),
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
			Transaction:  "test123",
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
				TrackNumber: "WBILMTESTTRACK",
				Price:       453,
				Rid:         "ab4219087a764ae0btest",
				Name:        "Mascaras",
				Sale:        30,
				Size:        "0",
				TotalPrice:  317,
				NmID:        2389212,
				Brand:       "Vivienne Sabo",
				Status:      202,
			},
		},
	}

	// Expectations
	mockUseCase.On("ValidateOrder", mock.Anything).Return(nil)
	mockUseCase.On("CreateOrder", mock.Anything, mock.Anything).Return(nil)

	// Execute
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	jsonData, _ := json.Marshal(order)
	c.Request = httptest.NewRequest("POST", "/orders", bytes.NewReader(jsonData))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CreateOrder(c)

	// Verify
	assert.Equal(t, http.StatusCreated, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestHandler_GetOrderByID_Success(t *testing.T) {
	mockUseCase := new(MockOrderUseCase)
	handler := handlers.NewHandler(mockUseCase)

	expectedOrder := &model.Order{
		OrderUID:    "test123",
		TrackNumber: "WBILMTESTTRACK",
	}

	mockUseCase.On("GetOrderByID", mock.Anything, "test123").Return(expectedOrder, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "id", Value: "test123"}}
	c.Request = httptest.NewRequest("GET", "/orders/test123", nil)

	handler.GetOrderByID(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestHandler_GetOrderByID_NotFound(t *testing.T) {
	mockUseCase := new(MockOrderUseCase)
	handler := handlers.NewHandler(mockUseCase)

	mockUseCase.On("GetOrderByID", mock.Anything, "nonexistent").Return(nil, fmt.Errorf("user not foud"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "id", Value: "nonexistent"}}
	c.Request = httptest.NewRequest("GET", "/orders/nonexistent", nil)

	handler.GetOrderByID(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockUseCase.AssertExpectations(t)
}
