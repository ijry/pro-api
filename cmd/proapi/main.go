// proapi 主程序入口。
package main

import (
	"context"
	"errors"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ijry/pro-api/internal/app"
	"github.com/ijry/pro-api/internal/app/config"
	"github.com/ijry/pro-api/internal/observability/logger"
	"github.com/ijry/pro-api/internal/observability/metrics"
	"github.com/ijry/pro-api/internal/server"
	"github.com/ijry/pro-api/internal/version"
	"go.uber.org/zap"
)

func main() {
	configPath := flag.String("config", "", "配置文件路径(可选,留空则仅用 ENV 与默认值)")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		_, _ = os.Stderr.WriteString("config error: " + err.Error() + "\n")
		os.Exit(1)
	}

	log, err := logger.New(cfg.Log.Level, cfg.Log.Format, nil)
	if err != nil {
		_, _ = os.Stderr.WriteString("logger error: " + err.Error() + "\n")
		os.Exit(1)
	}
	defer func() { _ = log.Sync() }()
	log.Info("proapi starting",
		zap.String("version", version.String()),
		zap.String("addr", cfg.Server.Addr),
		zap.Int("node_id", cfg.NodeID),
	)

	metrics.Init()

	ctx := context.Background()
	application, err := app.SetupBasic(ctx, cfg, log)
	if err != nil {
		log.Fatal("setup basic failed", zap.Error(err))
	}

	// 业务层 Wire(M1-02 ~ M1-13 各自 spec 实施时在这里加):
	// app.WireAuth(application)
	// app.WireToken(application)
	// ...

	engine := server.NewEngine(log)
	srv := &http.Server{
		Addr:         cfg.Server.Addr,
		Handler:      engine,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeoutMS) * time.Millisecond,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeoutMS) * time.Millisecond,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("http server error", zap.Error(err))
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Info("shutting down")
	shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		log.Error("http server shutdown failed", zap.Error(err))
	}
	if err := application.Shutdown(shutCtx); err != nil {
		log.Error("application shutdown reported errors", zap.Error(err))
	}
}
