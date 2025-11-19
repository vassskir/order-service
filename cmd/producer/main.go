package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)


type Order struct {
	OrderUID          string    `json:"order_uid"`
	TrackNumber       string    `json:"track_number"`
	Entry             string    `json:"entry"`
	Delivery          Delivery  `json:"delivery"`
	Payment           Payment   `json:"payment"`
	Items             []Item    `json:"items"`
	Locale            string    `json:"locale"`
	InternalSignature string    `json:"internal_signature"`
	CustomerID        string    `json:"customer_id"`
	DeliveryService   string    `json:"delivery_service"`
	Shardkey          string    `json:"shardkey"`
	SmID              int       `json:"sm_id"`
	DateCreated       time.Time `json:"date_created"`
	OofShard          string    `json:"oof_shard"`
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
	ChrtID      int64  `json:"chrt_id"`
	TrackNumber string `json:"track_number"`
	Price       int    `json:"price"`
	Rid         string `json:"rid"`
	Name        string `json:"name"`
	Sale        int    `json:"sale"`
	Size        string `json:"size"`
	TotalPrice  int    `json:"total_price"`
	NmID        int64  `json:"nm_id"`
	Brand       string `json:"brand"`
	Status      int    `json:"status"`
}

func main() {
	brokers := []string{"localhost:9092"}
	topic := "orders"

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: brokers,
		Topic:   topic,
	})
	defer writer.Close()

	order := createTestOrder()

	jsonData, err := json.Marshal(order)
	if err != nil {
		log.Fatalf("Failed to marshal order: %v", err)
	}

	err = writer.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte(order.OrderUID),
			Value: jsonData,
		},
	)
	if err != nil {
		log.Fatalf("Failed to write message: %v", err)
	}

	fmt.Printf("✅ Order sent to Kafka: %s\n", order.OrderUID)
	fmt.Printf("📦 Track number: %s\n", order.TrackNumber)
	fmt.Printf("👤 Customer: %s\n", order.Delivery.Name)
}

func createTestOrder() Order {
	now := time.Now()
	order := Order{
		OrderUID:          fmt.Sprintf("test-order-%d", now.Unix()),
		TrackNumber:       "WBILMTESTTRACK",
		Entry:             "WBIL",
		Locale:            "en",
		InternalSignature: "",
		CustomerID:        "test-customer",
		DeliveryService:   "meest",
		Shardkey:          "9",
		SmID:              99,
		DateCreated:       now,
		OofShard:          "1",
	}

	order.Delivery.Name = "Test Testov"
	order.Delivery.Phone = "+9720000000"
	order.Delivery.Zip = "2639809"
	order.Delivery.City = "Kiryat Mozkin"
	order.Delivery.Address = "Ploshad Mira 15"
	order.Delivery.Region = "Kraiot"
	order.Delivery.Email = "test@gmail.com"

	order.Payment.Transaction = order.OrderUID
	order.Payment.RequestID = ""
	order.Payment.Currency = "USD"
	order.Payment.Provider = "wbpay"
	order.Payment.Amount = 1817
	order.Payment.PaymentDt = now.Unix()
	order.Payment.Bank = "alpha"
	order.Payment.DeliveryCost = 1500
	order.Payment.GoodsTotal = 317
	order.Payment.CustomFee = 0

	order.Items = []Item{
		{
			ChrtID:      9934930,
			TrackNumber: order.TrackNumber,
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
	}
	return order
}