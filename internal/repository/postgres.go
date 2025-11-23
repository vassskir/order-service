package repository

import (
	"L0/internal/config"
	"L0/internal/interfaces"
	"L0/internal/logger"
	"L0/internal/models"
	"L0/internal/validation"
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

var _ interfaces.OrderRepository = (*PostgresRepository)(nil)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(cfg *config.DatabaseConfig) (*PostgresRepository, error) {
	connStr := cfg.GetDBConnectionString()

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Successfully connected to PostgreSQL")
	return &PostgresRepository{db: db}, nil
}

func (r *PostgresRepository) Close() error {
	return r.db.Close()
}

func (r *PostgresRepository) SaveOrder(order *models.Order) error {
	if err := validation.ValidateOrder(order); err != nil {
		return fmt.Errorf("order validation failed: %w", err)
	}

	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature, 
			customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (order_uid) DO NOTHING`,
		order.OrderUID, order.TrackNumber, order.Entry, order.Locale,
		order.InternalSignature, order.CustomerID, order.DeliveryService,
		order.Shardkey, order.SmID, order.DateCreated, order.OofShard,
	)
	if err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}

	_, err = tx.Exec(`
		INSERT INTO deliveries (order_uid, name, phone, zip, city, address, region, email)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (order_uid) DO NOTHING`,
		order.OrderUID, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip,
		order.Delivery.City, order.Delivery.Address, order.Delivery.Region, order.Delivery.Email,
	)
	if err != nil {
		return fmt.Errorf("failed to insert delivery: %w", err)
	}

	_, err = tx.Exec(`
		INSERT INTO payments (order_uid, transaction, request_id, currency, provider, 
			amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (order_uid) DO NOTHING`,
		order.OrderUID, order.Payment.Transaction, order.Payment.RequestID,
		order.Payment.Currency, order.Payment.Provider, order.Payment.Amount,
		order.Payment.PaymentDt, order.Payment.Bank, order.Payment.DeliveryCost,
		order.Payment.GoodsTotal, order.Payment.CustomFee,
	)
	if err != nil {
		return fmt.Errorf("failed to insert payment: %w", err)
	}

	for _, item := range order.Items {
		_, err = tx.Exec(`
			INSERT INTO items (order_uid, chrt_id, track_number, price, rid, name, 
				sale, size, total_price, nm_id, brand, status)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
			order.OrderUID, item.ChrtID, item.TrackNumber, item.Price, item.Rid,
			item.Name, item.Sale, item.Size, item.TotalPrice, item.NmID,
			item.Brand, item.Status,
		)
		if err != nil {
			return fmt.Errorf("failed to insert item: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *PostgresRepository) GetOrderByUID(orderUID string) (*models.Order, error) {
	var order models.Order

	err := r.db.QueryRow(`
		SELECT order_uid, track_number, entry, locale, internal_signature,
			   customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
		FROM orders WHERE order_uid = $1`, orderUID).Scan(
		&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale,
		&order.InternalSignature, &order.CustomerID, &order.DeliveryService,
		&order.Shardkey, &order.SmID, &order.DateCreated, &order.OofShard,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("order not found")
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	err = r.db.QueryRow(`
		SELECT name, phone, zip, city, address, region, email
		FROM deliveries WHERE order_uid = $1`, orderUID).Scan(
		&order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip,
		&order.Delivery.City, &order.Delivery.Address, &order.Delivery.Region,
		&order.Delivery.Email,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get delivery: %w", err)
	}

	err = r.db.QueryRow(`
		SELECT transaction, request_id, currency, provider, amount, payment_dt,
			   bank, delivery_cost, goods_total, custom_fee
		FROM payments WHERE order_uid = $1`, orderUID).Scan(
		&order.Payment.Transaction, &order.Payment.RequestID, &order.Payment.Currency,
		&order.Payment.Provider, &order.Payment.Amount, &order.Payment.PaymentDt,
		&order.Payment.Bank, &order.Payment.DeliveryCost, &order.Payment.GoodsTotal,
		&order.Payment.CustomFee,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	rows, err := r.db.Query(`
		SELECT chrt_id, track_number, price, rid, name, sale, size,
			   total_price, nm_id, brand, status
		FROM items WHERE order_uid = $1`, orderUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item models.Item
		err := rows.Scan(
			&item.ChrtID, &item.TrackNumber, &item.Price, &item.Rid, &item.Name,
			&item.Sale, &item.Size, &item.TotalPrice, &item.NmID, &item.Brand, &item.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan item: %w", err)
		}
		order.Items = append(order.Items, item)
	}

	return &order, nil
}

func (r *PostgresRepository) GetAllOrders() ([]*models.Order, error) {
	rows, err := r.db.Query("SELECT order_uid FROM orders")
	if err != nil {
		return nil, fmt.Errorf("failed to get order IDs: %w", err)
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		var orderUID string
		if err := rows.Scan(&orderUID); err != nil {
			return nil, fmt.Errorf("failed to scan order ID: %w", err)
		}

		order, err := r.GetOrderByUID(orderUID)
		if err != nil {
			logger.Warn("Failed to get order", map[string]interface{}{
				"order_uid": orderUID,
				"error":     err.Error(),
			})
			continue
		}
		orders = append(orders, order)
	}

	return orders, nil
}
