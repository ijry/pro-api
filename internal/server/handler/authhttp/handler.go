package authhttp

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ijry/pro-api/internal/audit"
	"github.com/ijry/pro-api/internal/auth"
	"github.com/ijry/pro-api/internal/auth/oauth"
	"github.com/ijry/pro-api/internal/auth/session"
	"github.com/ijry/pro-api/internal/group"
	"github.com/ijry/pro-api/internal/server/middleware"
	"github.com/ijry/pro-api/internal/setting"
	"github.com/ijry/pro-api/internal/user"
	"github.com/ijry/pro-api/internal/util/clock"
	"github.com/ijry/pro-api/pkg/apierr"
	"go.uber.org/zap"
)

// Deps 是 handler 模块的依赖。
type Deps struct {
	Auth         auth.Service
	User         user.Service
	Group        group.Service
	Session      session.Store
	OAuthRepo    oauth.Repository
	Setting      setting.Store
	Audit        audit.Logger
	Clock        clock.Clock
	Log          *zap.Logger
	CSRFKey      []byte
	CookieSecure bool // dev=false, prod=true
}

// CSRFWhitelist 是 spec §4.7 中需跳过 CSRF 校验的入口。
var CSRFWhitelist = []string{
	"/api/auth/register",
	"/api/auth/login",
	"/api/auth/logout",
	"/api/auth/email/send_code",
	"/api/auth/email/login",
	"/api/auth/password/forgot",
	"/api/auth/password/reset",
	"/api/auth/oauth/github/start",
	"/api/auth/oauth/github/callback",
	"/api/admin/auth/login",
	"/api/admin/auth/logout",
}

// RegisterRoutes 把所有 /api/auth/*、/api/user/*、/api/admin/* 挂到 eng 上。
func RegisterRoutes(eng *gin.Engine, d Deps) {
	h := &handlers{d: d}

	api := eng.Group("/api")

	// 公开入口(不走 SessionAuth、不走 CSRF)
	authG := api.Group("/auth")
	authG.POST("/register", h.register)
	authG.POST("/login", h.login)
	authG.POST("/logout", h.logout)
	authG.POST("/email/send_code", h.emailSendCode)
	authG.POST("/email/login", h.emailLogin)
	authG.POST("/password/forgot", h.passwordForgot)
	authG.POST("/password/reset", h.passwordReset)
	authG.GET("/oauth/github/start", h.githubStart)
	authG.GET("/oauth/github/callback", h.githubCallback)
	authG.GET("/oauth/:provider/start", h.oauthStart)
	authG.GET("/oauth/:provider/callback", h.oauthCallback)

	adminAuthG := api.Group("/admin/auth")
	adminAuthG.POST("/login", h.adminLogin)
	adminAuthG.POST("/logout", h.logout)

	// 用户侧(SessionAuth + CSRF)
	userG := api.Group("/user")
	userG.Use(middleware.SessionAuth(d.Session, d.Clock), middleware.CSRF(CSRFWhitelist))
	userG.GET("/profile", h.userProfile)
	userG.PATCH("/profile", h.userProfilePatch)
	userG.POST("/password", h.userPassword)
	userG.GET("/oauth/bindings", h.userOAuthBindings)
	userG.POST("/oauth/bindings/github", h.userBindGithub)
	userG.DELETE("/oauth/bindings/github", h.userUnbindGithub)
	userG.POST("/oauth/bindings/:provider", h.bindOAuth)
	userG.DELETE("/oauth/bindings/:provider", h.unbindOAuth)

	// 管理员侧(SessionAuth + CSRF + RoleGate(TenantAdmin))
	adminG := api.Group("/admin")
	adminG.Use(middleware.SessionAuth(d.Session, d.Clock), middleware.CSRF(CSRFWhitelist), middleware.RoleGate(user.RoleTenantAdmin))
	adminG.GET("/auth/me", h.adminMe)
	adminG.GET("/users", h.adminUsersList)
	adminG.GET("/users/:id", h.adminUserDetail)
	adminG.PATCH("/users/:id", h.adminUserPatch)
	adminG.POST("/users/:id/reset_password", h.adminUserResetPassword)
	adminG.DELETE("/users/:id", h.adminUserDelete)
}

type handlers struct{ d Deps }

// ----------------- helpers -----------------

func (h *handlers) setSessionCookies(c *gin.Context, sess *session.Session) {
	maxAge := int(time.Until(sess.ExpiresAt).Seconds())
	if maxAge < 0 {
		maxAge = 0
	}
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(middleware.CookieSession, sess.ID, maxAge, "/", "", h.d.CookieSecure, true)
	token := middleware.DeriveCSRFToken(h.d.CSRFKey, sess.ID)
	c.SetCookie(middleware.CookieCSRF, token, maxAge, "/", "", h.d.CookieSecure, false)
}

