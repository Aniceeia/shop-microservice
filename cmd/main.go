package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"shop-microservice/internal/di"

	"go.uber.org/fx"
)

// @title Shop Microservice API
// @version 1.0
// @description Microservice for order management with Kafka, PostgreSQL and caching
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8081
// @BasePath /api

// @tag.name orders
// @tag.description Order management operations

// @tag.name health
// @tag.description Health check endpoints

// @tag.name tests
// @tag.description System test endpoints
func main() {
	app := fx.New(di.Module)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := app.Start(ctx); err != nil {
		log.Fatal("Failed to start application:", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := app.Stop(shutdownCtx); err != nil {
		log.Fatal("Failed to stop application:", err)
	}
}
