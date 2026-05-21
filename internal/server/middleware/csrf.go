package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ijry/pro-api/pkg/apierr"
)

// CSRFMasterKey 用 HMAC 派生 csrf cookie 值(M1 不强求与运行时 master_key 一致,
// 只要进程内稳定即可;wire 层可注入)。
type CSRFMasterKey []byte

// CSRF 是双 Cookie 法 CSRF 中间件。
//
//   - 仅对 unsafe method(POST/PUT/PATCH/DELETE)校验
//   - 路由 whitelist 内的路径跳过(用于登录/注册等"未持有 csrf"入口)
//   - 校验逻辑:Header(X-CSRF-Token)== Cookie(proapi_csrf)
//   - GET/HEAD/OPTIONS 跳过
//
// 注意:登录成功后由 handler 负责 SetCookie proapi_csrf;此中间件不主动派发。
func CSRF(whitelist []string) gin.HandlerFunc {
	wl := make(map[string]struct{}, len(whitelist))
	for _, p := range whitelist {
		wl[p] = struct{}{}
	}
	return func(c *gin.Context) {
		method := strings.ToUpper(c.Request.Method)
		if method == "GET" || method == "HEAD" || method == "OPTIONS" {
			c.Next()
			return
		}
		path := c.Request.URL.Path
		if _, ok := wl[path]; ok {
			c.Next()
			return
		}
		// 仅前缀匹配:对 /api/auth/oauth/github/callback 等也跳过
		for prefix := range wl {
			if strings.HasSuffix(prefix, "/*") && strings.HasPrefix(path, strings.TrimSuffix(prefix, "/*")) {
				c.Next()
				return
			}
		}
		cookieVal, _ := c.Cookie(CookieCSRF)
		hdr := c.GetHeader(HeaderCSRF)
		if cookieVal == "" || hdr == "" {
			SetErr(c, apierr.New(apierr.CodeForbidden, "缺少 CSRF token"))
			return
		}
		if !hmac.Equal([]byte(cookieVal), []byte(hdr)) {
			SetErr(c, apierr.New(apierr.CodeForbidden, "CSRF token 不匹配"))
			return
		}
		c.Next()
	}
}

// DeriveCSRFToken 用 sessionID + masterKey 派生 csrf cookie 值。
func DeriveCSRFToken(masterKey []byte, sessionID string) string {
	mac := hmac.New(sha256.New, masterKey)
	mac.Write([]byte(sessionID))
	return hex.EncodeToString(mac.Sum(nil))[:32]
}
