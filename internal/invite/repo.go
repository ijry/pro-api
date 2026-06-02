package invite

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, r *Record) error
	ListByInviter(ctx context.Context, inviterID int64, limit, offset int) ([]*Record, error)
	CountByInviter(ctx context.Context, inviterID int64) (int64, error)
}

type gormRepo struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &gormRepo{db: db} }

func (r *gormRepo) Create(ctx context.Context, rec *Record) error {
	return r.db.WithContext(ctx).Create(rec).Error
}

func (r *gormRepo) ListByInviter(ctx context.Context, inviterID int64, limit, offset int) ([]*Record, error) {
	var list []*Record
	err := r.db.WithContext(ctx).
		Where("inviter_id = ?", inviterID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&list).Error
	return list, err
}

func (r *gormRepo) CountByInviter(ctx context.Context, inviterID int64) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&Record{}).
		Where("inviter_id = ?", inviterID).
		Count(&n).Error
	return n, err
}
