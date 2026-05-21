package session

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

// Repository 是 sessions 表(DB mirror)的 CRUD 抽象。
type Repository interface {
	Insert(ctx context.Context, s *DBSession) error
	Get(ctx context.Context, id string) (*DBSession, error)
	UpdateLastSeen(ctx context.Context, id string, lastSeen, expires time.Time) error
	MarkRevoked(ctx context.Context, id string, at time.Time) error
	MarkAllRevokedForUser(ctx context.Context, userID int64, at time.Time) error
	ListActive(ctx context.Context, now time.Time, limit int) ([]*DBSession, error)
	DeleteExpired(ctx context.Context, before time.Time, limit int) (int64, error)
}

type repo struct{ db *gorm.DB }

// NewRepository 构造 DB mirror 仓储。
func NewRepository(db *gorm.DB) Repository { return &repo{db: db} }

func (r *repo) Insert(ctx context.Context, s *DBSession) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *repo) Get(ctx context.Context, id string) (*DBSession, error) {
	var s DBSession
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&s).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *repo) UpdateLastSeen(ctx context.Context, id string, lastSeen, expires time.Time) error {
	return r.db.WithContext(ctx).Model(&DBSession{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"last_seen_at": lastSeen,
			"expires_at":   expires,
		}).Error
}

func (r *repo) MarkRevoked(ctx context.Context, id string, at time.Time) error {
	return r.db.WithContext(ctx).Model(&DBSession{}).
		Where("id = ? AND revoked_at IS NULL", id).
		Update("revoked_at", at).Error
}

func (r *repo) MarkAllRevokedForUser(ctx context.Context, userID int64, at time.Time) error {
	return r.db.WithContext(ctx).Model(&DBSession{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", at).Error
}

func (r *repo) ListActive(ctx context.Context, now time.Time, limit int) ([]*DBSession, error) {
	var items []*DBSession
	if limit <= 0 {
		limit = 1000
	}
	err := r.db.WithContext(ctx).
		Where("expires_at > ? AND revoked_at IS NULL", now).
		Limit(limit).
		Find(&items).Error
	return items, err
}

func (r *repo) DeleteExpired(ctx context.Context, before time.Time, limit int) (int64, error) {
	if limit <= 0 {
		limit = 1000
	}
	res := r.db.WithContext(ctx).
		Where("(expires_at < ?) OR (revoked_at IS NOT NULL AND revoked_at < ?)", before, before).
		Limit(limit).
		Delete(&DBSession{})
	return res.RowsAffected, res.Error
}
