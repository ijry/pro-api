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

// getFromDB 在 T07 阶段是 stub(返回 not found),T08 改为真实查询 DB。
func (s *store) getFromDB(ctx context.Context, key string) (json.RawMessage, bool) {
	return nil, false
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

// Put 在 T08 阶段补 DB 写入;T07 stub。
func (s *store) Put(ctx context.Context, key string, val any, actor int64) error {
	return fmt.Errorf("setting: Put not implemented in T07 (added in T08)")
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
