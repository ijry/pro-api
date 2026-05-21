package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/ijry/pro-api/internal/audit"
	"github.com/ijry/pro-api/internal/auth/oauth"
	"github.com/ijry/pro-api/internal/auth/oauth/github"
	"github.com/ijry/pro-api/internal/auth/password"
	"github.com/ijry/pro-api/internal/auth/session"
	"github.com/ijry/pro-api/internal/auth/verifycode"
	"github.com/ijry/pro-api/internal/group"
	"github.com/ijry/pro-api/internal/notify/email"
	"github.com/ijry/pro-api/internal/setting"
	"github.com/ijry/pro-api/internal/user"
	"github.com/ijry/pro-api/internal/util/clock"
	"github.com/ijry/pro-api/pkg/apierr"
	"go.uber.org/zap"
)

// User 直接复用 user 包的类型。
type User = user.User

// RegisterInput / RegisterResult。
type RegisterInput struct {
	Username string
	Email    string
	Password string
	IP       string
	UA       string
	Lang     string
}

// RegisterResult 是 Register 返回值。
type RegisterResult struct {
	UserID              int64
	EmailVerifyRequired bool
	Session             *session.Session // 验证不要求时直接登录
	User                *User
}

// LoginInput 密码登录入参。
type LoginInput struct {
	Identity string
	Password string
	IP       string
	UA       string
}

// EmailCodeLoginInput 邮箱验证码登录入参。
type EmailCodeLoginInput struct {
	Email string
	Code  string
	IP    string
	UA    string
}

// IDGenerator 与 audit 同形。
type IDGenerator interface {
	Generate() int64
}

// Service 是 auth 顶层接口。
type Service interface {
	Register(ctx context.Context, in RegisterInput) (*RegisterResult, error)
	Login(ctx context.Context, in LoginInput) (*session.Session, *User, error)
	EmailCodeLogin(ctx context.Context, in EmailCodeLoginInput) (*session.Session, *User, error)
	Logout(ctx context.Context, sessionID string) error
	SendEmailCode(ctx context.Context, purpose verifycode.Purpose, emailAddr, ip string) error
	ForgotPassword(ctx context.Context, emailAddr, ip string) error
	ResetPassword(ctx context.Context, emailAddr, code, newPlain string) error
	ChangePassword(ctx context.Context, userID int64, currentSessionID, oldPlain, newPlain string) error

	GithubOAuthStart(ctx context.Context, ip, ua, redirect string, bindUserID int64) (string, error)
	GithubOAuthCallback(ctx context.Context, code, state, ip, ua string) (*session.Session, *User, []byte /*payload*/, error)
	BindGithub(ctx context.Context, userID int64, code, state string) error
	UnbindGithub(ctx context.Context, userID int64) error

	GithubEnabled() bool
}

// Deps 是 NewService 的依赖。
type Deps struct {
	User           user.Service
	Group          group.Service
	Session        session.Store
	VerifyCode     verifycode.Store
	Mailer         email.Mailer
	GithubProvider github.Provider // 可空,表示 GitHub 未启用
	GithubState    oauth.StateStore
	OAuthRepo      oauth.Repository
	Setting        setting.Store
	Audit          audit.Logger
	Clock          clock.Clock
	IDGen          IDGenerator
	Log            *zap.Logger
}

// svc 默认实现。
type svc struct {
	deps Deps
}

// NewService 构造 auth.Service。
func NewService(deps Deps) Service {
	if deps.Clock == nil {
		deps.Clock = clock.Real
	}
	if deps.Log == nil {
		deps.Log = zap.NewNop()
	}
	if deps.Audit == nil {
		deps.Audit = audit.NewNoop()
	}
	return &svc{deps: deps}
}

// GithubEnabled 报告 GitHub OAuth 是否可用。
func (s *svc) GithubEnabled() bool {
	return s.deps.GithubProvider != nil
}

// --- Register ---

