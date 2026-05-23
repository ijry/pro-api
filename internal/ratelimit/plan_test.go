package ratelimit

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/ijry/pro-api/internal/setting"
)

// fakeSettingStore 是测试用的 setting.Store mock。
type fakeSettingStore struct {
	values map[string]any
}

func (f *fakeSettingStore) Get(_ context.Context, key string) (json.RawMessage, bool) {
	v, ok := f.values[key]
	if !ok {
		return nil, false
	}
	b, _ := json.Marshal(v)
	return b, true
}
func (f *fakeSettingStore) GetString(_ context.Context, k string, def string) string {
	if v, ok := f.values[k]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return def
}
func (f *fakeSettingStore) GetBool(_ context.Context, k string, def bool) bool {
	if v, ok := f.values[k]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return def
}
func (f *fakeSettingStore) GetInt(_ context.Context, k string, def int) int {
	if v, ok := f.values[k]; ok {
		switch x := v.(type) {
		case int:
			return x
		case int64:
			return int(x)
		case float64:
			return int(x)
		}
	}
	return def
}
func (f *fakeSettingStore) GetFloat(_ context.Context, k string, def float64) float64 {
	if v, ok := f.values[k]; ok {
		if x, ok := v.(float64); ok {
			return x
		}
	}
	return def
}
func (f *fakeSettingStore) GetJSON(_ context.Context, _ string, _ any) error { return nil }
func (f *fakeSettingStore) Put(_ context.Context, k string, v any, _ int64) error {
	f.values[k] = v
	return nil
}
func (f *fakeSettingStore) GetSecret(_ context.Context, _ string, _ setting.Decryptor) (string, error) {
	return "", nil
}
func (f *fakeSettingStore) ListAll(_ context.Context) ([]setting.Setting, error) { return nil, nil }
func (f *fakeSettingStore) Close() error                                          { return nil }

func newFakeSetting() *fakeSettingStore {
	return &fakeSettingStore{values: map[string]any{
		"ratelimit.user_default_rpm":  60,
		"ratelimit.user_default_tpm":  100000,
		"ratelimit.ip_rpm":            600,
		"ratelimit.model_default_rpm": 0,
		"ratelimit.model_default_tpm": 0,
		"ratelimit.window_seconds":    60,
	}}
}

func TestPlanRPM_AllDimensions_Generated(t *testing.T) {
	st := newFakeSetting()
	// 启用 model 维度
	st.values["ratelimit.model_default_rpm"] = 200
	p := NewPlanner(PlannerConfig{Setting: st})

	in := PlanInput{
		UserID:  42,
		TokenID: 7,
		GroupID: 1,
		IP:      "1.2.3.4",
		Model:   "gpt-4o",
	}
	checks := p.PlanRPM(context.Background(), in)
	if len(checks) != 4 {
		t.Fatalf("want 4 checks (user/token/ip/model), got %d: %+v", len(checks), checks)
	}
	wantOrder := []Dimension{DimUserRPM, DimTokenRPM, DimIPRPM, DimModelRPM}
	for i, c := range checks {
		if c.Dimension != wantOrder[i] {
			t.Errorf("checks[%d] dim=%s want %s", i, c.Dimension, wantOrder[i])
		}
		if c.Cost != 1 {
			t.Errorf("checks[%d] cost=%d want 1", i, c.Cost)
		}
	}
}

func TestPlanRPM_ZeroThresholds_Omitted(t *testing.T) {
	st := newFakeSetting()
	// model_default_rpm 已经 = 0
	p := NewPlanner(PlannerConfig{Setting: st})
	checks := p.PlanRPM(context.Background(), PlanInput{
		UserID: 1, TokenID: 1, IP: "1.2.3.4", Model: "gpt-4o",
	})
	for _, c := range checks {
		if c.Dimension == DimModelRPM {
			t.Fatal("model_rpm should be omitted when limit=0")
		}
	}
}

func TestPlanRPM_TokenOverride_TakesPriority(t *testing.T) {
	st := newFakeSetting()
	p := NewPlanner(PlannerConfig{Setting: st})
	checks := p.PlanRPM(context.Background(), PlanInput{
		UserID: 1, TokenID: 1, IP: "1.2.3.4",
		TokenRPMOverride: 120,
		GroupRatio:       0.8,
	})
	var userLim, tokenLim int
	for _, c := range checks {
		switch c.Dimension {
		case DimUserRPM:
			userLim = c.Limit
		case DimTokenRPM:
			tokenLim = c.Limit
		}
	}
	if tokenLim != 120 {
		t.Errorf("token_rpm want 120 got %d", tokenLim)
	}
	// user_rpm = 60 / 0.8 = 75
	if userLim != 75 {
		t.Errorf("user_rpm want 75 got %d", userLim)
	}
}

func TestPlanRPM_GroupRatio_AppliedOnlyToUserDim(t *testing.T) {
	st := newFakeSetting()
	st.values["ratelimit.user_default_rpm"] = 100
	p := NewPlanner(PlannerConfig{Setting: st})
	// token 无 override → 不应折算
	checks := p.PlanRPM(context.Background(), PlanInput{
		UserID: 1, TokenID: 1, IP: "1.2.3.4",
		GroupRatio: 0.5,
	})
	for _, c := range checks {
		if c.Dimension == DimUserRPM && c.Limit != 200 {
			t.Errorf("user_rpm want 200 got %d", c.Limit)
		}
	}
}

