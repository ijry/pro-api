package token

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/ijry/pro-api/internal/audit"
	"github.com/ijry/pro-api/internal/util/clock"
	"github.com/ijry/pro-api/pkg/apierr"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Config 是 New 的参数。
type Config struct {
	DB    *gorm.DB
	Cache *redis.Client
	Log   *zap.Logger
	Clock clock.Clock
	IDGen idGenerator
	Audit audit.Logger

	// 可选,0 表示用默认。
	CacheTTL         time.Duration // 默认 5min
	NegativeCacheTTL time.Duration // 默认 30s
	FlushInterval    time.Duration // 默认 30s
	FlushBatchSize   int           // 默认 200
	PrefixShowLen    int           // 默认 8
}

// service 组合 repo + cache + flusher + audit,实现 Store 接口。
type service struct {
	cfg    Config
	repo   *repo
	cache  *tokenCache
	flush  *flusher
	clk    clock.Clock
	log    *zap.Logger
	audit  audit.Logger
	closed bool
}

// New 构造 Store。内部启动 batch flusher 与 Pub/Sub 订阅 goroutine。
func New(cfg Config) (Store, error) {
	if cfg.DB == nil {
		return nil, errors.New("token: DB is required")
	}
	if cfg.IDGen == nil {
		return nil, errors.New("token: IDGen is required")
	}
	log := cfg.Log
	if log == nil {
		log = zap.NewNop()
	}
	clk := cfg.Clock
	if clk == nil {
		clk = clock.Real
	}
	auditLogger := cfg.Audit
	if auditLogger == nil {
		auditLogger = audit.NewNoop()
	}

	r := &repo{
		db:            cfg.DB,
		idgen:         cfg.IDGen,
		clk:           clk,
		prefixShowLen: cfg.PrefixShowLen,
	}
	c := newCache(cfg.Cache, log, cfg.CacheTTL, cfg.NegativeCacheTTL)
	f := newFlusher(r, c, clk, log, cfg.FlushInterval, cfg.FlushBatchSize)
	f.start()

	log.Info("token store: ready",
		zap.Duration("flush_interval", f.interval),
		zap.Duration("cache_ttl", durationOrDefault(cfg.CacheTTL, defaultCacheTTL)),
	)

	return &service{
		cfg:   cfg,
		repo:  r,
		cache: c,
		flush: f,
		clk:   clk,
		log:   log,
		audit: auditLogger,
	}, nil
}

func durationOrDefault(d, def time.Duration) time.Duration {
	if d <= 0 {
		return def
	}
	return d
}

// Authenticate 走 Redis cache → DB,命中后异步 TouchLastUsed。
func (s *service) Authenticate(ctx context.Context, plaintext string) (*View, error) {
	if !hasTokenPrefix(plaintext) {
		return nil, apierr.New(apierr.CodeInvalidToken, "token format invalid")
	}
	h := hashPlaintext(plaintext)

	// 1. Cache
	if v, st := s.cache.Get(ctx, h); st == cacheHit {
		if err := s.validateCachedView(v); err != nil {
			return nil, err
		}
		s.flush.TouchLastUsed(v.ID, s.clk.Now())
		return v, nil
	} else if st == cacheNegative {
		return nil, apierr.New(apierr.CodeInvalidToken, "token not found")
	}

	// 2. DB
	v, err := s.repo.Authenticate(ctx, plaintext)
	if err != nil {
		// 仅 "not found" 走负缓存;状态/过期错误不缓存(用户可能马上恢复)
		var ae *apierr.Error
		if errors.As(err, &ae) && ae.Code == apierr.CodeInvalidToken && ae.Message == "token not found" {
			s.cache.SetNegative(ctx, h)
		}
		return nil, err
	}
	s.cache.SetPositive(ctx, h, v)
	s.flush.TouchLastUsed(v.ID, s.clk.Now())
	return v, nil
}

// validateCachedView 在命中缓存后再次校验 status / expires_at,防止缓存与 DB 状态滞后。
func (s *service) validateCachedView(v *View) error {
	if v.Status == StatusDisabled {
		return apierr.New(apierr.CodeInvalidToken, "token disabled")
	}
	if v.ExpiresAt != nil && !v.ExpiresAt.After(s.clk.Now()) {
		return apierr.New(apierr.CodeTokenExpired, "token expired")
	}
	return nil
}

// hasTokenPrefix 与 repo.Authenticate 同步的前缀快速判定。
func hasTokenPrefix(s string) bool {
	return len(s) >= 4 && s[:3] == keyPrefixLit
}

// Create 创建 → audit 记录。
func (s *service) Create(ctx context.Context, in CreateInput) (string, *View, error) {
	plaintext, view, err := s.repo.Create(ctx, in)
	if err != nil {
		return "", nil, err
	}
	s.logAudit(ctx, view.UserID, "token.create", view.ID, nil, map[string]any{
		"name":       view.Name,
		"key_prefix": view.KeyPrefix,
		"rpm_limit":  view.RPMLimit,
		"tpm_limit":  view.TPMLimit,
	})
	return plaintext, view, nil
}

