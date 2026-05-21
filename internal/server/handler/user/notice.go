// Package user 集中用户侧 HTTP handler。
package user

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ijry/pro-api/internal/notice"
	"github.com/ijry/pro-api/internal/server/middleware"
	"github.com/ijry/pro-api/pkg/apierr"
)

// NoticeHandler 是用户公告 handler。
//
// 依赖:
//   - svc:notice.Service 业务编排
//   - userOf:从 gin.Context 解出用户 id;若 nil 或返回 0,handler 返回 CodeNotLoggedIn
type NoticeHandler struct {
	Svc    notice.Service
	UserOf func(c *gin.Context) int64
}

// NewNoticeHandler 构造。
func NewNoticeHandler(svc notice.Service, userOf func(c *gin.Context) int64) *NoticeHandler {
	if userOf == nil {
		userOf = func(*gin.Context) int64 { return 0 }
	}
	return &NoticeHandler{Svc: svc, UserOf: userOf}
}

// Register 把 4 个 user 路由挂到 r(预期已加 SessionAuth)。
func (h *NoticeHandler) Register(r gin.IRouter) {
	r.GET("/notices", h.List)
	r.GET("/notices/unread_count", h.UnreadCount)
	r.GET("/notices/:id", h.Get)
	r.POST("/notices/:id/read", h.Read)
}

// requireUser 若未登录返回 CodeNotLoggedIn,否则返用户 id。
func (h *NoticeHandler) requireUser(c *gin.Context) (int64, bool) {
	uid := h.UserOf(c)
	if uid <= 0 {
		middleware.SetErr(c, apierr.New(apierr.CodeNotLoggedIn, "请先登录"))
		return 0, false
	}
	return uid, true
}

// List GET /notices?page=&size=
func (h *NoticeHandler) List(c *gin.Context) {
	uid, ok := h.requireUser(c)
	if !ok {
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	items, total, err := h.Svc.ListForUser(c.Request.Context(), uid, page, size)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{
		"items": items,
		"total": total,
		"page":  page,
		"size":  size,
	}})
}

// Get GET /notices/:id
func (h *NoticeHandler) Get(c *gin.Context) {
	uid, ok := h.requireUser(c)
	if !ok {
		return
	}
	id, ok2 := parseID(c)
	if !ok2 {
		return
	}
	n, err := h.Svc.GetForUser(c.Request.Context(), uid, id)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": n})
}

// Read POST /notices/:id/read
func (h *NoticeHandler) Read(c *gin.Context) {
	uid, ok := h.requireUser(c)
	if !ok {
		return
	}
	id, ok2 := parseID(c)
	if !ok2 {
		return
	}
	if err := h.Svc.MarkRead(c.Request.Context(), uid, id); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// UnreadCount GET /notices/unread_count
func (h *NoticeHandler) UnreadCount(c *gin.Context) {
	uid, ok := h.requireUser(c)
	if !ok {
		return
	}
	n := h.Svc.UnreadCountForUser(c.Request.Context(), uid)
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"unread_count": n}})
}

// --- helpers ---

func parseID(c *gin.Context) (int64, bool) {
	s := c.Param("id")
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		middleware.SetErr(c, apierr.New(apierr.CodeInvalidParam, "id 必须为整数"))
		return 0, false
	}
	return id, true
}

func writeErr(c *gin.Context, err error) {
	var ae *apierr.Error
	if errors.As(err, &ae) {
		middleware.SetErr(c, ae)
		return
	}
	middleware.SetErr(c, apierr.Wrap(apierr.CodeInternal, "internal error", err))
}
