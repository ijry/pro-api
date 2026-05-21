package cache

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newTestRedis(t *testing.T) *redis.Client {
	t.Helper()
	s := miniredis.RunT(t)
	return redis.NewClient(&redis.Options{Addr: s.Addr()})
}

func TestLoadScript_SetsSHA(t *testing.T) {
	ctx := context.Background()
	rdb := newTestRedis(t)
	src := `return 1`
	l, err := LoadScript(ctx, rdb, "test", src)
	if err != nil {
		t.Fatal(err)
	}
	if len(l.SHA()) != 40 {
		t.Fatalf("want 40-char sha, got %d (%q)", len(l.SHA()), l.SHA())
	}
}

func TestLuaScript_Run_BasicReturn(t *testing.T) {
	ctx := context.Background()
	rdb := newTestRedis(t)
	src := `return 42`
	l, _ := LoadScript(ctx, rdb, "test", src)
	r, err := l.Run(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if v, ok := r.(int64); !ok || v != 42 {
		t.Fatalf("want int64(42), got %T(%v)", r, r)
	}
}

func TestLuaScript_Run_AutoFallbackOnNoScript(t *testing.T) {
	ctx := context.Background()
	rdb := newTestRedis(t)
	src := `return 100`
	l, _ := LoadScript(ctx, rdb, "test", src)
	if err := rdb.ScriptFlush(ctx).Err(); err != nil {
		t.Fatal(err)
	}
	r, err := l.Run(ctx, nil)
	if err != nil {
		t.Fatalf("Run after FLUSH failed: %v", err)
	}
	if v, _ := r.(int64); v != 100 {
		t.Fatalf("want 100, got %v", r)
	}
}

func TestLuaScript_RunInts_ParsesIntArray(t *testing.T) {
	ctx := context.Background()
	rdb := newTestRedis(t)
	src := `return {1, 2, 3}`
	l, _ := LoadScript(ctx, rdb, "test", src)
	r, err := l.RunInts(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(r) != 3 || r[0] != 1 || r[1] != 2 || r[2] != 3 {
		t.Fatalf("want [1 2 3], got %v", r)
	}
}

func TestLuaScript_Run_WithKeysAndArgs(t *testing.T) {
	ctx := context.Background()
	rdb := newTestRedis(t)
	src := `
		redis.call('SET', KEYS[1], ARGV[1])
		return redis.call('GET', KEYS[1])
	`
	l, _ := LoadScript(ctx, rdb, "set_get", src)
	r, err := l.Run(ctx, []string{"k1"}, "v1")
	if err != nil {
		t.Fatal(err)
	}
	if v, _ := r.(string); v != "v1" {
		t.Fatalf("want v1, got %v", r)
	}
}
