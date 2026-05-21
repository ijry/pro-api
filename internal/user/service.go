package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/ijry/pro-api/internal/group"
	"github.com/ijry/pro-api/internal/util/clock"
	"github.com/ijry/pro-api/pkg/apierr"
)

// CreateInput 描述一个新建用户。
type CreateInput struct {
	Username     string
	Email        *string
	PasswordHash *string
	DisplayName  *string
	Avatar       *string
	Role         int8
	Status       int8
	GroupID      *int64
}

// UpdateInput 描述 PATCH 用户字段;仅非 nil 字段会被更新。
type UpdateInput struct {
	DisplayName *string
	Avatar      *string
	Role        *int8
	Status      *int8
	GroupID     *int64
}

// IDGenerator 与 audit 包同形。
type IDGenerator interface {
	Generate() int64
}

// Service 是 users 业务接口。
type Service interface {
	Create(ctx context.Context, in CreateInput) (*User, error)
	GetByID(ctx context.Context, id int64) (*User, error)
	GetByUsername(ctx context.Context, name string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	List(ctx context.Context, f ListFilter) ([]*User, int64, error)
	Update(ctx context.Context, id int64, in UpdateInput) (*User, error)
	UpdatePasswordHash(ctx context.Context, id int64, hash string) error
	MarkEmailVerified(ctx context.Context, id int64) error
	TouchLogin(ctx context.Context, id int64, ip string) error
	Delete(ctx context.Context, id int64) error
}

// svc 是默认实现。
type svc struct {
	repo  Repository
	group group.Service
	idgen IDGenerator
	clock clock.Clock
}

// NewService 构造 user.Service。
func NewService(repo Repository, g group.Service, idg IDGenerator, c clock.Clock) Service {
	if c == nil {
		c = clock.Real
	}
	return &svc{repo: repo, group: g, idgen: idg, clock: c}
}

// Create 写 users 表;GroupID 缺省时回填 default 分组。
func (s *svc) Create(ctx context.Context, in CreateInput) (*User, error) {
	if in.Username == "" {
		return nil, apierr.New(apierr.CodeInvalidParam, "username 必填")
	}
	// 重复校验
	if in.Email != nil && *in.Email != "" {
		exists, err := s.repo.GetByEmail(ctx, *in.Email)
		if err != nil {
			return nil, apierr.Wrap(apierr.CodeDatabase, "user create: check email", err)
		}
		if exists != nil {
			return nil, apierr.New(apierr.CodeEmailRegistered, "邮箱已被注册")
		}
	}
	if exists, err := s.repo.GetByUsername(ctx, in.Username); err != nil {
		return nil, apierr.Wrap(apierr.CodeDatabase, "user create: check username", err)
	} else if exists != nil {
		return nil, apierr.New(apierr.CodeUsernameTaken, "用户名已被占用")
	}

	// 默认分组
	gid := in.GroupID
	if gid == nil {
		if s.group != nil {
			g, err := s.group.Default(ctx)
			if err != nil {
				return nil, apierr.Wrap(apierr.CodeInternal, "user create: default group", err)
			}
			tmp := g.ID
			gid = &tmp
		}
	}

	if s.idgen == nil {
		return nil, errors.New("user: idgen not configured")
	}
	now := s.clock.Now().UTC()
	u := &User{
		ID:           s.idgen.Generate(),
		Username:     in.Username,
		Email:        in.Email,
		PasswordHash: in.PasswordHash,
		DisplayName:  in.DisplayName,
		Avatar:       in.Avatar,
		Role:         in.Role,
		Status:       in.Status,
		GroupID:      gid,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.repo.Create(ctx, u); err != nil {
		return nil, apierr.Wrap(apierr.CodeDatabase, "user create insert", err)
	}
	return u, nil
}

func (s *svc) GetByID(ctx context.Context, id int64) (*User, error) {
	return s.repo.GetByID(ctx, id)
}
func (s *svc) GetByUsername(ctx context.Context, name string) (*User, error) {
	return s.repo.GetByUsername(ctx, name)
}
func (s *svc) GetByEmail(ctx context.Context, email string) (*User, error) {
	return s.repo.GetByEmail(ctx, email)
}
func (s *svc) List(ctx context.Context, f ListFilter) ([]*User, int64, error) {
	return s.repo.List(ctx, f)
}

func (s *svc) Update(ctx context.Context, id int64, in UpdateInput) (*User, error) {
	fields := map[string]any{"updated_at": s.clock.Now().UTC()}
	if in.DisplayName != nil {
		fields["display_name"] = *in.DisplayName
	}
	if in.Avatar != nil {
		fields["avatar"] = *in.Avatar
	}
	if in.Role != nil {
		fields["role"] = *in.Role
	}
	if in.Status != nil {
		fields["status"] = *in.Status
	}
	if in.GroupID != nil {
		fields["group_id"] = *in.GroupID
	}
	if err := s.repo.UpdateFields(ctx, id, fields); err != nil {
		return nil, apierr.Wrap(apierr.CodeDatabase, "user update", err)
	}
	return s.repo.GetByID(ctx, id)
}

func (s *svc) UpdatePasswordHash(ctx context.Context, id int64, hash string) error {
	if hash == "" {
		return errors.New("user: empty hash")
	}
	return s.repo.UpdateFields(ctx, id, map[string]any{
		"password_hash": hash,
		"updated_at":    s.clock.Now().UTC(),
	})
}

func (s *svc) MarkEmailVerified(ctx context.Context, id int64) error {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if u == nil {
		return fmt.Errorf("user: %d not found", id)
	}
	now := s.clock.Now().UTC()
	fields := map[string]any{
		"email_verified_at": now,
		"updated_at":        now,
	}
	if u.Status == StatusPendingEmailVerify {
		fields["status"] = StatusActive
	}
	return s.repo.UpdateFields(ctx, id, fields)
}

func (s *svc) TouchLogin(ctx context.Context, id int64, ip string) error {
	now := s.clock.Now().UTC()
	return s.repo.UpdateFields(ctx, id, map[string]any{
		"last_login_at": now,
		"last_login_ip": ip,
		"updated_at":    now,
	})
}

func (s *svc) Delete(ctx context.Context, id int64) error {
	return s.repo.SoftDelete(ctx, id)
}
