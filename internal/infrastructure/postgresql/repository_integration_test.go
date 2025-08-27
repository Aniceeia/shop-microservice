package postgresql

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"shop-microservice/internal/domain/model"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	if err := setup(); err != nil {
		log.Fatalf("Test setup failed: %v", err)
	}

	code := m.Run()

	teardown()
	os.Exit(code)
}

func setup() error {
	connStr := "host=localhost port=5433 user=orders_user password=orders_password dbname=orders_db sslmode=disable"
	time.Sleep(5 * time.Second)

	// Отладочная информация
	log.Printf("Connecting to database: host=localhost port=5432 user=orders_user dbname=orders_db")

	var err error
	testDB, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Пробуем подключиться несколько раз
	for i := 0; i < 10; i++ {
		if err := testDB.Ping(); err == nil {
			log.Printf("Database connection successful!")
			break
		}
		log.Printf("Waiting for database... (attempt %d/10): %v", i+1, err)
		time.Sleep(2 * time.Second)
	}

	if err := testDB.Ping(); err != nil {
		// Попробуем проверить, доступна ли БД вообще
		log.Printf("Trying to check database availability...")

		// Попробуем подключиться без указания базы данных
		connStrNoDB := "host=localhost port=5432 user=orders_user password=orders_password sslmode=disable"
		tempDB, tempErr := sql.Open("postgres", connStrNoDB)
		if tempErr == nil {
			defer tempDB.Close()
			if tempErr := tempDB.Ping(); tempErr == nil {
				log.Printf("Can connect to PostgreSQL but not to specific database")
				// Проверим существование базы данных
				rows, tempErr := tempDB.Query("SELECT datname FROM pg_database WHERE datname = 'orders_db'")
				if tempErr == nil {
					defer rows.Close()
					if rows.Next() {
						log.Printf("Database orders_db exists")
					} else {
						log.Printf("Database orders_db does not exist")
					}
				}
			}
		}

		return fmt.Errorf("database ping failed after retries: %w", err)
	}

	// Создаем тестовые таблицы
	if err := setupTestSchema(testDB); err != nil {
		return fmt.Errorf("failed to setup test schema: %w", err)
	}

	log.Println("Test database setup completed successfully")
	return nil
}

func teardown() {
	if testDB != nil {
		// Очищаем таблицы
		tables := []string{"items", "payments", "deliveries", "orders"}
		for _, table := range tables {
			testDB.Exec("DELETE FROM " + table)
		}
		testDB.Close()
	}
}

func setupTestSchema(db *sql.DB) error {
	schema := `
	DROP TABLE IF EXISTS items;
	DROP TABLE IF EXISTS payments;
	DROP TABLE IF EXISTS deliveries;
	DROP TABLE IF EXISTS orders;

	CREATE TABLE orders (
		order_uid VARCHAR(255) PRIMARY KEY,
		track_number VARCHAR(255),
		entry VARCHAR(255),
		locale VARCHAR(10),
		internal_signature VARCHAR(255),
		customer_id VARCHAR(255),
		delivery_service VARCHAR(255),
		shardkey VARCHAR(255),
		sm_id INTEGER,
		date_created TIMESTAMP,
		oof_shard VARCHAR(255)
	);

	CREATE TABLE deliveries (
		order_uid VARCHAR(255) PRIMARY KEY REFERENCES orders(order_uid) ON DELETE CASCADE,
		name VARCHAR(255),
		phone VARCHAR(255),
		zip VARCHAR(255),
		city VARCHAR(255),
		address VARCHAR(255),
		region VARCHAR(255),
		email VARCHAR(255)
	);

	CREATE TABLE payments (
		order_uid VARCHAR(255) PRIMARY KEY REFERENCES orders(order_uid) ON DELETE CASCADE,
		transaction VARCHAR(255),
		request_id VARCHAR(255),
		currency VARCHAR(10),
		provider VARCHAR(255),
		amount INTEGER,
		payment_dt BIGINT,
		bank VARCHAR(255),
		delivery_cost INTEGER,
		goods_total INTEGER,
		custom_fee INTEGER
	);

	CREATE TABLE items (
		order_uid VARCHAR(255) REFERENCES orders(order_uid) ON DELETE CASCADE,
		chrt_id INTEGER,
		track_number VARCHAR(255),
		price INTEGER,
		rid VARCHAR(255),
		name VARCHAR(255),
		sale INTEGER,
		size VARCHAR(255),
		total_price INTEGER,
		nm_id INTEGER,
		brand VARCHAR(255),
		status INTEGER,
		PRIMARY KEY (order_uid, chrt_id)
	);
	`

	_, err := db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	return nil
}

