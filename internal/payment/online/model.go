package online

import (
	"encoding/json"
	"time"
)

// OrderStatus is the payment order status.
type OrderStatus string

const (
	StatusPending  OrderStatus = "pending"
	StatusPaid     OrderStatus = "paid"
	StatusFailed   OrderStatus = "failed"
	StatusRefunded OrderStatus = "refunded"
)

// Order is the DB model for payment_orders.
type Order struct {
	ID              int64           `gorm:"primaryKey"`
	UserID          int64           `gorm:"not null;index:idx_payment_orders_user_id"`
	Provider        string          `gorm:"size:32;not null"`
	OutTradeNo      string          `gorm:"size:64;not null;uniqueIndex"`
	ProviderOrderID string          `gorm:"size:128;not null;default:''"`
	AmountCents     int64           `gorm:"not null"`
	Currency        string          `gorm:"size:8;not null;default:CNY"`
	Status          OrderStatus     `gorm:"size:16;not null;default:pending;index:idx_payment_orders_status"`
	Credits         int64           `gorm:"not null;default:0"`
	Meta            json.RawMessage `gorm:"type:json"`
	PaidAt          *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (Order) TableName() string { return "payment_orders" }
