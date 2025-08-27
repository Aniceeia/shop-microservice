package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"shop-microservice/internal/api"
	"shop-microservice/internal/infrastructure/postgresql"
	"strconv"

	_ "github.com/lib/pq"
)

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func main() {
	// Получаем переменные окружения
	dbHost := getEnv("DB_HOST", "postgres")
	dbPortStr := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "orders_user")
	dbPassword := getEnv("DB_PASSWORD", "orders_password")
	dbName := getEnv("DB_NAME", "orders_db")
	appPort := getEnv("APP_PORT", "8081")

	// Проверяем обязательные переменные
	if dbHost == "" || dbPortStr == "" || dbUser == "" || dbPassword == "" || dbName == "" {
		log.Fatal("Missing required database environment variables")
	}

	// Конвертируем порт в число
	dbPort, err := strconv.Atoi(dbPortStr)
	if err != nil {
		log.Fatal("Invalid DB_PORT:", err)
	}

	// Подключаемся к БД
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

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
	if appPort == "" {
		appPort = "8081" // значение по умолчанию
	}
	log.Printf("Server starting on :%s", appPort)
	if err := http.ListenAndServe(":"+appPort, router); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
