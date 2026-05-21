package group

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ijry/pro-api/internal/util/clock"
)

// CreateInput 建分组参数。
type CreateInput struct {
	Name        string
	DisplayName string
	Ratio       float64
	Priority    int16
}

// Service 是 group 业务接口。
type Service interface {
	Create(ctx context.Context, in CreateInput) (*Group, error)
	GetByID(ctx context.Context, id int64) (*Group, error)
	GetByName(ctx context.Context, name string) (*Group, error)
	Default(ctx context.Context) (*Group, error)
	List(ctx context.Context) ([]*Group, error)
	Update(ctx context.Context, id int64, in CreateInput) (*Group, error)
	Delete(ctx context.Context, id int64) error
	RatioFor(ctx context.Context, id int64) (float64, error)
}

// IDGenerator 与 audit 包同形,用 idgen.Generator 即可。
type IDGenerator interface {
	Generate() int64
}

var _ = clock.Real // 显式引用避免清理

// svc 默认实现。
type svc struct {
	repo  Repository
	clock clock.Clock
	idgen IDGenerator

	mu       sync.RWMutex
	ratioMap map[int64]ratioEntry
	ratioTTL time.Duration
}

type ratioEntry struct {
	v       float64
	expires time.Time
}

// NewService 构造 group.Service。
func NewService(repo Repository, c clock.Clock, idg IDGenerator) Service {
	if c == nil {
		c = clock.Real
	}
	return &svc{
		repo:     repo,
		clock:    c,
		idgen:    idg,
		ratioMap: make(map[int64]ratioEntry),
		ratioTTL: 5 * time.Minute,
	}
}

// idgenOrSnowflake 兜底:若 NewService 时未注入 idgen,Create 时报错。
func (s *svc) newID() (int64, error) {
	if s.idgen == nil {
		return 0, errors.New("group: idgen not configured")
	}
	return s.idgen.Generate(), nil
}

func (s *svc) Create(ctx context.Context, in CreateInput) (*Group, error) {
	if in.Name == "" {
		return nil, errors.New("group: name required")
	}
	id, err := s.newID()
	if err != nil {
		return nil, err
	}
	if in.Ratio <= 0 {
		in.Ratio = 1.0
	}
	now := s.clock.Now().UTC()
	g := &Group{
		ID:          id,
		Name:        in.Name,
		DisplayName: in.DisplayName,
		Ratio:       in.Ratio,
		Priority:    in.Priority,
		Status:      0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.repo.Create(ctx, g); err != nil {
		return nil, fmt.Errorf("group: create: %w", err)
	}
	return g, nil
}

func (s *svc) GetByID(ctx context.Context, id int64) (*Group, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *svc) GetByName(ctx context.Context, name string) (*Group, error) {
	return s.repo.GetByName(ctx, name)
}

func (s *svc) Default(ctx context.Context) (*Group, error) {
	g, err := s.repo.GetByName(ctx, DefaultGroupName)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, fmt.Errorf("group: default group %q not seeded", DefaultGroupName)
	}
	return g, nil
}

func (s *svc) List(ctx context.Context) ([]*Group, error) {
	return s.repo.List(ctx)
}

func (s *svc) Update(ctx context.Context, id int64, in CreateInput) (*Group, error) {
	fields := map[string]any{
		"display_name": in.DisplayName,
		"ratio":        in.Ratio,
		"priority":     in.Priority,
		"updated_at":   s.clock.Now().UTC(),
	}
	if in.Name != "" {
		fields["name"] = in.Name
	}
	if err := s.repo.UpdateFields(ctx, id, fields); err != nil {
		return nil, err
	}
	// 失效缓存
	s.mu.Lock()
	delete(s.ratioMap, id)
	s.mu.Unlock()
	return s.repo.GetByID(ctx, id)
}

func (s *svc) Delete(ctx context.Context, id int64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	s.mu.Lock()
	delete(s.ratioMap, id)
	s.mu.Unlock()
	return nil
}

// RatioFor 取分组倍率,5 分钟内重复命中 走本地缓存。
func (s *svc) RatioFor(ctx context.Context, id int64) (float64, error) {
	now := s.clock.Now()
	s.mu.RLock()
	if e, ok := s.ratioMap[id]; ok && now.Before(e.expires) {
		s.mu.RUnlock()
		return e.v, nil
	}
	s.mu.RUnlock()

	g, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return 0, err
	}
	if g == nil {
		return 1.0, nil // 找不到时按 1.0
	}
	s.mu.Lock()
	s.ratioMap[id] = ratioEntry{v: g.Ratio, expires: now.Add(s.ratioTTL)}
	s.mu.Unlock()
	return g.Ratio, nil
}

