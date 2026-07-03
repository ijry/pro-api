// proapi 主程序入口。
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ijry/pro-api/internal/account"
	accountwire "github.com/ijry/pro-api/internal/account/wire"
	"github.com/ijry/pro-api/internal/adapter"
	"github.com/ijry/pro-api/internal/adapterreg"
	"github.com/ijry/pro-api/internal/app"
	"github.com/ijry/pro-api/internal/app/config"
	authhwire "github.com/ijry/pro-api/internal/auth/wire"
	"github.com/ijry/pro-api/internal/billing"
	"github.com/ijry/pro-api/internal/channel"
	"github.com/ijry/pro-api/internal/group"
	ilog "github.com/ijry/pro-api/internal/log"
	"github.com/ijry/pro-api/internal/notice"
	"github.com/ijry/pro-api/internal/observability/logger"
	"github.com/ijry/pro-api/internal/observability/metrics"
	"github.com/ijry/pro-api/internal/payment"
	mmanual "github.com/ijry/pro-api/internal/payment/manual"
	monline "github.com/ijry/pro-api/internal/payment/online"
	"github.com/ijry/pro-api/internal/payment/redeem"
	"github.com/ijry/pro-api/internal/pricing"
	"github.com/ijry/pro-api/internal/ratelimit"
	"github.com/ijry/pro-api/internal/relay"
	"github.com/ijry/pro-api/internal/server"
	"github.com/ijry/pro-api/internal/server/handler/admin"
	"github.com/ijry/pro-api/internal/server/handler/logh"
	paymenthdr "github.com/ijry/pro-api/internal/server/handler/payment"
	publichdr "github.com/ijry/pro-api/internal/server/handler/public"
	relayhdr "github.com/ijry/pro-api/internal/server/handler/relay"
	tokenhdr "github.com/ijry/pro-api/internal/server/handler/token"
	userhdr "github.com/ijry/pro-api/internal/server/handler/user"
	"github.com/ijry/pro-api/internal/server/middleware"
	frontendstatic "github.com/ijry/pro-api/internal/server/static"
	"github.com/ijry/pro-api/internal/token"
	"github.com/ijry/pro-api/internal/version"
	"github.com/ijry/pro-api/internal/wallet"
	"github.com/ijry/pro-api/pkg/apierr"
	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func run() error {
	configPath := flag.String("config", "", "配置文件路径(可选,留空则仅用 ENV 与默认值)")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	log, err := logger.New(cfg.Log.Level, cfg.Log.Format, nil)
	if err != nil {
		return fmt.Errorf("logger: %w", err)
	}
	defer func() { _ = log.Sync() }()

	log.Info("proapi starting",
		zap.String("version", version.String()),
		zap.String("addr", cfg.Server.Addr),
		zap.Int("node_id", cfg.NodeID),
	)

	metrics.Init()

	ctx := context.Background()

	// ── 基础设施层 ────────────────────────────────────────────────
	application, err := app.SetupBasic(ctx, cfg, log)
	if err != nil {
		return fmt.Errorf("setup basic: %w", err)
	}

	// ── 钱包 ──────────────────────────────────────────────────────
	ws, err := wallet.New(wallet.Config{
		DB:    application.DB,
		Cache: application.Cache,
		Log:   application.Log,
		Clock: application.Clock,
		IDGen: application.IDGen,
		Audit: application.Audit,
	})
	if err != nil {
		return fmt.Errorf("wallet: %w", err)
	}
	application.WalletStore = ws
	application.AddCloser("wallet", ws.Close)

	// ── 用户/鉴权/分组 ──────────────────────────────────────────
	if err := authhwire.Wire(application); err != nil {
		return fmt.Errorf("auth wire: %w", err)
	}

	// ── API 令牌 ─────────────────────────────────────────────────
	if err := token.WireToken(application); err != nil {
		return fmt.Errorf("token wire: %w", err)
	}

	// ── 计费 ──────────────────────────────────────────────────────
	if err := billing.WireBilling(application); err != nil {
		return fmt.Errorf("billing wire: %w", err)
	}

	// ── 定价 ──────────────────────────────────────────────────────
	if err := pricing.WirePricing(ctx, application); err != nil {
		log.Warn("pricing wire failed, billing estimates will be zero", zap.Error(err))
	}

	// ── 公告 ──────────────────────────────────────────────────────
	if err := notice.WireNotice(application); err != nil {
		return fmt.Errorf("notice wire: %w", err)
	}

	// ── 限流 ──────────────────────────────────────────────────────
	if err := ratelimit.WireRateLimit(application); err != nil {
		return fmt.Errorf("ratelimit wire: %w", err)
	}

	// ── 日志 ──────────────────────────────────────────────────────
	if err := ilog.WireLog(ctx, application); err != nil {
		return fmt.Errorf("log wire: %w", err)
	}

	// ── 渠道 ──────────────────────────────────────────────────────
	if err := channel.WireChannel(ctx, application); err != nil {
		return fmt.Errorf("channel wire: %w", err)
	}

	// ── 号池 ──────────────────────────────────────────────────────
	if err := accountwire.WireAccount(ctx, application); err != nil {
		return fmt.Errorf("account wire: %w", err)
	}
	accountFacade := account.FacadeFrom(application)

	// ── 适配器 & Relay ───────────────────────────────────────────
	adapterReg := adapter.NewRegistry()
	adapterreg.WireAdapters(adapterReg, application.Tokenize)
	application.AdapterReg = adapterReg

	// accountFacade 实现 relay.AccountFacade(隐式接口);为 nil 时 relay 跳过号池分支。
	var af relay.AccountFacade
	if accountFacade != nil {
		af = accountFacade
	}
	relaySvc := relay.New(adapterReg, af)
	application.Relay = relaySvc

	// ── 支付:手动充值 + 兑换码 ──────────────────────────────────
	if err := mmanual.Wire(application); err != nil {
		return fmt.Errorf("manual payment wire: %w", err)
	}
	if err := redeem.Wire(application); err != nil {
		return fmt.Errorf("redeem wire: %w", err)
	}

	// ── 支付:在线支付(Stripe 等) ─────────────────────────────
	// providers 在此处留空;各支付渠道 provider 由后续 C1/C2 任务注入。
	// online.Service.HandleWebhook / CreateOrder 在 provider 为空时会返回
	// "unknown provider" 错误,这是预期行为。
	monline.Wire(application, nil, nil, nil)

	// ── 路由注册 ─────────────────────────────────────────────────
	engine := server.NewEngine(log)
	if err := wireRoutes(ctx, engine, application, log); err != nil {
		return fmt.Errorf("routes: %w", err)
	}
	if err := frontendstatic.RegisterEmbedded(engine); err != nil {
		return fmt.Errorf("static assets: %w", err)
	}

	// ── HTTP 服务器 ───────────────────────────────────────────────
	srv := &http.Server{
		Addr:         cfg.Server.Addr,
		Handler:      engine,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeoutMS) * time.Millisecond,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeoutMS) * time.Millisecond,
	}

	serverErr := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-stop:
		log.Info("shutting down")
	case err := <-serverErr:
		log.Error("http server error", zap.Error(err))
	}

	shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		log.Error("http server shutdown failed", zap.Error(err))
	}
	if err := application.Shutdown(shutCtx); err != nil {
		log.Error("application shutdown reported errors", zap.Error(err))
	}
	return nil
}

