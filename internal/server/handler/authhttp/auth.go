package authhttp

import (
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/ijry/pro-api/internal/auth"
	"github.com/ijry/pro-api/internal/auth/verifycode"
	"github.com/ijry/pro-api/pkg/apierr"
)

func netURLQueryEscape(s string) string { return url.QueryEscape(s) }

type registerReq struct {
	Username   string `json:"username"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	Lang       string `json:"lang"`
	InviteCode string `json:"invite_code"`
}

func (h *handlers) register(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		sendErr(c, apierr.New(apierr.CodeInvalidParam, "请求格式错误"))
		return
	}
	res, err := h.d.Auth.Register(c.Request.Context(), auth.RegisterInput{
		Username:   req.Username,
		Email:      req.Email,
		Password:   req.Password,
		IP:         clientIP(c),
		UA:         userAgent(c),
		Lang:       req.Lang,
		InviteCode: req.InviteCode,
	})
	if err != nil {
		sendErr(c, err)
		return
	}
	resp := gin.H{
		"user":                  userView(res.User, resolveGroupName(c.Request.Context(), h.d.Group, res.User.GroupID)),
		"session":               sessionView(res.Session),
		"email_verify_required": res.EmailVerifyRequired,
	}
	if res.Session != nil {
		h.setSessionCookies(c, res.Session)
	}
	if res.EmailVerifyRequired {
		resp["hint"] = "验证码已发送至 " + req.Email + ",5 分钟内有效"
	}
	c.JSON(200, resp)
}

type loginReq struct {
	Identity string `json:"identity"`
	Password string `json:"password"`
}

func (h *handlers) login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		sendErr(c, apierr.New(apierr.CodeInvalidParam, "请求格式错误"))
		return
	}
	sess, u, err := h.d.Auth.Login(c.Request.Context(), auth.LoginInput{
		Identity: req.Identity, Password: req.Password,
		IP: clientIP(c), UA: userAgent(c),
	})
	if err != nil {
		sendErr(c, err)
		return
	}
	h.setSessionCookies(c, sess)
	c.JSON(200, gin.H{
		"user":    userView(u, resolveGroupName(c.Request.Context(), h.d.Group, u.GroupID)),
		"session": sessionView(sess),
	})
}

// adminLogin 与 login 共用 Service.Login,只是结果额外校验 role>=tenant_admin。
func (h *handlers) adminLogin(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		sendErr(c, apierr.New(apierr.CodeInvalidParam, "请求格式错误"))
		return
	}
	sess, u, err := h.d.Auth.Login(c.Request.Context(), auth.LoginInput{
		Identity: req.Identity, Password: req.Password,
		IP: clientIP(c), UA: userAgent(c),
	})
	if err != nil {
		sendErr(c, err)
		return
	}
	// 仅允许管理员;否则立刻撤销刚发的 session
	if u.Role < roleTenantAdmin() {
		_ = h.d.Session.Revoke(c.Request.Context(), sess.ID)
		sendErr(c, apierr.New(apierr.CodeForbidden, "无后台权限"))
		return
	}
	h.setSessionCookies(c, sess)
	c.JSON(200, gin.H{
		"user":    userView(u, resolveGroupName(c.Request.Context(), h.d.Group, u.GroupID)),
		"session": sessionView(sess),
	})
}

func roleTenantAdmin() int8 { return 2 }

func (h *handlers) logout(c *gin.Context) {
	if v, err := c.Cookie("proapi_session"); err == nil && v != "" {
		_ = h.d.Auth.Logout(c.Request.Context(), v)
	}
	h.clearSessionCookies(c)
	c.JSON(200, gin.H{"ok": true})
}

type emailSendCodeReq struct {
	Email   string `json:"email"`
	Purpose string `json:"purpose"`
}

func (h *handlers) emailSendCode(c *gin.Context) {
	var req emailSendCodeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		sendErr(c, apierr.New(apierr.CodeInvalidParam, "请求格式错误"))
		return
	}
	purpose := verifycode.Purpose(req.Purpose)
	switch purpose {
	case verifycode.PurposeRegister, verifycode.PurposeLogin,
		verifycode.PurposePasswordReset, verifycode.PurposeBindEmail:
	default:
		sendErr(c, apierr.New(apierr.CodeInvalidParam, "purpose 无效"))
		return
	}
	if err := h.d.Auth.SendEmailCode(c.Request.Context(), purpose, req.Email, clientIP(c)); err != nil {
		sendErr(c, err)
		return
	}
	c.JSON(200, gin.H{"ok": true, "ttl_seconds": 300})
}

type emailLoginReq struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

func (h *handlers) emailLogin(c *gin.Context) {
	var req emailLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		sendErr(c, apierr.New(apierr.CodeInvalidParam, "请求格式错误"))
		return
	}
	sess, u, err := h.d.Auth.EmailCodeLogin(c.Request.Context(), auth.EmailCodeLoginInput{
		Email: req.Email, Code: req.Code,
		IP: clientIP(c), UA: userAgent(c),
	})
	if err != nil {
		sendErr(c, err)
		return
	}
	h.setSessionCookies(c, sess)
	c.JSON(200, gin.H{
		"user":    userView(u, resolveGroupName(c.Request.Context(), h.d.Group, u.GroupID)),
		"session": sessionView(sess),
	})
}

type passwordForgotReq struct {
	Email string `json:"email"`
}

func (h *handlers) passwordForgot(c *gin.Context) {
	var req passwordForgotReq
	if err := c.ShouldBindJSON(&req); err != nil {
		sendErr(c, apierr.New(apierr.CodeInvalidParam, "请求格式错误"))
		return
	}
	_ = h.d.Auth.ForgotPassword(c.Request.Context(), req.Email, clientIP(c))
	c.JSON(200, gin.H{"ok": true})
}

type passwordResetReq struct {
	Email       string `json:"email"`
	Code        string `json:"code"`
	NewPassword string `json:"new_password"`
}

func (h *handlers) passwordReset(c *gin.Context) {
	var req passwordResetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		sendErr(c, apierr.New(apierr.CodeInvalidParam, "请求格式错误"))
		return
	}
	if err := h.d.Auth.ResetPassword(c.Request.Context(), req.Email, req.Code, req.NewPassword); err != nil {
		sendErr(c, err)
		return
	}
	c.JSON(200, gin.H{"ok": true})
}

// --- GitHub OAuth ---

func (h *handlers) githubStart(c *gin.Context) {
	redirect := c.Query("redirect")
	url, err := h.d.Auth.GithubOAuthStart(c.Request.Context(), clientIP(c), userAgent(c), redirect, 0)
	if err != nil {
		sendErr(c, err)
		return
	}
	if c.Query("json") == "1" {
		c.JSON(200, gin.H{"url": url})
		return
	}
	c.Redirect(302, url)
}

func (h *handlers) githubCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")
	sess, u, _, err := h.d.Auth.GithubOAuthCallback(c.Request.Context(), code, state, clientIP(c), userAgent(c))
	if err != nil {
		// 失败重定向到 /login?oauth_error=...
		if c.Query("json") == "1" {
			sendErr(c, err)
			return
		}
		c.Redirect(302, "/login?oauth_error=github&oauth_message="+urlEncode(err.Error()))
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
	c.Redirect(302, "/")
}

func urlEncode(s string) string {
	return urlEscapePlus(s)
}

// urlEscapePlus 是简化的 query 编码,够用 OAuth 错误回跳。
func urlEscapePlus(s string) string {
	return netURLQueryEscape(s)
}
