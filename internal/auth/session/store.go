package session

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/ijry/pro-api/internal/util/clock"
	"github.com/ijry/pro-api/pkg/apierr"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// 常量。
const (
	SessionIDPrefix     = "sess_"
	SessionIDRandBytes  = 32
	revokeChannel       = "proapi:session:revoke"
	defaultMirrorBatch  = 100
	defaultMirrorTicker = 30 * time.Second
	defaultRestore      = true
)

// Store 是 session 存储抽象。
type Store interface {
	Create(ctx context.Context, userID int64, role int8, ip, ua string) (*Session, error)
	Get(ctx context.Context, id string) (*Session, error)
	Touch(ctx context.Context, id string) error
	Revoke(ctx context.Context, id string) error
	RevokeAllForUser(ctx context.Context, userID int64) error
	Close() error
}

// Config 是 New 的参数。
type Config struct {
	TTL              time.Duration
	Sliding          bool
	MirrorBatchSize  int           // 0 走默认 100
	MirrorBatchEvery time.Duration // 0 走默认 30s
	RestoreOnStart   *bool         // nil 走默认 true
}

// Deps 是 New 的依赖。
type Deps struct {
	DB    Repository
	Cache *redis.Client
	IDGen IDGenerator
	Clock clock.Clock
	Log   *zap.Logger
}

// IDGenerator 提供 audit / 兜底用 id。
type IDGenerator interface {
	Generate() int64
}

// store 实现 Store。
type store struct {
	repo  Repository
	rdb   *redis.Client
	clock clock.Clock
	log   *zap.Logger
	cfg   Config

	mu        sync.Mutex
	mirrorQ   []mirrorOp
	stopCh    chan struct{}
	stopWg    sync.WaitGroup
	mirrorWok bool
	pubsub    *redis.PubSub
}

type mirrorOp struct {
	id       string
	lastSeen time.Time
	expires  time.Time
}

// New 构造 Store。
func New(deps Deps, cfg Config) (Store, error) {
	if deps.DB == nil {
		return nil, errors.New("session: repo required")
	}
	if deps.Cache == nil {
		return nil, errors.New("session: redis required")
	}
	if cfg.TTL <= 0 {
		cfg.TTL = 30 * 24 * time.Hour
	}
	if cfg.MirrorBatchSize <= 0 {
		cfg.MirrorBatchSize = defaultMirrorBatch
	}
	if cfg.MirrorBatchEvery <= 0 {
		cfg.MirrorBatchEvery = defaultMirrorTicker
	}
	clk := deps.Clock
	if clk == nil {
		clk = clock.Real
	}
	log := deps.Log
	if log == nil {
		log = zap.NewNop()
	}

	s := &store{
		repo:   deps.DB,
		rdb:    deps.Cache,
		clock:  clk,
		log:    log,
		cfg:    cfg,
		stopCh: make(chan struct{}),
	}

	// 启动回放
	restore := defaultRestore
	if cfg.RestoreOnStart != nil {
		restore = *cfg.RestoreOnStart
	}
	if restore {
		if err := s.restoreFromDB(context.Background()); err != nil {
			log.Warn("session: restore from DB failed", zap.Error(err))
		}
	}

	// 启动 mirror 后台 goroutine
	s.stopWg.Add(1)
	go s.runMirror()
	s.mirrorWok = true

	// 启动 pubsub 订阅(目前仅为占位 + 日志,M1 不带本地缓存)
	s.pubsub = deps.Cache.Subscribe(context.Background(), revokeChannel)
	s.stopWg.Add(1)
	go s.runPubSub()

	return s, nil
}

// Create 创建 session。
func (s *store) Create(ctx context.Context, userID int64, role int8, ip, ua string) (*Session, error) {
	id, err := genID()
	if err != nil {
		return nil, apierr.Wrap(apierr.CodeInternal, "session: gen id", err)
	}
	now := s.clock.Now().UTC()
	exp := now.Add(s.cfg.TTL)

	if err := s.writeRedis(ctx, id, userID, role, ip, ua, now, now, exp); err != nil {
		return nil, err
	}

	// DB mirror(失败仅 log)
	if err := s.repo.Insert(ctx, &DBSession{
		ID: id, UserID: userID, IP: ip, UserAgent: ua,
		CreatedAt: now, LastSeenAt: now, ExpiresAt: exp,
	}); err != nil {
		s.log.Warn("session: db insert failed", zap.String("id", id), zap.Error(err))
	}

	return &Session{
		ID: id, UserID: userID, Role: role, IP: ip, UserAgent: ua,
		CreatedAt: now, LastSeen: now, ExpiresAt: exp,
	}, nil
}

