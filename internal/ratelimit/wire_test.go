package ratelimit

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/ijry/pro-api/internal/app"
	"github.com/ijry/pro-api/internal/util/clock"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func TestWireRateLimit_NilApp_ReturnsError(t *testing.T) {
	if err := WireRateLimit(nil); err == nil {
		t.Fatal("want error on nil app")
	}
}

func TestWireRateLimit_NilCache_ReturnsError(t *testing.T) {
	a := &app.Application{}
	if err := WireRateLimit(a); err == nil {
		t.Fatal("want error on nil cache")
	}
}

func TestWireRateLimit_NilSetting_ReturnsError(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()
	a := &app.Application{Cache: rdb}
	if err := WireRateLimit(a); err == nil {
		t.Fatal("want error on nil setting")
	}
}

func TestWireRateLimit_Success_PopulatesLimiterAndPlanner(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	st := newFakeSetting()
	a := &app.Application{
		Cache:   rdb,
		Setting: st,
		Log:     zap.NewNop(),
		Clock:   clock.Real,
	}
	if err := WireRateLimit(a); err != nil {
		t.Fatalf("wire: %v", err)
	}
	if a.Limiter == nil {
		t.Fatal("Limiter not set")
	}
	if _, ok := a.Limiter.(Limiter); !ok {
		t.Fatalf("Limiter is %T, want Limiter interface", a.Limiter)
	}
	if PlannerFrom(a) == nil {
		t.Fatal("PlannerFrom returned nil")
	}
	// shutdown should not error
	if err := a.Shutdown(context.Background()); err != nil {
		t.Errorf("shutdown: %v", err)
	}
	if PlannerFrom(a) != nil {
		t.Error("planner should be cleared after shutdown")
	}
}

func TestWireRateLimit_PubSub_PurgesPlannerCache(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	st := newFakeSetting()
	a := &app.Application{
		Cache:   rdb,
		Setting: st,
		Log:     zap.NewNop(),
		Clock:   clock.Real,
	}
	if err := WireRateLimit(a); err != nil {
		t.Fatal(err)
	}
	defer a.Shutdown(context.Background())

	// 等待 PubSub 订阅者 goroutine 就绪
	time.Sleep(100 * time.Millisecond)

	p := PlannerFrom(a)
	// 让 cache 填充
	_ = p.PlanRPM(context.Background(), PlanInput{UserID: 1, TokenID: 1, IP: "1.2.3.4"})
	if _, ok := p.cache.get("user_rpm_default"); !ok {
		t.Fatal("cache should be populated after first PlanRPM")
	}
	// 发送 invalidate
	if err := rdb.Publish(context.Background(), settingInvalidateChannel, "ratelimit.user_default_rpm").Err(); err != nil {
		t.Fatal(err)
	}
	// 等到 cache 清空(最多 1 秒)
	if !waitUntil(3*time.Second, func() bool {
		_, ok := p.cache.get("user_rpm_default")
		return !ok
	}) {
		t.Fatal("planner cache should be purged after pub/sub")
	}
}

func TestWireRateLimit_PubSub_IgnoresNonRatelimitKeys(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	st := newFakeSetting()
	a := &app.Application{Cache: rdb, Setting: st, Log: zap.NewNop(), Clock: clock.Real}
	if err := WireRateLimit(a); err != nil {
		t.Fatal(err)
	}
	defer a.Shutdown(context.Background())

	// 等待 PubSub 订阅者 goroutine 就绪
	time.Sleep(100 * time.Millisecond)

	p := PlannerFrom(a)
	_ = p.PlanRPM(context.Background(), PlanInput{UserID: 1, TokenID: 1, IP: "1.2.3.4"})
	if _, ok := p.cache.get("user_rpm_default"); !ok {
		t.Fatal("cache should be populated")
	}
	// 发送其它 key — 不应触发 purge
	_ = rdb.Publish(context.Background(), settingInvalidateChannel, "billing.something").Err()
	// 给 goroutine 一些时间执行(若处理错误也来得及)
	if waitUntil(200*time.Millisecond, func() bool {
		_, ok := p.cache.get("user_rpm_default")
		return !ok
	}) {
		t.Fatal("cache should NOT be purged for non-ratelimit keys")
	}
}

// waitUntil 轮询 fn,直到返回 true 或超时;返回是否成功满足。
func waitUntil(timeout time.Duration, fn func() bool) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if fn() {
			return true
		}
		time.Sleep(5 * time.Millisecond)
	}
	return false
}

var _ = errors.New