func TestPlanRPM_IPCanonicalized(t *testing.T) {
	st := newFakeSetting()
	p := NewPlanner(PlannerConfig{Setting: st})
	checks := p.PlanRPM(context.Background(), PlanInput{
		UserID: 1, TokenID: 1, IP: "1.2.3.4:5000",
	})
	found := false
	for _, c := range checks {
		if c.Dimension == DimIPRPM {
			found = true
			// 期望 key 含 "1.2.3.0/24"
			if !contains(c.Key, "1.2.3.0/24") {
				t.Errorf("ip key should contain 1.2.3.0/24; got %s", c.Key)
			}
		}
	}
	if !found {
		t.Fatal("ip_rpm not generated")
	}
}

func TestPlanRPM_EmptyModel_OmitsModelDim(t *testing.T) {
	st := newFakeSetting()
	st.values["ratelimit.model_default_rpm"] = 100
	p := NewPlanner(PlannerConfig{Setting: st})
	checks := p.PlanRPM(context.Background(), PlanInput{
		UserID: 1, TokenID: 1, IP: "1.2.3.4", Model: "",
	})
	for _, c := range checks {
		if c.Dimension == DimModelRPM {
			t.Fatal("model_rpm should be omitted when model is empty")
		}
	}
}

func TestPlanRPM_EmptyIP_OmitsIPDim(t *testing.T) {
	st := newFakeSetting()
	p := NewPlanner(PlannerConfig{Setting: st})
	checks := p.PlanRPM(context.Background(), PlanInput{
		UserID: 1, TokenID: 1, IP: "",
	})
	for _, c := range checks {
		if c.Dimension == DimIPRPM {
			t.Fatal("ip_rpm should be omitted when IP is empty")
		}
	}
}

func TestPlanRPM_ZeroUserID_OmitsUserDim(t *testing.T) {
	st := newFakeSetting()
	p := NewPlanner(PlannerConfig{Setting: st})
	checks := p.PlanRPM(context.Background(), PlanInput{
		UserID: 0, TokenID: 1, IP: "1.2.3.4",
	})
	for _, c := range checks {
		if c.Dimension == DimUserRPM {
			t.Fatal("user_rpm should be omitted when UserID=0")
		}
	}
}

func TestPlanRPM_ZeroTokenID_OmitsTokenDim(t *testing.T) {
	st := newFakeSetting()
	p := NewPlanner(PlannerConfig{Setting: st})
	checks := p.PlanRPM(context.Background(), PlanInput{
		UserID: 1, TokenID: 0, IP: "1.2.3.4",
	})
	for _, c := range checks {
		if c.Dimension == DimTokenRPM {
			t.Fatal("token_rpm should be omitted when TokenID=0")
		}
	}
}

func TestPlanTPM_CostZeroByDefault(t *testing.T) {
	st := newFakeSetting()
	p := NewPlanner(PlannerConfig{Setting: st})
	checks := p.PlanTPM(context.Background(), PlanInput{
		UserID: 1, TokenID: 1, IP: "1.2.3.4",
	})
	for _, c := range checks {
		if c.Cost != 0 {
			t.Errorf("PlanTPM check[%s].Cost=%d want 0", c.Dimension, c.Cost)
		}
		if !c.Dimension.IsTPM() {
			t.Errorf("PlanTPM produced non-TPM dim %s", c.Dimension)
		}
	}
}

func TestFillTPMCost_AllCostUpdated(t *testing.T) {
	in := []Check{
		{Dimension: DimUserTPM, Cost: 0},
		{Dimension: DimTokenTPM, Cost: 0},
	}
	out := FillTPMCost(in, 500)
	for _, c := range out {
		if c.Cost != 500 {
			t.Errorf("Cost=%d want 500", c.Cost)
		}
	}
	// 不可变性:原切片不变
	if in[0].Cost != 0 {
		t.Errorf("original slice mutated: in[0].Cost=%d", in[0].Cost)
	}
}

func TestPlanner_WindowSeconds_FromSetting(t *testing.T) {
	st := newFakeSetting()
	st.values["ratelimit.window_seconds"] = 30
	p := NewPlanner(PlannerConfig{Setting: st})
	checks := p.PlanRPM(context.Background(), PlanInput{
		UserID: 1, TokenID: 1, IP: "1.2.3.4",
	})
	for _, c := range checks {
		if c.Window != 30*time.Second {
			t.Errorf("Window=%v want 30s", c.Window)
		}
	}
}

func TestPlanner_PubSub_InvalidatesCache(t *testing.T) {
	st := newFakeSetting()
	p := NewPlanner(PlannerConfig{Setting: st, LocalTTL: time.Minute})
	// 先调一次填充 cache
	_ = p.PlanRPM(context.Background(), PlanInput{UserID: 1, TokenID: 1, IP: "1.2.3.4"})
	// 改 setting
	st.values["ratelimit.user_default_rpm"] = 200
	// purge 本地缓存
	p.cache.purge()
	checks := p.PlanRPM(context.Background(), PlanInput{UserID: 1, TokenID: 1, IP: "1.2.3.4"})
	var userLim int
	for _, c := range checks {
		if c.Dimension == DimUserRPM {
			userLim = c.Limit
		}
	}
	if userLim != 200 {
		t.Errorf("user_rpm after purge want 200 got %d", userLim)
	}
}

// helpers
func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
