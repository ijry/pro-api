package authhttp

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ijry/pro-api/internal/server/middleware"
	"github.com/ijry/pro-api/pkg/apierr"
)

// GET /api/user/oauth/bindings
func (h *handlers) userOAuthBindings(c *gin.Context) {
	uid := middleware.UserID(c)
	items, err := h.d.OAuthRepo.ListByUser(c.Request.Context(), uid)
	if err != nil {
		sendErr(c, err)
		return
	}
	out := make([]gin.H, 0, len(items))
	for _, b := range items {
		out = append(out, gin.H{
			"provider": b.Provider,
			"uid":      b.ProviderUID,
			"email":    b.Email,
			"bound_at": b.CreatedAt.Format(time.RFC3339Nano),
		})
	}
	c.JSON(200, gin.H{"bindings": out})
}

type bindGithubReq struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

// POST /api/user/oauth/bindings/github
func (h *handlers) userBindGithub(c *gin.Context) {
	uid := middleware.UserID(c)
	var req bindGithubReq
	if err := c.ShouldBindJSON(&req); err != nil {
		sendErr(c, apierr.New(apierr.CodeInvalidParam, "请求格式错误"))
		return
	}
	if err := h.d.Auth.BindGithub(c.Request.Context(), uid, req.Code, req.State); err != nil {
		sendErr(c, err)
		return
	}
	c.JSON(200, gin.H{"ok": true})
}

// DELETE /api/user/oauth/bindings/github
func (h *handlers) userUnbindGithub(c *gin.Context) {
	uid := middleware.UserID(c)
	if err := h.d.Auth.UnbindGithub(c.Request.Context(), uid); err != nil {
		sendErr(c, err)
		return
	}
	c.JSON(200, gin.H{"ok": true})
}