func (s *store) writeRedis(ctx context.Context, id string, userID int64, role int8, ip, ua string, created, lastSeen, exp time.Time) error {
	pipe := s.rdb.Pipeline()
	pipe.HSet(ctx, sessionKey(id), map[string]any{
		"user_id":    userID,
		"role":       role,
		"ip":         ip,
		"ua":         ua,
		"created_at": created.UnixMilli(),
		"last_seen":  lastSeen.UnixMilli(),
		"expires_at": exp.UnixMilli(),
	})
	pipe.PExpire(ctx, sessionKey(id), s.cfg.TTL)
	pipe.SAdd(ctx, userKey(userID), id)
	pipe.PExpire(ctx, userKey(userID), s.cfg.TTL)
	if _, err := pipe.Exec(ctx); err != nil {
		return apierr.Wrap(apierr.CodeCache, "session: redis write", err)
	}
	return nil
}

// Get 从 Redis 取 session。
func (s *store) Get(ctx context.Context, id string) (*Session, error) {
	if id == "" {
		return nil, nil
	}
	m, err := s.rdb.HGetAll(ctx, sessionKey(id)).Result()
	if err != nil {
		return nil, apierr.Wrap(apierr.CodeCache, "session: redis hgetall", err)
	}
	if len(m) == 0 {
		return nil, nil
	}
	sess, err := parseSession(id, m)
	if err != nil {
		s.log.Warn("session: parse failed", zap.String("id", id), zap.Error(err))
		return nil, nil
	}
	if s.clock.Now().After(sess.ExpiresAt) {
		// 立刻清掉
		_ = s.rdb.Del(ctx, sessionKey(id)).Err()
		return nil, nil
	}
	return sess, nil
}

// Touch 更新 last_seen 并(若 Sliding)滑动延期。
func (s *store) Touch(ctx context.Context, id string) error {
	if id == "" {
		return nil
	}
	sess, err := s.Get(ctx, id)
	if err != nil || sess == nil {
		return err
	}
	now := s.clock.Now().UTC()
	newExp := sess.ExpiresAt
	if s.cfg.Sliding {
		newExp = now.Add(s.cfg.TTL)
	}
	pipe := s.rdb.Pipeline()
	pipe.HSet(ctx, sessionKey(id), map[string]any{
		"last_seen":  now.UnixMilli(),
		"expires_at": newExp.UnixMilli(),
	})
	if s.cfg.Sliding {
		pipe.PExpire(ctx, sessionKey(id), s.cfg.TTL)
		pipe.PExpire(ctx, userKey(sess.UserID), s.cfg.TTL)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return apierr.Wrap(apierr.CodeCache, "session: touch", err)
	}
	// 入队 mirror
	s.enqueueMirror(mirrorOp{id: id, lastSeen: now, expires: newExp})
	return nil
}

// Revoke 删 Redis + 发布 + DB 标记。
func (s *store) Revoke(ctx context.Context, id string) error {
	if id == "" {
		return nil
	}
	// 先取 user_id 以维护 user set
	uidStr, _ := s.rdb.HGet(ctx, sessionKey(id), "user_id").Result()

	pipe := s.rdb.Pipeline()
	pipe.Del(ctx, sessionKey(id))
	if uidStr != "" {
		pipe.SRem(ctx, userKey(parseInt64(uidStr)), id)
	}
	pipe.Publish(ctx, revokeChannel, id)
	if _, err := pipe.Exec(ctx); err != nil {
		s.log.Warn("session: revoke redis", zap.Error(err))
	}
	if err := s.repo.MarkRevoked(ctx, id, s.clock.Now().UTC()); err != nil {
		s.log.Warn("session: db mark revoked", zap.String("id", id), zap.Error(err))
	}
	return nil
}

// RevokeAllForUser 强制下线该用户所有 session。
func (s *store) RevokeAllForUser(ctx context.Context, userID int64) error {
	ids, err := s.rdb.SMembers(ctx, userKey(userID)).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		s.log.Warn("session: smembers", zap.Error(err))
	}
	for _, id := range ids {
		_ = s.Revoke(ctx, id)
	}
	// 即使 Redis 空,也要保证 DB mirror 中所有未撤销的都打标
	if err := s.repo.MarkAllRevokedForUser(ctx, userID, s.clock.Now().UTC()); err != nil {
		s.log.Warn("session: db mark all revoked", zap.Int64("user_id", userID), zap.Error(err))
	}
	_ = s.rdb.Del(ctx, userKey(userID)).Err()
	return nil
}

