// Package admin 集中管理员侧 HTTP handler。
package admin

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ijry/pro-api/internal/notice"
	"github.com/ijry/pro-api/internal/server/middleware"
	"github.com/ijry/pro-api/pkg/apierr"
)

// NoticeHandler 是管理员公告 handler。
//
// 依赖:
//   - svc:notice.Service 业务编排
//   - actorOf:从 gin.Context 解出操作者 user id;若 nil 则用 0(测试或裸接入)
type NoticeHandler struct {
	Svc     notice.Service
	ActorOf func(c *gin.Context) int64
}

// NewNoticeHandler 构造 NoticeHandler。
func NewNoticeHandler(svc notice.Service, actorOf func(c *gin.Context) int64) *NoticeHandler {
	if actorOf == nil {
		actorOf = func(*gin.Context) int64 { return 0 }
	}
	return &NoticeHandler{Svc: svc, ActorOf: actorOf}
}

// Register 把 7 个 admin 路由挂到 r(预期已加 SessionAuth + RoleGate(admin))。
func (h *NoticeHandler) Register(r gin.IRouter) {
	r.POST("/notices", h.Create)
	r.GET("/notices", h.List)
	r.GET("/notices/:id", h.Get)
	r.PATCH("/notices/:id", h.Update)
	r.DELETE("/notices/:id", h.Delete)
	r.POST("/notices/:id/publish", h.Publish)
	r.POST("/notices/:id/unpublish", h.Unpublish)
}

// createReq 是 POST /notices 的 body。time 字段用 *time.Time 接 null。
type createReq struct {
	Title     string     `json:"title"`
	Content   string     `json:"content"`
	Level     string     `json:"level"`
	Target    string     `json:"target"`
	Pinned    bool       `json:"pinned"`
	PublishAt *time.Time `json:"publish_at"`
	ExpiresAt *time.Time `json:"expires_at"`
}

// Create POST /notices
func (h *NoticeHandler) Create(c *gin.Context) {
	var req createReq
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.SetErr(c, apierr.New(apierr.CodeInvalidParam, "请求体不合法"))
		return
	}
	in := notice.CreateInput{
		Title:     req.Title,
		Content:   req.Content,
		Level:     req.Level,
		Target:    req.Target,
		Pinned:    req.Pinned,
		PublishAt: req.PublishAt,
		ExpiresAt: req.ExpiresAt,
	}
	n, err := h.Svc.Create(c.Request.Context(), in, h.ActorOf(c))
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": n})
}

// List GET /notices?status=&page=&size=
func (h *NoticeHandler) List(c *gin.Context) {
	status := int8(-1)
	if s := c.Query("status"); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil {
			middleware.SetErr(c, apierr.New(apierr.CodeInvalidParam, "status 必须为整数"))
			return
		}
		status = int8(n)
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	items, total, err := h.Svc.List(c.Request.Context(), notice.ListFilter{Status: status, Page: page, Size: size})
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
	id, ok := parseID(c)
	if !ok {
		return
	}
	n, err := h.Svc.Get(c.Request.Context(), id)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": n})
}

// Update PATCH /notices/:id;接受 map[string]json.RawMessage 以支持 explicit null。
func (h *NoticeHandler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	raw := map[string]json.RawMessage{}
	if err := c.ShouldBindJSON(&raw); err != nil {
		middleware.SetErr(c, apierr.New(apierr.CodeInvalidParam, "请求体不合法"))
		return
	}
	if _, has := raw["status"]; has {
		middleware.SetErr(c, apierr.New(apierr.CodeInvalidParam, "禁止通过 PATCH 修改 status,请用 /publish 或 /unpublish"))
		return
	}
	patch, perr := parseUpdatePatch(raw)
	if perr != nil {
		middleware.SetErr(c, perr)
		return
	}
	n, err := h.Svc.Update(c.Request.Context(), id, patch, h.ActorOf(c))
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": n})
}

