package online

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// Repository is the CRUD interface for payment_orders.
type Repository interface {
	Create(ctx context.Context, o *Order) error
	FindByOutTradeNo(ctx context.Context, no string) (*Order, error)
	UpdateStatus(ctx context.Context, id int64, status OrderStatus, providerOrderID string, paidAt *time.Time) error
	ListByUser(ctx context.Context, userID int64, limit, offset int) ([]*Order, error)
}

type gormRepo struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &gormRepo{db: db} }

func (r *gormRepo) Create(ctx context.Context, o *Order) error {
	return r.db.WithContext(ctx).Create(o).Error
}

func (r *gormRepo) FindByOutTradeNo(ctx context.Context, no string) (*Order, error) {
	var o Order
	if err := r.db.WithContext(ctx).Where("out_trade_no = ?", no).First(&o).Error; err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *gormRepo) UpdateStatus(ctx context.Context, id int64, status OrderStatus, providerOrderID string, paidAt *time.Time) error {
	return r.db.WithContext(ctx).Model(&Order{}).Where("id = ?", id).Updates(map[string]any{
		"status":            status,
		"provider_order_id": providerOrderID,
		"paid_at":           paidAt,
	}).Error
}

func (r *gormRepo) ListByUser(ctx context.Context, userID int64, limit, offset int) ([]*Order, error) {
	var list []*Order
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&list).Error
	return list, err
}
