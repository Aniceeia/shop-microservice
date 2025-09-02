package unit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"shop-microservice/internal/api"
	"shop-microservice/internal/api/handlers"
	"shop-microservice/internal/domain/model"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockOrderUseCase struct{ mock.Mock }

func (m *mockOrderUseCase) CreateOrder(ctx context.Context, order *model.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *mockOrderUseCase) GetOrderByID(ctx context.Context, id string) (*model.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Order), args.Error(1)
}

func (m *mockOrderUseCase) GetAllOrders(ctx context.Context) ([]*model.Order, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Order), args.Error(1)
}

func (m *mockOrderUseCase) ValidateOrder(order *model.Order) error { return nil }

func (m *mockOrderUseCase) HealthCheck(ctx context.Context) (map[string]interface{}, error) {
	args := m.Called(ctx)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *mockOrderUseCase) Shutdown() { m.Called() }

func (m *mockOrderUseCase) GetMetrics() map[string]interface{} {
	args := m.Called()
	return args.Get(0).(map[string]interface{})
}

func TestRouter_Endpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUC := new(mockOrderUseCase)
	mockUC.On("HealthCheck", mock.Anything).Return(map[string]interface{}{
		"status": "healthy",
	}, nil)
	handler := handlers.NewHandler(mockUC)
	router := api.SetupRouter(handler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRouter_StaticFiles(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUC := new(mockOrderUseCase)
	handler := handlers.NewHandler(mockUC)
	router := api.SetupRouter(handler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
