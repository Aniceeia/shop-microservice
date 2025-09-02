package unit

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"shop-microservice/internal/api/handlers"
	"shop-microservice/internal/application/usecases"
	"shop-microservice/internal/domain/model"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type hcMockUseCase struct{ mock.Mock }

func (m *hcMockUseCase) CreateOrder(ctx context.Context, order *model.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *hcMockUseCase) GetOrderByID(ctx context.Context, id string) (*model.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Order), args.Error(1)
}

func (m *hcMockUseCase) GetAllOrders(ctx context.Context) ([]*model.Order, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Order), args.Error(1)
}

func (m *hcMockUseCase) ValidateOrder(order *model.Order) error { return nil }

func (m *hcMockUseCase) HealthCheck(ctx context.Context) (map[string]any, error) {
	args := m.Called(ctx)
	return args.Get(0).(map[string]any), args.Error(1)
}

func (m *hcMockUseCase) Shutdown() { m.Called() }

func (m *hcMockUseCase) GetMetrics() map[string]any {
	args := m.Called()
	return args.Get(0).(map[string]any)
}

func makeOrder() model.Order {
	return model.Order{
		OrderUID:    "test123",
		TrackNumber: "WBILMTESTTRACK",
		Entry:       "WBIL",
		CustomerID:  "test",
		DateCreated: time.Now(),
		Delivery:    model.Delivery{Name: "Test Testov", Phone: "+9720000000", Zip: "2639809", City: "Kiryat Mozkin", Address: "Ploshad Mira 15", Region: "Kraiot", Email: "test@gmail.com"},
		Payment:     model.Payment{Transaction: "test123", Currency: "USD", Provider: "wbpay", Amount: 1817, PaymentDt: time.Now().Unix(), Bank: "alpha", DeliveryCost: 1500, GoodsTotal: 317, CustomFee: 0},
		Items:       []model.Item{{ChrtID: 9934930, TrackNumber: "WBILMTESTTRACK", Price: 453, Rid: "ab4219087a764ae0btest", Name: "Mascaras", Sale: 30, Size: "0", TotalPrice: 317, NmID: 2389212, Brand: "Vivienne Sabo", Status: 202}},
	}
}

func TestHandler_CreateOrder_ValidationError(t *testing.T) {
	m := new(hcMockUseCase)
	h := handlers.NewHandler(m)
	m.On("CreateOrder", mock.Anything, mock.Anything).Return(usecases.ErrValidation)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	ord := makeOrder()
	buf, _ := json.Marshal(ord)
	c.Request = httptest.NewRequest("POST", "/api/orders", bytes.NewReader(buf))
	c.Request.Header.Set("Content-Type", "application/json")

	h.CreateOrder(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_CreateOrder_InternalError(t *testing.T) {
	m := new(hcMockUseCase)
	h := handlers.NewHandler(m)
	m.On("CreateOrder", mock.Anything, mock.Anything).Return(usecases.ErrServerClosed)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	ord := makeOrder()
	buf, _ := json.Marshal(ord)
	c.Request = httptest.NewRequest("POST", "/api/orders", bytes.NewReader(buf))
	c.Request.Header.Set("Content-Type", "application/json")

	h.CreateOrder(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandler_GetAllOrders_Error(t *testing.T) {
	m := new(hcMockUseCase)
	h := handlers.NewHandler(m)
	m.On("GetAllOrders", mock.Anything).Return(nil, errors.New("db"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/orders", nil)

	h.GetAllOrders(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandler_HealthCheck_ServiceUnavailable(t *testing.T) {
	m := new(hcMockUseCase)
	h := handlers.NewHandler(m)
	m.On("HealthCheck", mock.Anything).Return(map[string]any{"status": "unhealthy"}, errors.New("x"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/health", nil)

	h.HealthCheck(c)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestHandler_GetMetrics(t *testing.T) {
	m := new(hcMockUseCase)
	h := handlers.NewHandler(m)

	expectedMetrics := map[string]interface{}{
		"errors":             map[string]interface{}{"/test": float64(1)},
		"requests":           map[string]interface{}{"/test": float64(5)},
		"cache_hits":         float64(0),
		"cache_misses":       float64(0),
		"db_connections":     float64(0),
		"kafka_messages":     float64(0),
		"avg_response_times": map[string]interface{}{},
	}
	m.On("GetMetrics").Return(expectedMetrics)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/metrics", nil)

	h.GetMetrics(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedMetrics, response)

	m.AssertExpectations(t)
}
