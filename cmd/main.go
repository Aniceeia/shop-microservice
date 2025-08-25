package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"shop-microservice/internal/api"
	"shop-microservice/internal/infrastructure/postgresql"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "your_user"
	password = "your_password"
	dbname   = "orders_db"
)

func main() {
	// Подключаемся к БД
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Проверяем соединение
	if err := db.Ping(); err != nil {
		log.Fatal("Database ping failed:", err)
	}

	// Запускаем миграции
	if err := postgresql.RunMigrations(db); err != nil {
		log.Fatal("Migrations failed:", err)
	}

	// Инициализируем репозиторий
	repo := postgresql.NewOrderRepository(db)

	// Инициализируем handler
	handler := api.NewHandler(repo)

	// Настраиваем роутер
	router := api.SetupRouter(handler)

	// Запускаем сервер
	log.Println("Server starting on :8081")
	if err := http.ListenAndServe(":8081", router); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
