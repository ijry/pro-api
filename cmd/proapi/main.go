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

	"github.com/ijry/pro-api/internal/app"
	"github.com/ijry/pro-api/internal/app/config"
	"github.com/ijry/pro-api/internal/observability/logger"
	"github.com/ijry/pro-api/internal/observability/metrics"
	"github.com/ijry/pro-api/internal/server"
	"github.com/ijry/pro-api/internal/version"
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
	application, err := app.SetupBasic(ctx, cfg, log)
	if err != nil {
		return fmt.Errorf("setup basic: %w", err)
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
