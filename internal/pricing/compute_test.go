package pricing

import (
	"context"
	"testing"
)

// mockChannel 是一个测试用的 ChannelInfo,持有一个固定的 model→Ratios 映射。
type mockChannel struct {
	overrides map[string]Ratios
}

func (m *mockChannel) ModelOverrideFor(model string) (Ratios, bool) {
	r, ok := m.overrides[model]
	return r, ok
}

// mockCatalog 是一个测试用的 CatalogInfo。
type mockCatalog struct {
	m map[string]Catalog
}

func (c *mockCatalog) LookupCatalog(model string) Catalog {
	if c.m == nil {
		return Catalog{}
	}
	return c.m[model]
}

func TestRatioFor_CatalogDefault(t *testing.T) {
	cat := &mockCatalog{m: map[string]Catalog{
		"gpt-4o": {DefaultInputRatio: 1.0, DefaultOutputRatio: 3.0},
	}}
	s, _ := newTestService(t, false, func(c *Config) {
		c.Catalog = cat
	})
	r := s.RatioFor(ctx(), "gpt-4o", 0, nil)
	if r.Input != 1.0 || r.Output != 3.0 {
		t.Fatalf("bad ratios: %+v", r)
	}
}

func TestRatioFor_NilCachedRatio_FallbackToInput(t *testing.T) {
	cat := &mockCatalog{m: map[string]Catalog{
		"gpt-4o": {DefaultInputRatio: 1.5, DefaultOutputRatio: 4.0},
	}}
	s, _ := newTestService(t, false, func(c *Config) { c.Catalog = cat })
	r := s.RatioFor(ctx(), "gpt-4o", 0, nil)
	if r.Cached != 1.5 {
		t.Fatalf("cached fallback to input failed: %+v", r)
	}
	if r.Reasoning != 4.0 {
		t.Fatalf("reasoning fallback to output failed: %+v", r)
	}
}

func TestRatioFor_ChannelOverride(t *testing.T) {
	cat := &mockCatalog{m: map[string]Catalog{
		"gpt-4o": {DefaultInputRatio: 1.0, DefaultOutputRatio: 2.0},
	}}
	s, _ := newTestService(t, false, func(c *Config) { c.Catalog = cat })
	ch := &mockChannel{overrides: map[string]Ratios{
		"gpt-4o": {Input: 0.5, Output: 1.0},
	}}
	r := s.RatioFor(ctx(), "gpt-4o", 0, ch)
	if r.Input != 0.5 || r.Output != 1.0 {
		t.Fatalf("expected channel override, got %+v", r)
	}
}

func TestRatioFor_NilChannel_SkipsOverride(t *testing.T) {
	cat := &mockCatalog{m: map[string]Catalog{
		"gpt-4o": {DefaultInputRatio: 1.0, DefaultOutputRatio: 2.0},
	}}
	s, _ := newTestService(t, false, func(c *Config) { c.Catalog = cat })
	r := s.RatioFor(ctx(), "gpt-4o", 0, nil)
	if r.Input != 1.0 {
		t.Fatalf("expected catalog default, got %+v", r)
	}
}

func TestRatioFor_RuleOverridesChannel(t *testing.T) {
	cat := &mockCatalog{m: map[string]Catalog{
		"gpt-4o": {DefaultInputRatio: 1.0, DefaultOutputRatio: 2.0},
	}}
	s, _ := newTestService(t, false, func(c *Config) { c.Catalog = cat })
	_, _ = s.Create(ctx(), CreateInput{
		Scope: ScopeModel, Model: strPtr("gpt-4o"),
		InputRatio: ratioPtr(5.0), Priority: 100,
	}, 0)
	ch := &mockChannel{overrides: map[string]Ratios{
		"gpt-4o": {Input: 0.5, Output: 1.0},
	}}
	r := s.RatioFor(ctx(), "gpt-4o", 0, ch)
	if r.Input != 5.0 {
		t.Fatalf("expected rule override (5.0), got %+v", r)
	}
}

func TestRatioFor_GroupRatioApplied(t *testing.T) {
	cat := &mockCatalog{m: map[string]Catalog{
		"gpt-4o": {DefaultInputRatio: 1.0, DefaultOutputRatio: 2.0},
	}}
	s, _ := newTestService(t, false, func(c *Config) {
		c.Catalog = cat
		c.GroupRatio = func(_ context.Context, gid int64) float64 {
			if gid == 2 {
				return 0.8
			}
			return 1.0
		}
	})
	r := s.RatioFor(ctx(), "gpt-4o", 2, nil)
	if r.Group != 0.8 {
		t.Fatalf("expected group=0.8, got %v", r.Group)
	}
}

