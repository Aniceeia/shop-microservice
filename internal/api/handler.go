package api

import (
	"context"
	"log"
	"net/http"
	"shop-microservice/internal/domain/model"
	"shop-microservice/internal/domain/repositories"
	"shop-microservice/internal/infrastructure/cash"
	"shop-microservice/internal/infrastructure/kafka"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	repo     repositories.OrderRepository
	producer *kafka.Producer
	cash     *cash.Cash
}

func NewHandler(repo repositories.OrderRepository, producer *kafka.Producer, cash *cash.Cash) *Handler {
	return &Handler{
		repo:     repo,
		producer: producer,
		cash:     cash,
	}
}

func (h *Handler) CreateOrder(c *gin.Context) {
	var order model.Order

	if err := c.ShouldBindJSON(&order); err != nil {
		log.Printf("Invalid json payload: %w", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request payload",
			"details": err.Error(),
		})
		return
	}
	//limit values
	if order.OrderUID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order uid is required"})
		return
	}

	if order.TrackNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "truck number is required"})
		return
	}

	if order.Entry == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "entry is required"})
		return
	}

	if order.CustomerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "customer id is required"})
		return
	}

	if len(order.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "delivery name is required"})
		return
	}

	if order.Payment.Transaction == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payment transaction is required"})
		return
	}

	ctx := c.Request.Context()
	if err := h.repo.Save(ctx, &order); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.cash.Set(order.OrderUID, &order)

	if h.producer != nil {
		go func() {
			//ctx := context.Background()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := h.producer.Produce(ctx, order.OrderUID, order); err != nil {
				log.Printf("Failed to produce message to Kafka: %v", err)
			}
		}()
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Order created successfully",
		"order_uid": order.OrderUID,
		"order":     order, // Возвращаем созданный заказ
	})
}

// GetOrderByID возвращает заказ по ID (с использованием кэша)
func (h *Handler) GetOrderByID(c *gin.Context) {
	orderUID := c.Param("id")

	if order, exists := h.cash.Get(orderUID); exists {
		c.JSON(http.StatusOK, order)
		return
	}

	ctx := c.Request.Context()
	order, err := h.repo.FindByID(ctx, orderUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Order not found",
			"uid":   orderUID})
		return
	}

	h.cash.Set(orderUID, order)

	c.JSON(http.StatusOK, order)
}

// GetAllOrders возвращает все заказы (с использованием кэша)
func (h *Handler) GetAllOrders(c *gin.Context) {
	orders := h.cash.GetAll()
	if len(orders) > 0 {
		log.Printf("Returning orders from cache")
		c.JSON(http.StatusOK, orders)
		return
	}
	//if cash is empty load from bd
	ctx := c.Request.Context()
	dbOrders, err := h.repo.FindAll(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch orders",
		})
		return
	}
	for _, order := range dbOrders {
		if order != nil {
			h.cash.Set(order.OrderUID, order)
		}
	}
	log.Printf("Returning orders from database")
	c.JSON(http.StatusOK, orders)
}

// HealthCheck проверяет соединение с БД и состояние кэша
func (h *Handler) HealthCheck(c *gin.Context) {
	health := gin.H{
		"status":       "healthy",
		"cache_size":   h.cash.Size(),
		"cache_loaded": h.cash.Size() > 0,
		"timestamp":    time.Now().Format(time.RFC3339),
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	_, err := h.repo.FindAll(ctx)
	if err != nil {
		health["status"] = "unhealthy"
		health["database_error"] = err.Error()
		c.JSON(http.StatusServiceUnavailable, health)
		return
	}

	health["database"] = "connected"
	c.JSON(http.StatusOK, health)
}
