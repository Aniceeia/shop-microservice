package usecases

import (
	"context"
	"errors"
	"fmt"
	"log"
	"shop-microservice/internal/api/middleware"
	"shop-microservice/internal/domain/model"
	"shop-microservice/internal/domain/repositories"
	"strconv"
	"sync"
	"time"
)

var (
	ErrValidation   = errors.New("validation err")
	ErrServerClosed = errors.New("server is closed")
)

type OrderUseCase interface {
	CreateOrder(ctx context.Context, order *model.Order) error
	GetOrderByID(ctx context.Context, id string) (*model.Order, error)
	GetAllOrders(ctx context.Context) ([]*model.Order, error)
	HealthCheck(ctx context.Context) (map[string]interface{}, error)
	Shutdown()
}

type orderUseCase struct {
	repo            repositories.OrderRepository
	messageProducer repositories.MessageProducer
	cache           repositories.Cache
	metrics         repositories.Metrics
	producerChan    chan *model.Order
	workers         int
	wg              sync.WaitGroup
}

func NewOrderUseCase(
	repo repositories.OrderRepository,
	messageProducer repositories.MessageProducer,
	cache repositories.Cache,
	metrics repositories.Metrics,
	workers int,
	bufferSize int,
) OrderUseCase {
	if workers <= 0 {
		workers = 10
	}
	if bufferSize <= 0 {
		bufferSize = 1000
	}

	uc := &orderUseCase{
		repo:            repo,
		messageProducer: messageProducer,
		cache:           cache,
		metrics:         metrics,
		producerChan:    make(chan *model.Order, bufferSize),
		workers:         workers,
	}
	uc.startWorkers()
	return uc
}

func (uc *orderUseCase) startWorkers() {
	for i := 0; i < uc.workers; i++ {
		uc.wg.Add(1)
		go func(workerNumber int) {
			defer uc.wg.Done()
			uc.workerLoop(workerNumber)
		}(i)
	}
}

func (uc *orderUseCase) workerLoop(workerNumber int) {
	for order := range uc.producerChan {
		ctx := context.Background()
		if err := uc.messageProducer.ProduceOrder(ctx, order); err != nil {
			log.Printf("failed to produce message: worker: %d, ERROR: %v", workerNumber, err)
			middleware.WorkerErrors.WithLabelValues(
				strconv.Itoa(workerNumber),
			).Inc()
		} else {
			log.Printf("order successfully sent to message queue: worker: %d, order: %v", workerNumber, order.OrderUID)
		}
	}
}

func (uc *orderUseCase) CreateOrder(ctx context.Context, order *model.Order) error {
	if err := uc.validateOrderFields(order); err != nil {
		middleware.ValidationErrors.Inc()
		return err
	}

	uc.cache.Set(order.OrderUID, order)

	if err := uc.repo.Save(ctx, order); err != nil {
		return err
	}

	middleware.OrdersCreated.Inc()
	middleware.QueueSize.Set(float64(len(uc.producerChan)))
	middleware.QueueCapacity.Set(float64(cap(uc.producerChan)))

	select {
	case uc.producerChan <- order:
		return nil
	default:
		return ErrServerClosed
	}
}

func (uc *orderUseCase) GetOrderByID(ctx context.Context, id string) (*model.Order, error) {
	if order, exists := uc.cache.Get(id); exists {
		uc.metrics.IncrementCacheHit()
		middleware.OrdersRetrieved.Inc()
		return order, nil
	}
	uc.metrics.IncrementCacheMiss()

	order, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	uc.cache.Set(id, order)
	middleware.OrdersRetrieved.Inc()
	return order, nil
}

func (uc *orderUseCase) GetAllOrders(ctx context.Context) ([]*model.Order, error) {
	cachedOrders := uc.cache.GetAll()
	if len(cachedOrders) > 0 {
		return cachedOrders, nil
	}

	orders, err := uc.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	for _, order := range orders {
		uc.cache.Set(order.OrderUID, order)
	}

	return orders, nil
}

func (uc *orderUseCase) validateOrderFields(order *model.Order) error {
	if order.OrderUID == "" {
		return errors.New("order uid is required")
	}
	if order.TrackNumber == "" {
		return errors.New("track number is required")
	}
	if order.Entry == "" {
		return errors.New("entry is required")
	}
	if order.Delivery.Name == "" {
		return errors.New("delivery name is required")
	}
	if order.Payment.Amount <= 0 {
		return errors.New("payment amount must be positive")
	}
	if len(order.Items) == 0 {
		return errors.New("order must have at least one item")
	}
	return nil
}

func (uc *orderUseCase) HealthCheck(ctx context.Context) (map[string]interface{}, error) {
	health := map[string]any{
		"status":       "healthy",
		"timestamp":    time.Now().Unix(),
		"cache_size":   uc.cache.Size(),
		"workers":      uc.workers,
		"queue_size":   len(uc.producerChan),
		"cap":          cap(uc.producerChan),
		"db_connected": true,
	}

	if len(uc.producerChan) >= cap(uc.producerChan) {
		health["status"] = "unhealthy"
		health["error"] = "queue is full"
		return health, fmt.Errorf("queue is full")
	}

	return health, nil
}

func (uc *orderUseCase) Shutdown() {
	close(uc.producerChan)
	uc.wg.Wait()
	log.Printf("all workers stopped")
}
