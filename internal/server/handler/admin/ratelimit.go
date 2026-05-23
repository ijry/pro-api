package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ijry/pro-api/internal/audit"
	"github.com/ijry/pro-api/internal/ratelimit"
	"github.com/ijry/pro-api/internal/server/middleware"
	"github.com/ijry/pro-api/pkg/apierr"
	"go.uber.org/zap"
)

// RatelimitHandler 提供限流键的查询与重置 endpoint。
type RatelimitHandler struct {
	Limiter ratelimit.Limiter
	Audit   audit.Logger
	Log     *zap.Logger
	ActorOf func(c *gin.Context) int64
}

// NewRatelimitHandler 构造 RatelimitHandler。actorOf 为空时默认返回 0。
func NewRatelimitHandler(l ratelimit.Limiter, audi audit.Logger, log *zap.Logger, actorOf func(c *gin.Context) int64) *RatelimitHandler {
	if log == nil {
		log = zap.NewNop()
	}
	if actorOf == nil {
		actorOf = func(*gin.Context) int64 { return 0 }
	}
	return &RatelimitHandler{
		Limiter: l,
		Audit:   audi,
		Log:     log,
		ActorOf: actorOf,
	}
}

// GetStats godoc
//
// GET /api/admin/ratelimit/keys/:key/stats
// path 参数 :key 是完整 redis key(URL-encoded);Gin 自动 decode。
//
// 返回 {"key": string, "count": int, "oldest_at": rfc3339 | null}
func (h *RatelimitHandler) GetStats(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		middleware.SetErr(c, apierr.New(apierr.CodeMissingParam, "key is required"))
		return
	}
	count, oldest, err := h.Limiter.Stats(c.Request.Context(), key)
	if err != nil {
		middleware.SetErr(c, apierr.New(apierr.CodeCache, err.Error()))
		return
	}
	resp := gin.H{
		"key":       key,
		"count":     count,
		"oldest_at": nil,
	}
	if !oldest.IsZero() {
		resp["oldest_at"] = oldest.UTC()
	}
	c.JSON(http.StatusOK, resp)
}

// Reset godoc
//
// POST /api/admin/ratelimit/keys/:key/reset
// 删除限流 key;写 audit。
func (h *RatelimitHandler) Reset(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		middleware.SetErr(c, apierr.New(apierr.CodeMissingParam, "key is required"))
		return
	}
	if err := h.Limiter.Reset(c.Request.Context(), key); err != nil {
		middleware.SetErr(c, apierr.New(apierr.CodeCache, err.Error()))
		return
	}
	// audit
	if h.Audit != nil {
		actor := h.ActorOf(c)
		entry := audit.Entry{
			Action:     "ratelimit.reset",
			TargetType: "ratelimit_key",
			After:      []byte(`{"key":"` + key + `"}`),
			IP:         c.ClientIP(),
		}
		if actor != 0 {
			entry.ActorID = &actor
		}
		_ = h.Audit.Log(c.Request.Context(), entry)
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "key": key})
}