func (s *svc) Register(ctx context.Context, in RegisterInput) (*RegisterResult, error) {
	if !s.deps.Setting.GetBool(ctx, "auth.allow_register", true) {
		return nil, apierr.New(apierr.CodeForbidden, "当前不允许注册")
	}
	minLen := s.deps.Setting.GetInt(ctx, "auth.password.min_length", 8)
	requireMixed := s.deps.Setting.GetBool(ctx, "auth.password.require_mixed", false)
	if err := password.CheckStrength(in.Password, minLen, requireMixed); err != nil {
		return nil, err
	}
	if in.Email == "" {
		return nil, apierr.New(apierr.CodeInvalidParam, "email 必填")
	}
	hash, err := password.Hash(in.Password)
	if err != nil {
		return nil, apierr.Wrap(apierr.CodeInternal, "auth: hash", err)
	}
	verifyRequired := s.deps.Setting.GetBool(ctx, "auth.email_verification_required", false)
	status := user.StatusActive
	if verifyRequired {
		status = user.StatusPendingEmailVerify
	}
	u, err := s.deps.User.Create(ctx, user.CreateInput{
		Username:     in.Username,
		Email:        &in.Email,
		PasswordHash: &hash,
		Role:         user.RoleUser,
		Status:       status,
	})
	if err != nil {
		return nil, err
	}
	after, _ := json.Marshal(map[string]any{"username": in.Username, "email": in.Email})
	_ = s.deps.Audit.Log(ctx, audit.Entry{
		Action: "user.register", TargetType: "user", TargetID: &u.ID, After: after, IP: in.IP,
	})

	if verifyRequired {
		// 发码
		code, err := s.deps.VerifyCode.Generate(ctx, verifycode.PurposeRegister, in.Email, in.IP)
		if err != nil {
			return nil, err
		}
		_ = s.deps.Mailer.Send(ctx, email.Message{
			To: in.Email, Subject: "ProAPI 邮箱验证码",
			Body: fmt.Sprintf("您的验证码是 %s,5 分钟内有效。", code),
			Tag:  "verify_register",
		})
		return &RegisterResult{
			UserID: u.ID, EmailVerifyRequired: true, User: u,
		}, nil
	}
	sess, err := s.deps.Session.Create(ctx, u.ID, u.Role, in.IP, in.UA)
	if err != nil {
		return nil, err
	}
	_ = s.deps.User.TouchLogin(ctx, u.ID, in.IP)
	return &RegisterResult{UserID: u.ID, Session: sess, User: u}, nil
}

// --- Login ---

func (s *svc) Login(ctx context.Context, in LoginInput) (*session.Session, *User, error) {
	u, err := s.lookupUser(ctx, in.Identity)
	if err != nil {
		return nil, nil, err
	}
	if u == nil || u.PasswordHash == nil {
		return nil, nil, apierr.New(apierr.CodeWrongPassword, "邮箱或密码错误")
	}
	if err := password.Verify(*u.PasswordHash, in.Password); err != nil {
		if errors.Is(err, password.ErrMismatch) {
			return nil, nil, apierr.New(apierr.CodeWrongPassword, "邮箱或密码错误")
		}
		return nil, nil, apierr.Wrap(apierr.CodeInternal, "auth: verify password", err)
	}
	switch u.Status {
	case user.StatusDisabled:
		return nil, nil, apierr.New(apierr.CodeForbidden, "账号已被禁用")
	case user.StatusPendingEmailVerify:
		return nil, nil, apierr.New(apierr.CodeEmailNotVerified, "邮箱未验证")
	}
	sess, err := s.deps.Session.Create(ctx, u.ID, u.Role, in.IP, in.UA)
	if err != nil {
		return nil, nil, err
	}
	_ = s.deps.User.TouchLogin(ctx, u.ID, in.IP)
	after, _ := json.Marshal(map[string]any{"method": "password", "ip": in.IP})
	_ = s.deps.Audit.Log(ctx, audit.Entry{Action: "user.login", TargetType: "user", TargetID: &u.ID, After: after, IP: in.IP})
	return sess, u, nil
}

// EmailCodeLogin 用邮箱验证码登录,同时完成 verify。
func (s *svc) EmailCodeLogin(ctx context.Context, in EmailCodeLoginInput) (*session.Session, *User, error) {
	u, err := s.deps.User.GetByEmail(ctx, in.Email)
	if err != nil {
		return nil, nil, apierr.Wrap(apierr.CodeDatabase, "auth: get email", err)
	}
	if u == nil {
		return nil, nil, apierr.New(apierr.CodeWrongPassword, "邮箱或验证码错误")
	}
	if u.Status == user.StatusDisabled {
		return nil, nil, apierr.New(apierr.CodeForbidden, "账号已被禁用")
	}
	purpose := verifycode.PurposeLogin
	if u.Status == user.StatusPendingEmailVerify {
		purpose = verifycode.PurposeRegister
	}
	if err := s.deps.VerifyCode.Verify(ctx, purpose, in.Email, in.Code); err != nil {
		return nil, nil, err
	}
	if u.EmailVerifiedAt == nil {
		_ = s.deps.User.MarkEmailVerified(ctx, u.ID)
		_ = s.deps.Audit.Log(ctx, audit.Entry{
			Action: "user.email_verified", TargetType: "user", TargetID: &u.ID, IP: in.IP,
		})
		// 重新加载
		u, _ = s.deps.User.GetByID(ctx, u.ID)
	}
	sess, err := s.deps.Session.Create(ctx, u.ID, u.Role, in.IP, in.UA)
	if err != nil {
		return nil, nil, err
	}
	_ = s.deps.User.TouchLogin(ctx, u.ID, in.IP)
	after, _ := json.Marshal(map[string]any{"method": "email_code"})
	_ = s.deps.Audit.Log(ctx, audit.Entry{Action: "user.login", TargetType: "user", TargetID: &u.ID, After: after, IP: in.IP})
	return sess, u, nil
}

