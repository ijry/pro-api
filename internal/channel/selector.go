package channel

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/ijry/pro-api/pkg/apierr"
)

type channelSelector struct {
	cache   *channelCache
	breaker Breaker

	randMu sync.Mutex
	rnd    *rand.Rand
}

func newSelector(c *channelCache, b Breaker, seed int64) *channelSelector {
	return &channelSelector{
		cache:   c,
		breaker: b,
		rnd:     rand.New(rand.NewSource(seed)),
	}
}

// Select 选出最优渠道。
func (s *channelSelector) Select(_ context.Context, model string, hint SelectHint) (*Channel, error) {
	candidates := s.cache.ChannelsByModel(model)
	if len(candidates) == 0 {
		return nil, apierr.New(apierr.CodeNoChannel, "no channel configured for model "+model)
	}

	var filtered []*Channel
	for _, c := range candidates {
		if c.Status != 0 {
			continue
		}
		if inInt64Slice(hint.Excluded, c.ID) {
			continue
		}
		if s.breaker.State(c.ID) == StateOpen {
			continue
		}
		if !tagsMatch(c.Tags, hint.Tags) {
			continue
		}
		if hint.GroupID > 0 && c.GroupID != 0 && c.GroupID != hint.GroupID {
			continue
		}
		filtered = append(filtered, c)
	}
	if len(filtered) == 0 {
		return nil, apierr.New(apierr.CodeNoChannel,
			"no available channel for model "+model+" (all filtered)")
	}

	// 取最高优先级
	topPrio := filtered[0].Priority
	for _, c := range filtered {
		if c.Priority > topPrio {
			topPrio = c.Priority
		}
	}
	var pool []*Channel
	for _, c := range filtered {
		if c.Priority == topPrio {
			pool = append(pool, c)
		}
	}

	return s.weightedRandom(pool), nil
}

func (s *channelSelector) weightedRandom(pool []*Channel) *Channel {
	total := 0
	for _, c := range pool {
		total += maxInt(c.Weight, 1)
	}
	s.randMu.Lock()
	r := s.rnd.Intn(total)
	s.randMu.Unlock()
	for _, c := range pool {
		w := maxInt(c.Weight, 1)
		if r < w {
			return c
		}
		r -= w
	}
	return pool[len(pool)-1]
}

// ReportSuccess 转发给 breaker。
func (s *channelSelector) ReportSuccess(channelID int64, latency time.Duration) {
	s.breaker.RecordSuccess(channelID, latency)
}

// ReportFailure 转发给 breaker。
func (s *channelSelector) ReportFailure(channelID int64, err error) {
	s.breaker.RecordFailure(channelID, err)
}

// getBreaker 暴露内部 breaker 给 WithRetry。
func (s *channelSelector) getBreaker() Breaker { return s.breaker }
