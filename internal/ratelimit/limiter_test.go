package ratelimit

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ijry/pro-api/internal/util/clock"
	"go.uber.org/zap"
)

func newTestLimiter(t *testing.T) (*redisLimiter, *clock.Mock) {
	t.Helper()
	_, rdb := newTestRedis(t)
	mc := clock.NewMock(time.Unix(1716278400, 0))
	l, err := newRedisLimiter(context.Background(), Config{
		Cache: rdb,
		Log:   zap.NewNop(),
		Clock: mc,
	})
	if err != nil {
		t.Fatalf("newRedisLimiter: %v", err)
	}
	return l, mc
}

func TestAllowMulti_EmptyChecks_Allowed(t *testing.T) {
	l, _ := newTestLimiter(t)
	d := l.AllowMulti(context.Background(), nil)
	if !d.Allowed {
		t.Fatal("want allowed on empty checks")
	}
}

func TestAllowMulti_SingleDim_UnderLimit_Allowed(t *testing.T) {
	l, _ := newTestLimiter(t)
	checks := []Check{{
		Dimension: DimUserRPM,
		Key:       "ratelimit:user:1:rpm",
		Limit:     10,
		Window:    time.Minute,
		Cost:      1,
	}}
	for i := 0; i < 5; i++ {
		d := l.AllowMulti(context.Background(), checks)
		if !d.Allowed {
			t.Fatalf("iter=%d should be allowed", i)
		}
		if d.Per[0].Count != i+1 {
			t.Errorf("iter=%d count=%d want %d", i, d.Per[0].Count, i+1)
		}
	}
}

func TestAllowMulti_SingleDim_OverLimit_Denied(t *testing.T) {
	l, _ := newTestLimiter(t)
	checks := []Check{{
		Dimension: DimUserRPM, Key: "k1", Limit: 3, Window: time.Minute, Cost: 1,
	}}
	for i := 0; i < 3; i++ {
		if d := l.AllowMulti(context.Background(), checks); !d.Allowed {
			t.Fatalf("iter=%d should be allowed", i)
		}
	}
	d := l.AllowMulti(context.Background(), checks)
	if d.Allowed {
		t.Fatal("4th request should be denied")
	}
	if d.Dimension != DimUserRPM {
		t.Errorf("want denied dim=user_rpm got %s", d.Dimension)
	}
	if d.Limit != 3 {
		t.Errorf("want limit=3 got %d", d.Limit)
	}
}

func TestAllowMulti_ShortCircuit_DoesNotWriteLaterDims(t *testing.T) {
	l, mc := newTestLimiter(t)
	_ = mc
	// 先写满第一维度
	first := Check{Dimension: DimUserRPM, Key: "rpm:user:1", Limit: 1, Window: time.Minute, Cost: 1}
	if d := l.AllowMulti(context.Background(), []Check{first}); !d.Allowed {
		t.Fatal("first should allow")
	}
	// 此时再发组合检查;短路应不写后续 key
	d := l.AllowMulti(context.Background(), []Check{
		first,
		{Dimension: DimTokenRPM, Key: "rpm:token:1", Limit: 100, Window: time.Minute, Cost: 1},
		{Dimension: DimIPRPM, Key: "rpm:ip:1", Limit: 100, Window: time.Minute, Cost: 1},
	})
	if d.Allowed {
		t.Fatal("should be denied by first dim")
	}
	if d.Dimension != DimUserRPM {
		t.Errorf("want dim=user_rpm got %s", d.Dimension)
	}
	// 后续维度不应被写入
	count, _, _ := l.Stats(context.Background(), "rpm:token:1")
	if count != 0 {
		t.Errorf("token key should not be written; got count=%d", count)
	}
}

func TestAllowMulti_ZeroLimit_SkipsRedisCall(t *testing.T) {
	l, _ := newTestLimiter(t)
	checks := []Check{{Dimension: DimModelRPM, Key: "skip", Limit: 0, Window: time.Minute, Cost: 1}}
	d := l.AllowMulti(context.Background(), checks)
	if !d.Allowed {
		t.Fatal("limit=0 should allow")
	}
	count, _, _ := l.Stats(context.Background(), "skip")
	if count != 0 {
		t.Errorf("redis key should not be created; got %d", count)
	}
}

func TestAllowMulti_ZeroCost_SkipsRedisCall(t *testing.T) {
	l, _ := newTestLimiter(t)
	checks := []Check{{Dimension: DimUserRPM, Key: "ck", Limit: 10, Window: time.Minute, Cost: 0}}
	d := l.AllowMulti(context.Background(), checks)
	if !d.Allowed {
		t.Fatal("cost=0 should allow")
	}
	count, _, _ := l.Stats(context.Background(), "ck")
	if count != 0 {
		t.Errorf("cost=0 should not write")
	}
}