// Close 停后台 goroutine 并 flush mirror。
func (s *store) Close() error {
	s.mu.Lock()
	select {
	case <-s.stopCh:
	default:
		close(s.stopCh)
	}
	s.mu.Unlock()
	s.flushMirror(context.Background())
	if s.pubsub != nil {
		_ = s.pubsub.Close()
	}
	s.stopWg.Wait()
	return nil
}

// --- 内部:DB mirror ---

func (s *store) enqueueMirror(op mirrorOp) {
	s.mu.Lock()
	s.mirrorQ = append(s.mirrorQ, op)
	full := len(s.mirrorQ) >= s.cfg.MirrorBatchSize
	s.mu.Unlock()
	if full {
		s.flushMirror(context.Background())
	}
}

func (s *store) flushMirror(ctx context.Context) {
	s.mu.Lock()
	if len(s.mirrorQ) == 0 {
		s.mu.Unlock()
		return
	}
	ops := s.mirrorQ
	s.mirrorQ = nil
	s.mu.Unlock()

	// 合并相同 id 的最新一条
	latest := make(map[string]mirrorOp, len(ops))
	for _, op := range ops {
		latest[op.id] = op
	}
	for id, op := range latest {
		if err := s.repo.UpdateLastSeen(ctx, id, op.lastSeen, op.expires); err != nil {
			s.log.Warn("session: mirror update", zap.String("id", id), zap.Error(err))
		}
	}
}

func (s *store) runMirror() {
	defer s.stopWg.Done()
	ticker := s.clock.NewTicker(s.cfg.MirrorBatchEvery)
	defer ticker.Stop()
	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C():
			s.flushMirror(context.Background())
		}
	}
}

func (s *store) runPubSub() {
	defer s.stopWg.Done()
	if s.pubsub == nil {
		return
	}
	ch := s.pubsub.Channel()
	for {
		select {
		case <-s.stopCh:
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			s.log.Debug("session: revoke event", zap.String("id", msg.Payload))
		}
	}
}

// restoreFromDB 把未过期且未撤销的 session 写回 Redis。
func (s *store) restoreFromDB(ctx context.Context) error {
	now := s.clock.Now().UTC()
	items, err := s.repo.ListActive(ctx, now, 10000)
	if err != nil {
		return err
	}
	for _, it := range items {
		// 跳过 Redis 已有的
		exists, err := s.rdb.Exists(ctx, sessionKey(it.ID)).Result()
		if err == nil && exists > 0 {
			continue
		}
		_ = s.writeRedis(ctx, it.ID, it.UserID, 0, it.IP, it.UserAgent, it.CreatedAt, it.LastSeenAt, it.ExpiresAt)
	}
	s.log.Info("session: restored sessions from DB", zap.Int("count", len(items)))
	return nil
}

// --- helpers ---

func sessionKey(id string) string  { return "session:" + id }
func userKey(uid int64) string     { return fmt.Sprintf("session:user:%d", uid) }
func parseInt64(s string) int64    { v, _ := strconv.ParseInt(s, 10, 64); return v }
func parseInt8(s string) int8      { v, _ := strconv.ParseInt(s, 10, 8); return int8(v) }
func parseMillis(s string) time.Time {
	v, _ := strconv.ParseInt(s, 10, 64)
	return time.UnixMilli(v).UTC()
}

func parseSession(id string, m map[string]string) (*Session, error) {
	if len(m) == 0 {
		return nil, errors.New("empty hash")
	}
	return &Session{
		ID:        id,
		UserID:    parseInt64(m["user_id"]),
		Role:      parseInt8(m["role"]),
		IP:        m["ip"],
		UserAgent: m["ua"],
		CreatedAt: parseMillis(m["created_at"]),
		LastSeen:  parseMillis(m["last_seen"]),
		ExpiresAt: parseMillis(m["expires_at"]),
	}, nil
}

func genID() (string, error) {
	buf := make([]byte, SessionIDRandBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return SessionIDPrefix + base64.RawURLEncoding.EncodeToString(buf), nil
}

// ensure json import alive for future serialisation tests
var _ = json.RawMessage(nil)
