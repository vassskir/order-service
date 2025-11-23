package models

import (
	"time"
)

type Order struct {
	OrderUID          string    `json:"order_uid" db:"order_uid" validate:"required,min=1,max=255"`
	TrackNumber       string    `json:"track_number" db:"track_number" validate:"required,track_number"`
	Entry             string    `json:"entry" db:"entry" validate:"required,min=1,max=50"`
	Delivery          Delivery  `json:"delivery" db:"-"`
	Payment           Payment   `json:"payment" db:"-"`
	Items             []Item    `json:"items" db:"-"`
	Locale            string    `json:"locale" db:"locale" validate:"required,oneof=en ru"`
	InternalSignature string    `json:"internal_signature" db:"internal_signature" validate:"max=255"`
	CustomerID        string    `json:"customer_id" db:"customer_id" validate:"required,min=1,max=255"`
	DeliveryService   string    `json:"delivery_service" db:"delivery_service" validate:"required,min=1,max=100"`
	Shardkey          string    `json:"shardkey" db:"shardkey" validate:"max=50"`
	SmID              int       `json:"sm_id" db:"sm_id" validate:"min=0"`
	DateCreated       time.Time `json:"date_created" db:"date_created" validate:"required"`
	OofShard          string    `json:"oof_shard" db:"oof_shard" validate:"max=50"`
}

type Delivery struct {
	OrderUID string `json:"-" db:"order_uid"`
	Name     string `json:"name" db:"name" validate:"required,min=1,max=255"`
	Phone    string `json:"phone" db:"phone" validate:"required,phone"`
	Zip      string `json:"zip" db:"zip" validate:"max=50"`
	City     string `json:"city" db:"city" validate:"required,max=255"`
	Address  string `json:"address" db:"address" validate:"required"`
	Region   string `json:"region" db:"region" validate:"max=255"`
	Email    string `json:"email" db:"email" validate:"required,email"`
}

type Payment struct {
	OrderUID     string `json:"-" db:"order_uid"`
	Transaction  string `json:"transaction" db:"transaction" validate:"required,min=1,max=255"`
	RequestID    string `json:"request_id" db:"request_id" validate:"max=255"`
	Currency     string `json:"currency" db:"currency" validate:"required,iso4217"`
	Provider     string `json:"provider" db:"provider" validate:"required,min=1,max=100"`
	Amount       int    `json:"amount" db:"amount" validate:"min=0"`
	PaymentDt    int64  `json:"payment_dt" db:"payment_dt" validate:"min=0"`
	Bank         string `json:"bank" db:"bank" validate:"max=100"`
	DeliveryCost int    `json:"delivery_cost" db:"delivery_cost" validate:"min=0"`
	GoodsTotal   int    `json:"goods_total" db:"goods_total" validate:"min=0"`
	CustomFee    int    `json:"custom_fee" db:"custom_fee" validate:"min=0"`
}

type Item struct {
	ID          int    `json:"-" db:"id"`
	OrderUID    string `json:"-" db:"order_uid"`
	ChrtID      int64  `json:"chrt_id" db:"chrt_id" validate:"min=0"`
	TrackNumber string `json:"track_number" db:"track_number" validate:"required,track_number"`
	Price       int    `json:"price" db:"price" validate:"min=0"`
	Rid         string `json:"rid" db:"rid" validate:"required,max=255"`
	Name        string `json:"name" db:"name" validate:"required,max=255"`
	Sale        int    `json:"sale" db:"sale" validate:"min=0"`
	Size        string `json:"size" db:"size" validate:"max=50"`
	TotalPrice  int    `json:"total_price" db:"total_price" validate:"min=0"`
	NmID        int64  `json:"nm_id" db:"nm_id" validate:"min=0"`
	Brand       string `json:"brand" db:"brand" validate:"max=255"`
	Status      int    `json:"status" db:"status" validate:"min=0"`
}