func createTestOrder() *model.Order {
	return &model.Order{
		OrderUID:          "test-order-uid",
		TrackNumber:       "test-track",
		Entry:             "test-entry",
		Locale:            "en",
		InternalSignature: "test-signature",
		CustomerID:        "test-customer",
		DeliveryService:   "test-service",
		Shardkey:          "test-shard",
		SmID:              1,
		DateCreated:       time.Now(),
		OofShard:          "test-oof",
		Delivery: model.Delivery{
			Name:    "John Doe",
			Phone:   "+1234567890",
			Zip:     "123456",
			City:    "Moscow",
			Address: "Street 1",
			Region:  "Moscow",
			Email:   "john@example.com",
		},
		Payment: model.Payment{
			Transaction:  "test-transaction",
			RequestID:    "test-request",
			Currency:     "USD",
			Provider:     "test-provider",
			Amount:       1000,
			PaymentDt:    time.Now().Unix(),
			Bank:         "test-bank",
			DeliveryCost: 500,
			GoodsTotal:   500,
			CustomFee:    0,
		},
		Items: []model.Item{
			{
				ChrtID:      1,
				TrackNumber: "item-track-1",
				Price:       100,
				Rid:         "item-rid-1",
				Name:        "Test Item 1",
				Sale:        0,
				Size:        "M",
				TotalPrice:  100,
				NmID:        123,
				Brand:       "Test Brand",
				Status:      1,
			},
		},
	}
}

func TestOrderRepository_SaveAndFindByID(t *testing.T) {
	repo := NewOrderRepository(testDB)

	// Создаем тестовый заказ
	order := createTestOrder()

	// Сохраняем заказ
	err := repo.Save(order)
	require.NoError(t, err)

	// Ищем заказ по ID
	found, err := repo.FindByID(order.OrderUID)
	require.NoError(t, err)

	// Проверяем, что данные совпадают
	assert.Equal(t, order.OrderUID, found.OrderUID)
	assert.Equal(t, order.TrackNumber, found.TrackNumber)
	assert.Equal(t, order.Entry, found.Entry)
	assert.Equal(t, order.Locale, found.Locale)

	// Проверяем delivery
	assert.Equal(t, order.Delivery.Name, found.Delivery.Name)
	assert.Equal(t, order.Delivery.Phone, found.Delivery.Phone)
	assert.Equal(t, order.Delivery.Email, found.Delivery.Email)

	// Проверяем payment
	assert.Equal(t, order.Payment.Transaction, found.Payment.Transaction)
	assert.Equal(t, order.Payment.Amount, found.Payment.Amount)
	assert.Equal(t, order.Payment.Currency, found.Payment.Currency)

	// Проверяем items
	require.Len(t, found.Items, 1)
	assert.Equal(t, order.Items[0].Name, found.Items[0].Name)
	assert.Equal(t, order.Items[0].Price, found.Items[0].Price)
	assert.Equal(t, order.Items[0].Brand, found.Items[0].Brand)
}

// Добавьте остальные тестовые функции...

func TestOrderRepository_SaveAndFindByID_WithMultipleItems(t *testing.T) {
	repo := NewOrderRepository(testDB)

	// Создаем заказ с несколькими items
	order := createTestOrder()
	order.Items = []model.Item{
		{
			ChrtID:      1,
			TrackNumber: "item-track-1",
			Price:       100,
			Rid:         "item-rid-1",
			Name:        "Test Item 1",
			Sale:        0,
			Size:        "M",
			TotalPrice:  100,
			NmID:        123,
			Brand:       "Test Brand 1",
			Status:      1,
		},
		{
			ChrtID:      2,
			TrackNumber: "item-track-2",
			Price:       200,
			Rid:         "item-rid-2",
			Name:        "Test Item 2",
			Sale:        10,
			Size:        "L",
			TotalPrice:  180,
			NmID:        124,
			Brand:       "Test Brand 2",
			Status:      2,
		},
	}

	// Сохраняем заказ
	err := repo.Save(order)
	require.NoError(t, err)

	// Ищем заказ по ID
	found, err := repo.FindByID(order.OrderUID)
	require.NoError(t, err)

	// Проверяем items
	require.Len(t, found.Items, 2)
	assert.Equal(t, "Test Item 1", found.Items[0].Name)
	assert.Equal(t, "Test Item 2", found.Items[1].Name)
}

func TestOrderRepository_FindAll(t *testing.T) {
	repo := NewOrderRepository(testDB)

	// Сохраняем несколько заказов
	order1 := createTestOrder()
	order2 := createTestOrder()
	order2.OrderUID = "test-order-uid-2"
	order2.TrackNumber = "test-track-2"

	err := repo.Save(order1)
	require.NoError(t, err)

	err = repo.Save(order2)
	require.NoError(t, err)

	// Получаем все заказы
	orders, err := repo.FindAll()
	require.NoError(t, err)

	// Проверяем количество заказов
	assert.Len(t, orders, 2)

	// Проверяем, что оба заказа присутствуют
	orderUIDs := make(map[string]bool)
	for _, order := range orders {
		orderUIDs[order.OrderUID] = true
		// Проверяем, что у каждого заказа есть items
		assert.NotEmpty(t, order.Items)
	}

	assert.True(t, orderUIDs["test-order-uid"])
	assert.True(t, orderUIDs["test-order-uid-2"])
}

