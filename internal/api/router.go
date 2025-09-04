package api

import (
	"path/filepath"
	"runtime"
	"shop-microservice/internal/api/handlers"
	"shop-microservice/internal/api/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter(handler *handlers.Handler) *gin.Engine {
	router := gin.Default()

	router.Use(middleware.PrometheusMiddleware())

	_, filename, _, _ := runtime.Caller(0)
	rootDir := filepath.Join(filepath.Dir(filename), "../static")
	router.Static("/static", rootDir)

	router.GET("/", func(c *gin.Context) {
		c.File(filepath.Join(rootDir, "index.html"))
	})
	api := router.Group("/api")
	{
		api.POST("/orders", handler.CreateOrder)
		api.GET("/orders", handler.GetAllOrders)
		api.GET("/health", handler.HealthCheck)
		api.GET("/orders/:id", middleware.ValidateOrderIDMiddleware(), handler.GetOrderByID)
		router.GET("/metrics", middleware.GetPrometheusMetrics)
	}

	return router
}
