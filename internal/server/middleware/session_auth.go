package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/ijry/pro-api/internal/auth/session"
	"github.com/ijry/pro-api/internal/util/clock"
	"github.com/ijry/pro-api/pkg/apierr"
)

// SessionAuth 校验 session,注入 ctx,并触发 Sliding Touch(异步)。
func SessionAuth(store session.Store, _ clock.Clock) gin.HandlerFunc {
	return func(c *gin.Context) {
		sid := readSessionID(c)
		if sid == "" {
			SetErr(c, apierr.New(apierr.CodeNotLoggedIn, "请先登录"))
			return
		}
		sess, err := store.Get(c.Request.Context(), sid)
		if err != nil {
			SetErr(c, apierr.Wrap(apierr.CodeInternal, "session get", err))
			return
		}
		if sess == nil {
			SetErr(c, apierr.New(apierr.CodeSessionExpired, "会话已过期"))
			return
		}
		go func(id string) {
			_ = store.Touch(context.Background(), id)
		}(sid)
		c.Set(CtxKeyUserID, sess.UserID)
		c.Set(CtxKeyRole, sess.Role)
		c.Set(CtxKeySessionID, sess.ID)
		c.Next()
	}
}

// readSessionID 从 Cookie 优先,fallback Header 取 session id。
func readSessionID(c *gin.Context) string {
	if v, err := c.Cookie(CookieSession); err == nil && v != "" {
		return v
	}
	if v := c.GetHeader(HeaderSession); v != "" {
		return v
	}
	return ""
}

// UserID 从 ctx 取当前 user id。
func UserID(c *gin.Context) int64 {
	v, ok := c.Get(CtxKeyUserID)
	if !ok {
		return 0
	}
	id, _ := v.(int64)
	return id
}

// Role 从 ctx 取当前 role。
func Role(c *gin.Context) int8 {
	v, ok := c.Get(CtxKeyRole)
	if !ok {
		return 0
	}
	r, _ := v.(int8)
	return r
}

// SessionID 从 ctx 取当前 session id。
func SessionID(c *gin.Context) string {
	v, ok := c.Get(CtxKeySessionID)
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}
