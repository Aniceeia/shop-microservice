package di

import (
	"database/sql"
	"shop-microservice/internal/api"
	"shop-microservice/internal/infrastructure/cache"
	"shop-microservice/internal/infrastructure/kafka"
	"shop-microservice/internal/infrastructure/postgresql"
	"sync"
)

type Container struct {
	db *sql.DB
	//kafkaManager *kafka.KafkaManager
	repo     *postgresql.OrderRepository
	cache    *cache.Cache
	producer *kafka.Producer
	handler  *api.Handler
	workers  int
	mu       sync.Mutex
}

func NewContainer() *Container {
	return &Container{
		workers: 3,
	}
}

func (c *Container) GetDB() *sql.DB {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.db == nil {
		panic("DB not init")
	}
	return c.db
}

func (c *Container) SetDB(db *sql.DB) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.db = db
}

func (c *Container) GetRepository() *postgresql.OrderRepository {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.repo == nil {
		c.repo = postgresql.NewOrderRepository(c.GetDB())
	}
	return c.repo
}

func (c *Container) GetCache() *cache.Cache {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cache == nil {
		c.cache = cache.NewCash()
	}
	return c.cache
}

func (c *Container) GetKafkaProducer(brokers []string, topic string) *kafka.Producer {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.producer == nil {
		c.producer = kafka.NewProducer(kafka.ProducerConfig{
			Brokers: brokers,
			Topic:   topic,
		})
	}
	return c.producer
}

func (c *Container) GetHandler() *api.Handler {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.handler == nil {
		c.handler = api.NewHandler(
			c.GetRepository(),
			c.GetKafkaProducer([]string{"kafka:9092"}, "orders"),
			c.GetCache(),
			c.workers,
		)
	}
	return c.handler
}

func (c *Container) Shutdown() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.producer != nil {
		c.producer.Close()
	}
	if c.handler != nil {
		c.handler.Shutdown()
	}
	if c.db != nil {
		c.db.Close()
	}
}
