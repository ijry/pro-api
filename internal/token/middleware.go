package token

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ijry/pro-api/internal/server/middleware"
	"github.com/ijry/pro-api/pkg/apierr"
)

// Gin context key 命名。这些 key 用字符串而非自定义类型,便于跨包查询
// (Gin 自身只支持 string key)。
const (
	// CtxKeyToken 存放 *View。
	CtxKeyToken = "proapi:token"
	// CtxKeyUserID 存放当前 token 关联的 user_id(int64)。
	CtxKeyUserID = "proapi:user_id"
	// CtxKeyGroupID 存放当前 user 所属 group_id(int64),若 UserLookup 未注入则不会有该键。
	CtxKeyGroupID = "proapi:group_id"
)

// UserLookup 是 M1-02 注入的可选 helper,用于补全 ctx 中的 group_id。
// 注入 nil 时 TokenAuth 跳过 group 查询。
type UserLookup func(ctx context.Context, userID int64) (groupID int64, ok bool)

// TokenAuth 接受 "Authorization: Bearer pa-xxx",解析并注入 ctx。
//
//   - 失败 → SetErr(CodeInvalidToken / CodeTokenExpired / CodeIPNotAllowed) + Abort
//   - 成功 → c.Set(CtxKeyToken, view); c.Set(CtxKeyUserID, view.UserID)
//
// IP 白名单校验在此立即进行;模型白名单留给 handler 选 model 后调用。
func TokenAuth(s Store, userLookup UserLookup) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := c.GetHeader("Authorization")
		key, ok := extractBearer(raw)
		if !ok {
			middleware.SetErr(c, apierr.New(apierr.CodeInvalidToken, "missing or malformed bearer token"))
			return
		}
		view, err := s.Authenticate(c.Request.Context(), key)
		if err != nil {
			middleware.SetErr(c, toApiErr(err, apierr.CodeInvalidToken))
			return
		}
		ip := realIP(c.Request)
		if err := AssertIPAllowed(view, ip); err != nil {
			middleware.SetErr(c, toApiErr(err, apierr.CodeIPNotAllowed))
			return
		}
		c.Set(CtxKeyToken, view)
		c.Set(CtxKeyUserID, view.UserID)
		if userLookup != nil {
			if gid, ok := userLookup(c.Request.Context(), view.UserID); ok {
				c.Set(CtxKeyGroupID, gid)
			}
		}
		c.Next()
	}
}

// toApiErr 把已有 error 包装到 *apierr.Error;若已经是,直接返回。
func toApiErr(err error, fallback apierr.Code) *apierr.Error {
	if ae, ok := err.(*apierr.Error); ok {
		return ae
	}
	return apierr.New(fallback, err.Error())
}

// extractBearer 解析 "Bearer pa-xxx";Bearer 关键字大小写不敏感,pa- 前缀大小写敏感。
func extractBearer(h string) (string, bool) {
	const scheme = "Bearer "
	if len(h) < len(scheme) {
		return "", false
	}
	if !strings.EqualFold(h[:len(scheme)], scheme) {
		return "", false
	}
	s := strings.TrimSpace(h[len(scheme):])
	if !strings.HasPrefix(s, keyPrefixLit) {
		return "", false
	}
	return s, true
}

// realIP 取真实客户端 IP。M1 信任反向代理头(spec §11.13)。
//
//   - 优先 X-Forwarded-For 第一段
//   - 否则 X-Real-IP
//   - 否则 RemoteAddr(原始,可能含端口)
func realIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// 取第一段
		if idx := strings.Index(xff, ","); idx >= 0 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}
	if xr := r.Header.Get("X-Real-IP"); xr != "" {
		return strings.TrimSpace(xr)
	}
	return r.RemoteAddr
}

// FromContext 在下游 handler 取当前请求的令牌 View。
//
// 同时兼容 *gin.Context 与 context.Context(handler 把 gin.Context.Request.Context() 传下来时也能命中)。
func FromContext(c context.Context) (*View, bool) {
	if gc, ok := c.(*gin.Context); ok {
		v, ok := gc.Get(CtxKeyToken)
		if !ok {
			return nil, false
		}
		t, ok := v.(*View)
		return t, ok
	}
	v := c.Value(CtxKeyToken)
	t, ok := v.(*View)
	return t, ok
}

// UserIDFromContext 取当前 user id;0 表示未鉴权或未注入。
func UserIDFromContext(c context.Context) int64 {
	if gc, ok := c.(*gin.Context); ok {
		if v, ok := gc.Get(CtxKeyUserID); ok {
			if id, ok := v.(int64); ok {
				return id
			}
		}
		return 0
	}
	if v := c.Value(CtxKeyUserID); v != nil {
		if id, ok := v.(int64); ok {
			return id
		}
	}
	return 0
}

// GroupIDFromContext 取当前 group_id;0 表示无 UserLookup 注入或未鉴权。
func GroupIDFromContext(c context.Context) int64 {
	if gc, ok := c.(*gin.Context); ok {
		if v, ok := gc.Get(CtxKeyGroupID); ok {
			if id, ok := v.(int64); ok {
				return id
			}
		}
		return 0
	}
	if v := c.Value(CtxKeyGroupID); v != nil {
		if id, ok := v.(int64); ok {
			return id
		}
	}
	return 0
}
