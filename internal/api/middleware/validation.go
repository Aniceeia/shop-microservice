package middleware

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	orderIDRegex     = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	minOrderIDLength = 6
	maxOrderIDLength = 50
)

func ValidateOrderID(orderID string) error {
	if len(orderID) < minOrderIDLength {
		return fmt.Errorf("order ID must be at least %d characters long", minOrderIDLength)
	}

	if len(orderID) > maxOrderIDLength {
		return fmt.Errorf("order ID must be at most %d characters long", maxOrderIDLength)
	}

	if !orderIDRegex.MatchString(orderID) {
		return fmt.Errorf("order ID can only contain letters, numbers, underscores and hyphens")
	}

	// не менее 10
	if strings.Contains(orderID, "test") && len(orderID) < 10 {
		return fmt.Errorf("test order IDs should follow the pattern from test data")
	}

	//наличие только цифр (слишком простой ID)
	if isAllDigits(orderID) {
		return fmt.Errorf("order ID should not consist only of digits")
	}

	return nil
}

func isAllDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
