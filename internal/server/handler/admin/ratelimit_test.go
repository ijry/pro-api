package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/ijry/pro-api/internal/audit"
	"github.com/ijry/pro-api/internal/ratelimit"
	"github.com/ijry/pro-api/internal/server/middleware"
	"github.com/ijry/pro-api/internal/util/clock"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type capturedAudit struct {
	entries []audit.Entry
}

func (c *capturedAudit) Log(_ context.Context, e audit.Entry) error {
	c.entries = append(c.entries, e)
	return nil
}

func setupRLHandler(t *testing.T) (*gin.Engine, ratelimit.Limiter, *capturedAudit) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	l, err := ratelimit.NewRedisLimiter(context.Background(), ratelimit.Config{
		Cache: rdb, Log: zap.NewNop(), Clock: clock.Real,
	})
	if err != nil {
		t.Fatal(err)
	}
	au := &capturedAudit{}
	h := NewRatelimitHandler(l, au, zap.NewNop(), nil)

	r := gin.New()
	r.Use(middleware.ErrorResponse("json"))
	r.GET("/api/admin/ratelimit/keys/:key/stats", h.GetStats)
	r.POST("/api/admin/ratelimit/keys/:key/reset", h.Reset)
	return r, l, au
}

func TestRLHandler_GetStats_ValidKey_ReturnsCount(t *testing.T) {
	r, l, _ := setupRLHandler(t)
	// 先写入 3 个
	for i := 0; i < 3; i++ {
		_ = l.AllowMulti(context.Background(), []ratelimit.Check{
			{Dimension: ratelimit.DimUserRPM, Key: "user:42:rpm", Limit: 100, Window: time.Minute, Cost: 1},
		})
	}
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/admin/ratelimit/keys/ratelimit:user:42:rpm/stats", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200; got %d body=%s", w.Code, w.Body.String())
	}
	var body map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	if body["count"] != float64(3) {
		t.Errorf("want count=3; got %v", body["count"])
	}
}

func TestRLHandler_GetStats_MissingKey_400(t *testing.T) {
	r, _, _ := setupRLHandler(t)
	w := httptest.NewRecorder()
	// gin path param missing → 路由不会匹配。改用空字符串走 stats handler:
	// 直接调 handler 内部走 :key 为空。模拟方式:注册一个 catchall。
	r.GET("/empty/stats", func(c *gin.Context) {
		h := NewRatelimitHandler(nil, nil, nil, nil)
		h.GetStats(c)
	})
	req, _ := http.NewRequest(http.MethodGet, "/empty/stats", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400; got %d body=%s", w.Code, w.Body.String())
	}
}

func TestRLHandler_Reset_DeletesKeyAndAudits(t *testing.T) {
	r, l, au := setupRLHandler(t)
	// 先写入
	_ = l.AllowMulti(context.Background(), []ratelimit.Check{
		{Dimension: ratelimit.DimUserRPM, Key: "user:99:rpm", Limit: 100, Window: time.Minute, Cost: 1},
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/admin/ratelimit/keys/ratelimit:user:99:rpm/reset", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200; got %d body=%s", w.Code, w.Body.String())
	}
	count, _, _ := l.Stats(context.Background(), "ratelimit:user:99:rpm")
	if count != 0 {
		t.Errorf("want count=0 after reset; got %d", count)
	}
	// audit 写过一条
	if len(au.entries) != 1 {
		t.Fatalf("want 1 audit entry; got %d", len(au.entries))
	}
	if au.entries[0].Action != "ratelimit.reset" {
		t.Errorf("audit action=%s", au.entries[0].Action)
	}
}
