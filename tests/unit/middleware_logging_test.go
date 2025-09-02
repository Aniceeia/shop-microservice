package unit

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"shop-microservice/internal/api/middleware"
	"shop-microservice/internal/infrastructure/logger"
	"shop-microservice/internal/infrastructure/metrics"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestLoggingMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	log := logger.NewLogger(logger.INFO)
	metrics := metrics.NewMetrics()

	router := gin.New()
	router.Use(middleware.LoggingMiddleware(log, metrics))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", "test-agent")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	stats := metrics.GetStats()
	requests := stats["requests"].(map[string]interface{})
	assert.Equal(t, float64(1), requests["/test"])
}

func TestLoggingMiddleware_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)

	log := logger.NewLogger(logger.INFO)
	metrics := metrics.NewMetrics()

	router := gin.New()
	router.Use(middleware.LoggingMiddleware(log, metrics))

	router.GET("/error", func(c *gin.Context) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "test error"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/error", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	stats := metrics.GetStats()
	requests := stats["requests"].(map[string]interface{})
	errors := stats["errors"].(map[string]interface{})

	assert.Equal(t, float64(1), requests["/error"])
	assert.Equal(t, float64(1), errors["/error"])
}

func TestLoggingMiddleware_RequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	log := logger.NewLogger(logger.INFO)
	metrics := metrics.NewMetrics()

	router := gin.New()
	router.Use(middleware.LoggingMiddleware(log, metrics))

	router.GET("/test", func(c *gin.Context) {
		requestID := c.Request.Context().Value("request_id")
		assert.NotNil(t, requestID)
		c.JSON(http.StatusOK, gin.H{"request_id": requestID})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
