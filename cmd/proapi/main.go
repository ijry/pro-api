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
	"github.com/ijry/pro-api/internal/adapter"
	"github.com/ijry/pro-api/internal/adapterreg"
	"github.com/ijry/pro-api/internal/app"
	"github.com/ijry/pro-api/internal/app/config"
	authhwire "github.com/ijry/pro-api/internal/auth/wire"
	"github.com/ijry/pro-api/internal/billing"
	"github.com/ijry/pro-api/internal/channel"
	ilog "github.com/ijry/pro-api/internal/log"
	"github.com/ijry/pro-api/internal/notice"
	"github.com/ijry/pro-api/internal/observability/logger"
	"github.com/ijry/pro-api/internal/observability/metrics"
	mmanual "github.com/ijry/pro-api/internal/payment/manual"
	monline "github.com/ijry/pro-api/internal/payment/online"
	"github.com/ijry/pro-api/internal/payment/redeem"
	"github.com/ijry/pro-api/internal/payment"
	"github.com/ijry/pro-api/internal/ratelimit"
	"github.com/ijry/pro-api/internal/relay"
	"github.com/ijry/pro-api/internal/server"
	relayhdr "github.com/ijry/pro-api/internal/server/handler/relay"
	paymenthdr "github.com/ijry/pro-api/internal/server/handler/payment"
	"github.com/ijry/pro-api/internal/server/handler/logh"
	"github.com/ijry/pro-api/internal/server/middleware"
	"github.com/ijry/pro-api/internal/token"
	"github.com/ijry/pro-api/internal/version"
	"github.com/ijry/pro-api/internal/wallet"
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

	// ── 适配器 & Relay ───────────────────────────────────────────
	adapterReg := adapter.NewRegistry()
	adapterreg.WireAdapters(adapterReg, application.Tokenize)
	application.AdapterReg = adapterReg

	relaySvc := relay.New(adapterReg)
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
	if _, err := wireNoticeHandlers(a, adminGroup, userGroup, publicGroup); err != nil {
		log.Warn("notice handlers failed", zap.Error(err))
	}

	// 系统设置路由(M1-12)
	wireSettingHandler(a, adminGroup)

	// 日志 & 统计路由(M1-08)
	wireLogHandlers(a, adminGroup, userGroup)

	// 在线支付路由(M2a B4)
	wirePaymentRoutes(eng, a, log)

	// 代理路由(/v1/*) — TokenAuth 保护(M1-03/04)
	v1 := eng.Group("/v1", middleware.ErrorResponse("openai"))
	// TokenAuth 由 M1-03 提供;M1 先留 placeholder,实际路由由 relay handler 挂
	if relaySvc, ok := a.Relay.(*relay.Service); ok {
		relayH := relayhdr.New(relaySvc)
		relayH.Register(v1)
		// M2a: Anthropic 入口 /v1/messages
		relayH.RegisterAnthropicRoutes(v1)
		// M2a: Gemini 入口 /v1beta/models/:model/...
		v1beta := eng.Group("/v1beta", middleware.ErrorResponse("json"))
		relayH.RegisterGeminiRoutes(v1beta)
	}

	_ = ctx
	return nil
}

// wireNoticeHandlers 挂载公告 handler。
func wireNoticeHandlers(a *app.Application, adminG, userG, publicG *gin.RouterGroup) (bool, error) {
	// 注意:这些 Wire 函数在 M1-12 notice handler 中;重复调用不影响 NoticeSvc
	// 只是注册 handler 到 router group
	_ = a
	_ = adminG
	_ = userG
	_ = publicG
	// TODO: M1-12 handler 在 internal/server/handler/{admin,user,public}/
	// 路由已在 authhwire.RegisterRoutes 或各自 wire 内注册,这里保持空实现
	return true, nil
}

// wireSettingHandler 挂载系统设置 handler。
func wireSettingHandler(a *app.Application, adminG *gin.RouterGroup) {
	_ = a
	_ = adminG
	// TODO: internal/server/handler/admin.WireAdminSetting
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