// wireRoutes 把所有业务路由挂到 Gin 引擎。
func wireRoutes(ctx context.Context, eng *gin.Engine, a *app.Application, log *zap.Logger) error {
	// 通用认证/用户/管理员路由(M1-02)
	if err := authhwire.RegisterRoutes(eng, a); err != nil {
		return fmt.Errorf("auth routes: %w", err)
	}

	// 管理员错误响应中间件
	adminGroup := eng.Group("/api/admin", middleware.ErrorResponse("json"))
	userGroup := eng.Group("/api/user", middleware.ErrorResponse("json"))
	publicGroup := eng.Group("/api/public")

	// 公告路由(M1-12)
	wireNoticeHandlers(a, adminGroup, userGroup, publicGroup, log)

	// 系统设置路由(M1-12)
	wireSettingHandler(a, adminGroup, log)

	// 日志 & 统计路由(M1-08)
	wireLogHandlers(a, adminGroup, userGroup)

	// 号池路由(M2b P9)
	wireAccountHandler(a, adminGroup, log)

	// 在线支付路由(M2a B4)
	wirePaymentRoutes(eng, a, log)

	// 代理路由(/v1/*) — TokenAuth 保护 + GroupRatio + ratelimit
	var groupRatioLookup middleware.GroupRatioLookup
	if gs, ok := a.GroupSvc.(group.Service); ok {
		groupRatioLookup = func(ctx context.Context, id int64) (float64, error) {
			return gs.RatioFor(ctx, id)
		}
	}
	tokenStore, _ := a.TokenStore.(token.Store)
	v1 := eng.Group("/v1", middleware.ErrorResponse("openai"))
	if tokenStore != nil {
		v1.Use(token.TokenAuth(tokenStore))
	} else {
		log.Warn("token store not wired; /v1 routes are not token protected")
	}
	v1.Use(middleware.GroupRatioMiddleware(groupRatioLookup))
	if limiter, ok := a.Limiter.(ratelimit.Limiter); ok {
		if planner := ratelimit.PlannerFrom(a); planner != nil {
			v1.Use(ratelimit.Middleware(limiter, planner, a.Setting, log))
		}
	}
	chFacade := channel.FacadeFrom(a)
	if chFacade != nil {
		tokenhdr.NewModelsHandler(chFacade).Register(v1)
	}
	if relaySvc, ok := a.Relay.(*relay.Service); ok {
		var chSelector channel.Selector
		if chFacade != nil {
			chSelector = chFacade.Selector
		}
		relayH := relayhdr.New(relayhdr.Deps{
			Relay:   relaySvc,
			Pricing: pricing.PricingFrom(a),
			Biller:  billerFrom(a),
			LogSvc:  ilog.StoreFrom(a),
			Channel: chSelector,
		})
		relayH.Register(v1)
		// M2a: Anthropic 入口 /v1/messages
		relayH.RegisterAnthropicRoutes(v1)
		// M2a: Gemini 入口 /v1beta/models/:model/...
		v1beta := eng.Group("/v1beta", middleware.ErrorResponse("json"))
		if tokenStore != nil {
			v1beta.Use(token.TokenAuth(tokenStore))
		}
		v1beta.Use(middleware.GroupRatioMiddleware(groupRatioLookup))
		relayH.RegisterGeminiRoutes(v1beta)
	}

	// Playground 路由 — SessionAuth 保护(M2a E1)
	sessStore := authhwire.SessionStoreFrom(a)
	if sessStore != nil {
		pgH, err := userhdr.WirePlayground(a)
		if err != nil {
			log.Warn("playground handler wiring failed", zap.Error(err))
		} else {
			pgGroup := eng.Group("/api/user/playground",
				middleware.ErrorResponse("json"),
				middleware.SessionAuth(sessStore, a.Clock),
			)
			pgGroup.POST("/chat", pgH.Chat)
		}
	} else {
		log.Warn("session store not available; playground routes skipped")
	}

	// Invite 路由(邀请返佣)
	wireInviteRoutes(eng, a, log)

	// API Key 路由(/api/user/apikeys + /api/admin/apikeys)
	wireApikeyRoutes(eng, a, log)

	// 公开分组列表(供用户创建 apikey 时选择分组)
	wirePublicGroupsRoute(publicGroup, a)

	_ = ctx
	return nil
}

