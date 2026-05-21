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

	"github.com/proapi/proapi/internal/app/config"
	"github.com/proapi/proapi/internal/observability/logger"
	"github.com/proapi/proapi/internal/observability/metrics"
	"github.com/proapi/proapi/internal/server"
	"github.com/proapi/proapi/internal/version"
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

	// M0 只起 HTTP server,DB / Redis 在 M1 起业务模块时再接入并存到 app context。
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("graceful shutdown failed", zap.Error(err))
	}
}
