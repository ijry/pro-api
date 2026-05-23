package ratelimit

import (
	"context"
	"strconv"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

// newTestRedis 启动 miniredis 并返回 client + close func。
func newTestRedis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() {
		_ = rdb.Close()
	})
	return mr, rdb
}

func TestLoadScripts_Success(t *testing.T) {
	_, rdb := newTestRedis(t)
	s, err := loadScripts(context.Background(), rdb)
	if err != nil {
		t.Fatalf("loadScripts: %v", err)
	}
	if s.sliding == nil {
		t.Fatal("sliding script not loaded")
	}
	if s.sliding.SHA() == "" {
		t.Fatal("SHA missing")
	}
}

func TestLoadScripts_NilClient_ReturnsError(t *testing.T) {
	if _, err := loadScripts(context.Background(), nil); err == nil {
		t.Fatal("expected error on nil client")
	}
}

// TestSlidingWindow_LimitZero_AllowsWithoutWrite 直接调脚本验证 limit=0 路径
func TestSlidingWindow_LimitZero_AllowsWithoutWrite(t *testing.T) {
	mr, rdb := newTestRedis(t)
	s, err := loadScripts(context.Background(), rdb)
	if err != nil {
		t.Fatal(err)
	}
	res, err := s.sliding.RunInts(context.Background(),
		[]string{"ratelimit:user:1:rpm"},
		int64(60000), // window_ms
		int64(0),     // limit
		int64(1000),  // now_ms
		int64(1),     // cost
		int64(1),     // enforce
		"nonce123",
	)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	// {ok=1, count=0, limit=0, reset, wrote=0}
	if res[0] != 1 {
		t.Errorf("want ok=1, got %d", res[0])
	}
	if res[4] != 0 {
		t.Errorf("want wrote=0, got %d", res[4])
	}
	// key should not exist
	if mr.Exists("ratelimit:user:1:rpm") {
		t.Error("key should not be created when limit=0")
	}
}

func TestSlidingWindow_UnderLimit_AllowsAndWrites(t *testing.T) {
	mr, rdb := newTestRedis(t)
	s, _ := loadScripts(context.Background(), rdb)
	res, err := s.sliding.RunInts(context.Background(),
		[]string{"ratelimit:user:1:rpm"},
		int64(60000), int64(10), int64(1000), int64(1), int64(1), "n1",
	)
	if err != nil {
		t.Fatal(err)
	}
	if res[0] != 1 || res[1] != 1 || res[4] != 1 {
		t.Errorf("want ok=1 count=1 wrote=1; got %v", res)
	}
	if !mr.Exists("ratelimit:user:1:rpm") {
		t.Error("key should exist")
	}
}

func TestSlidingWindow_Gating_OverLimit_DoesNotWrite(t *testing.T) {
	mr, rdb := newTestRedis(t)
	s, _ := loadScripts(context.Background(), rdb)
	// 先写满
	for i := 0; i < 5; i++ {
		_, err := s.sliding.RunInts(context.Background(),
			[]string{"ratelimit:k"},
			int64(60000), int64(5), int64(1000+int64(i)), int64(1), int64(1),
			"n"+strconv.Itoa(i),
		)
		if err != nil {
			t.Fatal(err)
		}
	}
	// 第 6 次 gating=1 应被拒
	res, err := s.sliding.RunInts(context.Background(),
		[]string{"ratelimit:k"},
		int64(60000), int64(5), int64(2000), int64(1), int64(1), "n6",
	)
	if err != nil {
		t.Fatal(err)
	}
	if res[0] != 0 {
		t.Errorf("want ok=0 on gating over limit; got %v", res)
	}
	if res[4] != 0 {
		t.Errorf("want wrote=0; got %d", res[4])
	}
	// ZCARD 应仍为 5
	card, _ := rdb.ZCard(context.Background(), "ratelimit:k").Result()
	_ = mr
	if card != 5 {
		t.Errorf("want ZCARD=5; got %d", card)
	}
}

func TestSlidingWindow_Counting_OverLimit_StillWrites(t *testing.T) {
	_, rdb := newTestRedis(t)
	s, _ := loadScripts(context.Background(), rdb)
	res, err := s.sliding.RunInts(context.Background(),
		[]string{"ratelimit:k"},
		int64(60000), int64(2), int64(1000), int64(5), int64(0), "n",
	)
	if err != nil {
		t.Fatal(err)
	}
	// 写入 5 个,超 limit=2,ok=0 但 wrote=5
	if res[0] != 0 {
		t.Errorf("want ok=0 (over limit); got %v", res)
	}
	if res[4] != 5 {
		t.Errorf("want wrote=5; got %d", res[4])
	}
	card, _ := rdb.ZCard(context.Background(), "ratelimit:k").Result()
	if card != 5 {
		t.Errorf("want ZCARD=5; got %d", card)
	}
}

func TestSlidingWindow_ExpiredMembersRemoved(t *testing.T) {
	_, rdb := newTestRedis(t)
	s, _ := loadScripts(context.Background(), rdb)
	// t=0 写 1 个
	_, err := s.sliding.RunInts(context.Background(),
		[]string{"ratelimit:k"},
		int64(60000), int64(10), int64(0), int64(1), int64(1), "n1",
	)
	if err != nil {
		t.Fatal(err)
	}
	// t=window+1 再写 → 老成员应被 ZREMRANGEBYSCORE 清掉
	res, err := s.sliding.RunInts(context.Background(),
		[]string{"ratelimit:k"},
		int64(60000), int64(10), int64(61000), int64(1), int64(1), "n2",
	)
	if err != nil {
		t.Fatal(err)
	}
	if res[1] != 1 {
		t.Errorf("want count_after=1 (old expired); got %d", res[1])
	}
}

func TestSlidingWindow_CostZero_NotWritten(t *testing.T) {
	mr, rdb := newTestRedis(t)
	s, _ := loadScripts(context.Background(), rdb)
	res, err := s.sliding.RunInts(context.Background(),
		[]string{"ratelimit:k"},
		int64(60000), int64(10), int64(0), int64(0), int64(1), "n",
	)
	if err != nil {
		t.Fatal(err)
	}
	if res[0] != 1 || res[4] != 0 {
		t.Errorf("want ok=1 wrote=0; got %v", res)
	}
	if mr.Exists("ratelimit:k") {
		t.Error("key should not be created when cost=0")
	}
}
