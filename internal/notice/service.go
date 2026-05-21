package notice

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/ijry/pro-api/internal/audit"
	"github.com/ijry/pro-api/pkg/apierr"
	"go.uber.org/zap"
)

// IDGenerator 是对 idgen 的最小依赖。
type IDGenerator interface {
	Generate() int64
}

// Clock 是对 clock.Clock 的最小依赖(本包仅用 Now)。
type Clock interface {
	Now() time.Time
}

// Service 是 notice 业务接口。
type Service interface {
	// ===== 管理员 =====
	Create(ctx context.Context, in CreateInput, createdBy int64) (*Notice, error)
	List(ctx context.Context, filter ListFilter) ([]*Notice, int64, error)
	Get(ctx context.Context, id int64) (*Notice, error)
	Update(ctx context.Context, id int64, patch UpdatePatch, actor int64) (*Notice, error)
	Delete(ctx context.Context, id int64, actor int64) error
	Publish(ctx context.Context, id int64, actor int64) (*Notice, error)
	Unpublish(ctx context.Context, id int64, actor int64) (*Notice, error)

	// ===== 用户 =====
	ListForUser(ctx context.Context, userID int64, page, size int) ([]*UserNotice, int64, error)
	GetForUser(ctx context.Context, userID int64, id int64) (*UserNotice, error)
	MarkRead(ctx context.Context, userID, noticeID int64) error
	UnreadCountForUser(ctx context.Context, userID int64) int

	// ===== 公开 =====
	ListPublic(ctx context.Context, page, size int) ([]*Notice, int64, error)
}

// CreateInput 是 Create 的入参。
type CreateInput struct {
	Title     string
	Content   string
	Level     string
	Target    string
	Pinned    bool
	PublishAt *time.Time
	ExpiresAt *time.Time
}

// UpdatePatch 是 Update 的入参;指针字段 nil 表示不改;
// 对支持清空的 nullable 字段(publish_at / expires_at)用 *Null bool 表示清空意图。
type UpdatePatch struct {
	Title         *string
	Content       *string
	Level         *string
	Target        *string
	Pinned        *bool
	PublishAt     *time.Time
	ExpiresAt     *time.Time
	PublishAtNull bool
	ExpiresAtNull bool
}

// ListFilter 是后台分页过滤;Status<0 表示不筛。
type ListFilter struct {
	Status int8
	Page   int
	Size   int
}

// Config 是 NewService 的配置。
type Config struct {
	Repo   Repo
	Reader Reader
	IDGen  IDGenerator
	Clock  Clock
	Audit  audit.Logger
	Log    *zap.Logger
}

type service struct {
	repo   Repo
	reader Reader
	idgen  IDGenerator
	clock  Clock
	audit  audit.Logger
	log    *zap.Logger
}

// NewService 构造默认 Service。
func NewService(cfg Config) Service {
	log := cfg.Log
	if log == nil {
		log = zap.NewNop()
	}
	aud := cfg.Audit
	if aud == nil {
		aud = audit.NewNoop()
	}
	return &service{
		repo:   cfg.Repo,
		reader: cfg.Reader,
		idgen:  cfg.IDGen,
		clock:  cfg.Clock,
		audit:  aud,
		log:    log,
	}
}

// userTargets 用户视角可见 target 集合。
var userTargets = []string{TargetAll, TargetUser}

// publicTargets 公开视角可见 target 集合(仅 all)。
var publicTargets = []string{TargetAll}

// ===== 管理员 =====

func (s *service) Create(ctx context.Context, in CreateInput, createdBy int64) (*Notice, error) {
	if in.Title == "" {
		return nil, apierr.New(apierr.CodeMissingParam, "title 必填")
	}
	if in.Content == "" {
		return nil, apierr.New(apierr.CodeMissingParam, "content 必填")
	}
	if len(in.Title) > 128 {
		return nil, apierr.New(apierr.CodeInvalidParam, "title 长度超过 128")
	}
	if len(in.Content) > 65535 {
		return nil, apierr.New(apierr.CodeInvalidParam, "content 长度超过 65535")
	}
	level := in.Level
	if level == "" {
		level = LevelInfo
	}
	if !IsValidLevel(level) {
		return nil, apierr.New(apierr.CodeInvalidParam, "level 取值非法")
	}
	target := in.Target
	if target == "" {
		target = TargetAll
	}
	if !IsValidTarget(target) {
		return nil, apierr.New(apierr.CodeInvalidParam, "target 取值非法")
	}
	now := s.clock.Now()
	if in.ExpiresAt != nil && !in.ExpiresAt.After(now) {
		return nil, apierr.New(apierr.CodeInvalidParam, "expires_at 必须大于当前时间")
	}
	n := &Notice{
		ID:        s.idgen.Generate(),
		Title:     in.Title,
		Content:   in.Content,
		Level:     level,
		Target:    target,
		Status:    StatusDraft,
		Pinned:    in.Pinned,
		PublishAt: in.PublishAt,
		ExpiresAt: in.ExpiresAt,
		CreatedBy: createdBy,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.repo.Create(ctx, n); err != nil {
		return nil, apierr.Wrap(apierr.CodeDatabase, "create notice failed", err)
	}
	s.auditLog(ctx, "notice.create", n.ID, createdBy, nil, n)
	return n, nil
}

func (s *service) List(ctx context.Context, filter ListFilter) ([]*Notice, int64, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Size <= 0 {
		filter.Size = 20
	}
	if filter.Size > 100 {
		filter.Size = 100
	}
	items, total, err := s.repo.ListAdmin(ctx, filter.Status, filter.Page, filter.Size)
	if err != nil {
		return nil, 0, apierr.Wrap(apierr.CodeDatabase, "list notice failed", err)
	}
	return items, total, nil
}

func (s *service) Get(ctx context.Context, id int64) (*Notice, error) {
	n, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apierr.Wrap(apierr.CodeDatabase, "get notice failed", err)
	}
	if n == nil {
		return nil, apierr.New(apierr.CodeNotFound, "公告不存在")
	}
	return n, nil
}