func (h *handlers) clearSessionCookies(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(middleware.CookieSession, "", -1, "/", "", h.d.CookieSecure, true)
	c.SetCookie(middleware.CookieCSRF, "", -1, "/", "", h.d.CookieSecure, false)
}

func clientIP(c *gin.Context) string {
	return c.ClientIP()
}

func userAgent(c *gin.Context) string {
	return c.GetHeader("User-Agent")
}

func sendErr(c *gin.Context, err error) {
	if ae, ok := err.(*apierr.Error); ok {
		middleware.SetErr(c, ae)
		return
	}
	middleware.SetErr(c, apierr.Wrap(apierr.CodeInternal, "internal", err))
}

func userView(u *user.User, groupName string) gin.H {
	if u == nil {
		return nil
	}
	view := gin.H{
		"id":         u.ID,
		"username":   u.Username,
		"role":       u.Role,
		"status":     u.Status,
		"created_at": u.CreatedAt.Format(time.RFC3339Nano),
	}
	if u.Email != nil {
		view["email"] = *u.Email
	}
	if u.DisplayName != nil {
		view["display_name"] = *u.DisplayName
	}
	if u.Avatar != nil {
		view["avatar"] = *u.Avatar
	}
	if u.GroupID != nil {
		view["group_id"] = *u.GroupID
	}
	if groupName != "" {
		view["group_name"] = groupName
	}
	if u.EmailVerifiedAt != nil {
		view["email_verified_at"] = u.EmailVerifiedAt.Format(time.RFC3339Nano)
	}
	if u.LastLoginAt != nil {
		view["last_login_at"] = u.LastLoginAt.Format(time.RFC3339Nano)
	}
	return view
}

func sessionView(sess *session.Session) gin.H {
	if sess == nil {
		return nil
	}
	return gin.H{
		"id":         sess.ID,
		"expires_at": sess.ExpiresAt.Format(time.RFC3339Nano),
	}
}

// resolveGroupName 取 group 名;失败返回空串。
func resolveGroupName(ctx context.Context, gsvc group.Service, gid *int64) string {
	if gsvc == nil || gid == nil {
		return ""
	}
	g, _ := gsvc.GetByID(ctx, *gid)
	if g == nil {
		return ""
	}
	return g.Name
}

func parseInt64Param(c *gin.Context, key string) (int64, bool) {
	s := c.Param(key)
	if s == "" {
		return 0, false
	}
	var v int64
	_, err := scanInt64(s, &v)
	return v, err == nil
}

func scanInt64(s string, dst *int64) (int, error) {
	*dst = 0
	for i, c := range s {
		if c < '0' || c > '9' {
			return i, errInvalidInt
		}
		*dst = (*dst)*10 + int64(c-'0')
	}
	return len(s), nil
}

var errInvalidInt = apierr.New(apierr.CodeInvalidParam, "id 必须为正整数")

func optionalQueryInt(c *gin.Context, key string, def int) int {
	s := c.Query(key)
	if s == "" {
		return def
	}
	var v int64
	if _, err := scanInt64(s, &v); err != nil {
		return def
	}
	return int(v)
}

func optionalQueryInt8(c *gin.Context, key string) *int8 {
	s := c.Query(key)
	if s == "" {
		return nil
	}
	var v int64
	if _, err := scanInt64(s, &v); err != nil {
		return nil
	}
	r := int8(v)
	return &r
}

func optionalQueryInt64(c *gin.Context, key string) *int64 {
	s := c.Query(key)
	if s == "" {
		return nil
	}
	var v int64
	if _, err := scanInt64(s, &v); err != nil {
		return nil
	}
	return &v
}

func generateTempPassword() string {
	// 12 位 ASCII,简洁可读
	const charset = "abcdefghjkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789!@#$"
	b := make([]byte, 12)
	for i := range b {
		b[i] = charset[randIntn(len(charset))]
	}
	return string(b)
}

func randIntn(n int) int {
	var buf [8]byte
	_, _ = cryptoRandRead(buf[:])
	x := uint64(buf[0])<<56 | uint64(buf[1])<<48 | uint64(buf[2])<<40 | uint64(buf[3])<<32 |
		uint64(buf[4])<<24 | uint64(buf[5])<<16 | uint64(buf[6])<<8 | uint64(buf[7])
	return int(x % uint64(n))
}

// 用一个稳定的"读 8 随机字节"实现,便于测试时桩。
var cryptoRandRead = func(p []byte) (int, error) {
	return defaultRand(p)
}

// 占位防止 lint
var _ = strings.Contains
