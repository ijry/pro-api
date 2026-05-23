package ratelimit

import (
	"sync"
	"time"
)

// thresholdEntry 是单条本地阈值缓存。
type thresholdEntry struct {
	limit  int
	expire time.Time
}

// thresholdCache 是 Planner 用的本地阈值缓存(Redis Pub/Sub 失效后 30s 兜底)。
type thresholdCache struct {
	mu  sync.RWMutex
	m   map[string]thresholdEntry
	ttl time.Duration
}

// newThresholdCache 构造一个本地阈值缓存,默认 TTL 由 ttl 决定。
func newThresholdCache(ttl time.Duration) *thresholdCache {
	if ttl <= 0 {
		ttl = 30 * time.Second
	}
	return &thresholdCache{
		m:   make(map[string]thresholdEntry),
		ttl: ttl,
	}
}

// set 写入缓存。perKeyTTL=0 时使用 cache 默认 TTL。
func (c *thresholdCache) set(key string, limit int, perKeyTTL time.Duration) {
	ttl := perKeyTTL
	if ttl <= 0 {
		ttl = c.ttl
	}
	c.mu.Lock()
	c.m[key] = thresholdEntry{limit: limit, expire: time.Now().Add(ttl)}
	c.mu.Unlock()
}

// get 取缓存。命中返回 (limit, true);过期 / 不存在返回 (0, false)。
func (c *thresholdCache) get(key string) (int, bool) {
	c.mu.RLock()
	e, ok := c.m[key]
	c.mu.RUnlock()
	if !ok {
		return 0, false
	}
	if time.Now().After(e.expire) {
		// 过期清理(写锁短暂)
		c.mu.Lock()
		// 重新检查避免竞争
		if cur, ok2 := c.m[key]; ok2 && time.Now().After(cur.expire) {
			delete(c.m, key)
		}
		c.mu.Unlock()
		return 0, false
	}
	return e.limit, true
}

// purge 清空全部缓存。setting Pub/Sub 失效 ratelimit.* 时调用。
func (c *thresholdCache) purge() {
	c.mu.Lock()
	c.m = make(map[string]thresholdEntry)
	c.mu.Unlock()
}

// scaleByGroup 把 limit 按 group ratio 倒数放大(VIP 折扣 → 限流宽松)。
//
//	ratio = 0   → identity
//	ratio < 0   → identity(防御)
//	ratio >= 1  → identity(M1 不做缩小:大于 1 的 ratio 表示无折扣)
//	0 < ratio<1 → limit / ratio(向下取整)
//	limit = 0   → 0(不限永远不限)
func scaleByGroup(limit int, ratio float64) int {
	if limit <= 0 {
		return 0
	}
	if ratio <= 0 || ratio >= 1.0 {
		return limit
	}
	return int(float64(limit) / ratio)
}