func (s *service) Update(ctx context.Context, id int64, patch UpdatePatch, actor int64) (*Notice, error) {
	cur, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apierr.Wrap(apierr.CodeDatabase, "get notice failed", err)
	}
	if cur == nil {
		return nil, apierr.New(apierr.CodeNotFound, "公告不存在")
	}
	fields := map[string]any{}
	if patch.Title != nil {
		if *patch.Title == "" || len(*patch.Title) > 128 {
			return nil, apierr.New(apierr.CodeInvalidParam, "title 长度非法")
		}
		fields["title"] = *patch.Title
	}
	if patch.Content != nil {
		if *patch.Content == "" || len(*patch.Content) > 65535 {
			return nil, apierr.New(apierr.CodeInvalidParam, "content 长度非法")
		}
		fields["content"] = *patch.Content
	}
	if patch.Level != nil {
		if !IsValidLevel(*patch.Level) {
			return nil, apierr.New(apierr.CodeInvalidParam, "level 取值非法")
		}
		fields["level"] = *patch.Level
	}
	if patch.Target != nil {
		if !IsValidTarget(*patch.Target) {
			return nil, apierr.New(apierr.CodeInvalidParam, "target 取值非法")
		}
		fields["target"] = *patch.Target
	}
	if patch.Pinned != nil {
		fields["pinned"] = *patch.Pinned
	}
	if patch.PublishAtNull {
		fields["publish_at"] = nil
	} else if patch.PublishAt != nil {
		fields["publish_at"] = *patch.PublishAt
	}
	if patch.ExpiresAtNull {
		fields["expires_at"] = nil
	} else if patch.ExpiresAt != nil {
		now := s.clock.Now()
		if !patch.ExpiresAt.After(now) {
			return nil, apierr.New(apierr.CodeInvalidParam, "expires_at 必须大于当前时间")
		}
		fields["expires_at"] = *patch.ExpiresAt
	}
	if len(fields) == 0 {
		return cur, nil // 没有改动
	}
	fields["updated_at"] = s.clock.Now()
	if err := s.repo.Update(ctx, id, fields); err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, apierr.New(apierr.CodeNotFound, "公告不存在")
		}
		return nil, apierr.Wrap(apierr.CodeDatabase, "update notice failed", err)
	}
	after, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apierr.Wrap(apierr.CodeDatabase, "get notice failed", err)
	}
	s.auditLog(ctx, "notice.update", id, actor, cur, after)
	return after, nil
}

func (s *service) Delete(ctx context.Context, id int64, actor int64) error {
	cur, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return apierr.Wrap(apierr.CodeDatabase, "get notice failed", err)
	}
	if cur == nil {
		return apierr.New(apierr.CodeNotFound, "公告不存在")
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, ErrNotFound) {
			return apierr.New(apierr.CodeNotFound, "公告不存在")
		}
		return apierr.Wrap(apierr.CodeDatabase, "delete notice failed", err)
	}
	s.auditLog(ctx, "notice.delete", id, actor, cur, nil)
	return nil
}

func (s *service) Publish(ctx context.Context, id int64, actor int64) (*Notice, error) {
	cur, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apierr.Wrap(apierr.CodeDatabase, "get notice failed", err)
	}
	if cur == nil {
		return nil, apierr.New(apierr.CodeNotFound, "公告不存在")
	}
	if cur.Status == StatusPublished {
		return cur, nil // 幂等,不审计
	}
	now := s.clock.Now()
	fields := map[string]any{
		"status":     StatusPublished,
		"updated_at": now,
	}
	if cur.PublishAt == nil {
		fields["publish_at"] = now
	}
	if err := s.repo.Update(ctx, id, fields); err != nil {
		return nil, apierr.Wrap(apierr.CodeDatabase, "publish notice failed", err)
	}
	after, _ := s.repo.GetByID(ctx, id)
	s.auditLog(ctx, "notice.publish", id, actor, cur, after)
	return after, nil
}

