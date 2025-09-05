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

func main() {
	app := createApp()

	if err := startApp(app); err != nil {
		log.Fatal("Failed to start application:", err)
	}

	waitForShutdownSignal()

	if err := stopApp(app); err != nil {
		log.Fatal("Failed to stop application:", err)
	}
}

func createApp() *fx.App {
	return fx.New(di.Module)
}

func startApp(app *fx.App) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return app.Start(ctx)
}

func stopApp(app *fx.App) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return app.Stop(ctx)
}

func waitForShutdownSignal() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("closing...")
}
