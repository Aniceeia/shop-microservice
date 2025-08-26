package postgresql

import (
	"fmt"
	"shop-microservice/internal/domain/model"
	"time"
)

// ExampleOrderJSON возвращает пример JSON для тестирования
func ExampleOrderJSON() string {
	return `{
		"order_uid": "b563feb7b2b84b6test",
		"track_number": "WBILMTESTTRACK",
		"entry": "WBIL",
		"delivery": {
			"name": "Test Testov",
			"phone": "+9720000000",
			"zip": "2639809",
			"city": "Kiryat Mozkin",
			"address": "Ploshad Mira 15",
			"region": "Kraiot",
			"email": "test@gmail.com"
		},
		"payment": {
			"transaction": "b563feb7b2b84b6test",
			"request_id": "",
			"currency": "USD",
			"provider": "wbpay",
			"amount": 1817,
			"payment_dt": 1637907727,
			"bank": "alpha",
			"delivery_cost": 1500,
			"goods_total": 317,
			"custom_fee": 0
		},
		"items": [
			{
				"chrt_id": 9934930,
				"track_number": "WBILMTESTTRACK",
				"price": 453,
				"rid": "ab4219087a764ae0btest",
				"name": "Mascaras",
				"sale": 30,
				"size": "0",
				"total_price": 317,
				"nm_id": 2389212,
				"brand": "Vivienne Sabo",
				"status": 202
			}
		],
		"locale": "en",
		"internal_signature": "",
		"customer_id": "test",
		"delivery_service": "meest",
		"shardkey": "9",
		"sm_id": 99,
		"date_created": "2021-11-26T06:22:19Z",
		"oof_shard": "1"
	}`
}

// CreateBenchmarkOrder создает заказ для бенчмарк-тестов
func CreateBenchmarkOrder() *model.Order {
	order := createTestOrderDifferent()

	// Добавляем больше items для бенчмарков
	for i := 0; i < 100; i++ {
		order.Items = append(order.Items, model.Item{
			ChrtID:      i + 2,
			TrackNumber: fmt.Sprintf("item-track-%d", i+2),
			Price:       100 + i,
			Rid:         fmt.Sprintf("item-rid-%d", i+2),
			Name:        fmt.Sprintf("Test Item %d", i+2),
			Sale:        i % 30,
			Size:        "M",
			TotalPrice:  100 + i - (i % 30),
			NmID:        123 + i,
			Brand:       "Test Brand",
			Status:      (i % 3) + 1,
		})
	}

	return order
}

func createTestOrderDifferent() *model.Order {
	return &model.Order{
		OrderUID:          "test-order-uid",
		TrackNumber:       "test-track",
		Entry:             "test-entry",
		Locale:            "en",
		InternalSignature: "test-signature",
		CustomerID:        "test-customer",
		DeliveryService:   "test-service",
		Shardkey:          "test-shard",
		SmID:              1,
		DateCreated:       time.Now(),
		OofShard:          "test-oof",
		Delivery: model.Delivery{
			Name:    "John Doe",
			Phone:   "+1234567890",
			Zip:     "123456",
			City:    "Moscow",
			Address: "Street 1",
			Region:  "Moscow",
			Email:   "john@example.com",
		},
		Payment: model.Payment{
			Transaction:  "test-transaction",
			RequestID:    "test-request",
			Currency:     "USD",
			Provider:     "test-provider",
			Amount:       1000,
			PaymentDt:    time.Now().Unix(),
			Bank:         "test-bank",
			DeliveryCost: 500,
			GoodsTotal:   500,
			CustomFee:    0,
		},
		Items: []model.Item{
			{
				ChrtID:      1,
				TrackNumber: "item-track-1",
				Price:       100,
				Rid:         "item-rid-1",
				Name:        "Test Item 1",
				Sale:        0,
				Size:        "M",
				TotalPrice:  100,
				NmID:        123,
				Brand:       "Test Brand",
				Status:      1,
			},
		},
	}
}
