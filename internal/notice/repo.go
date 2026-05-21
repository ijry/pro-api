package notice

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

// ErrNotFound 表示公告不存在(用于 Repo.Update / Repo.Delete)。
var ErrNotFound = errors.New("notice: not found")

// Repo 是 notices 表的最小仓储接口。
type Repo interface {
	Create(ctx context.Context, n *Notice) error
	GetByID(ctx context.Context, id int64) (*Notice, error)
	Update(ctx context.Context, id int64, fields map[string]any) error
	Delete(ctx context.Context, id int64) error

	ListAdmin(ctx context.Context, status int8, page, size int) ([]*Notice, int64, error)
	ListVisibleForUser(ctx context.Context, targets []string, now time.Time, page, size int) ([]*Notice, int64, error)
	CountVisibleForUser(ctx context.Context, targets []string, now time.Time) (int64, error)
	VisibleIDsForUser(ctx context.Context, targets []string, now time.Time) ([]int64, error)
}

type repo struct {
	db *gorm.DB
}

// NewRepo 构造默认 GORM 实现。
func NewRepo(db *gorm.DB) Repo { return &repo{db: db} }

// Create 插入一条 notice;调用方负责填充 ID/CreatedAt/UpdatedAt。
func (r *repo) Create(ctx context.Context, n *Notice) error {
	return r.db.WithContext(ctx).Create(n).Error
}

// GetByID 返回单条;不存在返 (nil, nil)。
func (r *repo) GetByID(ctx context.Context, id int64) (*Notice, error) {
	var n Notice
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&n).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &n, nil
}

// Update 仅更新 fields 中给出的列;若没有任何行被影响,返 ErrNotFound。
func (r *repo) Update(ctx context.Context, id int64, fields map[string]any) error {
	if len(fields) == 0 {
		return nil
	}
	res := r.db.WithContext(ctx).Model(&Notice{}).Where("id = ?", id).Updates(fields)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete 按 id 删除;不存在返 ErrNotFound。
func (r *repo) Delete(ctx context.Context, id int64) error {
	res := r.db.WithContext(ctx).Where("id = ?", id).Delete(&Notice{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

const (
	adminOrderBy = "pinned DESC, COALESCE(publish_at, created_at) DESC, id DESC"
)

// ListAdmin status<0 表示不筛。
func (r *repo) ListAdmin(ctx context.Context, status int8, page, size int) ([]*Notice, int64, error) {
	q := r.db.WithContext(ctx).Model(&Notice{})
	if status >= 0 {
		q = q.Where("status = ?", status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return nil, 0, nil
	}
	offset := computeOffset(page, size)
	limit := size
	if limit <= 0 {
		limit = 20
	}
	var items []*Notice
	if err := q.Order(adminOrderBy).Offset(offset).Limit(limit).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// visibleWhere 拼接"对用户可见"的过滤条件;返 *gorm.DB。
func (r *repo) visibleWhere(ctx context.Context, targets []string, now time.Time) *gorm.DB {
	return r.db.WithContext(ctx).Model(&Notice{}).
		Where("status = ?", StatusPublished).
		Where("(publish_at IS NULL OR publish_at <= ?)", now).
		Where("(expires_at IS NULL OR expires_at > ?)", now).
		Where("target IN ?", targets)
}

func (r *repo) ListVisibleForUser(ctx context.Context, targets []string, now time.Time, page, size int) ([]*Notice, int64, error) {
	q := r.visibleWhere(ctx, targets, now)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return nil, 0, nil
	}
	offset := computeOffset(page, size)
	limit := size
	if limit <= 0 {
		limit = 20
	}
	var items []*Notice
	if err := q.Order(adminOrderBy).Offset(offset).Limit(limit).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *repo) CountVisibleForUser(ctx context.Context, targets []string, now time.Time) (int64, error) {
	var total int64
	if err := r.visibleWhere(ctx, targets, now).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (r *repo) VisibleIDsForUser(ctx context.Context, targets []string, now time.Time) ([]int64, error) {
	var ids []int64
	if err := r.visibleWhere(ctx, targets, now).
		Order("id DESC").
		Pluck("id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}

// computeOffset 在 page<=0 或 size<=0 时安全 fallback 到 0。
func computeOffset(page, size int) int {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	return (page - 1) * size
}
