package token

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ijry/pro-api/internal/server/middleware"
	tokensvc "github.com/ijry/pro-api/internal/token"
	"github.com/ijry/pro-api/pkg/apierr"
)

// SessionUserExtractor 从请求中抽取当前登录用户 id。
// 由调用方(handler 注册时)注入,通常由 M1-02 SessionAuth 中间件填充 ctx 后,这里读取。
//
// 返回 0 视为"未登录" → CodeNotLoggedIn(401)。
type SessionUserExtractor func(c *gin.Context) int64

// UserHandler 提供 /api/user/tokens/* 路由的 handler。
type UserHandler struct {
	store    tokensvc.Store
	userOf   SessionUserExtractor
	maxRPM   int // 0 = 不强制上限
	maxTPM   int // 0 = 不强制上限
}

// NewUserHandler 构造。userOf 必填。
//
//   - maxRPM / maxTPM > 0 时,Create / Update 中 RPM/TPM 超过会被拒(spec 留接口)
//   - = 0 视为不限
func NewUserHandler(store tokensvc.Store, userOf SessionUserExtractor) *UserHandler {
	return &UserHandler{store: store, userOf: userOf}
}

// Register 把路由挂到给定 group(已应用 SessionAuth)。
func (h *UserHandler) Register(g *gin.RouterGroup) {
	g.GET("", h.List)
	g.POST("", h.Create)
	g.GET("/:id", h.Detail)
	g.PATCH("/:id", h.Patch)
	g.DELETE("/:id", h.Delete)
	g.POST("/:id/regenerate", h.Regenerate)
}

// === handlers ===

// List GET /api/user/tokens
func (h *UserHandler) List(c *gin.Context) {
	uid := h.mustUser(c)
	if uid == 0 {
		return
	}
	page, size := parsePagination(c)
	filter := tokensvc.ListFilter{
		UserID:  uid,
		Keyword: c.Query("keyword"),
		Page:    page,
		Size:    size,
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

// Create POST /api/user/tokens
func (h *UserHandler) Create(c *gin.Context) {
	uid := h.mustUser(c)
	if uid == 0 {
		return
	}
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.SetErr(c, apierr.New(apierr.CodeInvalidParam, err.Error()))
		return
	}
	if err := validateCreate(&req); err != nil {
		middleware.SetErr(c, err)
		return
	}
	in := tokensvc.CreateInput{
		UserID:        uid,
		Name:          req.Name,
		QuotaLimit:    req.QuotaLimit,
		AllowedModels: req.AllowedModels,
		AllowedIPs:    req.AllowedIPs,
		RPMLimit:      req.RPMLimit,
		TPMLimit:      req.TPMLimit,
		ExpiresAt:     req.ExpiresAt,
	}
	plaintext, view, err := h.store.Create(c.Request.Context(), in)
	if err != nil {
		middleware.SetErr(c, toAPIErr(err, apierr.CodeInternal))
		return
	}
	c.JSON(http.StatusOK, CreateResponse{
		View:         toDTO(view),
		PlaintextKey: plaintext,
		Warning:      plaintextWarning,
	})
}

// Detail GET /api/user/tokens/:id
func (h *UserHandler) Detail(c *gin.Context) {
	uid := h.mustUser(c)
	if uid == 0 {
		return
	}
	id, ok := parseID(c)
	if !ok {
		return
	}
	view, err := h.fetchOwnToken(c.Request.Context(), id, uid)
	if err != nil {
		middleware.SetErr(c, toAPIErr(err, apierr.CodeInternal))
		return
	}
	c.JSON(http.StatusOK, toDTO(view))
}

// Patch PATCH /api/user/tokens/:id
func (h *UserHandler) Patch(c *gin.Context) {
	uid := h.mustUser(c)
	if uid == 0 {
		return
	}
	id, ok := parseID(c)
	if !ok {
		return
	}
	if _, err := h.fetchOwnToken(c.Request.Context(), id, uid); err != nil {
		middleware.SetErr(c, toAPIErr(err, apierr.CodeInternal))
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

// Delete DELETE /api/user/tokens/:id — 软禁用(Revoke)。
func (h *UserHandler) Delete(c *gin.Context) {
	uid := h.mustUser(c)
	if uid == 0 {
		return
	}
	id, ok := parseID(c)
	if !ok {
		return
	}
	if _, err := h.fetchOwnToken(c.Request.Context(), id, uid); err != nil {
		middleware.SetErr(c, toAPIErr(err, apierr.CodeInternal))
		return
	}
	if err := h.store.Revoke(c.Request.Context(), id); err != nil {
		middleware.SetErr(c, toAPIErr(err, apierr.CodeInternal))
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "revoked"})
}

// Regenerate POST /api/user/tokens/:id/regenerate
func (h *UserHandler) Regenerate(c *gin.Context) {
	uid := h.mustUser(c)
	if uid == 0 {
		return
	}
	id, ok := parseID(c)
	if !ok {
		return
	}
	if _, err := h.fetchOwnToken(c.Request.Context(), id, uid); err != nil {
		middleware.SetErr(c, toAPIErr(err, apierr.CodeInternal))
		return
	}
	plaintext, view, err := h.store.Regenerate(c.Request.Context(), id)
	if err != nil {
		middleware.SetErr(c, toAPIErr(err, apierr.CodeInternal))
		return
	}
	c.JSON(http.StatusOK, CreateResponse{
		View:         toDTO(view),
		PlaintextKey: plaintext,
		Warning:      plaintextWarning,
	})
}

// === internal helpers ===

func (h *UserHandler) mustUser(c *gin.Context) int64 {
	if h.userOf == nil {
		middleware.SetErr(c, apierr.New(apierr.CodeNotLoggedIn, "session not configured"))
		return 0
	}
	uid := h.userOf(c)
	if uid == 0 {
		middleware.SetErr(c, apierr.New(apierr.CodeNotLoggedIn, "login required"))
		return 0
	}
	return uid
}

// fetchOwnToken 取 token,确保 view.UserID == uid,否则 CodeForbidden。
func (h *UserHandler) fetchOwnToken(ctx context.Context, id, uid int64) (*tokensvc.View, error) {
	view, err := h.store.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if view.UserID != uid {
		return nil, apierr.New(apierr.CodeForbidden, "token belongs to another user")
	}
	return view, nil
}

func parseID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		middleware.SetErr(c, apierr.New(apierr.CodeInvalidParam, "invalid id"))
		return 0, false
	}
	return id, true
}

