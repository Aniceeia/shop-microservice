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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	if err := h.repo.Save(ctx, &order); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Добавляем в кэш
	h.cash.Set(order.OrderUID, &order)

	// Отправляем сообщение в Kafka
	if h.producer != nil {
		go func() {
			ctx := context.Background()
			if err := h.producer.Produce(ctx, order.OrderUID, order); err != nil {
				log.Printf("Failed to produce message to Kafka: %v", err)
			}
		}()
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Order created successfully", "order_uid": order.OrderUID})
}

// GetOrderByID возвращает заказ по ID (с использованием кэша)
func (h *Handler) GetOrderByID(c *gin.Context) {
	orderUID := c.Param("id")

	if order, exists := h.cash.Get(orderUID); exists {
		c.JSON(http.StatusOK, order)
		return
	}

	// Если нет в кэше, ищем в БД
	ctx := c.Request.Context()
	order, err := h.repo.FindByID(ctx, orderUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	// Сохраняем в кэш для будущих запросов
	h.cash.Set(orderUID, order)

	c.JSON(http.StatusOK, order)
}

// GetAllOrders возвращает все заказы (с использованием кэша)
func (h *Handler) GetAllOrders(c *gin.Context) {
	// Можно использовать кэш или БД в зависимости от требований
	// Здесь используем кэш для скорости
	orders := h.cash.GetAll()
	c.JSON(http.StatusOK, orders)
}

// HealthCheck проверяет соединение с БД и состояние кэша
func (h *Handler) HealthCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	_, err := h.repo.FindAll(ctx)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":     "unhealthy",
			"error":      err.Error(),
			"cache_size": h.cash.Size(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":       "healthy",
		"cache_size":   h.cash.Size(),
		"cache_loaded": h.cash.Size() > 0,
	})
}
