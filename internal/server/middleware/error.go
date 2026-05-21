package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/proapi/proapi/pkg/apierr"
)

const errCtxKey = "proapi:err"

// SetErr 由 handler 调用,在 Gin context 中标记错误并 abort。
func SetErr(c *gin.Context, e *apierr.Error) {
	c.Set(errCtxKey, e)
	c.Abort()
}

// ErrorResponse 是 trailing 中间件,在 handler return 后渲染错误响应。
//
//	format == "openai" → 走 OpenAI 错误协议 {"error": {message, type, code}}
//	format == "json"   → 走管理 API 格式 {code, message, details}
func ErrorResponse(format string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		v, ok := c.Get(errCtxKey)
		if !ok {
			return
		}
		e, ok := v.(*apierr.Error)
		if !ok {
			return
		}
		switch format {
		case "openai":
			c.JSON(e.HTTPStatus, gin.H{
				"error": gin.H{
					"message": e.Message,
					"type":    openAIType(e.Code),
					"code":    openAICode(e.Code),
				},
			})
		default:
			c.JSON(e.HTTPStatus, gin.H{
				"code":    int(e.Code),
				"message": e.Message,
				"details": e.Details,
			})
		}
	}
}
