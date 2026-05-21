package setting

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ErrNotFound 表示 key 不存在(任何层)。
var ErrNotFound = errors.New("setting: key not found")

const (
	redisKeyPrefix    = "setting:"
	redisInvalidateCh = "proapi:setting:invalidate"
	redisTTL          = 10 * time.Minute
	localDefaultTTL   = 60 * time.Second
)

// Store 是 setting 的公开接口。
type Store interface {
	Get(ctx context.Context, key string) (json.RawMessage, bool)
	GetString(ctx context.Context, key string, def string) string
	GetBool(ctx context.Context, key string, def bool) bool
	GetInt(ctx context.Context, key string, def int) int
	GetFloat(ctx context.Context, key string, def float64) float64
	GetJSON(ctx context.Context, key string, dest any) error
	Put(ctx context.Context, key string, val any, actor int64) error
	// GetSecret 读取 key,若 value 是 ENC(...) 形态字符串,自动解密为明文。
	// 非加密 / 非字符串 / key 不存在 → 返对应错误(ErrNotFound / ErrNotEncrypted / 解密错)。
	GetSecret(ctx context.Context, key string, dec Decryptor) (string, error)
	// ListAll 返回所有 setting 行(按 key 升序)。仅供管理后台列表使用。
	ListAll(ctx context.Context) ([]Setting, error)
	Close() error
}

// Config 是 New 的参数。
type Config struct {
	DB       *gorm.DB
	Cache    *redis.Client
	Log      *zap.Logger
	LocalTTL time.Duration // 0 = 用默认 60s
}

// store 是 Store 的默认实现。
type store struct {
	db     *gorm.DB
	rdb    *redis.Client
	log    *zap.Logger
	local  sync.Map
	ttl    time.Duration
	sub    *redis.PubSub
	stopCh chan struct{}
	stopWg sync.WaitGroup
}

type cachedValue struct {
	raw json.RawMessage
	ts  time.Time
}

func redisKey(key string) string { return redisKeyPrefix + key }

// Get 三层查找:local → redis → DB。
func (s *store) Get(ctx context.Context, key string) (json.RawMessage, bool) {
	if v, ok := s.local.Load(key); ok {
		cv := v.(cachedValue)
		if time.Since(cv.ts) <= s.ttl {
			return cv.raw, true
		}
	}
	raw, err := s.rdb.Get(ctx, redisKey(key)).Bytes()
	if err == nil {
		s.local.Store(key, cachedValue{raw: raw, ts: time.Now()})
		return raw, true
	}
	if !errors.Is(err, redis.Nil) {
		s.log.Warn("setting: redis get failed", zap.String("key", key), zap.Error(err))
	}
	return s.getFromDB(ctx, key)
}

// getFromDB 查 DB,命中后回填 Redis。
func (s *store) getFromDB(ctx context.Context, key string) (json.RawMessage, bool) {
	if s.db == nil {
		return nil, false
	}
	var row Setting
	err := s.db.WithContext(ctx).Where("`key` = ?", key).First(&row).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			s.log.Warn("setting: db get failed", zap.String("key", key), zap.Error(err))
		}
		return nil, false
	}
	if err := s.rdb.Set(ctx, redisKey(key), []byte(row.Value), redisTTL).Err(); err != nil {
		s.log.Warn("setting: redis backfill failed", zap.String("key", key), zap.Error(err))
	}
	s.local.Store(key, cachedValue{raw: row.Value, ts: time.Now()})
	return row.Value, true
}

// GetString 从 JSON-encoded value 中解出字符串。
func (s *store) GetString(ctx context.Context, key string, def string) string {
	v, ok := s.Get(ctx, key)
	if !ok {
		return def
	}
	var out string
	if err := json.Unmarshal(v, &out); err != nil {
		out = string(v)
	}
	return out
}

