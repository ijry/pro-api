package pricing

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/ijry/pro-api/pkg/apierr"
	"go.uber.org/zap"
)

// ruleCache 是规则索引。
//
// 数据结构:atomic.Value 持有 []*Rule 全量切片;读 O(1),写整体替换(无锁读)。
// 规则总量 < 1k,O(N) 扫描可接受(spec §4.8)。
type ruleCache struct {
	mu    sync.Mutex // 写时持锁;读不持锁(走 atomic)
	rules atomic.Value
}

// newRuleCache 构造空 cache。
func newRuleCache() *ruleCache {
	c := &ruleCache{}
	c.rules.Store([]*Rule(nil))
	return c
}

// All 返回所有 enabled 规则的切片快照。
func (c *ruleCache) All() []*Rule {
	v := c.rules.Load()
	if v == nil {
		return nil
	}
	return v.([]*Rule)
}

// Set 整体替换(写者持锁,避免重叠写)。
func (c *ruleCache) Set(rules []*Rule) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.rules.Store(rules)
}

// Refresh 从 DB 重载缓存(只取 enabled 规则)。
func (s *service) Refresh(ctx context.Context) error {
	var rows []*Rule
	if err := s.db.WithContext(ctx).
		Where("status = ?", RuleStatusEnabled).
		Order("priority ASC, id DESC").
		Find(&rows).Error; err != nil {
		return apierr.New(apierr.CodeDatabase, err.Error())
	}
	s.cache.Set(rows)
	s.log.Debug("pricing: rules refreshed", zap.Int("count", len(rows)))
	return nil
}
