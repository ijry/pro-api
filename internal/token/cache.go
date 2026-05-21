package token

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// 缓存常量。Pub/Sub channel 命名见 spec §4.3。
const (
	cacheKeyPrefix        = "token:"
	cacheInvalidateCh     = "proapi:token:invalidate"
	cacheLastUsedCh       = "proapi:token:lastused"
	negativeMarker        = "0"
	defaultCacheTTL       = 5 * time.Minute
	defaultNegativeTTL    = 30 * time.Second
	defaultFlushInterval  = 30 * time.Second
	defaultFlushBatchSize = 200
)

func cacheKey(hash string) string { return cacheKeyPrefix + hash }

// cacheValue 是 Redis 中 token 的 JSON 快照(去掉敏感字段)。
// 缓存中不存 LastUsedAt(spec §11.12 决策:cache 不承担 last_used 实时性)。
type cacheValue struct {
	ID            int64      `json:"id"`
	UserID        int64      `json:"user_id"`
	Name          string     `json:"name"`
	KeyPrefix     string     `json:"key_prefix"`
	QuotaLimit    *int64     `json:"quota_limit,omitempty"`
	QuotaUsed     int64      `json:"quota_used"`
	AllowedModels []string   `json:"allowed_models"`
	AllowedIPs    []string   `json:"allowed_ips"`
	RPMLimit      int        `json:"rpm_limit"`
	TPMLimit      int        `json:"tpm_limit"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	Status        int8       `json:"status"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

func toCacheValue(v *View) cacheValue {
	return cacheValue{
		ID:            v.ID,
		UserID:        v.UserID,
		Name:          v.Name,
		KeyPrefix:     v.KeyPrefix,
		QuotaLimit:    v.QuotaLimit,
		QuotaUsed:     v.QuotaUsed,
		AllowedModels: v.AllowedModels,
		AllowedIPs:    v.AllowedIPs,
		RPMLimit:      v.RPMLimit,
		TPMLimit:      v.TPMLimit,
		ExpiresAt:     v.ExpiresAt,
		Status:        v.Status,
		CreatedAt:     v.CreatedAt,
		UpdatedAt:     v.UpdatedAt,
	}
}

func (cv cacheValue) toView() *View {
	return &View{
		ID:            cv.ID,
		UserID:        cv.UserID,
		Name:          cv.Name,
		KeyPrefix:     cv.KeyPrefix,
		QuotaLimit:    cv.QuotaLimit,
		QuotaUsed:     cv.QuotaUsed,
		AllowedModels: cv.AllowedModels,
		AllowedIPs:    cv.AllowedIPs,
		RPMLimit:      cv.RPMLimit,
		TPMLimit:      cv.TPMLimit,
		ExpiresAt:     cv.ExpiresAt,
		Status:        cv.Status,
		CreatedAt:     cv.CreatedAt,
		UpdatedAt:     cv.UpdatedAt,
	}
}

// cacheHitMiss 是 cache.Get 的三态结果。
type cacheHitMiss int

const (
	cacheMiss     cacheHitMiss = iota // 完全未命中,调用方走 DB
	cacheHit                          // 命中正向缓存
	cacheNegative                     // 命中负缓存(此 key 已确认不存在),调用方直接返回 CodeInvalidToken
)

// tokenCache 封装 Redis 读写 + Pub/Sub 失效订阅。
type tokenCache struct {
	rdb           *redis.Client
	log           *zap.Logger
	ttl           time.Duration
	negativeTTL   time.Duration
	stopCh        chan struct{}
	sub           *redis.PubSub
	wg            sync.WaitGroup
	publishCtx    context.Context
	publishCancel context.CancelFunc
}

// newCache 构造 tokenCache 并启动 Pub/Sub 订阅。
//
//   - rdb 为 nil 时返回 nil(意味着不启用缓存,所有读写走 DB);适用于测试。
//   - ttl/negativeTTL <= 0 走默认值。
func newCache(rdb *redis.Client, log *zap.Logger, ttl, negativeTTL time.Duration) *tokenCache {
	if rdb == nil {
		return nil
	}
	if ttl <= 0 {
		ttl = defaultCacheTTL
	}
	if negativeTTL <= 0 {
		negativeTTL = defaultNegativeTTL
	}
	if log == nil {
		log = zap.NewNop()
	}
	publishCtx, publishCancel := context.WithCancel(context.Background())
	c := &tokenCache{
		rdb:           rdb,
		log:           log,
		ttl:           ttl,
		negativeTTL:   negativeTTL,
		stopCh:        make(chan struct{}),
		publishCtx:    publishCtx,
		publishCancel: publishCancel,
	}
	c.sub = rdb.Subscribe(publishCtx, cacheInvalidateCh, cacheLastUsedCh)
	c.wg.Add(1)
	go c.runSubscriber()
	return c
}

// runSubscriber 接收失效广播(本实例不维护本地 LRU,仅清除自己的 Redis 缓存键 — spec §4.3 说明)。
// invalidate 收到 key_hash 时 DEL Redis 键(同一 Redis,理论上其他实例 publish 已经 DEL,这里幂等再 DEL 一次容错)。
// lastused 当前不做任何事(spec §11.12)— 留接口供未来扩展。
func (c *tokenCache) runSubscriber() {
	defer c.wg.Done()
	ch := c.sub.Channel()
	for {
		select {
		case <-c.stopCh:
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			switch msg.Channel {
			case cacheInvalidateCh:
				_ = c.rdb.Del(c.publishCtx, cacheKey(msg.Payload)).Err()
			case cacheLastUsedCh:
				// no-op:cache 不承担 last_used 同步
			}
		}
	}
}

// Get 按 sha256(plaintext) 查 cache。
//
//   - hit  : 反序列化 View,返回 cacheHit
//   - negative: 命中负缓存,返回 cacheNegative
//   - miss : 完全无键,返回 cacheMiss
//   - error: Redis 错(非 redis.Nil)— 视为 miss,记日志
func (c *tokenCache) Get(ctx context.Context, hash string) (*View, cacheHitMiss) {
	if c == nil {
		return nil, cacheMiss
	}
	raw, err := c.rdb.Get(ctx, cacheKey(hash)).Bytes()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			c.log.Warn("token: cache get failed", zap.Error(err))
		}
		return nil, cacheMiss
	}
	if string(raw) == negativeMarker {
		return nil, cacheNegative
	}
	var cv cacheValue
	if err := json.Unmarshal(raw, &cv); err != nil {
		c.log.Warn("token: cache value unmarshal failed", zap.Error(err))
		return nil, cacheMiss
	}
	return cv.toView(), cacheHit
}

