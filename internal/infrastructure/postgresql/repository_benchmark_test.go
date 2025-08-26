package postgresql

import (
	"database/sql"
	"fmt"
	"shop-microservice/internal/domain/model"
	"testing"
	"time"
)

func BenchmarkOrderRepository_Save(b *testing.B) {
	db := setupBenchmarkDB()
	defer db.Close()

	repo := NewOrderRepository(db)
	order := CreateBenchmarkOrder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Меняем UID для каждого сохранения
		order.OrderUID = fmt.Sprintf("benchmark-order-%d", i)
		err := repo.Save(order)
		if err != nil {
			b.Fatalf("Save failed: %v", err)
		}
	}
}

func BenchmarkOrderRepository_FindByID(b *testing.B) {
	db := setupBenchmarkDB()
	defer db.Close()

	repo := NewOrderRepository(db)

	// Сначала сохраняем заказ для поиска
	order := CreateBenchmarkOrder()
	err := repo.Save(order)
	if err != nil {
		b.Fatalf("Setup failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.FindByID(order.OrderUID)
		if err != nil {
			b.Fatalf("FindByID failed: %v", err)
		}
	}
}

func BenchmarkOrderRepository_FindAll(b *testing.B) {
	db := setupBenchmarkDB()
	defer db.Close()

	repo := NewOrderRepository(db)

	// Сохраняем несколько заказов
	for i := 0; i < 100; i++ {
		order := createTestOrder()
		order.OrderUID = fmt.Sprintf("benchmark-order-%d", i)
		err := repo.Save(order)
		if err != nil {
			b.Fatalf("Setup failed: %v", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.FindAll()
		if err != nil {
			b.Fatalf("FindAll failed: %v", err)
		}
	}
}

func setupBenchmarkDB() *sql.DB {
	connStr := "host=localhost port=5433 user=orders_user password=orders_password dbname=orders_db sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}

	if err := db.Ping(); err != nil {
		panic(err)
	}

	// Очищаем таблицы перед бенчмарком
	tables := []string{"items", "payments", "deliveries", "orders"}
	for _, table := range tables {
		db.Exec("DELETE FROM " + table)
	}

	return db
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
