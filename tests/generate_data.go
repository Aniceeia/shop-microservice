package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"
)

type Order struct {
	OrderUID          string   `json:"order_uid"`
	TrackNumber       string   `json:"track_number"`
	Entry             string   `json:"entry"`
	Delivery          Delivery `json:"delivery"`
	Payment           Payment  `json:"payment"`
	Items             []Item   `json:"items"`
	Locale            string   `json:"locale"`
	InternalSignature string   `json:"internal_signature"`
	CustomerID        string   `json:"customer_id"`
	DeliveryService   string   `json:"delivery_service"`
	Shardkey          string   `json:"shardkey"`
	SmID              int      `json:"sm_id"`
	DateCreated       string   `json:"date_created"`
	OofShard          string   `json:"oof_shard"`
}

type Delivery struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Zip     string `json:"zip"`
	City    string `json:"city"`
	Address string `json:"address"`
	Region  string `json:"region"`
	Email   string `json:"email"`
}

type Payment struct {
	Transaction  string `json:"transaction"`
	RequestID    string `json:"request_id"`
	Currency     string `json:"currency"`
	Provider     string `json:"provider"`
	Amount       int    `json:"amount"`
	PaymentDt    int64  `json:"payment_dt"`
	Bank         string `json:"bank"`
	DeliveryCost int    `json:"delivery_cost"`
	GoodsTotal   int    `json:"goods_total"`
	CustomFee    int    `json:"custom_fee"`
}

type Item struct {
	ChrtID      int    `json:"chrt_id"`
	TrackNumber string `json:"track_number"`
	Price       int    `json:"price"`
	Rid         string `json:"rid"`
	Name        string `json:"name"`
	Sale        int    `json:"sale"`
	Size        string `json:"size"`
	TotalPrice  int    `json:"total_price"`
	NmID        int    `json:"nm_id"`
	Brand       string `json:"brand"`
	Status      int    `json:"status"`
}

func generateRandomOrder() Order {
	rand.Seed(time.Now().UnixNano())

	order := Order{
		OrderUID:          fmt.Sprintf("test%d", rand.Intn(1000000)),
		TrackNumber:       fmt.Sprintf("WBILM%d", rand.Intn(1000000)),
		Entry:             "WBIL",
		Locale:            "en",
		InternalSignature: "",
		CustomerID:        fmt.Sprintf("customer%d", rand.Intn(1000)),
		DeliveryService:   "meest",
		Shardkey:          fmt.Sprintf("%d", rand.Intn(10)),
		SmID:              rand.Intn(100),
		DateCreated:       time.Now().Format(time.RFC3339),
		OofShard:          "1",
	}

	// Delivery
	order.Delivery.Name = fmt.Sprintf("Test User%d", rand.Intn(1000))
	order.Delivery.Phone = fmt.Sprintf("+972%d", rand.Intn(1000000000))
	order.Delivery.Zip = fmt.Sprintf("%d", rand.Intn(10000000))
	order.Delivery.City = "Test City"
	order.Delivery.Address = fmt.Sprintf("Test Address %d", rand.Intn(100))
	order.Delivery.Region = "Test Region"
	order.Delivery.Email = fmt.Sprintf("test%d@gmail.com", rand.Intn(1000))

	// Payment
	order.Payment.Transaction = order.OrderUID
	order.Payment.RequestID = ""
	order.Payment.Currency = "USD"
	order.Payment.Provider = "wbpay"
	order.Payment.Amount = rand.Intn(10000) + 1000
	order.Payment.PaymentDt = time.Now().Unix()
	order.Payment.Bank = "alpha"
	order.Payment.DeliveryCost = rand.Intn(500) + 100
	order.Payment.GoodsTotal = order.Payment.Amount - order.Payment.DeliveryCost
	order.Payment.CustomFee = 0

	// Items
	itemsCount := rand.Intn(3) + 1
	for i := 0; i < itemsCount; i++ {
		item := Item{
			ChrtID:      rand.Intn(10000000),
			TrackNumber: order.TrackNumber,
			Price:       rand.Intn(1000) + 100,
			Rid:         fmt.Sprintf("rid%d", rand.Intn(1000000)),
			Name:        fmt.Sprintf("Product %d", i+1),
			Sale:        rand.Intn(50),
			Size:        "0",
			TotalPrice:  rand.Intn(900) + 100,
			NmID:        rand.Intn(1000000),
			Brand:       "Test Brand",
			Status:      202,
		}
		order.Items = append(order.Items, item)
	}

	return order
}

func generateTestData(count int, outputDir string) error {
	for i := 0; i < count; i++ {
		order := generateRandomOrder()

		data, err := json.MarshalIndent(order, "", "  ")
		if err != nil {
			return err
		}

		filename := fmt.Sprintf("%s/order_%d.json", outputDir, i)
		err = os.WriteFile(filename, data, 0644)
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	err := os.MkdirAll("tests/fixtures/orders", 0755)
	if err != nil {
		panic(err)
	}

	err = generateTestData(100000, "tests/fixtures/orders")
	if err != nil {
		panic(err)
	}

	fmt.Println("Generated 100000 test orders in tests/fixtures/orders/")
}
