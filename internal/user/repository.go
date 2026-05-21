package user

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"
)

// ListFilter 描述 List 查询参数。
type ListFilter struct {
	Keyword string // 模糊匹配 username / email / display_name
	Role    *int8
	Status  *int8
	GroupID *int64
	Page    int    // 1-based,缺省 1
	Size    int    // 默认 20,上限 100
	OrderBy string // "created_at_desc" / "last_login_desc";默认 created_at_desc
}

// Repository 是 users 表 CRUD 抽象。
type Repository interface {
	Create(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id int64) (*User, error)
	GetByUsername(ctx context.Context, name string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	List(ctx context.Context, f ListFilter) ([]*User, int64, error)
	UpdateFields(ctx context.Context, id int64, fields map[string]any) error
	SoftDelete(ctx context.Context, id int64) error
}

// repo 默认实现。
type repo struct {
	db *gorm.DB
}

// NewRepository 构造仓储。
func NewRepository(db *gorm.DB) Repository { return &repo{db: db} }

func (r *repo) Create(ctx context.Context, u *User) error {
	return r.db.WithContext(ctx).Create(u).Error
}

func (r *repo) GetByID(ctx context.Context, id int64) (*User, error) {
	var u User
	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *repo) GetByUsername(ctx context.Context, name string) (*User, error) {
	var u User
	err := r.db.WithContext(ctx).
		Where("username = ? AND deleted_at IS NULL", name).
		First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *repo) GetByEmail(ctx context.Context, email string) (*User, error) {
	var u User
	err := r.db.WithContext(ctx).
		Where("email = ? AND deleted_at IS NULL", email).
		First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *repo) List(ctx context.Context, f ListFilter) ([]*User, int64, error) {
	q := r.db.WithContext(ctx).Model(&User{}).Where("deleted_at IS NULL")
	if strings.TrimSpace(f.Keyword) != "" {
		k := "%" + strings.TrimSpace(f.Keyword) + "%"
		q = q.Where("username LIKE ? OR email LIKE ? OR display_name LIKE ?", k, k, k)
	}
	if f.Role != nil {
		q = q.Where("role = ?", *f.Role)
	}
	if f.Status != nil {
		q = q.Where("status = ?", *f.Status)
	}
	if f.GroupID != nil {
		q = q.Where("group_id = ?", *f.GroupID)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := f.Page
	if page < 1 {
		page = 1
	}
	size := f.Size
	if size <= 0 {
		size = 20
	}
	if size > 100 {
		size = 100
	}

	order := "created_at DESC"
	switch f.OrderBy {
	case "last_login_desc":
		order = "last_login_at DESC"
	case "created_at_asc":
		order = "created_at ASC"
	}
	var items []*User
	if err := q.Order(order).Offset((page - 1) * size).Limit(size).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *repo) UpdateFields(ctx context.Context, id int64, fields map[string]any) error {
	if len(fields) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&User{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(fields).Error
}

func (r *repo) SoftDelete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&User{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]any{
			"deleted_at": gormNow(),
			"status":     StatusDisabled,
		}).Error
}

// gormNow 返回当前 UTC,gorm 写入。
func gormNow() any { return nowUTC() }