// Delete DELETE /notices/:id
func (h *NoticeHandler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.Svc.Delete(c.Request.Context(), id, h.ActorOf(c)); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// Publish POST /notices/:id/publish
func (h *NoticeHandler) Publish(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	n, err := h.Svc.Publish(c.Request.Context(), id, h.ActorOf(c))
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": n})
}

// Unpublish POST /notices/:id/unpublish
func (h *NoticeHandler) Unpublish(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	n, err := h.Svc.Unpublish(c.Request.Context(), id, h.ActorOf(c))
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": n})
}

// --- helpers ---

// parseID 解 :id 路径参数;失败时已 SetErr(并返回 ok=false)。
func parseID(c *gin.Context) (int64, bool) {
	s := c.Param("id")
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		middleware.SetErr(c, apierr.New(apierr.CodeInvalidParam, "id 必须为整数"))
		return 0, false
	}
	return id, true
}

// writeErr 把 service 返回的 error 转 apierr 写到 ctx;非 apierr 包内部错则视作 CodeInternal。
func writeErr(c *gin.Context, err error) {
	var ae *apierr.Error
	if asAPIErr(err, &ae) {
		middleware.SetErr(c, ae)
		return
	}
	middleware.SetErr(c, apierr.Wrap(apierr.CodeInternal, "internal error", err))
}

// asAPIErr 类似 errors.As,把 err 解到 *apierr.Error。
func asAPIErr(err error, dst **apierr.Error) bool {
	if err == nil {
		return false
	}
	if ae, ok := err.(*apierr.Error); ok {
		*dst = ae
		return true
	}
	type unwrapper interface{ Unwrap() error }
	for cur := err; cur != nil; {
		u, ok := cur.(unwrapper)
		if !ok {
			break
		}
		next := u.Unwrap()
		if ae, ok := next.(*apierr.Error); ok {
			*dst = ae
			return true
		}
		cur = next
	}
	return false
}

// parseUpdatePatch 将 raw body 转 notice.UpdatePatch;支持 publish_at/expires_at 显式 null。
func parseUpdatePatch(raw map[string]json.RawMessage) (notice.UpdatePatch, *apierr.Error) {
	var patch notice.UpdatePatch
	getStrPtr := func(key string) (*string, *apierr.Error) {
		v, ok := raw[key]
		if !ok {
			return nil, nil
		}
		if string(v) == "null" {
			return nil, apierr.New(apierr.CodeInvalidParam, key+" 不允许为 null")
		}
		var s string
		if err := json.Unmarshal(v, &s); err != nil {
			return nil, apierr.New(apierr.CodeInvalidParam, key+" 必须为字符串")
		}
		return &s, nil
	}
	if p, e := getStrPtr("title"); e != nil {
		return patch, e
	} else {
		patch.Title = p
	}
	if p, e := getStrPtr("content"); e != nil {
		return patch, e
	} else {
		patch.Content = p
	}
	if p, e := getStrPtr("level"); e != nil {
		return patch, e
	} else {
		patch.Level = p
	}
	if p, e := getStrPtr("target"); e != nil {
		return patch, e
	} else {
		patch.Target = p
	}
	if v, ok := raw["pinned"]; ok {
		if string(v) == "null" {
			return patch, apierr.New(apierr.CodeInvalidParam, "pinned 不允许为 null")
		}
		var b bool
		if err := json.Unmarshal(v, &b); err != nil {
			return patch, apierr.New(apierr.CodeInvalidParam, "pinned 必须为布尔值")
		}
		patch.Pinned = &b
	}
	parseTime := func(key string, setVal func(*time.Time), setNull func()) *apierr.Error {
		v, ok := raw[key]
		if !ok {
			return nil
		}
		if string(v) == "null" {
			setNull()
			return nil
		}
		var t time.Time
		if err := json.Unmarshal(v, &t); err != nil {
			return apierr.New(apierr.CodeInvalidParam, key+" 必须是 RFC3339 时间或 null")
		}
		setVal(&t)
		return nil
	}
	if e := parseTime("publish_at",
		func(t *time.Time) { patch.PublishAt = t },
		func() { patch.PublishAtNull = true }); e != nil {
		return patch, e
	}
	if e := parseTime("expires_at",
		func(t *time.Time) { patch.ExpiresAt = t },
		func() { patch.ExpiresAtNull = true }); e != nil {
		return patch, e
	}
	return patch, nil
}
