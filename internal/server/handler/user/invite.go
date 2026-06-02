package user

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ijry/pro-api/internal/invite"
	"github.com/ijry/pro-api/internal/server/middleware"
	"github.com/ijry/pro-api/pkg/apierr"
)

// InviteHandler serves the user-side invite endpoints.
type InviteHandler struct {
	Svc    *invite.Service
	UserOf func(c *gin.Context) int64
}

// NewInviteHandler constructs an InviteHandler.
func NewInviteHandler(svc *invite.Service, userOf func(c *gin.Context) int64) *InviteHandler {
	if userOf == nil {
		userOf = func(*gin.Context) int64 { return 0 }
	}
	return &InviteHandler{Svc: svc, UserOf: userOf}
}

// Register mounts the 3 invite routes onto r.
func (h *InviteHandler) Register(r gin.IRouter) {
	r.GET("/me", h.Me)
	r.GET("/invitees", h.Invitees)
	r.GET("/records", h.Records)
}

func (h *InviteHandler) requireUser(c *gin.Context) (int64, bool) {
	uid := h.UserOf(c)
	if uid <= 0 {
		middleware.SetErr(c, apierr.New(apierr.CodeNotLoggedIn, "请先登录"))
		return 0, false
	}
	return uid, true
}

// Me GET /invites/me
func (h *InviteHandler) Me(c *gin.Context) {
	uid, ok := h.requireUser(c)
	if !ok {
		return
	}
	resp, err := h.Svc.GetSummary(c.Request.Context(), uid)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// Invitees GET /invites/invitees?page=&size=
func (h *InviteHandler) Invitees(c *gin.Context) {
	uid, ok := h.requireUser(c)
	if !ok {
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	items, total, err := h.Svc.ListInvitees(c.Request.Context(), uid, page, size)
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

// Records GET /invites/records?page=&size=
func (h *InviteHandler) Records(c *gin.Context) {
	uid, ok := h.requireUser(c)
	if !ok {
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	items, total, err := h.Svc.ListRecords(c.Request.Context(), uid, page, size)
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
