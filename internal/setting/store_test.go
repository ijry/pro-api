package setting

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func newStoreForTest(t *testing.T) (*store, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	s := &store{
		rdb: rdb,
		log: zap.NewNop(),
		ttl: 60 * time.Second,
	}
	return s, mr
}

func TestGet_NotFound_ReturnsFalse(t *testing.T) {
	s, _ := newStoreForTest(t)
	if _, ok := s.Get(context.Background(), "nope"); ok {
		t.Fatal("want not ok")
	}
}

func TestGet_AfterRedisSet_ReturnsValue(t *testing.T) {
	s, mr := newStoreForTest(t)
	if err := mr.Set("setting:foo", `"bar"`); err != nil {
		t.Fatal(err)
	}
	v, ok := s.Get(context.Background(), "foo")
	if !ok {
		t.Fatal("want ok")
	}
	if string(v) != `"bar"` {
		t.Fatalf("want \"bar\", got %s", v)
	}
}

func TestGet_HitsLocalCacheOnSecondCall(t *testing.T) {
	s, mr := newStoreForTest(t)
	_ = mr.Set("setting:foo", `42`)
	_, _ = s.Get(context.Background(), "foo")
	mr.Del("setting:foo")
	v, ok := s.Get(context.Background(), "foo")
	if !ok || string(v) != `42` {
		t.Fatalf("local cache miss: ok=%v v=%s", ok, v)
	}
}

func TestGet_LocalCacheExpiresAfterTTL(t *testing.T) {
	s, mr := newStoreForTest(t)
	s.ttl = 50 * time.Millisecond
	_ = mr.Set("setting:foo", `42`)
	_, _ = s.Get(context.Background(), "foo")
	mr.Del("setting:foo")
	time.Sleep(80 * time.Millisecond)
	if _, ok := s.Get(context.Background(), "foo"); ok {
		t.Fatal("expected expired local cache to miss")
	}
}

func TestGetString_UsesDefaultWhenMissing(t *testing.T) {
	s, _ := newStoreForTest(t)
	got := s.GetString(context.Background(), "missing", "fallback")
	if got != "fallback" {
		t.Fatalf("want fallback, got %s", got)
	}
}

func TestGetBool_True(t *testing.T) {
	s, mr := newStoreForTest(t)
	_ = mr.Set("setting:flag", `true`)
	if !s.GetBool(context.Background(), "flag", false) {
		t.Fatal("want true")
	}
}

func TestGetInt_FromJSONNumber(t *testing.T) {
	s, mr := newStoreForTest(t)
	_ = mr.Set("setting:n", `123`)
	if got := s.GetInt(context.Background(), "n", 0); got != 123 {
		t.Fatalf("want 123, got %d", got)
	}
}

func TestInvalidateLocal_ClearsEntry(t *testing.T) {
	s, _ := newStoreForTest(t)
	s.local.Store("foo", cachedValue{
		raw: []byte(`"cached"`),
		ts:  time.Now(),
	})
	s.invalidateLocal("foo")
	if _, ok := s.local.Load("foo"); ok {
		t.Fatal("expected local cache cleared")
	}
}