// wireNoticeHandlers 挂载公告 handler:
//   - public:无鉴权(/api/public/notices)
//   - user:SessionAuth(/api/user/notices)
//   - admin:SessionAuth + RoleGate(2)(/api/admin/notices)
func wireNoticeHandlers(a *app.Application, adminG, userG, publicG *gin.RouterGroup, log *zap.Logger) {
	if notice.ServiceFrom(a) == nil {
		log.Warn("notice service not wired; notice routes skipped")
		return
	}
	actorOf := func(c *gin.Context) int64 { return middleware.UserID(c) }

	// public:无鉴权
	if h, err := publichdr.WirePublicNotice(a); err != nil {
		log.Warn("public notice handler wiring failed", zap.Error(err))
	} else {
		h.Register(publicG)
	}

	sessStore := authhwire.SessionStoreFrom(a)
	if sessStore == nil {
		log.Warn("session store not available; admin/user notice routes skipped")
		return
	}
	// user:登录可见
	if h, err := userhdr.WireUserNotice(a, actorOf); err != nil {
		log.Warn("user notice handler wiring failed", zap.Error(err))
	} else {
		h.Register(userG.Group("", middleware.SessionAuth(sessStore, a.Clock)))
	}
	// admin:RoleGate(2)
	if h, err := admin.WireAdminNotice(a, actorOf); err != nil {
		log.Warn("admin notice handler wiring failed", zap.Error(err))
	} else {
		h.Register(adminG.Group("", middleware.SessionAuth(sessStore, a.Clock), middleware.RoleGate(2)))
	}
}

