package redeem

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

// ErrNotFound 表示 code 不存在。
var ErrNotFound = errors.New("redeem: code not found")

// ListFilter 是 List / Export 的过滤参数。
type ListFilter struct {
	BatchNo  string
	Status   *int8 // nil = 不限
	Page     int   // 1-based;ListAll 时忽略
	PageSize int   // 默认 20,最大 100;ListAll 时忽略
}

// Repo 是 redeem_codes 表的仓储接口。
type Repo interface {
	// BatchInsert 批量插入,任一冲突整批回滚。
	BatchInsert(ctx context.Context, codes []*Code) error

	GetByID(ctx context.Context, id int64) (*Code, error)
	GetByHash(ctx context.Context, hash string) (*Code, error)

	// UpdateToUsed 仅在当前状态是 unused 时把 code 改为 used,
	// 并填 used_by / used_at。返回受影响行数。
	UpdateToUsed(ctx context.Context, id int64, userID int64, at time.Time) (int64, error)

	// UpdateToDisabledBulk 批量把 ids 中状态仍为 unused 的转为 disabled,
	// 返回成功转换的行数(used / disabled 不再变更)。
	UpdateToDisabledBulk(ctx context.Context, ids []int64) (int64, error)

	// RollbackUsedToUnused 把状态从 used 回滚到 unused(补偿用)。
	// 仅当当前状态为 used 时生效。
	RollbackUsedToUnused(ctx context.Context, id int64) error

	List(ctx context.Context, f ListFilter) ([]*Code, int64, error)
	ListAll(ctx context.Context, f ListFilter) ([]*Code, error)

	// DB 暴露原始 *gorm.DB,供 service 用 Transaction(...) 包装 SELECT FOR UPDATE。
	DB() *gorm.DB
}

type repo struct {
	db *gorm.DB
}

// NewRepository 构造默认 GORM 实现。
func NewRepository(db *gorm.DB) Repo { return &repo{db: db} }

func (r *repo) DB() *gorm.DB { return r.db }

// BatchInsert 批量插入(GORM CreateInBatches 默认 100/batch)。
func (r *repo) BatchInsert(ctx context.Context, codes []*Code) error {
	if len(codes) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).CreateInBatches(codes, 100).Error
}

// GetByID 查单条,不存在返 (nil, nil)。
func (r *repo) GetByID(ctx context.Context, id int64) (*Code, error) {
	var c Code
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&c).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

// GetByHash 按 hash 查;不存在返 (nil, nil)。
func (r *repo) GetByHash(ctx context.Context, hash string) (*Code, error) {
	var c Code
	err := r.db.WithContext(ctx).Where("code_hash = ?", hash).First(&c).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

// UpdateToUsed 把 unused → used。Where 兜底原子性。
func (r *repo) UpdateToUsed(ctx context.Context, id int64, userID int64, at time.Time) (int64, error) {
	res := r.db.WithContext(ctx).Model(&Code{}).
		Where("id = ? AND status = ?", id, StatusUnused).
		Updates(map[string]any{
			"status":  StatusUsed,
			"used_by": userID,
			"used_at": at,
		})
	if res.Error != nil {
		return 0, res.Error
	}
	return res.RowsAffected, nil
}

// UpdateToDisabledBulk 把 ids 中 unused 的转 disabled,used/disabled 跳过。
func (r *repo) UpdateToDisabledBulk(ctx context.Context, ids []int64) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	res := r.db.WithContext(ctx).Model(&Code{}).
		Where("id IN ? AND status = ?", ids, StatusUnused).
		Updates(map[string]any{
			"status": StatusDisabled,
		})
	if res.Error != nil {
		return 0, res.Error
	}
	return res.RowsAffected, nil
}

// RollbackUsedToUnused 把 status 从 used 回滚到 unused(补偿用)。
func (r *repo) RollbackUsedToUnused(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&Code{}).
		Where("id = ? AND status = ?", id, StatusUsed).
		Updates(map[string]any{
			"status":  StatusUnused,
			"used_by": nil,
			"used_at": nil,
		}).Error
}

// listQuery 把 filter 拼成 gorm.DB(不含分页)。
func (r *repo) listQuery(ctx context.Context, f ListFilter) *gorm.DB {
	q := r.db.WithContext(ctx).Model(&Code{})
	if f.BatchNo != "" {
		q = q.Where("batch_no = ?", f.BatchNo)
	}
	if f.Status != nil {
		q = q.Where("status = ?", *f.Status)
	}
	return q
}

// List 分页列出。
func (r *repo) List(ctx context.Context, f ListFilter) ([]*Code, int64, error) {
	q := r.listQuery(ctx, f)
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
	var items []*Code
	if err := q.Order("created_at DESC, id DESC").Offset(offset).Limit(size).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// ListAll 不分页(用于 Export);调用方需理解可能返回大量行,建议流式 use。
func (r *repo) ListAll(ctx context.Context, f ListFilter) ([]*Code, error) {
	var items []*Code
	if err := r.listQuery(ctx, f).Order("created_at ASC, id ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}
