package unit

import (
	"testing"
	"time"

	"shop-microservice/internal/domain/model"

	"github.com/stretchr/testify/assert"
)

func TestValidateOrder_Valid(t *testing.T) {
	order := model.Order{
		OrderUID:    "test123ddfwQf",
		TrackNumber: "WBILMTESTTRACK",
		Entry:       "WBIL",
		CustomerID:  "test",
		DateCreated: time.Now(),
		Delivery: model.Delivery{
			Name:    "Test Testov",
			Phone:   "+9720000000",
			Zip:     "2639809",
			City:    "Kiryat Mozkin",
			Address: "Ploshad Mira 15",
			Region:  "Kraiot",
			Email:   "test@gmail.com",
		},
		Payment: model.Payment{
			Transaction:  "test123",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       1817,
			PaymentDt:    time.Now().Unix(),
			Bank:         "alpha",
			DeliveryCost: 1500,
			GoodsTotal:   317,
			CustomFee:    0,
		},
		Items: []model.Item{
			{
				ChrtID:      9934930,
				TrackNumber: "WBILMTESTTRACK",
				Price:       453,
				Rid:         "ab4219087a764ae0btest",
				Name:        "Mascaras",
				Sale:        30,
				Size:        "0",
				TotalPrice:  317,
				NmID:        2389212,
				Brand:       "Vivienne Sabo",
				Status:      202,
			},
		},
	}

	err := validateOrder(&order)
	assert.NoError(t, err)
}

func validateOrder(order *model.Order) error {
	if order.OrderUID == "" {
		return assert.AnError
	}
	if order.TrackNumber == "" {
		return assert.AnError
	}
	if order.Entry == "" {
		return assert.AnError
	}
	if order.Delivery.Name == "" {
		return assert.AnError
	}
	if order.Payment.Amount <= 0 {
		return assert.AnError
	}
	if len(order.Items) == 0 {
		return assert.AnError
	}
	return nil
}