// Logout 撤销 session。
func (s *svc) Logout(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return nil
	}
	return s.deps.Session.Revoke(ctx, sessionID)
}

// SendEmailCode 单独发码(用于"再发一次")。
func (s *svc) SendEmailCode(ctx context.Context, purpose verifycode.Purpose, emailAddr, ip string) error {
	if emailAddr == "" {
		return apierr.New(apierr.CodeInvalidParam, "email 为空")
	}
	code, err := s.deps.VerifyCode.Generate(ctx, purpose, emailAddr, ip)
	if err != nil {
		return err
	}
	_ = s.deps.Mailer.Send(ctx, email.Message{
		To: emailAddr, Subject: "ProAPI 验证码",
		Body: fmt.Sprintf("您的验证码是 %s,5 分钟内有效。", code),
		Tag:  "verify_" + string(purpose),
	})
	return nil
}

// ForgotPassword 发"重置链接"邮件。无论 email 是否注册都返回 ok。
func (s *svc) ForgotPassword(ctx context.Context, emailAddr, ip string) error {
	if emailAddr == "" {
		return apierr.New(apierr.CodeInvalidParam, "email 为空")
	}
	u, err := s.deps.User.GetByEmail(ctx, emailAddr)
	if err != nil {
		return apierr.Wrap(apierr.CodeDatabase, "auth: get email", err)
	}
	if u == nil {
		return nil // 静默忽略
	}
	code, err := s.deps.VerifyCode.Generate(ctx, verifycode.PurposePasswordReset, emailAddr, ip)
	if err != nil {
		// 节流命中也返回 ok 防枚举
		var ae *apierr.Error
		if errors.As(err, &ae) && ae.Code == apierr.CodeRateLimitUser {
			return nil
		}
		return err
	}
	_ = s.deps.Mailer.Send(ctx, email.Message{
		To: emailAddr, Subject: "ProAPI 密码重置",
		Body: fmt.Sprintf("您的密码重置验证码是 %s,5 分钟内有效。", code),
		Tag:  "password_reset",
	})
	return nil
}

// ResetPassword 校验码 + 改 hash + RevokeAllForUser。
func (s *svc) ResetPassword(ctx context.Context, emailAddr, code, newPlain string) error {
	minLen := s.deps.Setting.GetInt(ctx, "auth.password.min_length", 8)
	requireMixed := s.deps.Setting.GetBool(ctx, "auth.password.require_mixed", false)
	if err := password.CheckStrength(newPlain, minLen, requireMixed); err != nil {
		return err
	}
	u, err := s.deps.User.GetByEmail(ctx, emailAddr)
	if err != nil {
		return apierr.Wrap(apierr.CodeDatabase, "auth: get email", err)
	}
	if u == nil {
		return apierr.New(apierr.CodeCaptchaInvalid, "验证码错误或已过期")
	}
	if err := s.deps.VerifyCode.Verify(ctx, verifycode.PurposePasswordReset, emailAddr, code); err != nil {
		return err
	}
	hash, err := password.Hash(newPlain)
	if err != nil {
		return apierr.Wrap(apierr.CodeInternal, "auth: hash", err)
	}
	if err := s.deps.User.UpdatePasswordHash(ctx, u.ID, hash); err != nil {
		return apierr.Wrap(apierr.CodeDatabase, "auth: update hash", err)
	}
	_ = s.deps.Session.RevokeAllForUser(ctx, u.ID)
	_ = s.deps.Audit.Log(ctx, audit.Entry{Action: "user.reset_password", TargetType: "user", TargetID: &u.ID})
	return nil
}

