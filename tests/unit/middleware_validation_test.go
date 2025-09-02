package unit

import (
	"testing"

	"shop-microservice/internal/api/middleware"

	"github.com/stretchr/testify/assert"
)

func TestValidateOrderID_Success(t *testing.T) {
	assert.NoError(t, middleware.ValidateOrderID("order-ABC_123"))
	assert.NoError(t, middleware.ValidateOrderID("uid_123456"))
}

func TestValidateOrderID_Failures(t *testing.T) {
	assert.Error(t, middleware.ValidateOrderID("short"))
	long := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	assert.Error(t, middleware.ValidateOrderID(long))
	assert.Error(t, middleware.ValidateOrderID("bad space"))
	assert.Error(t, middleware.ValidateOrderID("1234567"))
	assert.Error(t, middleware.ValidateOrderID("test123"))
}
