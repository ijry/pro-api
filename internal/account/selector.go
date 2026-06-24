package account

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ijry/pro-api/internal/channel"
)

type selectorImpl struct {
	repo    Repo
	breaker Breaker
	rnd     *rand.Rand
	mu      sync.Mutex
}

// NewSelector 构造 4-策略 Selector(top_k / round_robin / weighted / max_remaining)。
// breaker 可为 nil(只用于 ReportFailure 路径)。seed 用于 weighted 随机选号。
func NewSelector(r Repo, b Breaker, seed int64) Selector {
	return &selectorImpl{repo: r, breaker: b, rnd: rand.New(rand.NewSource(seed))}
}

func (s *selectorImpl) Select(ctx context.Context, ch *channel.Channel, hint SelectHint) (*Account, error) {
	pool, err := s.poolFor(ctx, ch)
	if err != nil {
		return nil, err
	}
	cands := make([]*Account, 0, len(pool))
	for _, a := range pool {
		if a.Status != StatusActive {
			continue
		}
		if containsInt64(hint.Excluded, a.ID) {
			continue
		}
		cands = append(cands, a)
	}
	if len(cands) == 0 {
		return nil, ErrPoolEmpty
	}

	strategy := ch.AccountStrategy
	if strategy == "" {
		strategy = "top_k"
	}
	switch strategy {
	case "top_k":
		return s.topK(cands, int(ch.AccountTopK)), nil
	case "round_robin":
		return s.roundRobin(cands, ch.ID), nil
	case "weighted":
		return s.weighted(cands), nil
	case "max_remaining":
		return s.maxRemaining(cands), nil
	default:
		return s.topK(cands, int(ch.AccountTopK)), nil
	}
}

func (s *selectorImpl) topK(cands []*Account, k int) *Account {
	sort.SliceStable(cands, func(i, j int) bool {
		return remPct(cands[i]) > remPct(cands[j])
	})
	if k <= 0 || k > len(cands) {
		k = len(cands)
	}
	return s.weighted(cands[:k])
}

func (s *selectorImpl) maxRemaining(cands []*Account) *Account {
	best := cands[0]
	for _, c := range cands[1:] {
		if remPct(c) > remPct(best) {
			best = c
		}
	}
	return best
}

func (s *selectorImpl) weighted(cands []*Account) *Account {
	// 不写入 c.Weight:repo 返回的 *Account 可能在并发 Select 间共享,直接改字段会触发竞态。
	weightOf := func(c *Account) int {
		if c.Weight <= 0 {
			return 1
		}
		return c.Weight
	}
	total := 0
	for _, c := range cands {
		total += weightOf(c)
	}
	s.mu.Lock()
	n := s.rnd.Intn(total)
	s.mu.Unlock()
	acc := 0
	for _, c := range cands {
		acc += weightOf(c)
		if n < acc {
			return c
		}
	}
	return cands[len(cands)-1]
}

var rrCursor sync.Map // channelID -> int

func (s *selectorImpl) roundRobin(cands []*Account, channelID int64) *Account {
	sort.Slice(cands, func(i, j int) bool { return cands[i].ID < cands[j].ID })
	cur, _ := rrCursor.LoadOrStore(channelID, 0)
	n := cur.(int) % len(cands)
	rrCursor.Store(channelID, n+1)
	return cands[n]
}

func (s *selectorImpl) poolFor(ctx context.Context, ch *channel.Channel) ([]*Account, error) {
	return s.repo.ListByChannel(ctx, ch.ID)
}

func remPct(a *Account) float64 {
	if a.Quota5hTotal == nil || a.Quota5hRemaining == nil || *a.Quota5hTotal == 0 {
		return 1.0
	}
	return float64(*a.Quota5hRemaining) / float64(*a.Quota5hTotal)
}

func containsInt64(xs []int64, x int64) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}

func (s *selectorImpl) ReportSuccess(accountID int64, _ time.Duration) {
	// 使用 ResetFailures(map-based),否则 GORM Updates(struct) 会跳过 ConsecFailures=0 的零值写入。
	_ = s.repo.ResetFailures(context.Background(), accountID)
}

func (s *selectorImpl) ReportFailure(accountID int64, err error, headers http.Header) {
	if s.breaker == nil {
		return
	}
	ctx := context.Background()
	switch classifyFailure(err) {
	case failRateLimited:
		reset := parseResetAt(headers)
		if reset.IsZero() {
			reset = time.Now().Add(5 * time.Minute)
		}
		_ = s.breaker.MarkCooldown(ctx, accountID, reset, err.Error())
	case failTokenExpired:
		_ = s.breaker.MarkExpired(ctx, accountID, err.Error())
	case failInvalidCred:
		_ = s.breaker.MarkInvalid(ctx, accountID, err.Error())
	default:
		n, _ := s.breaker.IncConsecFailure(ctx, accountID)
		if n >= 5 {
			_ = s.breaker.MarkCooldown(ctx, accountID, time.Now().Add(60*time.Second), "consec fail >=5")
		}
	}
}

type failClass int

const (
	failUnknown failClass = iota
	failRateLimited
	failTokenExpired
	failInvalidCred
)

func classifyFailure(err error) failClass {
	if err == nil {
		return failUnknown
	}
	s := strings.ToLower(err.Error())
	switch {
	case strings.Contains(s, "429") || strings.Contains(s, "rate limit"):
		return failRateLimited
	case strings.Contains(s, "401") || strings.Contains(s, "token expired") || strings.Contains(s, "invalid_token"):
		return failTokenExpired
	case strings.Contains(s, "403") || strings.Contains(s, "invalid_grant") || strings.Contains(s, "revoked"):
		return failInvalidCred
	}
	return failUnknown
}

func parseResetAt(h http.Header) time.Time {
	for _, k := range []string{"anthropic-ratelimit-tokens-reset", "x-ratelimit-reset-tokens", "retry-after"} {
		v := h.Get(k)
		if v == "" {
			continue
		}
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			return t
		}
		if d, ok := parseSeconds(v); ok {
			return time.Now().Add(d)
		}
	}
	return time.Time{}
}

func parseSeconds(s string) (time.Duration, bool) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	if err != nil {
		return 0, false
	}
	return time.Duration(n) * time.Second, true
}
