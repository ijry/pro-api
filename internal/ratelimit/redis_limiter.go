package ratelimit

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/ijry/pro-api/internal/util/clock"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Config 是 NewRedisLimiter 的参数。
type Config struct {
	Cache *redis.Client
	Log   *zap.Logger
	Clock clock.Clock
}

// redisLimiter 是 Limiter 的默认实现。
type redisLimiter struct {
	rdb     *redis.Client
	log     *zap.Logger
	clock   clock.Clock
	scripts *scripts

	nonceSeed uint64
	nonceCtr  uint64

	metrics struct {
		allowed   atomic.Uint64
		denied    atomic.Uint64
		failopen  atomic.Uint64
		tpmWrites atomic.Uint64
	}
}

// NewRedisLimiter 构造一个 Limiter。Cache 必传;Clock / Log 缺省为 Real / Nop。
func NewRedisLimiter(ctx context.Context, cfg Config) (Limiter, error) {
	return newRedisLimiter(ctx, cfg)
}

func newRedisLimiter(ctx context.Context, cfg Config) (*redisLimiter, error) {
	if cfg.Cache == nil {
		return nil, fmt.Errorf("ratelimit: cache is required")
	}
	log := cfg.Log
	if log == nil {
		log = zap.NewNop()
	}
	clk := cfg.Clock
	if clk == nil {
		clk = clock.Real
	}
	s, err := loadScripts(ctx, cfg.Cache)
	if err != nil {
		return nil, err
	}
	// 随机 nonce seed:用 crypto/rand 生成,避免多实例 / 多次启动撞
	seed, err := randomUint64()
	if err != nil {
		// 不致命:fallback 用 unix nano(测试环境也够用)
		seed = uint64(clk.Now().UnixNano())
	}
	return &redisLimiter{
		rdb:       cfg.Cache,
		log:       log,
		clock:     clk,
		scripts:   s,
		nonceSeed: seed,
	}, nil
}

// AllowMulti 短路语义:遇到第一个被拒立即返回。
func (l *redisLimiter) AllowMulti(ctx context.Context, checks []Check) Decision {
	if len(checks) == 0 {
		return Decision{Allowed: true}
	}
	if ctx == nil {
		ctx = context.Background()
	}
	per := make([]PerDimDecision, 0, len(checks))
	now := l.clock.Now()
	for i, ck := range checks {
		nonce := l.makeNonce(i)
		if ck.Limit <= 0 || ck.Cost == 0 {
			per = append(per, PerDimDecision{
				Dimension: ck.Dimension,
				Allowed:   true,
				Count:     0,
				Limit:     ck.Limit,
				Reset:     now.Add(ck.Window),
			})
			continue
		}
		res, err := l.scripts.sliding.RunInts(ctx,
			[]string{fullKey(ck.Key)},
			int64(ck.Window/time.Millisecond),
			int64(ck.Limit),
			now.UnixMilli(),
			int64(ck.Cost),
			int64(1), // enforce = 1 (gating)
			nonce,
		)
		if err != nil {
			l.metrics.failopen.Add(1)
			l.log.Warn("ratelimit: sliding script failed, fail-open",
				zap.Error(err),
				zap.String("dim", string(ck.Dimension)),
				zap.String("key", ck.Key),
			)
			per = append(per, PerDimDecision{
				Dimension: ck.Dimension,
				Allowed:   true,
				Count:     0,
				Limit:     ck.Limit,
				Reset:     now.Add(ck.Window),
			})
			continue
		}
		ok := res[0] == 1
		count := int(res[1])
		lim := int(res[2])
		reset := time.UnixMilli(res[3])
		pd := PerDimDecision{
			Dimension: ck.Dimension,
			Allowed:   ok,
			Count:     count,
			Limit:     lim,
			Reset:     reset,
		}
		per = append(per, pd)
		if !ok {
			l.metrics.denied.Add(1)
			ckCopy := ck
			remaining := lim - count
			if remaining < 0 {
				remaining = 0
			}
			return Decision{
				Allowed:   false,
				Denied:    &ckCopy,
				Dimension: ck.Dimension,
				Remaining: remaining,
				Limit:     lim,
				Reset:     reset,
				Per:       per,
			}
		}
	}
	// 全部通过 — 找最紧张的维度填顶层字段
	tightest := selectTightest(per)
	l.metrics.allowed.Add(1)
	remaining := tightest.Limit - tightest.Count
	if remaining < 0 {
		remaining = 0
	}
	return Decision{
		Allowed:   true,
		Dimension: tightest.Dimension,
		Remaining: remaining,
		Limit:     tightest.Limit,
		Reset:     tightest.Reset,
		Per:       per,
	}
}

