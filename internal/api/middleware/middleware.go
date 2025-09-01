package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func ValidateOrderIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderID := c.Param("id")
		if orderID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Order ID is required",
			})
			c.Abort()
			return
		}

		if err := ValidateOrderID(orderID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid order ID",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
