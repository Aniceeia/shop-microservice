package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"shop-microservice/internal/api"
	"shop-microservice/internal/infrastructure/cash"
	"shop-microservice/internal/infrastructure/kafka"
	"shop-microservice/internal/infrastructure/postgresql"
	"strconv"
	"strings"
	"time"

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
	dbHost := getEnv("DB_HOST", "postgres")
	dbPortStr := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "orders_user")
	dbPassword := getEnv("DB_PASSWORD", "orders_password")
	dbName := getEnv("DB_NAME", "orders_db")
	appPort := getEnv("APP_PORT", "8081")

	kafkaBrokers := getEnv("KAFKA_BROKERS", "kafka:9092")
	kafkaTopic := getEnv("KAFKA_TOPIC", "orders")

	if dbHost == "" || dbPortStr == "" || dbUser == "" || dbPassword == "" || dbName == "" {
		log.Fatal("Missing required database environment variables")
	}

	dbPort, err := strconv.Atoi(dbPortStr)
	if err != nil {
		log.Fatal("Invalid DB_PORT:", err)
	}

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Database ping failed:", err)
	}

	if err := postgresql.RunMigrations(db); err != nil {
		log.Fatal("Migrations failed:", err)
	}

	brokers := strings.Split(kafkaBrokers, ",")
	kafkaManager := kafka.NewKafkaManager(brokers)

	log.Println("Waiting for Kafka to be available...")
	if err := kafkaManager.WaitForKafka(30 * time.Second); err != nil {
		log.Fatal("Kafka not available:", err)
	}

	if err := kafkaManager.CreateTopicIfNotExists(kafkaTopic, 3, 1); err != nil {
		log.Printf("Warning: failed to create topic: %v", err)
	}

	repo := postgresql.NewOrderRepository(db)
	cash := cash.NewCash()

	kafkaProducer := kafka.NewProducer(kafka.ProducerConfig{
		Brokers: brokers,
		Topic:   kafkaTopic,
	})
	defer kafkaProducer.Close()

	if err := cash.WarmUp(repo); err != nil {
		log.Printf("Warning: cache warm-up failed: %v", err)
	} else {
		log.Printf("Cache initialized with %d orders", cash.Size())
	}

	handler := api.NewHandler(repo, kafkaProducer, cash)

	router := api.SetupRouter(handler)

	if appPort == "" {
		appPort = "8081"
	}

	log.Printf("Server starting on :%s", appPort)
	if err := http.ListenAndServe(":"+appPort, router); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