// ChangePassword 已登录改密。RevokeAllForUser(除当前 session)。
func (s *svc) ChangePassword(ctx context.Context, userID int64, currentSessionID, oldPlain, newPlain string) error {
	u, err := s.deps.User.GetByID(ctx, userID)
	if err != nil {
		return apierr.Wrap(apierr.CodeDatabase, "auth: get user", err)
	}
	if u == nil {
		return apierr.New(apierr.CodeForbidden, "用户不存在")
	}
	if u.PasswordHash == nil {
		// OAuth-only 账号"修改密码"实质是"设置密码"
	} else if err := password.Verify(*u.PasswordHash, oldPlain); err != nil {
		return apierr.New(apierr.CodeWrongPassword, "原密码错误")
	}
	minLen := s.deps.Setting.GetInt(ctx, "auth.password.min_length", 8)
	requireMixed := s.deps.Setting.GetBool(ctx, "auth.password.require_mixed", false)
	if err := password.CheckStrength(newPlain, minLen, requireMixed); err != nil {
		return err
	}
	hash, err := password.Hash(newPlain)
	if err != nil {
		return apierr.Wrap(apierr.CodeInternal, "auth: hash", err)
	}
	if err := s.deps.User.UpdatePasswordHash(ctx, u.ID, hash); err != nil {
		return apierr.Wrap(apierr.CodeDatabase, "auth: update hash", err)
	}
	// 简化:RevokeAllForUser 把当前也撤掉;handler 层重发 cookie
	// 更细致的"保留当前 session"需要 SessionStore 支持 excludeID,M1 暂不做。
	_ = s.deps.Session.RevokeAllForUser(ctx, u.ID)
	_ = s.deps.Audit.Log(ctx, audit.Entry{Action: "user.change_password", TargetType: "user", TargetID: &u.ID})
	return nil
}

// --- GitHub OAuth ---

type githubStatePayload struct {
	BindUserID int64  `json:"bind_user_id,omitempty"`
	Redirect   string `json:"redirect,omitempty"`
	IP         string `json:"ip,omitempty"`
	UA         string `json:"ua,omitempty"`
	IssuedAt   int64  `json:"issued_at,omitempty"`
}

// GithubOAuthStart 生成跳转 URL。
func (s *svc) GithubOAuthStart(ctx context.Context, ip, ua, redirect string, bindUserID int64) (string, error) {
	if s.deps.GithubProvider == nil {
		return "", apierr.New(apierr.CodeForbidden, "GitHub 登录未启用")
	}
	cfg := s.readGithubConfig(ctx)
	if cfg.RedirectURL == "" {
		return "", apierr.New(apierr.CodeForbidden, "GitHub OAuth redirect_url 未配置")
	}
	payload := githubStatePayload{
		BindUserID: bindUserID, Redirect: redirect, IP: ip, UA: ua,
		IssuedAt: s.deps.Clock.Now().Unix(),
	}
	b, _ := json.Marshal(payload)
	state, err := s.deps.GithubState.Issue(ctx, "github", b)
	if err != nil {
		return "", err
	}
	return s.deps.GithubProvider.BuildAuthURL(ctx, state, cfg.RedirectURL)
}

