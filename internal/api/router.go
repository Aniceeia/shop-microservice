package api

import (
	"path/filepath"
	"runtime"

	"github.com/gin-gonic/gin"
)

func SetupRouter(handler *Handler) *gin.Engine {
	router := gin.Default()

	// Serve static files
	_, filename, _, _ := runtime.Caller(0)
	rootDir := filepath.Join(filepath.Dir(filename), "../../../static")
	router.Static("/static", rootDir)

	router.GET("/", func(c *gin.Context) {
		c.File(filepath.Join(rootDir, "index.html"))
	})

	api := router.Group("/api")
	{
		api.POST("/orders", handler.CreateOrder)
		api.GET("/orders/:id", handler.GetOrderByID)
		api.GET("/orders", handler.GetAllOrders)
		api.GET("/health", handler.HealthCheck)
	}

	return router
}
