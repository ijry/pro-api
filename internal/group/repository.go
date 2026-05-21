package group

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

// Repository 是 user_groups CRUD 接口。
type Repository interface {
	Create(ctx context.Context, g *Group) error
	GetByID(ctx context.Context, id int64) (*Group, error)
	GetByName(ctx context.Context, name string) (*Group, error)
	List(ctx context.Context) ([]*Group, error)
	UpdateFields(ctx context.Context, id int64, fields map[string]any) error
	Delete(ctx context.Context, id int64) error
}

type repo struct{ db *gorm.DB }

// NewRepository 构造仓储。
func NewRepository(db *gorm.DB) Repository { return &repo{db: db} }

func (r *repo) Create(ctx context.Context, g *Group) error {
	return r.db.WithContext(ctx).Create(g).Error
}

func (r *repo) GetByID(ctx context.Context, id int64) (*Group, error) {
	var g Group
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&g).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &g, nil
}

func (r *repo) GetByName(ctx context.Context, name string) (*Group, error) {
	var g Group
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&g).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &g, nil
}

func (r *repo) List(ctx context.Context) ([]*Group, error) {
	var items []*Group
	if err := r.db.WithContext(ctx).Order("priority DESC, id ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *repo) UpdateFields(ctx context.Context, id int64, fields map[string]any) error {
	if len(fields) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&Group{}).Where("id = ?", id).Updates(fields).Error
}

func (r *repo) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&Group{}).Error
}
