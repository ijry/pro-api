package ratelimit

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/ijry/pro-api/internal/server/middleware"
	"github.com/ijry/pro-api/internal/util/clock"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func newTestServer(t *testing.T, enabled bool, in PlanInput, handler gin.HandlerFunc) (*gin.Engine, *miniredis.Miniredis, Limiter) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	l, err := NewRedisLimiter(context.Background(), Config{
		Cache: rdb, Log: zap.NewNop(), Clock: clock.Real,
	})
	if err != nil {
		t.Fatal(err)
	}
	st := &fakeSettingStore{values: map[string]any{
		"ratelimit.enabled":           enabled,
		"ratelimit.user_default_rpm":  60,
		"ratelimit.user_default_tpm":  100000,
		"ratelimit.ip_rpm":            600,
		"ratelimit.model_default_rpm": 0,
		"ratelimit.model_default_tpm": 0,
		"ratelimit.window_seconds":    60,
	}}
	p := NewPlanner(PlannerConfig{Setting: st})

	r := gin.New()
	r.Use(middleware.ErrorResponse("openai"))
	r.Use(func(c *gin.Context) {
		// 把 PlanInput 注入到 ctx
		setPlanInput(c, in)
		c.Next()
	})
	r.Use(Middleware(l, p, st, zap.NewNop()))
	r.Any("/v1/test", handler)
	return r, mr, l
}

const ctxPlanInputKey = "test:plan_input"

func setPlanInput(c *gin.Context, in PlanInput) { c.Set(ctxPlanInputKey, in) }

// 给 Middleware 提供 PlanInput 来源的测试桩。
func init() {
	contextPlanResolver = func(c *gin.Context) PlanInput {
		v, _ := c.Get(ctxPlanInputKey)
		in, _ := v.(PlanInput)
		return in
	}
}

func TestMiddleware_Disabled_DirectlyPasses(t *testing.T) {
	r, mr, _ := newTestServer(t, false, PlanInput{UserID: 1, TokenID: 1, IP: "1.2.3.4"}, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/test", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200 got %d", w.Code)
	}
	// 不应有 X-RateLimit-* 头
	for k := range w.Header() {
		if len(k) >= 13 && k[:13] == "X-Ratelimit-L" {
			t.Errorf("unexpected header %s when disabled", k)
		}
	}
	// redis 应无 key
	keys := mr.Keys()
	if len(keys) > 0 {
		t.Errorf("redis should be empty; got keys %v", keys)
	}
}

func TestMiddleware_RPMLimit_Returns429(t *testing.T) {
	r, _, _ := newTestServer(t, true, PlanInput{
		UserID: 100, TokenID: 1, IP: "10.0.0.1",
		TokenRPMOverride: 1, // 命中第 2 次
	}, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	// 第 1 次:通过
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodPost, "/v1/test", nil)
	r.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("first req want 200 got %d body=%s", w1.Code, w1.Body.String())
	}
	// 第 2 次:被拒
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodPost, "/v1/test", nil)
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusTooManyRequests {
		t.Fatalf("second req want 429 got %d body=%s", w2.Code, w2.Body.String())
	}
	if w2.Header().Get("X-RateLimit-Limit-Token-RPM") != "1" {
		t.Errorf("missing X-RateLimit-Limit-Token-RPM; got %v", w2.Header())
	}
	if w2.Header().Get("Retry-After") == "" {
		t.Error("missing Retry-After")
	}
	// OpenAI 错误格式
	var body map[string]any
	_ = json.Unmarshal(w2.Body.Bytes(), &body)
	if body["error"] == nil {
		t.Errorf("want openai error envelope; body=%s", w2.Body.String())
	}
}

func TestMiddleware_TPM_ConsumedAfterHandler(t *testing.T) {
	r, _, l := newTestServer(t, true, PlanInput{
		UserID: 200, TokenID: 2, IP: "10.0.0.2",
		TokenTPMOverride: 1000,
	}, func(c *gin.Context) {
		// handler 写入 total_tokens
		c.Set(CtxKeyTotalTokens, 500)
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/v1/test", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200 got %d", w.Code)
	}
	// user_tpm 应增加 500
	count, _, _ := l.Stats(context.Background(), "ratelimit:user:200:tpm")
	if count != 500 {
		t.Errorf("user_tpm count want 500 got %d", count)
	}
	// token_tpm 同样 500
	count2, _, _ := l.Stats(context.Background(), "ratelimit:token:2:tpm")
	if count2 != 500 {
		t.Errorf("token_tpm count want 500 got %d", count2)
	}
}

func TestMiddleware_NoTokens_NoTPMConsume(t *testing.T) {
	r, _, l := newTestServer(t, true, PlanInput{
		UserID: 300, TokenID: 3, IP: "10.0.0.3",
		TokenTPMOverride: 1000,
	}, func(c *gin.Context) {
		// 不设置 total_tokens
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/v1/test", nil)
	r.ServeHTTP(w, req)
	count, _, _ := l.Stats(context.Background(), "ratelimit:user:300:tpm")
	if count != 0 {
		t.Errorf("want count=0 (no tokens consumed); got %d", count)
	}
}

func TestMiddleware_AllowedRequest_WritesObservabilityHeaders(t *testing.T) {
	r, _, _ := newTestServer(t, true, PlanInput{
		UserID: 400, TokenID: 4, IP: "10.0.0.4",
		TokenRPMOverride: 10,
	}, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/v1/test", nil)
	r.ServeHTTP(w, req)
	if w.Header().Get("X-RateLimit-Reset") == "" {
		t.Error("missing X-RateLimit-Reset on allowed")
	}
}

// 限流 dimension code 映射
func TestCodeForDimension(t *testing.T) {
	cases := []struct{ d Dimension }{
		{DimUserRPM}, {DimTokenRPM}, {DimIPRPM}, {DimModelRPM},
	}
	for _, c := range cases {
		if int(codeForDimension(c.d)) == 0 {
			t.Errorf("dim %s mapped to 0 code", c.d)
		}
	}
}

// helper: middleware 应处理 context.Background 输入(避免 ctx 被 cancel 后 ConsumeTPM 失败)
func TestMiddleware_TPMUsesDetachedCtx(t *testing.T) {
	// 调用方 ctx cancel 不应导致 ConsumeTPM 失败 → ConsumeTPM 走独立 context
	r, _, l := newTestServer(t, true, PlanInput{
		UserID: 500, TokenID: 5, IP: "10.0.0.5",
		TokenTPMOverride: 100,
	}, func(c *gin.Context) {
		c.Set(CtxKeyTotalTokens, 10)
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/v1/test", nil)
	// 显式 cancel request ctx(模拟客户端断连)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_ = time.Millisecond
	req = req.WithContext(ctx)
	r.ServeHTTP(w, req)
	// 即使 cancel,ConsumeTPM 仍应写入(因为用 detached ctx)
	count, _, _ := l.Stats(context.Background(), "ratelimit:user:500:tpm")
	if count != 10 {
		t.Errorf("want detached ctx to still consume tpm; count=%d", count)
	}
}
