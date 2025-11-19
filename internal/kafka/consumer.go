package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"L0/internal/interfaces"
	"L0/internal/logger"
	"L0/internal/models"
	"L0/internal/retry"
	"L0/internal/validation"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader    *kafka.Reader
	repo      interfaces.OrderRepository
	cache     interfaces.OrderCache
	topic     string
	groupID   string
	brokers   []string
}

func NewConsumer(brokers []string, topic string, groupID string, repo interfaces.OrderRepository, cache interfaces.OrderCache) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          topic,
		GroupID:        groupID,
		MinBytes:       10e3, 
		MaxBytes:       10e6, 
		CommitInterval: time.Second,
		StartOffset:    kafka.FirstOffset,
	})

	return &Consumer{
		reader:  reader,
		repo:    repo,
		cache:   cache,
		topic:   topic,
		groupID: groupID,
		brokers: brokers,
	}
}

func (c *Consumer) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			logger.Info("Stopping Kafka consumer")
			return
		default:
			c.processMessage(ctx)
		}
	}
}

func (c *Consumer) processMessage(ctx context.Context) {
	msgCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	msg, err := c.reader.ReadMessage(msgCtx)
	if err != nil {
		logger.Warn("Error reading Kafka message", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	logger.Info("Received Kafka message", map[string]interface{}{
		"topic":     msg.Topic,
		"partition": msg.Partition,
		"offset":    msg.Offset,
	})

	if err := c.handleOrderMessage(msg.Value); err != nil {
		logger.Error("Error handling order message", err, map[string]interface{}{
			"offset": msg.Offset,
		})
		return
	}

	logger.Info("Successfully processed message", map[string]interface{}{
		"offset": msg.Offset,
	})
}

func (c *Consumer) handleOrderMessage(data []byte) error {
	var order models.Order
	if err := json.Unmarshal(data, &order); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	if err := validation.ValidateOrder(&order); err != nil {
		return fmt.Errorf("order validation failed: %w", err)
	}

	ctx := context.Background()
	err := retry.Retry(ctx, "save_order_to_db", retry.DefaultConfig, func() error {
		return c.repo.SaveOrder(&order)
	})
	
	if err != nil {
		return fmt.Errorf("failed to save order to database after retries: %w", err)
	}

	c.cache.Set(&order)

	logger.Info("Order processed successfully", map[string]interface{}{
		"order_uid":    order.OrderUID,
		"track_number": order.TrackNumber,
	})
	return nil
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}