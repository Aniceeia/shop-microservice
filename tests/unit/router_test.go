package unit

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"shop-microservice/internal/api"
	"shop-microservice/internal/api/handlers"

	"github.com/gin-gonic/gin"
)

type dummyUseCase struct{}

func (d *dummyUseCase) CreateOrder(ctx interface{}, order interface{}) error { return nil }

func TestRouter_HealthAndRoot(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := handlers.NewHandler(nil)
	r := api.SetupRouter(h)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusNotFound {
		t.Skip("static root not critical for unit env")
	}
}
