package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func SetupRouter(handler *Handler) *gin.Engine {
	router := gin.Default()

	// API routes
	api := router.Group("/api")
	{
		api.POST("/orders", handler.CreateOrder)
		api.GET("/orders/:id", handler.GetOrderByID)
		api.GET("/orders", handler.GetAllOrders)
		api.GET("/health", handler.HealthCheck)
	}

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "order service is running",
		})
	})

	return router
}
