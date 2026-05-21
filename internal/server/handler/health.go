// Package handler 集中 HTTP handler。
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ijry/pro-api/internal/version"
)

// Health 返回 200 + 服务版本。M0 阶段不做依赖检查;M1 起补 DB/Redis 探活。
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"version": version.String(),
	})
}
