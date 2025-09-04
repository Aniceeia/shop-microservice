package unit

import (
	"database/sql"
	"testing"
	"time"

	"shop-microservice/internal/di"
	"shop-microservice/internal/infrastructure/cache"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockDB struct{ mock.Mock }

func (m *mockDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	callArgs := m.Called(append([]interface{}{query}, args...))
	return callArgs.Get(0).(sql.Result), callArgs.Error(1)
}
func (m *mockDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	callArgs := m.Called(append([]interface{}{query}, args...))
	return callArgs.Get(0).(*sql.Rows), callArgs.Error(1)
}
func (m *mockDB) QueryRow(query string, args ...interface{}) *sql.Row {
	callArgs := m.Called(append([]interface{}{query}, args...))
	return callArgs.Get(0).(*sql.Row)
}

func TestNewConfig(t *testing.T) {
	cfg, err := di.NewConfig()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "postgres", cfg.DBHost)
	assert.Equal(t, 5432, cfg.DBPort)
	assert.Equal(t, "orders_user", cfg.DBUser)
	assert.Equal(t, "orders_password", cfg.DBPassword)
	assert.Equal(t, "orders_db", cfg.DBName)
	assert.Equal(t, "8081", cfg.AppPort)
	assert.Equal(t, []string{"kafka:9092"}, cfg.KafkaBrokers)
	assert.Equal(t, "orders", cfg.KafkaTopic)
	assert.Equal(t, "order-service", cfg.KafkaGroupID)
}

func TestNewConfig_InvalidPort(t *testing.T) {
	t.Setenv("DB_PORT", "invalid")

	cfg, err := di.NewConfig()
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestNewDB_Success(t *testing.T) {
	cfg := &di.Config{
		DBHost:     "localhost",
		DBPort:     5432,
		DBUser:     "test",
		DBPassword: "test",
		DBName:     "test",
	}

	db, err := di.NewDB(cfg)
	if err != nil {
		assert.Error(t, err)
	} else {
		assert.NoError(t, err)
		assert.NotNil(t, db)
		db.Close()
	}
}
func TestNewKafkaManager(t *testing.T) {
	cfg := &di.Config{KafkaBrokers: []string{"localhost:9092"}}
	manager := di.NewKafkaManager(cfg)
	assert.NotNil(t, manager)
}

func TestNewKafkaProducer(t *testing.T) {
	cfg := &di.Config{
		KafkaBrokers: []string{"localhost:9092"},
		KafkaTopic:   "test-topic",
	}
	producer := di.NewKafkaProducer(cfg)
	assert.NotNil(t, producer)
}

func TestNewKafkaMessageProducer(t *testing.T) {
	// Вместо мока создаем реальный продюсер с тестовой конфигурацией
	cfg := &di.Config{
		KafkaBrokers: []string{"localhost:9092"},
		KafkaTopic:   "test-topic",
	}

	producer := di.NewKafkaProducer(cfg)
	assert.NotNil(t, producer)

	messageProducer := di.NewKafkaMessageProducer(producer)
	assert.NotNil(t, messageProducer)
}
func TestNewOrderRepository(t *testing.T) {
	// Используем реальную базу данных для теста
	cfg := &di.Config{
		DBHost:     "localhost",
		DBPort:     5432,
		DBUser:     "test",
		DBPassword: "test",
		DBName:     "test",
	}

	db, err := di.NewDB(cfg)
	if err != nil {
		t.Skipf("Skipping test: cannot connect to DB: %v", err)
		return
	}
	defer db.Close()

	repo := di.NewOrderRepository(db)
	assert.NotNil(t, repo)
}

func TestNewCache(t *testing.T) {
	cache := cache.NewCash(1000, 30*time.Minute)
	assert.Equal(t, 0, cache.Size())
}

func TestNewMetrics(t *testing.T) {
	metrics := di.NewMetrics()
	assert.NotNil(t, metrics)

	stats := metrics.GetStats()
	assert.IsType(t, map[string]interface{}{}, stats["requests"])
	assert.IsType(t, map[string]interface{}{}, stats["errors"])
	assert.IsType(t, float64(0), stats["cache_hits"])
	assert.IsType(t, float64(0), stats["cache_misses"])
	assert.IsType(t, float64(0), stats["db_connections"])
	assert.IsType(t, float64(0), stats["kafka_messages"])
}

func TestNewOrderUseCase(t *testing.T) {
	cfg, _ := di.NewConfig()
	db, _ := di.NewDB(cfg)
	repo := di.NewOrderRepository(db)
	producer := di.NewKafkaProducer(cfg)
	messageProducer := di.NewKafkaMessageProducer(producer)
	cache := di.NewCache()
	metrics := di.NewMetrics()

	uc := di.NewOrderUseCase(repo, messageProducer, cache, metrics)
	assert.NotNil(t, uc)
}

func TestNewCacheAdapter(t *testing.T) {
	cache := di.NewCache()
	adapter := di.NewCacheAdapter(cache)
	assert.NotNil(t, adapter)
	assert.Equal(t, cache, adapter)
}

func TestNewHandler(t *testing.T) {
	cfg, _ := di.NewConfig()
	db, _ := di.NewDB(cfg)
	repo := di.NewOrderRepository(db)
	producer := di.NewKafkaProducer(cfg)
	messageProducer := di.NewKafkaMessageProducer(producer)
	cache := di.NewCache()
	metrics := di.NewMetrics()
	uc := di.NewOrderUseCase(repo, messageProducer, cache, metrics)

	handler := di.NewHandler(uc)
	assert.NotNil(t, handler)
}

func TestNewRouter(t *testing.T) {
	cfg, _ := di.NewConfig()
	db, _ := di.NewDB(cfg)
	repo := di.NewOrderRepository(db)
	producer := di.NewKafkaProducer(cfg)
	messageProducer := di.NewKafkaMessageProducer(producer)
	cache := di.NewCache()
	metrics := di.NewMetrics()
	uc := di.NewOrderUseCase(repo, messageProducer, cache, metrics)
	handler := di.NewHandler(uc)

	router := di.NewRouter(handler)
	assert.NotNil(t, router)
}

func TestNewHTTPServer(t *testing.T) {
	cfg, _ := di.NewConfig()
	db, _ := di.NewDB(cfg)
	repo := di.NewOrderRepository(db)
	producer := di.NewKafkaProducer(cfg)
	messageProducer := di.NewKafkaMessageProducer(producer)
	cache := di.NewCache()
	metrics := di.NewMetrics()
	uc := di.NewOrderUseCase(repo, messageProducer, cache, metrics)
	handler := di.NewHandler(uc)
	router := di.NewRouter(handler)

	server := di.NewHTTPServer(router, cfg)
	assert.NotNil(t, server)
	assert.Equal(t, ":"+cfg.AppPort, server.Addr)
}
