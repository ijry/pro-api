package oauth

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

// Repository 是 oauth_bindings CRUD 抽象。
type Repository interface {
	Create(ctx context.Context, b *Binding) error
	FindByProviderUID(ctx context.Context, provider, uid string) (*Binding, error)
	ListByUser(ctx context.Context, userID int64) ([]*Binding, error)
	GetByUserProvider(ctx context.Context, userID int64, provider string) (*Binding, error)
	DeleteByUserProvider(ctx context.Context, userID int64, provider string) error
	Update(ctx context.Context, b *Binding) error
}

type repo struct{ db *gorm.DB }

// NewRepository 构造仓储。
func NewRepository(db *gorm.DB) Repository { return &repo{db: db} }

func (r *repo) Create(ctx context.Context, b *Binding) error {
	return r.db.WithContext(ctx).Create(b).Error
}

func (r *repo) FindByProviderUID(ctx context.Context, provider, uid string) (*Binding, error) {
	var b Binding
	err := r.db.WithContext(ctx).
		Where("provider = ? AND provider_uid = ?", provider, uid).
		First(&b).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &b, nil
}

func (r *repo) ListByUser(ctx context.Context, userID int64) ([]*Binding, error) {
	var items []*Binding
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&items).Error
	return items, err
}

func (r *repo) GetByUserProvider(ctx context.Context, userID int64, provider string) (*Binding, error) {
	var b Binding
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND provider = ?", userID, provider).
		First(&b).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &b, nil
}

func (r *repo) DeleteByUserProvider(ctx context.Context, userID int64, provider string) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND provider = ?", userID, provider).
		Delete(&Binding{}).Error
}

func (r *repo) Update(ctx context.Context, b *Binding) error {
	return r.db.WithContext(ctx).Save(b).Error
}