func TestCompute_BasicFormula(t *testing.T) {
	cat := &mockCatalog{m: map[string]Catalog{
		"gpt-4o": {DefaultInputRatio: 2.0, DefaultOutputRatio: 5.0},
	}}
	s, _ := newTestService(t, false, func(c *Config) { c.Catalog = cat })
	res := s.Compute(ctx(), ComputeInput{
		Model: "gpt-4o", InputTokens: 1000, OutputTokens: 500,
	})
	// 1000*2 + 500*5 = 2000 + 2500 = 4500 * 1.0 group ratio = 4500
	if res.Quota != 4500 {
		t.Fatalf("quota = %d, want 4500", res.Quota)
	}
}

func TestCompute_WithGroupRatio(t *testing.T) {
	cat := &mockCatalog{m: map[string]Catalog{
		"gpt-4o": {DefaultInputRatio: 2.0, DefaultOutputRatio: 5.0},
	}}
	s, _ := newTestService(t, false, func(c *Config) {
		c.Catalog = cat
		c.GroupRatio = func(_ context.Context, _ int64) float64 { return 0.8 }
	})
	res := s.Compute(ctx(), ComputeInput{
		Model: "gpt-4o", GroupID: 2, InputTokens: 1000, OutputTokens: 500,
	})
	// 4500 * 0.8 = 3600
	if res.Quota != 3600 {
		t.Fatalf("quota = %d, want 3600", res.Quota)
	}
}

func TestCompute_CeilingForFraction(t *testing.T) {
	cat := &mockCatalog{m: map[string]Catalog{
		"x": {DefaultInputRatio: 0.001, DefaultOutputRatio: 0.001},
	}}
	s, _ := newTestService(t, false, func(c *Config) { c.Catalog = cat })
	res := s.Compute(ctx(), ComputeInput{Model: "x", InputTokens: 1, OutputTokens: 1})
	if res.Quota != 1 {
		t.Fatalf("expected quota=1 (ceil of 0.002), got %d", res.Quota)
	}
}

func TestCompute_TokensZero_QuotaZero(t *testing.T) {
	cat := &mockCatalog{m: map[string]Catalog{
		"x": {DefaultInputRatio: 1.0, DefaultOutputRatio: 1.0},
	}}
	s, _ := newTestService(t, false, func(c *Config) { c.Catalog = cat })
	res := s.Compute(ctx(), ComputeInput{Model: "x"})
	if res.Quota != 0 {
		t.Fatalf("expected 0, got %d", res.Quota)
	}
}

func TestEstimateMax_UsesMaxOut(t *testing.T) {
	cat := &mockCatalog{m: map[string]Catalog{
		"x": {DefaultInputRatio: 1.0, DefaultOutputRatio: 2.0},
	}}
	s, _ := newTestService(t, false, func(c *Config) { c.Catalog = cat })
	q := s.EstimateMax(ctx(), "x", EstimateInput{InputTokens: 100, MaxOutTokens: 50})
	// 100*1 + 50*2 = 200
	if q != 200 {
		t.Fatalf("estimate = %d, want 200", q)
	}
}

func TestEstimateMax_MaxOutZero_UsesDefault(t *testing.T) {
	cat := &mockCatalog{m: map[string]Catalog{
		"x": {DefaultInputRatio: 1.0, DefaultOutputRatio: 0},
	}}
	s, _ := newTestService(t, false, func(c *Config) {
		c.Catalog = cat
		c.DefaultMaxOut = 4096
	})
	q := s.EstimateMax(ctx(), "x", EstimateInput{InputTokens: 1, MaxOutTokens: 0})
	// 1*1 + 4096*0(因为 cat 没设 output)= 1
	if q != 1 {
		t.Fatalf("estimate = %d, want 1", q)
	}
}

func TestDefaultMaxOut_FromCatalog(t *testing.T) {
	cat := &mockCatalog{m: map[string]Catalog{
		"x": {MaxInputTokens: 8000},
	}}
	s, _ := newTestService(t, false, func(c *Config) { c.Catalog = cat })
	if v := s.DefaultMaxOut(ctx(), "x"); v != 4000 {
		t.Fatalf("DefaultMaxOut = %d, want 4000", v)
	}
}

func TestDefaultMaxOut_CapAt4096(t *testing.T) {
	cat := &mockCatalog{m: map[string]Catalog{
		"x": {MaxInputTokens: 1000000},
	}}
	s, _ := newTestService(t, false, func(c *Config) { c.Catalog = cat })
	if v := s.DefaultMaxOut(ctx(), "x"); v != 4096 {
		t.Fatalf("DefaultMaxOut = %d, want 4096 (capped)", v)
	}
}

func TestDefaultMaxOut_NoCatalog_FallbackDefault(t *testing.T) {
	s, _ := newTestService(t, false, func(c *Config) {
		c.DefaultMaxOut = 4096
	})
	if v := s.DefaultMaxOut(ctx(), "anything"); v != 4096 {
		t.Fatalf("DefaultMaxOut = %d, want 4096", v)
	}
}
