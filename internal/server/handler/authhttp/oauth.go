package authhttp

import (
	"net/http"
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

// GET /api/auth/oauth/:provider/start
func (h *handlers) oauthStart(c *gin.Context) {
	provider := c.Param("provider")
	redirect := c.Query("redirect_url")
	authURL, err := h.d.Auth.OAuthStart(c.Request.Context(), provider, clientIP(c), userAgent(c), redirect, 0)
	if err != nil {
		sendErr(c, err)
		return
	}
	c.Redirect(http.StatusFound, authURL)
}

// GET /api/auth/oauth/:provider/callback
func (h *handlers) oauthCallback(c *gin.Context) {
	provider := c.Param("provider")
	code := c.Query("code")
	state := c.Query("state")
	sess, u, _, err := h.d.Auth.OAuthCallback(c.Request.Context(), provider, code, state, clientIP(c), userAgent(c))
	if err != nil {
		if c.Query("json") == "1" {
			sendErr(c, err)
			return
		}
		c.Redirect(http.StatusFound, "/login?oauth_error="+provider+"&oauth_message="+urlEncode(err.Error()))
		return
	}
	h.setSessionCookies(c, sess)
	if c.Query("json") == "1" {
		c.JSON(200, gin.H{
			"user":    userView(u, resolveGroupName(c.Request.Context(), h.d.Group, u.GroupID)),
			"session": sessionView(sess),
		})
		return
	}
	c.Redirect(http.StatusFound, "/")
}

type bindOAuthReq struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

// POST /api/auth/oauth/:provider/bind
func (h *handlers) bindOAuth(c *gin.Context) {
	provider := c.Param("provider")
	uid := middleware.UserID(c)
	var req bindOAuthReq
	if err := c.ShouldBindJSON(&req); err != nil {
		sendErr(c, apierr.New(apierr.CodeInvalidParam, "请求格式错误"))
		return
	}
	if err := h.d.Auth.BindOAuth(c.Request.Context(), uid, provider, req.Code, req.State); err != nil {
		sendErr(c, err)
		return
	}
	c.JSON(200, gin.H{"ok": true})
}

// DELETE /api/auth/oauth/:provider/bind
func (h *handlers) unbindOAuth(c *gin.Context) {
	provider := c.Param("provider")
	uid := middleware.UserID(c)
	if err := h.d.Auth.UnbindOAuth(c.Request.Context(), uid, provider); err != nil {
		sendErr(c, err)
		return
	}
	c.JSON(200, gin.H{"ok": true})
}
