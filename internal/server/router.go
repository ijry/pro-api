// Package server 装配 Gin 引擎与路由。
package server

import (
	"github.com/gin-gonic/gin"
	"github.com/ijry/pro-api/internal/observability/metrics"
	"github.com/ijry/pro-api/internal/server/handler"
	"github.com/ijry/pro-api/internal/server/middleware"
	"go.uber.org/zap"
)

// NewEngine 构造可启动的 Gin 引擎。本 spec 完成的中间件:
//
//	RequestID → AccessLog → Recovery → /healthz → /metrics
//
// 业务路由在 M1-02 ~ M1-13 各自 Wire 函数里 register。
func NewEngine(log *zap.Logger) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(
		middleware.RequestID(),
		middleware.AccessLog(log),
		middleware.Recovery(log),
	)

	r.GET("/healthz", handler.Health)
	r.GET("/metrics", gin.WrapH(metrics.HTTPHandler()))
	return r
}
