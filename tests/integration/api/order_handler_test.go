package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"shop-microservice/internal/api/handlers"
	"shop-microservice/internal/domain/model"
)

func TestOrderHandler_CreateOrder(t *testing.T) {
	handler := handlers.NewHandler(mockOrderService{})

	data, err := os.ReadFile("tests/fixtures/orders/order_0.json")
	if err != nil {
		t.Fatalf("Failed to read test data: %v", err)
	}

	req := httptest.NewRequest("POST", "/orders", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	handler.CreateOrder(w, req)

	// Verify
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}
}

type mockOrderService struct{}

func (m mockOrderService) CreateOrder(ctx context.Context, order *model.Order) error {
	return nil
}

func (m mockOrderService) GetOrderByID(id string) (*model.Order, error) {
	return &model.Order{}, nil
}
