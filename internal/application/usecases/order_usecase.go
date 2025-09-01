package usecases

import (
	"context"
	"errors"
	"fmt"
	"log"
	"shop-microservice/internal/domain/model"
	"shop-microservice/internal/domain/repositories"
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
	ValidateOrder(order *model.Order) error
	HealthCheck(ctx context.Context) (map[string]interface{}, error)
	Shutdown()
}

type orderUseCase struct {
	repo            repositories.OrderRepository
	messageProducer repositories.MessageProducer
	cache           repositories.Cache
	producerChan    chan *model.Order
	workers         int
	wg              sync.WaitGroup
}

func NewOrderUseCase(
	repo repositories.OrderRepository,
	messageProducer repositories.MessageProducer,
	cache repositories.Cache,
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
		} else {
			log.Printf("order successfully sent to message queue: worker: %d, order: %v", workerNumber, order.OrderUID)
		}
	}
}

func (uc *orderUseCase) CreateOrder(ctx context.Context, order *model.Order) error {
	if err := uc.ValidateOrder(order); err != nil {
		return err
	}

	// Сохраняем в кэш
	uc.cache.Set(order.OrderUID, order)

	// Сохраняем в базу данных
	if err := uc.repo.Save(ctx, order); err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}

	// Отправляем в канал для асинхронной обработки
	select {
	case uc.producerChan <- order:
		return nil
	default:
		return errors.New("channel is full, message dropped")
	}
}

func (uc *orderUseCase) GetOrderByID(ctx context.Context, id string) (*model.Order, error) {
	// Пытаемся получить из кэша
	if order, exists := uc.cache.Get(id); exists {
		return order, nil
	}

	// Если нет в кэше, ищем в базе
	order, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}

	// Сохраняем в кэш для будущих запросов
	uc.cache.Set(id, order)
	return order, nil
}

func (uc *orderUseCase) GetAllOrders(ctx context.Context) ([]*model.Order, error) {
	// Пытаемся получить из кэша
	cachedOrders := uc.cache.GetAll()
	if len(cachedOrders) > 0 {
		return cachedOrders, nil
	}

	// Если нет в кэше, загружаем из базы
	orders, err := uc.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch orders: %w", err)
	}

	// Сохраняем все заказы в кэш
	for _, order := range orders {
		if order != nil {
			uc.cache.Set(order.OrderUID, order)
		}
	}

	return orders, nil
}

func (uc *orderUseCase) ValidateOrder(order *model.Order) error {
	if order.OrderUID == "" {
		return errors.New("order uid is required")
	}
	if order.TrackNumber == "" {
		return errors.New("track number is required")
	}
	if order.Entry == "" {
		return errors.New("entry is required")
	}
	if order.CustomerID == "" {
		return errors.New("customer id is required")
	}
	if len(order.Items) == 0 {
		return errors.New("items are required")
	}
	if order.Payment.Transaction == "" {
		return errors.New("payment transaction is required")
	}
	return nil
}

func (uc *orderUseCase) HealthCheck(ctx context.Context) (map[string]interface{}, error) {
	health := map[string]any{
		"status":       "healthy",
		"cache_size":   uc.cache.Size(),
		"cache_loaded": uc.cache.Size() > 0,
		"timestamp":    time.Now().Format(time.RFC3339),
	}

	// Проверяем подключение к базе данных
	_, err := uc.repo.FindAll(ctx)
	if err != nil {
		health["status"] = "unhealthy"
		health["database_error"] = err.Error()
		return health, err
	}

	health["database"] = "connected"
	return health, nil
}

func (uc *orderUseCase) Shutdown() {
	close(uc.producerChan)
	uc.wg.Wait()
	log.Printf("all workers stopped")
}
