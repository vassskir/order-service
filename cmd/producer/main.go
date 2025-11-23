package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jaswdr/faker"
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
	interval := flag.Duration("interval", 10*time.Second, "Interval between order generation")
	count := flag.Int("count", 1, "Number of orders to generate (0 for continuous)")
	brokers := flag.String("brokers", "localhost:9092", "Kafka brokers (comma separated)")
	topic := flag.String("topic", "orders", "Kafka topic")
	flag.Parse()

	brokersList := []string{*brokers}

	writer := &kafka.Writer{
		Addr:                   kafka.TCP(brokersList...),
		Topic:                  *topic,
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,
	}
	defer writer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		log.Println("Received shutdown signal")
		cancel()
	}()

	fake := faker.New()
	ticker := time.NewTicker(*interval)
	defer ticker.Stop()

	ordersGenerated := 0
	log.Printf("Starting order producer (interval: %v, count: %d)", *interval, *count)

	for {
		select {
		case <-ctx.Done():
			log.Println("Producer stopped")
			return
		case <-ticker.C:
			if *count > 0 && ordersGenerated >= *count {
				log.Printf("Generated %d orders, stopping", *count)
				return
			}

			order := generateTestOrder(fake)
			jsonData, err := json.Marshal(order)
			if err != nil {
				log.Printf("Failed to marshal order: %v", err)
				continue
			}

			err = writer.WriteMessages(ctx,
				kafka.Message{
					Key:   []byte(order.OrderUID),
					Value: jsonData,
				},
			)
			if err != nil {
				log.Printf("Failed to write message: %v", err)
				continue
			}

			ordersGenerated++
			fmt.Printf("Order #%d sent: %s\n", ordersGenerated, order.OrderUID)
			fmt.Printf("Track: %s, Customer: %s\n", order.TrackNumber, order.Delivery.Name)
			fmt.Printf("Amount: %d %s, Items: %d\n", order.Payment.Amount, order.Payment.Currency, len(order.Items))
			fmt.Println("---")
		}
	}
}

func generateTestOrder(fake faker.Faker) Order {
	orderUID := fake.UUID().V4()
	orderTrackNumber := "WBIL" + fake.UUID().V4()[:8]

	delivery := Delivery{
		Name:    fake.Person().Name(),
		Phone:   "+" + fake.Numerify("###########"),
		Zip:     fake.Address().PostCode(),
		City:    fake.Address().City(),
		Address: fake.Address().Address(),
		Region:  fake.Address().State(),
		Email:   fake.Internet().Email(),
	}

	amount := fake.IntBetween(1000, 100000)

	// Фиксируем список валют которые проходят валидацию
	currencies := []string{"USD", "EUR", "RUB", "GBP", "JPY"}

	payment := Payment{
		Transaction:  orderUID,
		RequestID:    fake.UUID().V4(),
		Currency:     currencies[fake.IntBetween(0, len(currencies)-1)], // Используем фиксированные валюты
		Provider:     fake.Company().Name(),
		Amount:       amount,
		PaymentDt:    time.Now().Unix(),
		Bank:         fake.Company().Name(),
		DeliveryCost: fake.IntBetween(500, 5000),
		GoodsTotal:   amount - fake.IntBetween(100, 1000),
		CustomFee:    fake.IntBetween(0, 500),
	}

	sizes := []string{"S", "M", "L", "XL", "0"}
	locales := []string{"en", "ru"}
	deliveryServices := []string{"meest", "russianpost", "dhl", "fedex"}

	var items []Item
	numItems := fake.IntBetween(1, 3)
	for i := 0; i < numItems; i++ {
		price := fake.IntBetween(100, 5000)
		sale := fake.IntBetween(0, 50)

		item := Item{
			ChrtID:      fake.Int64(),
			TrackNumber: orderTrackNumber,
			Price:       price,
			Rid:         fake.UUID().V4(),
			Name:        "Product " + fake.Lorem().Word(),
			Sale:        sale,
			Size:        sizes[fake.IntBetween(0, len(sizes)-1)],
			TotalPrice:  price - (price * sale / 100),
			NmID:        fake.Int64(),
			Brand:       fake.Company().Name(),
			Status:      fake.IntBetween(200, 204),
		}
		items = append(items, item)
	}

	return Order{
		OrderUID:          orderUID,
		TrackNumber:       orderTrackNumber,
		Entry:             "WBIL",
		Delivery:          delivery,
		Payment:           payment,
		Items:             items,
		Locale:            locales[fake.IntBetween(0, len(locales)-1)],
		InternalSignature: "",
		CustomerID:        fake.UUID().V4(),
		DeliveryService:   deliveryServices[fake.IntBetween(0, len(deliveryServices)-1)],
		Shardkey:          fmt.Sprintf("%d", fake.IntBetween(1, 10)),
		SmID:              fake.IntBetween(1, 100),
		DateCreated:       time.Now(),
		OofShard:          fmt.Sprintf("%d", fake.IntBetween(1, 5)),
	}
}
