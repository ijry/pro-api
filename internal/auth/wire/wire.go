// Package wire 装配 user/group/auth/session/oauth 服务到 application,
// 并提供 RegisterRoutes 函数把 /api/auth/* + /api/user/* + /api/admin/* 路由
// 挂到 Gin 引擎。
//
// 单独成包(internal/auth/wire)是为了打破 internal/auth ↔ internal/server/handler
// 的循环 import。
package wire

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ijry/pro-api/internal/app"
	"github.com/ijry/pro-api/internal/auth"
	"github.com/ijry/pro-api/internal/auth/oauth"
	"github.com/ijry/pro-api/internal/auth/oauth/github"
	"github.com/ijry/pro-api/internal/auth/session"
	"github.com/ijry/pro-api/internal/auth/verifycode"
	"github.com/ijry/pro-api/internal/group"
	"github.com/ijry/pro-api/internal/notify/email"
	"github.com/ijry/pro-api/internal/server/handler/authhttp"
	"github.com/ijry/pro-api/internal/user"
	"go.uber.org/zap"
)

// sessStoreRegistry 保留每个 Application 对应的 session.Store,供 RegisterRoutes 取回。
var (
	storeMu           sync.RWMutex
	sessStoreRegistry = map[*app.Application]session.Store{}
)

// Wire 装配 user/group/auth/session/oauth 等服务到 application。
//
// 由 cmd/proapi/main.go 调用,在 SetupBasic 之后。
//
// 装配完毕后:
//   - app.UserSvc / app.GroupSvc / app.AuthSvc 已填好(类型分别是
//     user.Service / group.Service / auth.Service)
//   - 内部 registry 也保留了 session.Store,供 RegisterRoutes 取回
//
// 之后应再调 RegisterRoutes(eng, app) 把路由挂到 Gin 引擎。
func Wire(a *app.Application) error {
	ctx := context.Background()

	groupSvc := group.NewService(group.NewRepository(a.DB), a.Clock, a.IDGen)
	userSvc := user.NewService(user.NewRepository(a.DB), groupSvc, a.IDGen, a.Clock)
	a.UserSvc = userSvc
	a.GroupSvc = groupSvc

	ttlDays := a.Setting.GetInt(ctx, "session.ttl_days", 30)
	sliding := a.Setting.GetBool(ctx, "session.sliding", true)
	sessStore, err := session.New(session.Deps{
		DB: session.NewRepository(a.DB), Cache: a.Cache,
		IDGen: a.IDGen, Clock: a.Clock, Log: a.Log,
	}, session.Config{
		TTL: time.Duration(ttlDays) * 24 * time.Hour, Sliding: sliding,
	})
	if err != nil {
		return fmt.Errorf("auth wire: session: %w", err)
	}
	a.AddCloser("session-store", sessStore.Close)
	storeMu.Lock()
	sessStoreRegistry[a] = sessStore
	storeMu.Unlock()

	vc := verifycode.New(a.Cache, a.Clock, a.Log, verifycode.Config{})
	stateStore := oauth.NewStateStore(a.Cache, a.Clock)
	oauthRepo := oauth.NewRepository(a.DB)
	mailer := email.NewStub(a.Log, a.Audit)

	var ghProvider github.Provider
	if cfg := readGithubOAuth(ctx, a); cfg != nil {
		ghProvider = github.New(*cfg)
	}

	authSvc := auth.NewService(auth.Deps{
		User: userSvc, Group: groupSvc, Session: sessStore, VerifyCode: vc,
		Mailer: mailer, GithubProvider: ghProvider, GithubState: stateStore,
		OAuthRepo: oauthRepo,
		Setting:   a.Setting, Audit: a.Audit,
		Clock: a.Clock, IDGen: a.IDGen, Log: a.Log,
	})
	a.AuthSvc = authSvc

	a.Log.Info("auth service ready", zap.Bool("github_enabled", authSvc.GithubEnabled()))
	return nil
}

// readGithubOAuth 读 setting 并解密 client_secret;空配置返回 nil。
//
// setting 结构:
//
//	{"client_id":"...","client_secret":"ENC(v1,...)","redirect_url":"..."}
func readGithubOAuth(ctx context.Context, a *app.Application) *github.Config {
	var raw struct {
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		RedirectURL  string `json:"redirect_url"`
	}
	if err := a.Setting.GetJSON(ctx, "auth.github_oauth", &raw); err != nil {
		return nil
	}
	if raw.ClientID == "" || raw.ClientSecret == "" {
		return nil
	}
	secret := raw.ClientSecret
	if a.Crypto != nil {
		if dec, err := a.Crypto.Decrypt(raw.ClientSecret); err == nil {
			secret = dec
		}
	}
	return &github.Config{
		ClientID:     raw.ClientID,
		ClientSecret: secret,
	}
}

// RegisterRoutes 把 /api/auth/* + /api/user/* + /api/admin/* 三组路由挂到 eng。
//
// 调用方负责传入已经 Use 过 RequestID/AccessLog/Recovery/ErrorResponse 的 engine。
func RegisterRoutes(eng *gin.Engine, a *app.Application) error {
	authSvc, ok := a.AuthSvc.(auth.Service)
	if !ok {
		return errors.New("auth: AuthSvc not wired")
	}
	userSvc, ok := a.UserSvc.(user.Service)
	if !ok {
		return errors.New("auth: UserSvc not wired")
	}
	groupSvc, ok := a.GroupSvc.(group.Service)
	if !ok {
		return errors.New("auth: GroupSvc not wired")
	}
	storeMu.RLock()
	sessStore, ok := sessStoreRegistry[a]
	storeMu.RUnlock()
	if !ok {
		return errors.New("auth: session store not wired")
	}
	authhttp.RegisterRoutes(eng, authhttp.Deps{
		Auth: authSvc, User: userSvc, Group: groupSvc, Session: sessStore,
		OAuthRepo: oauth.NewRepository(a.DB),
		Setting:   a.Setting,
		Audit:     a.Audit,
		Clock:     a.Clock,
		Log:       a.Log,
		CSRFKey:   deriveCSRFKey(a),
		// 生产环境(prod)由 caller 在调用前覆写 cookie_secure;此处默认 false 兼容本地 dev
		CookieSecure: false,
	})
	return nil
}

// deriveCSRFKey 派生一个稳定 32 字节 HMAC key。
//
// 当前用 SHA-256(salt || master_key);master_key 缺失时退化为固定 salt
// 的 hash(只在开发/测试有意义)。
func deriveCSRFKey(a *app.Application) []byte {
	salt := []byte("proapi:csrf-master-salt")
	h := sha256.New()
	h.Write(salt)
	if a.Config != nil && a.Config.MasterKey != "" {
		h.Write([]byte(a.Config.MasterKey))
	}
	return h.Sum(nil)
}
