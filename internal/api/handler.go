package api

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"shop-microservice/internal/domain/model"
	"shop-microservice/internal/domain/repositories"
	"shop-microservice/internal/infrastructure/cache"
	"shop-microservice/internal/infrastructure/kafka"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	repo         repositories.OrderRepository
	producer     *kafka.Producer
	cache        *cache.Cache
	producerChan chan *model.Order
	workers      int
	wg           sync.WaitGroup
}

func NewHandler(repo repositories.OrderRepository, producer *kafka.Producer, cache *cache.Cache, workers int) *Handler {
	if workers <= 0 {
		workers = 10
	}
	h := &Handler{
		repo:         repo,
		producer:     producer,
		cache:        cache,
		producerChan: make(chan *model.Order, 1000),
		workers:      workers,
	}
	h.startWorkers()
	return h
}

func (h *Handler) startWorkers() {
	for i := 0; i < h.workers; i++ {
		h.wg.Add(1)
		go func(workerNumber int) {
			defer h.wg.Done()
			h.workerLoop(workerNumber)
		}(i)
	}
}

func (h *Handler) workerLoop(workerNumber int) {
	for order := range h.producerChan {
		ctx := context.Background()
		if err := h.producer.Produce(ctx, order.OrderUID, order); err != nil {
			log.Printf("failed to produce to Kafka: worker: %d, ERROR: %v", workerNumber, err)
		} else {
			log.Printf("order successfully produced: worker: %d, order: %v", workerNumber, order.OrderUID)
		}
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
	if err := h.validateOrder(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Internal server error"})
		return
	}

	h.cache.Set(order.OrderUID, &order)

	ctx := c.Request.Context()
	if err := h.repo.Save(ctx, &order); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save order"})
		return
	}

	if err := h.orderToChannel(&order); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "channel is full, message dropped"})
		//return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Order created successfully",
		"order_uid": order.OrderUID,
	})
}

func (h *Handler) orderToChannel(order *model.Order) error {
	select {
	case h.producerChan <- order:
		return nil
	default:
		return errors.New("creation error, message dropped")
	}
}

func (h *Handler) validateOrder(order *model.Order) error {
	fail := func(err string) error {
		return fmt.Errorf("validation err: %v", err)
	}
	if order.OrderUID == "" {
		return fail("order uid is required")
	}
	if order.TrackNumber == "" {
		return fail("track number is required")
	}
	if order.Entry == "" {
		return fail("entry is required")
	}
	if order.CustomerID == "" {
		return fail("customer id is required")
	}
	if len(order.Items) == 0 {
		return fail("items are required")
	}
	if order.Payment.Transaction == "" {
		return fail("payment transaction is required")
	}
	return nil
}

func (h *Handler) GetOrderByID(c *gin.Context) {
	orderUID := c.Param("id")

	if order, exists := h.cache.Get(orderUID); exists {
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

	h.cache.Set(orderUID, order)

	c.JSON(http.StatusOK, order)
}

func (h *Handler) GetAllOrders(c *gin.Context) {
	orders := h.cache.GetAll()
	if len(orders) > 0 {
		log.Printf("Returning orders from cache")
		c.JSON(http.StatusOK, orders)
		return
	}
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
			h.cache.Set(order.OrderUID, order)
		}
	}
	log.Printf("Returning orders from database")
	c.JSON(http.StatusOK, orders)
}

func (h *Handler) HealthCheck(c *gin.Context) {
	health := gin.H{
		"status":       "healthy",
		"cache_size":   h.cache.Size(),
		"cache_loaded": h.cache.Size() > 0,
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

func (h *Handler) Shutdown() {
	close(h.producerChan)
	h.wg.Wait()
	log.Printf("all workers stopped")
}
