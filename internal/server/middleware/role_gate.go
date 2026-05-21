package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/ijry/pro-api/pkg/apierr"
)

// RoleGate 校验当前 ctx 的 role 必须 >= minRole。
func RoleGate(minRole int8) gin.HandlerFunc {
	return func(c *gin.Context) {
		if Role(c) < minRole {
			SetErr(c, apierr.New(apierr.CodeForbidden, "权限不足"))
			return
		}
		c.Next()
	}
}
