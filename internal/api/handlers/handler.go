package handlers

import (
	"errors"
	"net/http"
	"shop-microservice/internal/application/usecases"
	"shop-microservice/internal/domain/model"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	useCase usecases.OrderUseCase
}

func NewHandler(useCase usecases.OrderUseCase) *Handler {
	return &Handler{
		useCase: useCase,
	}
}

func (h *Handler) CreateOrder(c *gin.Context) {
	var order model.Order

	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	ctx := c.Request.Context()
	if err := h.useCase.CreateOrder(ctx, &order); err != nil {
		if errors.Is(err, usecases.ErrValidation) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create order"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Order created successfully",
		"order_uid": order.OrderUID,
	})
}

func (h *Handler) GetOrderByID(c *gin.Context) {
	orderUID := c.Param("id")

	ctx := c.Request.Context()
	order, err := h.useCase.GetOrderByID(ctx, orderUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Order not found",
			"uid":   orderUID,
		})
		return
	}

	c.JSON(http.StatusOK, order)
}

func (h *Handler) GetAllOrders(c *gin.Context) {
	ctx := c.Request.Context()
	orders, err := h.useCase.GetAllOrders(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch orders",
		})
		return
	}

	c.JSON(http.StatusOK, orders)
}

func (h *Handler) HealthCheck(c *gin.Context) {
	ctx := c.Request.Context()
	health, err := h.useCase.HealthCheck(ctx)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, health)
		return
	}

	c.JSON(http.StatusOK, health)
}

func (h *Handler) Shutdown() {
	h.useCase.Shutdown()
}