func (s *service) Unpublish(ctx context.Context, id int64, actor int64) (*Notice, error) {
	cur, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apierr.Wrap(apierr.CodeDatabase, "get notice failed", err)
	}
	if cur == nil {
		return nil, apierr.New(apierr.CodeNotFound, "公告不存在")
	}
	if cur.Status == StatusArchived {
		return cur, nil // 幂等
	}
	now := s.clock.Now()
	if err := s.repo.Update(ctx, id, map[string]any{
		"status":     StatusArchived,
		"updated_at": now,
	}); err != nil {
		return nil, apierr.Wrap(apierr.CodeDatabase, "unpublish notice failed", err)
	}
	after, _ := s.repo.GetByID(ctx, id)
	s.auditLog(ctx, "notice.unpublish", id, actor, cur, after)
	return after, nil
}

// ===== 用户 =====

func (s *service) ListForUser(ctx context.Context, userID int64, page, size int) ([]*UserNotice, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	if size > 100 {
		size = 100
	}
	now := s.clock.Now()
	items, total, err := s.repo.ListVisibleForUser(ctx, userTargets, now, page, size)
	if err != nil {
		return nil, 0, apierr.Wrap(apierr.CodeDatabase, "list notice failed", err)
	}
	readSet, err := s.reader.ReadSet(ctx, userID)
	if err != nil {
		s.log.Warn("notice: ReadSet failed", zap.Int64("user_id", userID), zap.Error(err))
		readSet = map[int64]struct{}{}
	}
	out := make([]*UserNotice, len(items))
	for i, n := range items {
		_, isRead := readSet[n.ID]
		out[i] = ToUserNotice(n, isRead)
	}
	return out, total, nil
}

func (s *service) GetForUser(ctx context.Context, userID int64, id int64) (*UserNotice, error) {
	n, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apierr.Wrap(apierr.CodeDatabase, "get notice failed", err)
	}
	if n == nil {
		return nil, apierr.New(apierr.CodeNotFound, "公告不存在")
	}
	now := s.clock.Now()
	if !isVisibleForTargets(n, userTargets, now) {
		return nil, apierr.New(apierr.CodeNotFound, "公告不存在")
	}
	isRead, err := s.reader.IsRead(ctx, userID, n.ID)
	if err != nil {
		s.log.Warn("notice: IsRead failed", zap.Int64("user_id", userID), zap.Error(err))
		isRead = false
	}
	return ToUserNotice(n, isRead), nil
}

func (s *service) MarkRead(ctx context.Context, userID, noticeID int64) error {
	if err := s.reader.MarkRead(ctx, userID, noticeID); err != nil {
		return apierr.Wrap(apierr.CodeCache, "mark read failed", err)
	}
	return nil
}

func (s *service) UnreadCountForUser(ctx context.Context, userID int64) int {
	now := s.clock.Now()
	ids, err := s.repo.VisibleIDsForUser(ctx, userTargets, now)
	if err != nil {
		s.log.Warn("notice: VisibleIDsForUser failed", zap.Error(err))
		return 0
	}
	if len(ids) == 0 {
		return 0
	}
	n, err := s.reader.UnreadCount(ctx, userID, ids)
	if err != nil {
		s.log.Warn("notice: UnreadCount failed", zap.Error(err))
		return 0
	}
	return n
}

// ===== 公开 =====

func (s *service) ListPublic(ctx context.Context, page, size int) ([]*Notice, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	if size > 100 {
		size = 100
	}
	now := s.clock.Now()
	items, total, err := s.repo.ListVisibleForUser(ctx, publicTargets, now, page, size)
	if err != nil {
		return nil, 0, apierr.Wrap(apierr.CodeDatabase, "list notice failed", err)
	}
	return items, total, nil
}

// isVisibleForTargets 在内存里判断 notice 对该目标视角是否可见。
func isVisibleForTargets(n *Notice, targets []string, now time.Time) bool {
	if n.Status != StatusPublished {
		return false
	}
	if n.PublishAt != nil && n.PublishAt.After(now) {
		return false
	}
	if n.ExpiresAt != nil && !n.ExpiresAt.After(now) {
		return false
	}
	for _, t := range targets {
		if n.Target == t {
			return true
		}
	}
	return false
}

// auditLog 写一条审计;失败仅 warn(audit 内部已吞错)。
func (s *service) auditLog(ctx context.Context, action string, id int64, actor int64, before, after any) {
	var beforeJSON, afterJSON json.RawMessage
	if before != nil {
		if b, err := json.Marshal(before); err == nil {
			beforeJSON = b
		}
	}
	if after != nil {
		if b, err := json.Marshal(after); err == nil {
			afterJSON = b
		}
	}
	actorRef := actor
	idRef := id
	entry := audit.Entry{
		Action:     action,
		TargetType: "notice",
		TargetID:   &idRef,
		ActorID:    &actorRef,
		Before:     beforeJSON,
		After:      afterJSON,
	}
	_ = s.audit.Log(ctx, entry)
}
