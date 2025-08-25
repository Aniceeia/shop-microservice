package api

import "github.com/gin-gonic/gin"

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

	return router
}