func TestOrderRepository_Save_UpdateExisting(t *testing.T) {
	repo := NewOrderRepository(testDB)

	// Создаем и сохраняем заказ
	order := createTestOrder()
	err := repo.Save(order)
	require.NoError(t, err)

	// Обновляем заказ
	order.TrackNumber = "updated-track"
	order.Delivery.Name = "Updated Name"
	order.Items[0].Price = 999

	err = repo.Save(order)
	require.NoError(t, err)

	// Проверяем, что заказ обновился
	found, err := repo.FindByID(order.OrderUID)
	require.NoError(t, err)

	assert.Equal(t, "updated-track", found.TrackNumber)
	assert.Equal(t, "Updated Name", found.Delivery.Name)
	assert.Equal(t, 999, found.Items[0].Price)
}

func TestOrderRepository_WithExampleJSON(t *testing.T) {
	repo := NewOrderRepository(testDB)

	// Создаем заказ из примера JSON
	order := &model.Order{
		OrderUID:          "b563feb7b2b84b6test",
		TrackNumber:       "WBILMTESTTRACK",
		Entry:             "WBIL",
		Locale:            "en",
		InternalSignature: "",
		CustomerID:        "test",
		DeliveryService:   "meest",
		Shardkey:          "9",
		SmID:              99,
		DateCreated:       time.Date(2021, 11, 26, 6, 22, 19, 0, time.UTC),
		OofShard:          "1",
		Delivery: model.Delivery{
			Name:    "Test Testov",
			Phone:   "+9720000000",
			Zip:     "2639809",
			City:    "Kiryat Mozkin",
			Address: "Ploshad Mira 15",
			Region:  "Kraiot",
			Email:   "test@gmail.com",
		},
		Payment: model.Payment{
			Transaction:  "b563feb7b2b84b6test",
			RequestID:    "",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       1817,
			PaymentDt:    1637907727,
			Bank:         "alpha",
			DeliveryCost: 1500,
			GoodsTotal:   317,
			CustomFee:    0,
		},
		Items: []model.Item{
			{
				ChrtID:      9934930,
				TrackNumber: "WBILMTESTTRACK",
				Price:       453,
				Rid:         "ab4219087a764ae0btest",
				Name:        "Mascaras",
				Sale:        30,
				Size:        "0",
				TotalPrice:  317,
				NmID:        2389212,
				Brand:       "Vivienne Sabo",
				Status:      202,
			},
		},
	}

	// Сохраняем заказ
	err := repo.Save(order)
	require.NoError(t, err)

	// Проверяем, что заказ сохранился корректно
	found, err := repo.FindByID(order.OrderUID)
	require.NoError(t, err)

	// Проверяем основные поля
	assert.Equal(t, "b563feb7b2b84b6test", found.OrderUID)
	assert.Equal(t, "WBILMTESTTRACK", found.TrackNumber)
	assert.Equal(t, "WBIL", found.Entry)

	// Проверяем delivery
	assert.Equal(t, "Test Testov", found.Delivery.Name)
	assert.Equal(t, "+9720000000", found.Delivery.Phone)
	assert.Equal(t, "test@gmail.com", found.Delivery.Email)

	// Проверяем payment
	assert.Equal(t, "b563feb7b2b84b6test", found.Payment.Transaction)
	assert.Equal(t, 1817, found.Payment.Amount)
	assert.Equal(t, "USD", found.Payment.Currency)

	// Проверяем items
	require.Len(t, found.Items, 1)
	assert.Equal(t, "Mascaras", found.Items[0].Name)
	assert.Equal(t, 453, found.Items[0].Price)
	assert.Equal(t, "Vivienne Sabo", found.Items[0].Brand)
}

// func createTestOrder() *model.Order {
// 	return &model.Order{
// 		OrderUID:          "test-order-uid",
// 		TrackNumber:       "test-track",
// 		Entry:             "test-entry",
// 		Locale:            "en",
// 		InternalSignature: "test-signature",
// 		CustomerID:        "test-customer",
// 		DeliveryService:   "test-service",
// 		Shardkey:          "test-shard",
// 		SmID:              1,
// 		DateCreated:       time.Now(),
// 		OofShard:          "test-oof",
// 		Delivery: model.Delivery{
// 			Name:    "John Doe",
// 			Phone:   "+1234567890",
// 			Zip:     "123456",
// 			City:    "Moscow",
// 			Address: "Street 1",
// 			Region:  "Moscow",
// 			Email:   "john@example.com",
// 		},
// 		Payment: model.Payment{
// 			Transaction:  "test-transaction",
// 			RequestID:    "test-request",
// 			Currency:     "USD",
// 			Provider:     "test-provider",
// 			Amount:       1000,
// 			PaymentDt:    time.Now().Unix(),
// 			Bank:         "test-bank",
// 			DeliveryCost: 500,
// 			GoodsTotal:   500,
// 			CustomFee:    0,
// 		},
// 		Items: []model.Item{
// 			{
// 				ChrtID:      1,
// 				TrackNumber: "item-track-1",
// 				Price:       100,
// 				Rid:         "item-rid-1",
// 				Name:        "Test Item 1",
// 				Sale:        0,
// 				Size:        "M",
// 				TotalPrice:  100,
// 				NmID:        123,
// 				Brand:       "Test Brand",
// 				Status:      1,
// 			},
// 		},
// 	}
// }
