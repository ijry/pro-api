package channel

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/ijry/pro-api/internal/util/clock"
	"errors"
	"github.com/ijry/pro-api/pkg/apierr"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	defaultWindowSeconds    = 60
	defaultFailThreshold    = 5
	defaultCoolDownSeconds  = 30
	breakerRefreshTick      = 5 * time.Second
	breakerKeyTTL           = 24 * time.Hour
	defaultProbeTimeout     = 30 * time.Second
)

type breakerConfig struct {
	WindowSeconds   int
	FailThreshold   int
	CoolDownSeconds int
}

type localBreakerState struct {
	state        BreakerState
	openedAt     time.Time
	consecFail   int
	lastProbeAt  time.Time
	successCount int
	failureCount int
}

type redisBreaker struct {
	rdb    *redis.Client
	clock  clock.Clock
	log    *zap.Logger

	cfgMu sync.RWMutex
	cfg   breakerConfig

	mu     sync.RWMutex
	states map[int64]localBreakerState

	pubsub *redis.PubSub
	stop   chan struct{}
	wg     sync.WaitGroup
}

func newBreaker(rdb *redis.Client, clk clock.Clock, log *zap.Logger) *redisBreaker {
	return &redisBreaker{
		rdb:    rdb,
		clock:  clk,
		log:    log,
		cfg:    breakerConfig{WindowSeconds: defaultWindowSeconds, FailThreshold: defaultFailThreshold, CoolDownSeconds: defaultCoolDownSeconds},
		states: make(map[int64]localBreakerState),
		stop:   make(chan struct{}),
	}
}

// Start 启动后台 goroutine（订阅 Pub/Sub + cron 扫 OPEN）。
func (b *redisBreaker) Start(ctx context.Context) {
	b.pubsub = b.rdb.Subscribe(ctx, "proapi:channel:breaker")
	b.wg.Add(1)
	go b.loop(ctx)
	b.log.Info("breaker subscribed pubsub=proapi:channel:breaker")
}

func (b *redisBreaker) loop(ctx context.Context) {
	defer b.wg.Done()
	sub := b.pubsub.Channel()
	ticker := time.NewTicker(breakerRefreshTick)
	defer ticker.Stop()
	for {
		select {
		case <-b.stop:
			return
		case msg, ok := <-sub:
			if !ok {
				return
			}
			b.handlePubSub(msg)
		case <-ticker.C:
			b.scanOpenForTransition(ctx)
		}
	}
}

