// Package server 装配 Gin 引擎与路由。
package server

import (
	"github.com/gin-gonic/gin"
	"github.com/proapi/proapi/internal/server/handler"
	"github.com/proapi/proapi/internal/server/middleware"
	"go.uber.org/zap"
)

// NewEngine 构造可启动的 Gin 引擎。后续业务路由在 M1 起按模块注册。
func NewEngine(log *zap.Logger) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(middleware.RequestID(), middleware.Recovery(log))

	r.GET("/healthz", handler.Health)
	// /metrics 由 T12 注册
	return r
}
