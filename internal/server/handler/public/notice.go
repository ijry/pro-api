// Package public 集中公开(无鉴权)HTTP handler。
package public

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ijry/pro-api/internal/notice"
	"github.com/ijry/pro-api/internal/server/middleware"
	"github.com/ijry/pro-api/pkg/apierr"
)

// NoticeHandler 是公开公告 handler。
type NoticeHandler struct {
	Svc notice.Service
}

// NewNoticeHandler 构造。
func NewNoticeHandler(svc notice.Service) *NoticeHandler {
	return &NoticeHandler{Svc: svc}
}

// Register 把 1 个公开路由挂到 r。
func (h *NoticeHandler) Register(r gin.IRouter) {
	r.GET("/notices", h.List)
}

// List GET /notices?page=&size= — target=all 且 status=Published 且未过期。
func (h *NoticeHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	items, total, err := h.Svc.ListPublic(c.Request.Context(), page, size)
	if err != nil {
		var ae *apierr.Error
		if errors.As(err, &ae) {
			middleware.SetErr(c, ae)
			return
		}
		middleware.SetErr(c, apierr.Wrap(apierr.CodeInternal, "internal error", err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{
		"items": items,
		"total": total,
		"page":  page,
		"size":  size,
	}})
}