func (b *redisBreaker) handlePubSub(msg *redis.Message) {
	var payload struct {
		ChannelID int64        `json:"channel_id"`
		State     BreakerState `json:"state"`
		TS        int64        `json:"ts"`
	}
	if err := json.Unmarshal([]byte(msg.Payload), &payload); err != nil {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	s := b.states[payload.ChannelID]
	s.state = payload.State
	if payload.State == StateOpen {
		s.openedAt = time.UnixMilli(payload.TS)
	}
	if payload.State == StateClosed {
		s.consecFail = 0
		s.openedAt = time.Time{}
	}
	b.states[payload.ChannelID] = s
}

func (b *redisBreaker) scanOpenForTransition(ctx context.Context) {
	b.cfgMu.RLock()
	coolDown := time.Duration(b.cfg.CoolDownSeconds) * time.Second
	b.cfgMu.RUnlock()

	now := b.clock.Now()
	b.mu.RLock()
	var transit []int64
	for id, s := range b.states {
		if s.state == StateOpen && now.Sub(s.openedAt) >= coolDown {
			transit = append(transit, id)
		}
	}
	b.mu.RUnlock()

	for _, id := range transit {
		b.tryHalfOpen(ctx, id)
	}
}

func (b *redisBreaker) tryHalfOpen(ctx context.Context, channelID int64) {
	lockKey := fmt.Sprintf("channel:breaker:transition_lock:%d", channelID)
	ok, err := b.rdb.SetNX(ctx, lockKey, "1", 10*time.Second).Result()
	if err != nil || !ok {
		return
	}
	defer b.rdb.Del(ctx, lockKey)

	// 二次检查 Redis state
	state, _ := b.rdb.HGet(ctx, redisKeyBreaker(channelID), "state").Result()
	if state != string(StateOpen) {
		return
	}
	b.doTransition(ctx, channelID, StateHalfOpen, "cool_down_elapsed")
}

func redisKeyBreaker(id int64) string {
	return fmt.Sprintf("channel:breaker:%d", id)
}

func (b *redisBreaker) doTransition(ctx context.Context, channelID int64, to BreakerState, reason string) {
	key := redisKeyBreaker(channelID)
	now := b.clock.Now().UnixMilli()

	pipe := b.rdb.TxPipeline()
	pipe.HSet(ctx, key, "state", string(to))
	switch to {
	case StateOpen:
		pipe.HSet(ctx, key, "opened_at", now)
	case StateClosed:
		pipe.HSet(ctx, key, "opened_at", 0)
		pipe.HSet(ctx, key, "consec_fail", 0)
		pipe.HSet(ctx, key, "last_probe_at", 0)
	case StateHalfOpen:
		pipe.HSet(ctx, key, "last_probe_at", 0)
	}
	pipe.Expire(ctx, key, breakerKeyTTL)
	if _, err := pipe.Exec(ctx); err != nil {
		b.log.Error("breaker: transition redis failed", zap.Error(err))
	}

	b.mu.Lock()
	s := b.states[channelID]
	s.state = to
	if to == StateOpen {
		s.openedAt = b.clock.Now()
	}
	if to == StateClosed {
		s.consecFail = 0
		s.openedAt = time.Time{}
	}
	b.states[channelID] = s
	b.mu.Unlock()

	// 广播
	payload, _ := json.Marshal(map[string]any{
		"channel_id": channelID,
		"state":      to,
		"ts":         now,
		"reason":     reason,
	})
	if err := b.rdb.Publish(ctx, "proapi:channel:breaker", payload).Err(); err != nil {
		b.log.Warn("breaker: publish failed", zap.Error(err))
	}
	b.log.Info("breaker transitioned",
		zap.Int64("channel_id", channelID),
		zap.String("to", string(to)),
		zap.String("reason", reason))
}

// State 返回当前熔断状态（读本地缓存，默认 closed）。
func (b *redisBreaker) State(channelID int64) BreakerState {
	b.mu.RLock()
	s, ok := b.states[channelID]
	b.mu.RUnlock()
	if !ok {
		return StateClosed
	}
	return s.state
}

// Snapshot 返回完整快照。
func (b *redisBreaker) Snapshot(channelID int64) HealthSnapshot {
	b.mu.RLock()
	s := b.states[channelID]
	b.mu.RUnlock()

	snap := HealthSnapshot{
		ChannelID:  channelID,
		State:      s.state,
		ConsecFail: s.consecFail,
		Recent: Counters{
			Success: s.successCount,
			Failure: s.failureCount,
			Window:  b.cfg.WindowSeconds,
		},
	}
	if !s.openedAt.IsZero() {
		snap.OpenedAt = &s.openedAt
	}
	if !s.lastProbeAt.IsZero() {
		snap.LastProbeAt = &s.lastProbeAt
	}
	return snap
}

// RecordSuccess 记录成功，可能触发 HALF_OPEN→CLOSED。
func (b *redisBreaker) RecordSuccess(channelID int64, _ time.Duration) {
	ctx := context.Background()
	key := redisKeyBreaker(channelID)
	pipe := b.rdb.TxPipeline()
	pipe.HSet(ctx, key, "consec_fail", 0)
	pipe.HIncrBy(ctx, key, "success_count", 1)
	cmd := pipe.HGet(ctx, key, "state")
	_, _ = pipe.Exec(ctx)

	b.mu.Lock()
	s := b.states[channelID]
	s.consecFail = 0
	s.successCount++
	b.states[channelID] = s
	b.mu.Unlock()

	if cmd.Val() == string(StateHalfOpen) {
		b.doTransition(ctx, channelID, StateClosed, "half_open probe succeeded")
	}
}

// RecordFailure 记录失败，可能触发 CLOSED→OPEN 或 HALF_OPEN→OPEN。
func (b *redisBreaker) RecordFailure(channelID int64, err error) {
	if err == nil {
		return
	}
	var e *apierr.Error; ok := errors.As(err, &e)
	if !ok {
		return
	}
	switch e.Code {
	case apierr.CodeUpstreamError,
		apierr.CodeUpstreamTimeout,
		apierr.CodeUpstreamUnavail,
		apierr.CodeUpstreamRateLimit:
		// 计入
	default:
		return
	}

	ctx := context.Background()
	key := redisKeyBreaker(channelID)

	pipe := b.rdb.TxPipeline()
	cmdFail := pipe.HIncrBy(ctx, key, "consec_fail", 1)
	pipe.HIncrBy(ctx, key, "failure_count", 1)
	cmdState := pipe.HGet(ctx, key, "state")
	if _, err2 := pipe.Exec(ctx); err2 != nil && err2 != redis.Nil {
		b.log.Warn("breaker: record failure pipe failed", zap.Error(err2))
		return
	}
	state := cmdState.Val()
	if state == "" {
		state = string(StateClosed)
	}
	consec := int(cmdFail.Val())

	b.mu.Lock()
	s := b.states[channelID]
	s.consecFail = consec
	s.failureCount++
	b.states[channelID] = s
	b.mu.Unlock()

	b.cfgMu.RLock()
	threshold := b.cfg.FailThreshold
	b.cfgMu.RUnlock()

	if state == string(StateClosed) && consec >= threshold {
		b.doTransition(ctx, channelID, StateOpen, "consec_fail >= threshold")
	} else if state == string(StateHalfOpen) {
		b.doTransition(ctx, channelID, StateOpen, "half_open probe failed")
	}
}

// AcquireProbe 在 HALF_OPEN 状态下抢 probe 令牌。
func (b *redisBreaker) AcquireProbe(ctx context.Context, channelID int64, timeout time.Duration) (bool, func(), error) {
	lockKey := fmt.Sprintf("channel:probe_lock:%d", channelID)
	val := strconv.FormatInt(b.clock.Now().UnixNano(), 10) + "-" + strconv.FormatInt(rand.Int63(), 10)

	ok, err := b.rdb.SetNX(ctx, lockKey, val, timeout).Result()
	if err != nil {
		return false, nil, apierr.New(apierr.CodeCache, "acquire probe: "+err.Error())
	}
	if !ok {
		return false, nil, nil
	}

	release := func() {
		script := `if redis.call("get", KEYS[1]) == ARGV[1] then return redis.call("del", KEYS[1]) else return 0 end`
		_, _ = b.rdb.Eval(context.Background(), script, []string{lockKey}, val).Result()
	}
	return true, release, nil
}

// ForceTransition 强制状态切换（管理员 enable 时调用）。
func (b *redisBreaker) ForceTransition(channelID int64, to BreakerState) error {
	b.doTransition(context.Background(), channelID, to, "force_transition")
	return nil
}

// ClearState 清除 Redis 中的 breaker 状态（channel 删除时）。
func (b *redisBreaker) ClearState(ctx context.Context, channelID int64) {
	b.rdb.Del(ctx, redisKeyBreaker(channelID))
	b.mu.Lock()
	delete(b.states, channelID)
	b.mu.Unlock()
}

// Close 停止后台 goroutine。
func (b *redisBreaker) Close() error {
	close(b.stop)
	b.wg.Wait()
	if b.pubsub != nil {
		return b.pubsub.Close()
	}
	return nil
}
