package api

import (
	"context"
	"log"
	"net/http"
	"shop-microservice/internal/domain/model"
	"shop-microservice/internal/domain/repositories"
	"shop-microservice/internal/infrastructure/kafka"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	repo     repositories.OrderRepository
	producer *kafka.Producer
}

func NewHandler(repo repositories.OrderRepository, producer *kafka.Producer) *Handler {
	return &Handler{
		repo:     repo,
		producer: producer,
	}
}

// CreateOrder создает новый заказ
func (h *Handler) CreateOrder(c *gin.Context) {
	var order model.Order

	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.Save(&order); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Отправляем сообщение в Kafka
	if h.producer != nil {
		go func() {
			ctx := context.Background()
			if err := h.producer.Produce(ctx, order.OrderUID, order); err != nil {
				// Логируем ошибку, но не прерываем выполнение
				log.Printf("Failed to produce message to Kafka: %v", err)
			}
		}()
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Order created successfully", "order_uid": order.OrderUID})
}

// GetOrderByID возвращает заказ по ID
func (h *Handler) GetOrderByID(c *gin.Context) {
	orderUID := c.Param("id")

	order, err := h.repo.FindByID(orderUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	c.JSON(http.StatusOK, order)
}

// GetAllOrders возвращает все заказы
func (h *Handler) GetAllOrders(c *gin.Context) {
	orders, err := h.repo.FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, orders)
}

// HealthCheck проверяет соединение с БД
func (h *Handler) HealthCheck(c *gin.Context) {
	// Простой запрос чтобы проверить соединение
	_, err := h.repo.FindAll()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}