func parsePagination(c *gin.Context) (page, size int) {
	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	if page <= 0 {
		page = 1
	}
	size, _ = strconv.Atoi(c.DefaultQuery("size", "20"))
	if size <= 0 {
		size = 20
	}
	if size > 100 {
		size = 100
	}
	return
}

func parseStatusQuery(c *gin.Context) (int8, bool) {
	s := c.Query("status")
	if s == "" {
		return 0, false
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, false
	}
	return int8(n), true
}

func toAPIErr(err error, fallback apierr.Code) *apierr.Error {
	var ae *apierr.Error
	if errors.As(err, &ae) {
		return ae
	}
	return apierr.New(fallback, err.Error())
}

// 名称合法字符集:1-64 char。
var nameRe = regexp.MustCompile(`^[^\x00-\x1f]{1,64}$`)

// 模型 pattern 合法字符:[a-zA-Z0-9._\-*]
var modelPatternRe = regexp.MustCompile(`^[a-zA-Z0-9._\-*]{1,128}$`)

func validateCreate(r *CreateRequest) *apierr.Error {
	if !nameRe.MatchString(r.Name) {
		return apierr.New(apierr.CodeInvalidParam, "name must be 1-64 chars and not contain control chars")
	}
	if r.QuotaLimit != nil && *r.QuotaLimit < 0 {
		return apierr.New(apierr.CodeInvalidParam, "quota_limit must be >= 0")
	}
	for _, m := range r.AllowedModels {
		if !modelPatternRe.MatchString(m) {
			return apierr.New(apierr.CodeInvalidParam, "invalid model pattern: "+m)
		}
	}
	if r.RPMLimit < 0 || r.TPMLimit < 0 {
		return apierr.New(apierr.CodeInvalidParam, "rpm/tpm limit must be >= 0")
	}
	if r.ExpiresAt != nil && r.ExpiresAt.Before(time.Now().Add(time.Minute)) {
		return apierr.New(apierr.CodeInvalidParam, "expires_at must be at least 1 minute in the future")
	}
	return nil
}

func validatePatch(r *PatchRequest) *apierr.Error {
	if r.Name != nil && !nameRe.MatchString(*r.Name) {
		return apierr.New(apierr.CodeInvalidParam, "name must be 1-64 chars and not contain control chars")
	}
	if r.AllowedModels != nil {
		for _, m := range *r.AllowedModels {
			if !modelPatternRe.MatchString(m) {
				return apierr.New(apierr.CodeInvalidParam, "invalid model pattern: "+m)
			}
		}
	}
	if r.RPMLimit != nil && *r.RPMLimit < 0 {
		return apierr.New(apierr.CodeInvalidParam, "rpm_limit must be >= 0")
	}
	if r.TPMLimit != nil && *r.TPMLimit < 0 {
		return apierr.New(apierr.CodeInvalidParam, "tpm_limit must be >= 0")
	}
	if r.Status != nil && *r.Status != 0 && *r.Status != 1 {
		return apierr.New(apierr.CodeInvalidParam, "status only 0/1 allowed via API")
	}
	if r.ExpiresAt != nil && !r.ClearExpiresAt && r.ExpiresAt.Before(time.Now().Add(time.Minute)) {
		return apierr.New(apierr.CodeInvalidParam, "expires_at must be at least 1 minute in the future")
	}
	return nil
}