// GithubOAuthCallback 处理回调,返回 sess + user + payload(供 handler 取 redirect)。
func (s *svc) GithubOAuthCallback(ctx context.Context, code, state, ip, ua string) (*session.Session, *User, []byte, error) {
	if s.deps.GithubProvider == nil {
		return nil, nil, nil, apierr.New(apierr.CodeForbidden, "GitHub 登录未启用")
	}
	payload, err := s.deps.GithubState.Consume(ctx, "github", state)
	if err != nil {
		return nil, nil, nil, err
	}
	var p githubStatePayload
	_ = json.Unmarshal(payload, &p)

	cfg := s.readGithubConfig(ctx)
	info, _, err := s.deps.GithubProvider.Exchange(ctx, code, cfg.RedirectURL)
	if err != nil {
		return nil, nil, payload, err
	}
	if info.ID == 0 {
		return nil, nil, payload, apierr.New(apierr.CodeUpstreamError, "GitHub 返回 user 缺少 id")
	}
	providerUID := fmt.Sprintf("%d", info.ID)

	binding, err := s.deps.OAuthRepo.FindByProviderUID(ctx, "github", providerUID)
	if err != nil {
		return nil, nil, payload, apierr.Wrap(apierr.CodeDatabase, "oauth lookup binding", err)
	}

	var u *User
	if binding != nil {
		u, err = s.deps.User.GetByID(ctx, binding.UserID)
		if err != nil {
			return nil, nil, payload, apierr.Wrap(apierr.CodeDatabase, "oauth user", err)
		}
		if u == nil || u.Status == user.StatusDisabled {
			return nil, nil, payload, apierr.New(apierr.CodeForbidden, "账号不可用")
		}
	} else if info.Email != "" {
		if existing, _ := s.deps.User.GetByEmail(ctx, info.Email); existing != nil {
			if existing.Status == user.StatusDisabled {
				return nil, nil, payload, apierr.New(apierr.CodeForbidden, "账号不可用")
			}
			u = existing
			if err := s.createBinding(ctx, u.ID, providerUID, info); err != nil {
				return nil, nil, payload, err
			}
			after, _ := json.Marshal(map[string]any{"provider": "github"})
			_ = s.deps.Audit.Log(ctx, audit.Entry{Action: "oauth.bind", TargetType: "user", TargetID: &u.ID, ActorID: &u.ID, After: after})
		}
	}
	if u == nil {
		// 自动建账号
		var err error
		u, err = s.autoCreateGithubUser(ctx, info, ip)
		if err != nil {
			return nil, nil, payload, err
		}
		if err := s.createBinding(ctx, u.ID, providerUID, info); err != nil {
			return nil, nil, payload, err
		}
	}

	sess, err := s.deps.Session.Create(ctx, u.ID, u.Role, ip, ua)
	if err != nil {
		return nil, nil, payload, err
	}
	_ = s.deps.User.TouchLogin(ctx, u.ID, ip)
	after, _ := json.Marshal(map[string]any{"method": "github"})
	_ = s.deps.Audit.Log(ctx, audit.Entry{Action: "user.login", TargetType: "user", TargetID: &u.ID, After: after, IP: ip})
	return sess, u, payload, nil
}

// BindGithub 已登录的用户绑定 GitHub。
func (s *svc) BindGithub(ctx context.Context, userID int64, code, state string) error {
	if s.deps.GithubProvider == nil {
		return apierr.New(apierr.CodeForbidden, "GitHub 登录未启用")
	}
	payload, err := s.deps.GithubState.Consume(ctx, "github", state)
	if err != nil {
		return err
	}
	var p githubStatePayload
	_ = json.Unmarshal(payload, &p)
	if p.BindUserID != userID {
		return apierr.New(apierr.CodeForbidden, "state 与当前用户不匹配")
	}
	cfg := s.readGithubConfig(ctx)
	info, _, err := s.deps.GithubProvider.Exchange(ctx, code, cfg.RedirectURL)
	if err != nil {
		return err
	}
	providerUID := fmt.Sprintf("%d", info.ID)
	existing, _ := s.deps.OAuthRepo.FindByProviderUID(ctx, "github", providerUID)
	if existing != nil && existing.UserID != userID {
		return apierr.New(apierr.CodeForbidden, "此 GitHub 已绑定其他账号")
	}
	if existing != nil && existing.UserID == userID {
		return nil // idempotent
	}
	if err := s.createBinding(ctx, userID, providerUID, info); err != nil {
		return err
	}
	after, _ := json.Marshal(map[string]any{"provider": "github", "uid": providerUID})
	_ = s.deps.Audit.Log(ctx, audit.Entry{Action: "oauth.bind", TargetType: "user", TargetID: &userID, ActorID: &userID, After: after})
	return nil
}

// UnbindGithub 解绑 GitHub。
func (s *svc) UnbindGithub(ctx context.Context, userID int64) error {
	u, err := s.deps.User.GetByID(ctx, userID)
	if err != nil {
		return apierr.Wrap(apierr.CodeDatabase, "auth: get user", err)
	}
	if u == nil {
		return apierr.New(apierr.CodeForbidden, "用户不存在")
	}
	if u.PasswordHash == nil {
		return apierr.New(apierr.CodeForbidden, "解绑前请先设置密码,以免无法登录")
	}
	if err := s.deps.OAuthRepo.DeleteByUserProvider(ctx, userID, "github"); err != nil {
		return apierr.Wrap(apierr.CodeDatabase, "auth: delete binding", err)
	}
	after, _ := json.Marshal(map[string]any{"provider": "github"})
	_ = s.deps.Audit.Log(ctx, audit.Entry{Action: "oauth.unbind", TargetType: "user", TargetID: &userID, ActorID: &userID, After: after})
	return nil
}

