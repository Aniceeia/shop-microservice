package unit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"shop-microservice/internal/api/middleware"
)

func TestValidateOrderID_Success(t *testing.T) {
	validIDs := []string{"validsds123", "testqwe456", "order78qwe9"}

	for _, id := range validIDs {
		err := middleware.ValidateOrderID(id)
		assert.NoError(t, err, "ID %s should be valid", id)
	}
}

func TestValidateOrderID_Failures(t *testing.T) {
	invalidIDs := []string{"", "   ", "invalid@id", "id with spaces"}

	for _, id := range invalidIDs {
		err := middleware.ValidateOrderID(id)
		assert.Error(t, err, "ID %s should be invalid", id)
	}
}
