package manual

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

// ErrNotFound 表示充值单不存在(用于 Repo.Update / 状态机)。
var ErrNotFound = errors.New("manual: recharge not found")

// ListFilter 是 List 的过滤参数。
type ListFilter struct {
	UserID   int64 // 0 = 不限(admin)
	Status   *int8 // nil = 不限
	Page     int   // 1-based
	PageSize int   // 默认 20,最大 100
}

// Repo 是 manual_recharges 表的仓储接口。
type Repo interface {
	Create(ctx context.Context, n *Recharge) error
	GetByID(ctx context.Context, id int64) (*Recharge, error)
	Update(ctx context.Context, id int64, fields map[string]any) error
	UpdateStatusFromPending(ctx context.Context, id int64, fields map[string]any) (int64, error)
	List(ctx context.Context, f ListFilter) ([]*Recharge, int64, error)
}

type repo struct {
	db *gorm.DB
}

// NewRepository 构造默认 GORM 实现。
func NewRepository(db *gorm.DB) Repo { return &repo{db: db} }

// Create 插入一条 recharge;调用方负责填充 ID/CreatedAt/UpdatedAt。
func (r *repo) Create(ctx context.Context, n *Recharge) error {
	return r.db.WithContext(ctx).Create(n).Error
}

// GetByID 返回单条;不存在返 (nil, nil)。
func (r *repo) GetByID(ctx context.Context, id int64) (*Recharge, error) {
	var n Recharge
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&n).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &n, nil
}

// Update 仅更新 fields 中给出的列;没行被影响返 ErrNotFound。
func (r *repo) Update(ctx context.Context, id int64, fields map[string]any) error {
	if len(fields) == 0 {
		return nil
	}
	res := r.db.WithContext(ctx).Model(&Recharge{}).Where("id = ?", id).Updates(fields)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// UpdateStatusFromPending 只在当前状态是 pending 时更新。
//
// 返回 affected rows:1 = 更新成功;0 = 状态不是 pending 或 id 不存在。
// 用于 Approve / Reject / Cancel 等状态机迁移。
func (r *repo) UpdateStatusFromPending(ctx context.Context, id int64, fields map[string]any) (int64, error) {
	if len(fields) == 0 {
		return 0, nil
	}
	res := r.db.WithContext(ctx).Model(&Recharge{}).
		Where("id = ? AND status = ?", id, StatusPending).
		Updates(fields)
	if res.Error != nil {
		return 0, res.Error
	}
	return res.RowsAffected, nil
}

// List 按 filter 分页查询;返回结果 + 总数。
//
// 排序:created_at DESC, id DESC(管理员视角和用户视角一致)。
func (r *repo) List(ctx context.Context, f ListFilter) ([]*Recharge, int64, error) {
	q := r.db.WithContext(ctx).Model(&Recharge{})
	if f.UserID > 0 {
		q = q.Where("user_id = ?", f.UserID)
	}
	if f.Status != nil {
		q = q.Where("status = ?", *f.Status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return nil, 0, nil
	}
	page, size := f.Page, f.PageSize
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	if size > 100 {
		size = 100
	}
	offset := (page - 1) * size
	var items []*Recharge
	if err := q.Order("created_at DESC, id DESC").Offset(offset).Limit(size).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}
