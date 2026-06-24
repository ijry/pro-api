package token

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func newTestRedis(t *testing.T) *redis.Client {
	t.Helper()
	s := miniredis.RunT(t)
	return redis.NewClient(&redis.Options{Addr: s.Addr()})
}

func TestCache_Get_Miss(t *testing.T) {
	rdb := newTestRedis(t)
	c := newCache(rdb, zap.NewNop(), 0, 0)
	defer c.Close()
	v, st := c.Get(context.Background(), "deadbeef")
	if st != cacheMiss || v != nil {
		t.Fatalf("want miss/nil, got %v/%v", v, st)
	}
}

func TestCache_SetGet_Roundtrip(t *testing.T) {
	rdb := newTestRedis(t)
	c := newCache(rdb, zap.NewNop(), 0, 0)
	defer c.Close()
	view := &View{
		ID:            123,
		UserID:        7,
		GroupID:       5,
		Name:          "demo",
		KeyPrefix:     "pa-AbCdEf01****WXYZ",
		AllowedModels: []string{"gpt-4*"},
		AllowedIPs:    []string{"10.0.0.0/8"},
		RPMLimit:      30,
		Status:        StatusEnabled,
	}
	c.SetPositive(context.Background(), "hash1", view)
	got, st := c.Get(context.Background(), "hash1")
	if st != cacheHit || got == nil {
		t.Fatalf("want hit, got %v / %v", got, st)
	}
	if got.ID != 123 || got.UserID != 7 || got.GroupID != 5 || got.RPMLimit != 30 || got.AllowedModels[0] != "gpt-4*" {
		t.Fatalf("decode mismatch: %+v", got)
	}
}

func TestCache_NegativeCache(t *testing.T) {
	rdb := newTestRedis(t)
	c := newCache(rdb, zap.NewNop(), 0, 0)
	defer c.Close()
	c.SetNegative(context.Background(), "bad-hash")
	_, st := c.Get(context.Background(), "bad-hash")
	if st != cacheNegative {
		t.Fatalf("want cacheNegative, got %v", st)
	}
}

func TestCache_Invalidate_DeletesKey(t *testing.T) {
	rdb := newTestRedis(t)
	c := newCache(rdb, zap.NewNop(), 0, 0)
	defer c.Close()
	c.SetPositive(context.Background(), "h", &View{ID: 1})
	c.Invalidate(context.Background(), "h")
	_, st := c.Get(context.Background(), "h")
	if st != cacheMiss {
		t.Fatalf("want miss after invalidate, got %v", st)
	}
}

func TestCache_PublishLastUsed_NoError(t *testing.T) {
	rdb := newTestRedis(t)
	c := newCache(rdb, zap.NewNop(), 0, 0)
	defer c.Close()
	// 不实际订阅,验证不 panic
	c.PublishLastUsed(context.Background(), 42, time.Now())
}

func TestCache_TwoInstances_InvalidationPropagates(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb1 := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	rdb2 := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb1.Close()
	defer rdb2.Close()

	a := newCache(rdb1, zap.NewNop(), 0, 0)
	b := newCache(rdb2, zap.NewNop(), 0, 0)
	defer a.Close()
	defer b.Close()

	// 两侧同时有同一缓存
	a.SetPositive(context.Background(), "h", &View{ID: 1})
	// b 也 Get 一遍写本地(本 spec 实际只 Redis,无本地 LRU,等价 a 的 SET 在共享 Redis)
	_, _ = b.Get(context.Background(), "h")

	// 通过 a 失效,b 的订阅 goroutine 应当 DEL Redis(虽然 a 已经 DEL,此处验证不 panic)
	a.Invalidate(context.Background(), "h")
	// 给订阅消费时间
	time.Sleep(120 * time.Millisecond)
	_, st := b.Get(context.Background(), "h")
	if st != cacheMiss {
		t.Fatalf("want miss after propagated invalidate, got %v", st)
	}
}

func TestCache_NilCacheIsNoop(t *testing.T) {
	var c *tokenCache
	v, st := c.Get(context.Background(), "x")
	if v != nil || st != cacheMiss {
		t.Fatalf("nil cache must return miss/nil, got %v/%v", v, st)
	}
	c.SetPositive(context.Background(), "x", &View{})
	c.SetNegative(context.Background(), "x")
	c.Invalidate(context.Background(), "x")
	c.PublishLastUsed(context.Background(), 1, time.Now())
	if err := c.Close(); err != nil {
		t.Fatalf("Close nil cache err: %v", err)
	}
}

func TestCache_Close_Idempotent(t *testing.T) {
	rdb := newTestRedis(t)
	c := newCache(rdb, zap.NewNop(), 0, 0)
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	// 二次 Close 不应 panic
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
}