// wireSettingHandler 挂载系统设置 handler(SessionAuth + RoleGate(2),/api/admin/settings)。
// mailer 暂传 nil:test_smtp 走 stub,待邮件子系统接入后注入真实 Mailer。
func wireSettingHandler(a *app.Application, adminG *gin.RouterGroup, log *zap.Logger) {
	sessStore := authhwire.SessionStoreFrom(a)
	if sessStore == nil {
		log.Warn("session store not available; admin setting routes skipped")
		return
	}
	actorOf := func(c *gin.Context) int64 { return middleware.UserID(c) }
	h, err := admin.WireAdminSetting(a, nil, actorOf)
	if err != nil {
		log.Warn("admin setting handler wiring failed", zap.Error(err))
		return
	}
	h.Register(adminG.Group("", middleware.SessionAuth(sessStore, a.Clock), middleware.RoleGate(2)))
}

// wireAccountHandler 挂载号池 admin 路由。
//
// 路由组先套 SessionAuth + RoleGate(2)(admin);
// /accounts/:id/credentials/peek 额外套 RoleGate(3)(superadmin),
// 因为该 endpoint 返回明文凭证。
func wireAccountHandler(a *app.Application, adminG *gin.RouterGroup, log *zap.Logger) {
	if a == nil {
		return
	}
	sessStore := authhwire.SessionStoreFrom(a)
	if sessStore == nil {
		log.Warn("session store not available; admin account routes skipped")
		return
	}
	actorOf := func(c *gin.Context) int64 { return middleware.UserID(c) }
	h, err := admin.WireAdminAccount(a, actorOf)
	if err != nil {
		log.Warn("admin account handler wiring failed", zap.Error(err))
		return
	}
	// adminGroup 已带 ErrorResponse("json");再加 SessionAuth + RoleGate(2)。
	accG := adminG.Group("",
		middleware.SessionAuth(sessStore, a.Clock),
		middleware.RoleGate(2),
	)
	// OAuth callback 不带 session,依赖一次性 state 校验。
	adminG.GET("/accounts/oauth/callback", h.OAuthCallback)
	// 普通 admin 可访问的 endpoint。
	accG.GET("/accounts", h.List)
	accG.GET("/accounts/stats/overview", h.Stats)
	accG.POST("/accounts/oauth/start", h.OAuthStart)
	accG.GET("/accounts/:id", h.Get)
	accG.POST("/accounts", h.Create)
	accG.POST("/accounts/import", h.Import)
	accG.PATCH("/accounts/:id", h.Patch)
	accG.DELETE("/accounts/:id", h.Delete)
	accG.POST("/accounts/:id/enable", h.Enable)
	accG.POST("/accounts/:id/disable", h.Disable)
	accG.POST("/accounts/:id/test", h.Test)
	accG.POST("/accounts/:id/refresh_token", h.RefreshToken)
	accG.POST("/accounts/:id/clear_cooldown", h.ClearCooldown)
	accG.GET("/accounts/:id/events", h.Events)
	// 唯一需要 superadmin 的 endpoint。
	superG := accG.Group("", middleware.RoleGate(3))
	superG.GET("/accounts/:id/credentials/peek", h.PeekCredentials)
}

// wireLogHandlers 挂载日志 handler。
func wireLogHandlers(a *app.Application, adminG, userG *gin.RouterGroup) {
	logStore := ilog.StoreFrom(a)
	if logStore == nil {
		return
	}
	adminH := logh.NewAdmin(logStore, nil, nil)
	adminH.Register(adminG.Group("/logs"))

	userH := logh.NewUser(logStore)
	userH.Register(userG.Group("/logs"))
}