// ConsumeTPM 仅扣 TPM 维度,不阻塞调用方;即使超额也写入。
func (l *redisLimiter) ConsumeTPM(ctx context.Context, checks []Check) error {
	if len(checks) == 0 {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	now := l.clock.Now()
	var lastErr error
	for i, ck := range checks {
		if ck.Limit <= 0 || ck.Cost == 0 {
			continue
		}
		nonce := l.makeNonce(i)
		_, err := l.scripts.sliding.RunInts(ctx,
			[]string{fullKey(ck.Key)},
			int64(ck.Window/time.Millisecond),
			int64(ck.Limit),
			now.UnixMilli(),
			int64(ck.Cost),
			int64(0), // enforce=0,counting
			nonce,
		)
		if err != nil {
			l.metrics.failopen.Add(1)
			l.log.Warn("ratelimit: tpm consume failed",
				zap.Error(err),
				zap.String("dim", string(ck.Dimension)),
				zap.String("key", ck.Key),
			)
			lastErr = err
			continue
		}
		l.metrics.tpmWrites.Add(uint64(ck.Cost))
	}
	return lastErr
}

// Stats 返回某 redis key 的当前窗口内计数 + 最早成员时间。
// 注意:Stats 不应用窗口过滤(实现简单);若需精确滑动窗口口径,使用 Lua 脚本走 cost=0 路径。
func (l *redisLimiter) Stats(ctx context.Context, key string) (int, time.Time, error) {
	if key == "" {
		return 0, time.Time{}, errors.New("ratelimit: key required")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	fk := fullKey(key)
	count, err := l.rdb.ZCard(ctx, fk).Result()
	if err != nil {
		return 0, time.Time{}, fmt.Errorf("ratelimit: zcard: %w", err)
	}
	if count == 0 {
		return 0, time.Time{}, nil
	}
	// 取最早一条:ZRANGE 0 0 WITHSCORES
	first, err := l.rdb.ZRangeWithScores(ctx, fk, 0, 0).Result()
	if err != nil {
		return int(count), time.Time{}, fmt.Errorf("ratelimit: zrange: %w", err)
	}
	if len(first) == 0 {
		return int(count), time.Time{}, nil
	}
	oldest := time.UnixMilli(int64(first[0].Score))
	return int(count), oldest, nil
}

// Reset 删除 key。
func (l *redisLimiter) Reset(ctx context.Context, key string) error {
	if key == "" {
		return errors.New("ratelimit: key required")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return l.rdb.Del(ctx, fullKey(key)).Err()
}

// MetricsSnapshot 暴露 4 个计数器的当前值。
func (l *redisLimiter) MetricsSnapshot() map[string]uint64 {
	return map[string]uint64{
		"allowed":    l.metrics.allowed.Load(),
		"denied":     l.metrics.denied.Load(),
		"failopen":   l.metrics.failopen.Load(),
		"tpm_writes": l.metrics.tpmWrites.Load(),
	}
}

// Close 是占位关闭,limiter 自身无后台 goroutine。
func (l *redisLimiter) Close() error { return nil }

// makeNonce 生成本次调用某维度的唯一 nonce 字符串。
// 用 seed 异或全局计数器 + 维度 index,避免并发同一秒撞。
func (l *redisLimiter) makeNonce(dimIndex int) string {
	v := atomic.AddUint64(&l.nonceCtr, 1) ^ l.nonceSeed
	return strconv.FormatUint(v, 36) + "-" + strconv.Itoa(dimIndex)
}

// fullKey 把 Key 拼到 redis 前缀。
func fullKey(key string) string {
	if strings.HasPrefix(key, "ratelimit:") {
		return key
	}
	return "ratelimit:" + key
}

// selectTightest 找 remaining 最小(count 最高占比)的维度;同 ratio 取列表中较后的。
func selectTightest(per []PerDimDecision) PerDimDecision {
	if len(per) == 0 {
		return PerDimDecision{}
	}
	best := per[0]
	bestRatio := computeUseRatio(best)
	for _, p := range per[1:] {
		r := computeUseRatio(p)
		if r >= bestRatio {
			best = p
			bestRatio = r
		}
	}
	return best
}

// computeUseRatio = count / limit;limit=0 视为 0(不算紧张)。
func computeUseRatio(p PerDimDecision) float64 {
	if p.Limit <= 0 {
		return 0
	}
	return float64(p.Count) / float64(p.Limit)
}

// randomUint64 生成一个随机 uint64,用于 nonce seed。
func randomUint64() (uint64, error) {
	max := new(big.Int).Lsh(big.NewInt(1), 64)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return 0, err
	}
	// big.Int 不直接转 uint64,走 bytes
	buf := make([]byte, 8)
	b := n.Bytes()
	copy(buf[8-len(b):], b)
	return binary.BigEndian.Uint64(buf), nil
}