// SetPositive 写入正向缓存。
func (c *tokenCache) SetPositive(ctx context.Context, hash string, v *View) {
	if c == nil {
		return
	}
	raw, err := json.Marshal(toCacheValue(v))
	if err != nil {
		c.log.Warn("token: cache value marshal failed", zap.Error(err))
		return
	}
	if err := c.rdb.Set(ctx, cacheKey(hash), raw, c.ttl).Err(); err != nil {
		c.log.Warn("token: cache set failed", zap.Error(err))
	}
}

// SetNegative 写入负缓存,防止字典爆破打爆 DB。
func (c *tokenCache) SetNegative(ctx context.Context, hash string) {
	if c == nil {
		return
	}
	if err := c.rdb.Set(ctx, cacheKey(hash), negativeMarker, c.negativeTTL).Err(); err != nil {
		c.log.Warn("token: cache set negative failed", zap.Error(err))
	}
}

// Invalidate 显式失效 + 广播。
func (c *tokenCache) Invalidate(ctx context.Context, hash string) {
	if c == nil || hash == "" {
		return
	}
	if err := c.rdb.Del(ctx, cacheKey(hash)).Err(); err != nil {
		c.log.Warn("token: cache del failed", zap.Error(err))
	}
	if err := c.rdb.Publish(ctx, cacheInvalidateCh, hash).Err(); err != nil {
		c.log.Warn("token: cache publish invalidate failed", zap.Error(err))
	}
}

// PublishLastUsed 广播 last_used 事件,供其他实例感知(M1 内部不消费,留接口)。
// payload 形如 "{token_id}:{unix_ms}"。
func (c *tokenCache) PublishLastUsed(ctx context.Context, tokenID int64, t time.Time) {
	if c == nil {
		return
	}
	payload := fmt.Sprintf("%d:%d", tokenID, t.UnixMilli())
	if err := c.rdb.Publish(ctx, cacheLastUsedCh, payload).Err(); err != nil {
		c.log.Warn("token: cache publish lastused failed", zap.Error(err))
	}
}

// Close 停止订阅 goroutine。
func (c *tokenCache) Close() error {
	if c == nil {
		return nil
	}
	select {
	case <-c.stopCh:
		// already closed
	default:
		close(c.stopCh)
	}
	c.publishCancel()
	if c.sub != nil {
		_ = c.sub.Close()
	}
	c.wg.Wait()
	return nil
}