// wirePaymentRoutes 挂载在线支付 handler。
//
// 用户侧路由需要 SessionAuth;Webhook 路由公开(无 auth)。
func wirePaymentRoutes(eng *gin.Engine, a *app.Application, log *zap.Logger) {
	h := payment.HolderFrom(a.PaymentSvc)
	onlineSvc, _ := h.Online.(*monline.Service)
	if onlineSvc == nil {
		log.Warn("online payment service not wired; payment routes skipped")
		return
	}

	payH := paymenthdr.New(paymenthdr.Deps{
		Online: onlineSvc,
		Log:    log,
	})

	// 用户侧:需要 SessionAuth
	sessStore := authhwire.SessionStoreFrom(a)
	if sessStore != nil {
		userPayG := eng.Group("/api/user/payment",
			middleware.ErrorResponse("json"),
			middleware.SessionAuth(sessStore, a.Clock),
		)
		userPayG.POST("/create", payH.CreateOrder)
		userPayG.GET("/orders", payH.ListOrders)
	} else {
		log.Warn("session store not available; user payment routes skipped")
	}

	// Webhook:公开,无需认证
	payWebhookG := eng.Group("/api/pay", middleware.ErrorResponse("json"))
	payWebhookG.POST("/webhook/:provider", payH.Webhook)
}

// billerFrom extracts billing.Biller from app.Biller.
func billerFrom(a *app.Application) billing.Biller {
	b, _ := a.Biller.(billing.Biller)
	return b
}

// wireInviteRoutes mounts GET /api/user/invites/{me,invitees,records} — all require SessionAuth.
func wireInviteRoutes(eng *gin.Engine, a *app.Application, log *zap.Logger) {
	sessStore := authhwire.SessionStoreFrom(a)
	if sessStore == nil {
		log.Warn("session store not available; invite routes skipped")
		return
	}
	userOf := func(c *gin.Context) int64 { return middleware.UserID(c) }
	h, err := userhdr.WireInvite(a, userOf)
	if err != nil {
		log.Warn("invite handler wiring failed", zap.Error(err))
		return
	}
	inviteG := eng.Group("/api/user/invites",
		middleware.ErrorResponse("json"),
		middleware.SessionAuth(sessStore, a.Clock),
	)
	h.Register(inviteG)
}

// wireApikeyRoutes mounts /api/user/apikeys/* and /api/admin/apikeys/*.
func wireApikeyRoutes(eng *gin.Engine, a *app.Application, log *zap.Logger) {
	ts, ok := a.TokenStore.(token.Store)
	if !ok || ts == nil {
		log.Warn("token store not wired; apikey routes skipped")
		return
	}
	sessStore := authhwire.SessionStoreFrom(a)
	if sessStore == nil {
		log.Warn("session store not available; apikey routes skipped")
		return
	}

	// 用户侧:SessionAuth + CSRF(CSRF 在 authhttp 注册的 userG 上已挂;此处独立 group 需自带)
	userOf := func(c *gin.Context) int64 { return middleware.UserID(c) }
	userH := tokenhdr.NewUserHandler(ts, userOf)
	userApikeyG := eng.Group("/api/user/apikeys",
		middleware.ErrorResponse("json"),
		middleware.SessionAuth(sessStore, a.Clock),
	)
	userH.Register(userApikeyG)

	// 管理员侧:SessionAuth + RoleGate(2)
	adminH := tokenhdr.NewAdminHandler(ts)
	adminApikeyG := eng.Group("/api/admin/apikeys",
		middleware.ErrorResponse("json"),
		middleware.SessionAuth(sessStore, a.Clock),
		middleware.RoleGate(2),
	)
	adminH.Register(adminApikeyG)
}

// wirePublicGroupsRoute mounts GET /api/public/groups — returns active groups for self-selection.
func wirePublicGroupsRoute(publicG *gin.RouterGroup, a *app.Application) {
	gs, ok := a.GroupSvc.(group.Service)
	if !ok || gs == nil {
		return
	}
	publicG.GET("/groups", func(c *gin.Context) {
		groups, err := gs.List(c.Request.Context())
		if err != nil {
			middleware.SetErr(c, apierr.Wrap(apierr.CodeInternal, "list groups", err))
			return
		}
		type groupItem struct {
			ID          int64   `json:"id"`
			Name        string  `json:"name"`
			DisplayName string  `json:"display_name"`
			Ratio       float64 `json:"ratio"`
		}
		items := make([]groupItem, 0, len(groups))
		for _, g := range groups {
			if g.Status != 0 {
				continue
			}
			items = append(items, groupItem{
				ID:          g.ID,
				Name:        g.Name,
				DisplayName: g.DisplayName,
				Ratio:       g.Ratio,
			})
		}
		c.JSON(http.StatusOK, gin.H{"items": items})
	})
}
