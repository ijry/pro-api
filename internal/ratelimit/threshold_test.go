package ratelimit

import (
	"testing"
	"time"
)

func TestScaleByGroup_Identity(t *testing.T) {
	if got := scaleByGroup(100, 1.0); got != 100 {
		t.Fatalf("ratio=1.0 want 100 got %d", got)
	}
}

func TestScaleByGroup_VIPRatio_Increases(t *testing.T) {
	// ratio=0.8 → 1/0.8 = 1.25x → 125
	if got := scaleByGroup(100, 0.8); got != 125 {
		t.Fatalf("ratio=0.8 want 125 got %d", got)
	}
}

func TestScaleByGroup_ZeroRatio_TreatedAsIdentity(t *testing.T) {
	if got := scaleByGroup(100, 0); got != 100 {
		t.Fatalf("ratio=0 want 100 got %d", got)
	}
}

func TestScaleByGroup_NegativeRatio_TreatedAsIdentity(t *testing.T) {
	if got := scaleByGroup(100, -1); got != 100 {
		t.Fatalf("ratio=-1 want 100 got %d", got)
	}
}

func TestScaleByGroup_RatioAboveOne_NotAmplified(t *testing.T) {
	// ratio>=1.0 表示无折扣 / 加价场景。M1 限流上限不应变小,直接 identity。
	if got := scaleByGroup(100, 1.5); got != 100 {
		t.Fatalf("ratio=1.5 want 100 got %d", got)
	}
}

func TestScaleByGroup_ZeroLimit_StaysZero(t *testing.T) {
	if got := scaleByGroup(0, 0.5); got != 0 {
		t.Fatalf("limit=0 want 0 got %d", got)
	}
}

func TestThresholdCache_HitWithinTTL(t *testing.T) {
	c := newThresholdCache(100 * time.Millisecond)
	c.set("k", 60, 0)
	v, ok := c.get("k")
	if !ok || v != 60 {
		t.Fatalf("want hit 60 got %d ok=%v", v, ok)
	}
}

func TestThresholdCache_MissAfterTTL(t *testing.T) {
	c := newThresholdCache(10 * time.Millisecond)
	c.set("k", 60, 0)
	time.Sleep(20 * time.Millisecond)
	if _, ok := c.get("k"); ok {
		t.Fatalf("want miss after TTL")
	}
}

func TestThresholdCache_Purge_ClearsAll(t *testing.T) {
	c := newThresholdCache(time.Minute)
	c.set("a", 1, 0)
	c.set("b", 2, 0)
	c.purge()
	if _, ok := c.get("a"); ok {
		t.Fatalf("a should be gone after purge")
	}
	if _, ok := c.get("b"); ok {
		t.Fatalf("b should be gone after purge")
	}
}

func TestThresholdCache_PerKeyTTLOverride(t *testing.T) {
	c := newThresholdCache(time.Minute)
	c.set("k", 60, 10*time.Millisecond) // per-key TTL 覆盖默认值
	time.Sleep(20 * time.Millisecond)
	if _, ok := c.get("k"); ok {
		t.Fatalf("want miss after per-key TTL")
	}
}