// GetBool 取布尔。
func (s *store) GetBool(ctx context.Context, key string, def bool) bool {
	v, ok := s.Get(ctx, key)
	if !ok {
		return def
	}
	var out bool
	if err := json.Unmarshal(v, &out); err != nil {
		switch string(v) {
		case "true":
			return true
		case "false":
			return false
		}
		return def
	}
	return out
}

// GetInt 取整数。
func (s *store) GetInt(ctx context.Context, key string, def int) int {
	v, ok := s.Get(ctx, key)
	if !ok {
		return def
	}
	var out int
	if err := json.Unmarshal(v, &out); err != nil {
		if n, err2 := strconv.Atoi(string(v)); err2 == nil {
			return n
		}
		return def
	}
	return out
}

// GetFloat 取浮点。
func (s *store) GetFloat(ctx context.Context, key string, def float64) float64 {
	v, ok := s.Get(ctx, key)
	if !ok {
		return def
	}
	var out float64
	if err := json.Unmarshal(v, &out); err != nil {
		if n, err2 := strconv.ParseFloat(string(v), 64); err2 == nil {
			return n
		}
		return def
	}
	return out
}

// GetJSON 把 value 反序列化到 dest。
func (s *store) GetJSON(ctx context.Context, key string, dest any) error {
	v, ok := s.Get(ctx, key)
	if !ok {
		return ErrNotFound
	}
	return json.Unmarshal(v, dest)
}

// Put 写 DB(UPSERT)→ DEL Redis → PUBLISH 失效。
// val 任意 Go 值,内部 json.Marshal。actor=0 表示系统写入。
func (s *store) Put(ctx context.Context, key string, val any, actor int64) error {
	if s.db == nil {
		return fmt.Errorf("setting: DB not configured")
	}
	raw, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("setting: marshal value: %w", err)
	}
	rec := Setting{
		Key:       key,
		Value:     raw,
		UpdatedAt: time.Now().UTC(),
	}
	if actor != 0 {
		rec.UpdatedBy = &actor
	}
	if err := s.db.WithContext(ctx).Save(&rec).Error; err != nil {
		return fmt.Errorf("setting: db upsert %s: %w", key, err)
	}
	if err := s.rdb.Del(ctx, redisKey(key)).Err(); err != nil {
		s.log.Warn("setting: redis del failed", zap.String("key", key), zap.Error(err))
	}
	s.invalidateLocal(key)
	if err := s.rdb.Publish(ctx, redisInvalidateCh, key).Err(); err != nil {
		s.log.Warn("setting: publish failed", zap.String("key", key), zap.Error(err))
	}
	return nil
}

// Close 关停所有后台 goroutine。
func (s *store) Close() error {
	if s.stopCh != nil {
		select {
		case <-s.stopCh:
		default:
			close(s.stopCh)
		}
	}
	if s.sub != nil {
		_ = s.sub.Close()
	}
	s.stopWg.Wait()
	return nil
}

// invalidateLocal 清掉本地一级缓存某个 key。
func (s *store) invalidateLocal(key string) {
	s.local.Delete(key)
}

// New 启动一个 Store 实例并订阅 Pub/Sub 失效。
// ctx 用于初始化阶段(订阅 Subscribe);后台 goroutine 内部用独立 context。
func New(ctx context.Context, cfg Config) (Store, error) {
	ttl := cfg.LocalTTL
	if ttl == 0 {
		ttl = localDefaultTTL
	}
	log := cfg.Log
	if log == nil {
		log = zap.NewNop()
	}
	s := &store{
		db:     cfg.DB,
		rdb:    cfg.Cache,
		log:    log,
		ttl:    ttl,
		stopCh: make(chan struct{}),
	}
	if cfg.Cache != nil {
		s.sub = cfg.Cache.Subscribe(ctx, redisInvalidateCh)
		s.stopWg.Add(1)
		go s.runInvalidator()
	}
	return s, nil
}

func (s *store) runInvalidator() {
	defer s.stopWg.Done()
	ch := s.sub.Channel()
	for {
		select {
		case <-s.stopCh:
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			s.invalidateLocal(msg.Payload)
		}
	}
}