func (s *service) List(ctx context.Context, f ListFilter) ([]*View, int64, error) {
	return s.repo.List(ctx, f)
}

func (s *service) Get(ctx context.Context, id int64) (*View, error) {
	return s.repo.Get(ctx, id)
}

// Update 后从 DB 拿最新 row 的 key_hash 失效缓存。
func (s *service) Update(ctx context.Context, id int64, p UpdatePatch) (*View, error) {
	before, err := s.repo.getRowByID(ctx, id)
	if err != nil {
		return nil, err
	}
	updated, err := s.repo.Update(ctx, id, p)
	if err != nil {
		return nil, err
	}
	// 任何字段改动都失效缓存(简单粗暴,避免漏失效导致脏读)
	s.cache.Invalidate(ctx, before.KeyHash)
	s.logAudit(ctx, updated.UserID, "token.update", updated.ID, beforeAfterFromRow(before), patchToMap(p))
	return updated, nil
}

// Revoke 软禁用 + 失效缓存 + audit。
func (s *service) Revoke(ctx context.Context, id int64) error {
	view, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	oldHash, err := s.repo.Revoke(ctx, id)
	if err != nil {
		return err
	}
	s.cache.Invalidate(ctx, oldHash)
	s.logAudit(ctx, view.UserID, "token.revoke", id, map[string]any{"status": int(view.Status)}, map[string]any{"status": int(StatusDisabled)})
	return nil
}

// Regenerate 替换 key + 失效旧 hash + audit。
func (s *service) Regenerate(ctx context.Context, id int64) (string, *View, error) {
	plaintext, view, oldHash, err := s.repo.Regenerate(ctx, id)
	if err != nil {
		return "", nil, err
	}
	s.cache.Invalidate(ctx, oldHash)
	s.logAudit(ctx, view.UserID, "token.regenerate", id,
		map[string]any{"hash_prefix": truncateHash(oldHash)},
		map[string]any{"key_prefix": view.KeyPrefix},
	)
	return plaintext, view, nil
}

// IncrementUsage 透传到 flusher。
func (s *service) IncrementUsage(tokenID int64, delta int64) {
	s.flush.IncrementUsage(tokenID, delta)
}

// TouchLastUsed 透传到 flusher。
func (s *service) TouchLastUsed(tokenID int64, t time.Time) {
	s.flush.TouchLastUsed(tokenID, t)
}

// Close 关停 flusher(final flush)+ cache 订阅。
func (s *service) Close() error {
	if s.closed {
		return nil
	}
	s.closed = true
	var firstErr error
	if err := s.flush.Close(); err != nil {
		firstErr = err
	}
	if err := s.cache.Close(); err != nil && firstErr == nil {
		firstErr = err
	}
	s.log.Info("token store: closed")
	return firstErr
}

// === audit helpers ===

func (s *service) logAudit(ctx context.Context, actorID int64, action string, targetID int64, before, after map[string]any) {
	if s.audit == nil {
		return
	}
	entry := audit.Entry{
		Action:     action,
		TargetType: "api_token",
	}
	if actorID > 0 {
		entry.ActorID = &actorID
	}
	if targetID > 0 {
		entry.TargetID = &targetID
	}
	if before != nil {
		if b, err := json.Marshal(before); err == nil {
			entry.Before = b
		}
	}
	if after != nil {
		if b, err := json.Marshal(after); err == nil {
			entry.After = b
		}
	}
	_ = s.audit.Log(ctx, entry)
}

func beforeAfterFromRow(t *Token) map[string]any {
	view := t.ToView()
	return map[string]any{
		"name":       view.Name,
		"status":     int(view.Status),
		"rpm_limit":  view.RPMLimit,
		"tpm_limit":  view.TPMLimit,
		"key_prefix": view.KeyPrefix,
	}
}

func patchToMap(p UpdatePatch) map[string]any {
	m := map[string]any{}
	if p.Name != nil {
		m["name"] = *p.Name
	}
	if p.RPMLimit != nil {
		m["rpm_limit"] = *p.RPMLimit
	}
	if p.TPMLimit != nil {
		m["tpm_limit"] = *p.TPMLimit
	}
	if p.QuotaLimit != nil {
		m["quota_limit"] = *p.QuotaLimit
	}
	if p.ClearQuotaLimit {
		m["quota_limit"] = nil
	}
	if p.AllowedModels != nil {
		m["allowed_models"] = *p.AllowedModels
	}
	if p.AllowedIPs != nil {
		m["allowed_ips"] = *p.AllowedIPs
	}
	if p.ExpiresAt != nil {
		m["expires_at"] = *p.ExpiresAt
	}
	if p.ClearExpiresAt {
		m["expires_at"] = nil
	}
	if p.Status != nil {
		m["status"] = int(*p.Status)
	}
	return m
}

func truncateHash(h string) string {
	if len(h) > 8 {
		return h[:8]
	}
	return h
}
