package postgresql

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
)

func BenchmarkOrderRepository_Save(b *testing.B) {
	db := setupBenchmarkDB()
	defer db.Close()

	ctx := context.Background()

	repo := NewOrderRepository(db)
	order := CreateBenchmarkOrder()

	for i := 0; b.Loop(); i++ {
		order.OrderUID = fmt.Sprintf("benchmark-order-%d", i)
		err := repo.Save(ctx, order)
		if err != nil {
			b.Fatalf("Save failed: %v", err)
		}
	}
}

func BenchmarkOrderRepository_FindByID(b *testing.B) {
	db := setupBenchmarkDB()
	defer db.Close()

	ctx := context.Background()

	repo := NewOrderRepository(db)

	order := CreateBenchmarkOrder()
	err := repo.Save(ctx, order)
	if err != nil {
		b.Fatalf("Setup failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.FindByID(ctx, order.OrderUID)
		if err != nil {
			b.Fatalf("FindByID failed: %v", err)
		}
	}
}

func BenchmarkOrderRepository_FindAll(b *testing.B) {
	db := setupBenchmarkDB()
	defer db.Close()

	repo := NewOrderRepository(db)
	ctx := context.Background()

	for i := range 100 {
		order := createTestOrder()
		order.OrderUID = fmt.Sprintf("benchmark-order-%d", i)
		err := repo.Save(ctx, order)
		if err != nil {
			b.Fatalf("Setup failed: %v", err)
		}
	}

	for b.Loop() {
		_, err := repo.FindAll(ctx)
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

	tables := []string{"items", "payments", "deliveries", "orders"}
	for _, table := range tables {
		db.Exec("DELETE FROM " + table)
	}

	return db
}