func TestAllowMulti_MultiDim_AllPass_ReturnsTightest(t *testing.T) {
	l, _ := newTestLimiter(t)
	// 三维度:user_rpm 还有 99,token_rpm 还有 9,ip_rpm 还有 599
	checks := []Check{
		{Dimension: DimUserRPM, Key: "u", Limit: 100, Window: time.Minute, Cost: 1},
		{Dimension: DimTokenRPM, Key: "t", Limit: 10, Window: time.Minute, Cost: 1},
		{Dimension: DimIPRPM, Key: "i", Limit: 600, Window: time.Minute, Cost: 1},
	}
	d := l.AllowMulti(context.Background(), checks)
	if !d.Allowed {
		t.Fatal("should allow")
	}
	if d.Dimension != DimTokenRPM {
		t.Errorf("want tightest=token_rpm got %s", d.Dimension)
	}
	if d.Remaining != 9 {
		t.Errorf("want remaining=9 got %d", d.Remaining)
	}
}

func TestAllowMulti_FailOpen_OnRedisError(t *testing.T) {
	mr, rdb := newTestRedis(t)
	l, err := newRedisLimiter(context.Background(), Config{Cache: rdb, Log: zap.NewNop(), Clock: clock.Real})
	if err != nil {
		t.Fatal(err)
	}
	mr.Close() // 关掉 redis
	d := l.AllowMulti(context.Background(), []Check{
		{Dimension: DimUserRPM, Key: "x", Limit: 10, Window: time.Minute, Cost: 1},
	})
	if !d.Allowed {
		t.Fatal("fail-open should allow")
	}
	if l.MetricsSnapshot()["failopen"] == 0 {
		t.Error("failopen counter should increment")
	}
}

func TestConsumeTPM_AlwaysWrites_EvenWhenOverLimit(t *testing.T) {
	l, _ := newTestLimiter(t)
	checks := []Check{{Dimension: DimUserTPM, Key: "u:tpm", Limit: 100, Window: time.Minute, Cost: 200}}
	if err := l.ConsumeTPM(context.Background(), checks); err != nil {
		t.Fatalf("ConsumeTPM: %v", err)
	}
	count, _, _ := l.Stats(context.Background(), "u:tpm")
	if count != 200 {
		t.Errorf("want count=200 got %d", count)
	}
}

func TestConsumeTPM_ZeroLimit_NoOp(t *testing.T) {
	l, _ := newTestLimiter(t)
	checks := []Check{{Dimension: DimUserTPM, Key: "x", Limit: 0, Window: time.Minute, Cost: 100}}
	if err := l.ConsumeTPM(context.Background(), checks); err != nil {
		t.Fatal(err)
	}
	count, _, _ := l.Stats(context.Background(), "x")
	if count != 0 {
		t.Errorf("want count=0 got %d", count)
	}
}

func TestStats_ReturnsCountAndOldest(t *testing.T) {
	l, _ := newTestLimiter(t)
	for i := 0; i < 3; i++ {
		l.AllowMulti(context.Background(), []Check{
			{Dimension: DimUserRPM, Key: "k", Limit: 10, Window: time.Minute, Cost: 1},
		})
	}
	count, oldest, err := l.Stats(context.Background(), "k")
	if err != nil {
		t.Fatal(err)
	}
	if count != 3 {
		t.Errorf("want count=3 got %d", count)
	}
	if oldest.IsZero() {
		t.Error("oldest should not be zero")
	}
}

func TestStats_MissingKey_ReturnsZero(t *testing.T) {
	l, _ := newTestLimiter(t)
	count, oldest, err := l.Stats(context.Background(), "nope")
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Errorf("want count=0 got %d", count)
	}
	if !oldest.IsZero() {
		t.Error("oldest should be zero on miss")
	}
}

func TestReset_DeletesKey(t *testing.T) {
	l, _ := newTestLimiter(t)
	l.AllowMulti(context.Background(), []Check{
		{Dimension: DimUserRPM, Key: "kk", Limit: 10, Window: time.Minute, Cost: 1},
	})
	if err := l.Reset(context.Background(), "kk"); err != nil {
		t.Fatal(err)
	}
	count, _, _ := l.Stats(context.Background(), "kk")
	if count != 0 {
		t.Errorf("want count=0 after reset; got %d", count)
	}
}

func TestNonceUniqueness_ConcurrentWrites(t *testing.T) {
	l, _ := newTestLimiter(t)
	// 100 个 cost=1 同时写,limit 足够大,ZCARD 应严格 = 100(无 ZADD 因成员重复被忽略)
	const N = 100
	errCh := make(chan error, N)
	for i := 0; i < N; i++ {
		go func() {
			d := l.AllowMulti(context.Background(), []Check{
				{Dimension: DimUserRPM, Key: "nonce", Limit: 1000, Window: time.Minute, Cost: 1},
			})
			if !d.Allowed {
				errCh <- errors.New("unexpectedly denied")
				return
			}
			errCh <- nil
		}()
	}
	for i := 0; i < N; i++ {
		if err := <-errCh; err != nil {
			t.Fatal(err)
		}
	}
	count, _, _ := l.Stats(context.Background(), "nonce")
	if count != N {
		t.Errorf("want count=%d, got %d (nonce collision?)", N, count)
	}
}

func TestNewRedisLimiter_NilCache_ReturnsError(t *testing.T) {
	if _, err := newRedisLimiter(context.Background(), Config{Cache: nil}); err == nil {
		t.Fatal("expected error on nil cache")
	}
}