// --- internal helpers ---

// lookupUser 按 identity(email 含 @ 或 username)查 User。
func (s *svc) lookupUser(ctx context.Context, identity string) (*User, error) {
	identity = strings.TrimSpace(identity)
	if identity == "" {
		return nil, nil
	}
	if strings.Contains(identity, "@") {
		u, err := s.deps.User.GetByEmail(ctx, identity)
		if err != nil {
			return nil, apierr.Wrap(apierr.CodeDatabase, "auth: get email", err)
		}
		return u, nil
	}
	u, err := s.deps.User.GetByUsername(ctx, identity)
	if err != nil {
		return nil, apierr.Wrap(apierr.CodeDatabase, "auth: get username", err)
	}
	return u, nil
}

// autoCreateGithubUser 自动建账号。
func (s *svc) autoCreateGithubUser(ctx context.Context, info *github.UserInfo, ip string) (*User, error) {
	base := info.Login
	if base == "" {
		base = fmt.Sprintf("gh_%d", info.ID)
	}
	username, err := s.pickUsername(ctx, base)
	if err != nil {
		return nil, err
	}
	var emailPtr *string
	if info.Email != "" {
		emailPtr = &info.Email
	}
	var avatarPtr *string
	if info.Avatar != "" {
		avatarPtr = &info.Avatar
	}
	in := user.CreateInput{
		Username:    username,
		Email:       emailPtr,
		Avatar:      avatarPtr,
		DisplayName: ptrIfNonEmpty(info.Name),
		Role:        user.RoleUser,
		Status:      user.StatusActive, // GitHub 验证过邮箱,直接 active
	}
	u, err := s.deps.User.Create(ctx, in)
	if err != nil {
		// 邮箱重复时改为不带邮箱再建一次(保险)
		var ae *apierr.Error
		if errors.As(err, &ae) && ae.Code == apierr.CodeEmailRegistered {
			in.Email = nil
			u, err = s.deps.User.Create(ctx, in)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	after, _ := json.Marshal(map[string]any{"method": "github", "username": username})
	_ = s.deps.Audit.Log(ctx, audit.Entry{Action: "user.register", TargetType: "user", TargetID: &u.ID, After: after, IP: ip})
	return u, nil
}

// pickUsername 重试附 _2/_3 直到不冲突。
func (s *svc) pickUsername(ctx context.Context, base string) (string, error) {
	base = strings.TrimSpace(base)
	if base == "" {
		base = "user"
	}
	candidate := base
	for i := 2; i < 100; i++ {
		exists, err := s.deps.User.GetByUsername(ctx, candidate)
		if err != nil {
			return "", apierr.Wrap(apierr.CodeDatabase, "auth: get username", err)
		}
		if exists == nil {
			return candidate, nil
		}
		candidate = fmt.Sprintf("%s_%d", base, i)
	}
	return fmt.Sprintf("%s_%d", base, s.deps.Clock.Now().UnixNano()%10000), nil
}

func (s *svc) createBinding(ctx context.Context, userID int64, providerUID string, info *github.UserInfo) error {
	if s.deps.IDGen == nil {
		return errors.New("auth: idgen not configured")
	}
	now := s.deps.Clock.Now().UTC()
	prof, _ := json.Marshal(map[string]any{
		"login": info.Login, "name": info.Name, "avatar_url": info.Avatar,
	})
	return s.deps.OAuthRepo.Create(ctx, &oauth.Binding{
		ID: s.deps.IDGen.Generate(), UserID: userID,
		Provider: "github", ProviderUID: providerUID, Email: info.Email,
		Profile:   prof,
		CreatedAt: now, UpdatedAt: now,
	})
}

type githubConfig struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURL  string `json:"redirect_url"`
}

// readGithubConfig 读 setting "auth.github_oauth"。
//
// 注意:本 spec 不解密 client_secret(由 wire 层解密后注入 provider);
// 这里只读 redirect_url。
func (s *svc) readGithubConfig(ctx context.Context) githubConfig {
	var c githubConfig
	if err := s.deps.Setting.GetJSON(ctx, "auth.github_oauth", &c); err != nil {
		s.deps.Log.Debug("auth: github_oauth setting missing", zap.Error(err))
	}
	return c
}

// ptrIfNonEmpty 把非空字符串转指针。
func ptrIfNonEmpty(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}

// 防止 url 包未使用
var _ = url.PathEscape

// 防止 time 包未使用
var _ = time.Second
