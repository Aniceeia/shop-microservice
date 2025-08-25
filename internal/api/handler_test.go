package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"shop-microservice/internal/domain/model"
	"shop-microservice/internal/domain/repositories"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestHandler() (*Handler, *repositories.MockOrderRepository) {
	mockRepo := repositories.NewMockOrderRepository()
	handler := NewHandler(mockRepo)
	return handler, mockRepo
}

func TestCreateOrderHandler(t *testing.T) {
	handler, mockRepo := setupTestHandler()
	router := gin.Default()
	router.POST("/api/orders", handler.CreateOrder)

	// Тестовые данные
	testOrder := model.Order{
		OrderUID:    "test123",
		TrackNumber: "WBILMTESTTRACK",
		// ... остальные поля
	}

	jsonData, _ := json.Marshal(testOrder)

	req := httptest.NewRequest("POST", "/api/orders", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	// Проверяем что заказ сохранился в моке
	savedOrder, err := mockRepo.FindByID("test123")
	assert.NoError(t, err)
	assert.Equal(t, "test123", savedOrder.OrderUID)
}

func TestGetOrderHandler(t *testing.T) {
	handler, mockRepo := setupTestHandler()
	router := gin.Default()
	router.GET("/api/orders/:id", handler.GetOrderByID)

	// Предварительно сохраняем заказ
	testOrder := &model.Order{
		OrderUID:    "test123",
		TrackNumber: "WBILMTESTTRACK",
		// ... остальные поля
	}
	mockRepo.Save(testOrder)

	req := httptest.NewRequest("GET", "/api/orders/test123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response model.Order
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "test123", response.OrderUID)
}
