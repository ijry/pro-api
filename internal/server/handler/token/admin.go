package token

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ijry/pro-api/internal/server/middleware"
	tokensvc "github.com/ijry/pro-api/internal/token"
	"github.com/ijry/pro-api/pkg/apierr"
)

// AdminHandler 提供 /api/admin/apikeys/* 路由。
//
// 与 UserHandler 区别:
//
//   - List 接受 ?user_id= 过滤,user_id=0 / 不传 时列全部
//   - Patch 不限制"必须是自己的 token"
//   - Delete 直接 Revoke 任意 token
//   - 没有 regenerate 端点(spec §11.10 决策)
//
// 调用方必须在 group 上挂 SessionAuth + RoleGate(>= 2)。
type AdminHandler struct {
	store tokensvc.Store
}

// NewAdminHandler 构造。
func NewAdminHandler(store tokensvc.Store) *AdminHandler {
	return &AdminHandler{store: store}
}

// Register 把路由挂到给定 group。
func (h *AdminHandler) Register(g *gin.RouterGroup) {
	g.GET("", h.List)
	g.GET("/:id", h.Detail)
	g.PATCH("/:id", h.Patch)
	g.DELETE("/:id", h.Delete)
}

// List GET /api/admin/apikeys
func (h *AdminHandler) List(c *gin.Context) {
	page, size := parsePagination(c)
	filter := tokensvc.ListFilter{
		Keyword: c.Query("keyword"),
		Page:    page,
		Size:    size,
	}
	if uid := c.Query("user_id"); uid != "" {
		if n, err := strconv.ParseInt(uid, 10, 64); err == nil && n > 0 {
			filter.UserID = n
		}
	}
	if st, ok := parseStatusQuery(c); ok {
		filter.Status = &st
	}
	items, total, err := h.store.List(c.Request.Context(), filter)
	if err != nil {
		middleware.SetErr(c, toAPIErr(err, apierr.CodeInternal))
		return
	}
	dtos := make([]ViewDTO, len(items))
	for i, v := range items {
		dtos[i] = toDTO(v)
	}
	c.JSON(http.StatusOK, ListResponse{Items: dtos, Total: total, Page: page, Size: size})
}

// Detail GET /api/admin/apikeys/:id
func (h *AdminHandler) Detail(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	view, err := h.store.Get(c.Request.Context(), id)
	if err != nil {
		middleware.SetErr(c, toAPIErr(err, apierr.CodeInternal))
		return
	}
	c.JSON(http.StatusOK, toDTO(view))
}

// Patch PATCH /api/admin/apikeys/:id
func (h *AdminHandler) Patch(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req PatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.SetErr(c, apierr.New(apierr.CodeInvalidParam, err.Error()))
		return
	}
	if err := validatePatch(&req); err != nil {
		middleware.SetErr(c, err)
		return
	}
	updated, err := h.store.Update(c.Request.Context(), id, req.toUpdatePatch())
	if err != nil {
		middleware.SetErr(c, toAPIErr(err, apierr.CodeInternal))
		return
	}
	c.JSON(http.StatusOK, toDTO(updated))
}

// Delete DELETE /api/admin/apikeys/:id
func (h *AdminHandler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.store.Revoke(c.Request.Context(), id); err != nil {
		middleware.SetErr(c, toAPIErr(err, apierr.CodeInternal))
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "revoked"})
}
