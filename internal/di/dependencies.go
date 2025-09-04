package di

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"shop-microservice/internal/api"
	"shop-microservice/internal/api/handlers"
	"shop-microservice/internal/api/middleware"
	"shop-microservice/internal/application/usecases"
	"shop-microservice/internal/domain/model"
	"shop-microservice/internal/domain/repositories"
	cachepkg "shop-microservice/internal/infrastructure/cache"
	"shop-microservice/internal/infrastructure/kafka"
	"shop-microservice/internal/infrastructure/metrics"
	"shop-microservice/internal/infrastructure/postgresql"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		NewConfig,
		NewDBPool,
		NewKafkaManager,
		NewKafkaProducer,
		NewKafkaMessageProducer,
		NewOrderRepository,
		NewCache,
		NewCacheAdapter,
		fx.Annotate(NewMetrics, fx.As(new(repositories.Metrics))),
		NewOrderUseCase,
		NewHandler,
		NewRouter,
		NewHTTPServer,
	),
	fx.Invoke(
		RunMigrations,
		PreloadCacheFromDB,
		StartKafkaConsumer,
		RegisterHooks,
	),
)

type Config struct {
	DBHost       string
	DBPort       int
	DBUser       string
	DBPassword   string
	DBName       string
	AppPort      string
	KafkaBrokers []string
	KafkaTopic   string
	KafkaGroupID string
}

func NewConfig() (*Config, error) {
	dbPort, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %w", err)
	}

	return &Config{
		DBHost:       getEnv("DB_HOST", "postgres"),
		DBPort:       dbPort,
		DBUser:       getEnv("DB_USER", "orders_user"),
		DBPassword:   getEnv("DB_PASSWORD", "orders_password"),
		DBName:       getEnv("DB_NAME", "orders_db"),
		AppPort:      getEnv("APP_PORT", "8081"),
		KafkaBrokers: strings.Split(getEnv("KAFKA_BROKERS", "kafka:9092"), ","),
		KafkaTopic:   getEnv("KAFKA_TOPIC", "orders"),
		KafkaGroupID: getEnv("KAFKA_GROUP_ID", "order-service"),
	}, nil
}

func NewDBPool(cfg *Config) (*pgxpool.Pool, error) {
	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)

	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	config.MaxConns = 50
	config.MinConns = 5
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	var pool *pgxpool.Pool
	// retry
	for i := range 5 {
		pool, err = pgxpool.NewWithConfig(context.Background(), config)
		if err == nil {
			if err := pool.Ping(context.Background()); err == nil {
				middleware.DBConnections.Set(1)
				return pool, nil
			}
		}
		log.Printf("Attempt %d: failed to connect to database, retrying...", i+1)
		time.Sleep(2 * time.Second)
	}

	middleware.DBConnections.Set(0)
	return nil, fmt.Errorf("failed to connect to database after 5 attempts: %w", err)
}

func NewKafkaManager(cfg *Config) *kafka.KafkaManager {
	return kafka.NewKafkaManager(cfg.KafkaBrokers)
}

func NewKafkaProducer(cfg *Config) *kafka.Producer {
	return kafka.NewProducer(kafka.ProducerConfig{
		Brokers: cfg.KafkaBrokers,
		Topic:   cfg.KafkaTopic,
	})
}

func NewKafkaMessageProducer(producer *kafka.Producer) repositories.MessageProducer {
	return kafka.NewKafkaMessageProducer(producer)
}

func NewOrderRepository(pool *pgxpool.Pool) *postgresql.OrderRepository {
	return postgresql.NewOrderRepository(pool)
}

func NewCache() *cachepkg.Cache {
	return cachepkg.NewCash(1000, 30*time.Minute)
}

func NewCacheAdapter(cache *cachepkg.Cache) repositories.Cache {
	return cache
}

func NewMetrics() *metrics.Metrics {
	return metrics.NewMetrics()
}

func NewOrderUseCase(
	repo *postgresql.OrderRepository,
	messageProducer repositories.MessageProducer,
	cache repositories.Cache,
	metrics repositories.Metrics,
) usecases.OrderUseCase {
	return usecases.NewOrderUseCase(
		repo,
		messageProducer,
		cache,
		metrics,
		3,
		1000,
	)
}

func NewHandler(useCase usecases.OrderUseCase) *handlers.Handler {
	return handlers.NewHandler(useCase)
}

func NewRouter(handler *handlers.Handler) *gin.Engine {
	return api.SetupRouter(handler)
}

func NewHTTPServer(router *gin.Engine, cfg *Config) *http.Server {
	return &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: router,
	}
}

func RunMigrations(pool *pgxpool.Pool, lc fx.Lifecycle) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return postgresql.RunMigrations(pool)
		},
	})
}

func PreloadCacheFromDB(cache repositories.Cache, repo *postgresql.OrderRepository, lc fx.Lifecycle) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := cache.PreloadCacheFromDB(repo); err != nil {
				log.Printf("Warning: cache warm-up failed: %v", err)
				return nil
			}
			log.Printf("Cache initialized with %d orders", cache.Size())
			return nil
		},
		OnStop: func(ctx context.Context) error {
			if c, ok := cache.(*cachepkg.Cache); ok {
				c.Stop()
			}
			return nil
		},
	})
}

func StartKafkaConsumer(
	cfg *Config,
	repo *postgresql.OrderRepository,
	cache *cachepkg.Cache,
	km *kafka.KafkaManager,
	lc fx.Lifecycle,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			//retry
			if err := km.TryToConnectKafkaFor(30 * time.Second); err != nil {
				return fmt.Errorf("kafka not available: %w", err)
			}

			if err := km.CreateTopicIfNotExists(cfg.KafkaTopic, 3, 1); err != nil {
				log.Printf("Warning: failed to create topic: %v", err)
			}

			go func() {
				consumer := kafka.NewConsumer(kafka.ConsumerConfig{
					Brokers:     cfg.KafkaBrokers,
					Topic:       cfg.KafkaTopic,
					GroupID:     cfg.KafkaGroupID,
					StartOffset: kafka.FirstOffset,
				})
				defer consumer.Close()

				handler := func(key string, value []byte) error {
					var order model.Order
					if err := json.Unmarshal(value, &order); err != nil {
						log.Printf("Failed to unmarshal order: %v", err)
						return err
					}

					ctx := context.Background()
					if err := repo.Save(ctx, &order); err != nil {
						log.Printf("Failed to save order to DB: %v", err)
						return err
					}

					cache.Set(order.OrderUID, &order)
					log.Printf("Order %s processed from Kafka", order.OrderUID)
					return nil
				}

				if err := consumer.Consume(context.Background(), handler); err != nil {
					log.Printf("Kafka consumer error: %v", err)
				}
			}()

			return nil
		},
	})
}

func RegisterHooks(
	server *http.Server,
	pool *pgxpool.Pool,
	producer *kafka.Producer,
	handler *handlers.Handler,
	lc fx.Lifecycle,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				log.Printf("Server starting on %s", server.Addr)
				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Fatal("Failed to start server:", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			handler.Shutdown()
			producer.Close()
			pool.Close()
			return server.Shutdown(ctx)
		},
	})
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
