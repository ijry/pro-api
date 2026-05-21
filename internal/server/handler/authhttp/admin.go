package authhttp

import (
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ijry/pro-api/internal/audit"
	"github.com/ijry/pro-api/internal/auth/password"
	"github.com/ijry/pro-api/internal/server/middleware"
	"github.com/ijry/pro-api/internal/user"
	"github.com/ijry/pro-api/pkg/apierr"
)

// GET /api/admin/auth/me
func (h *handlers) adminMe(c *gin.Context) {
	uid := middleware.UserID(c)
	u, err := h.d.User.GetByID(c.Request.Context(), uid)
	if err != nil {
		sendErr(c, err)
		return
	}
	if u == nil {
		sendErr(c, apierr.New(apierr.CodeForbidden, "用户不存在"))
		return
	}
	c.JSON(200, userView(u, resolveGroupName(c.Request.Context(), h.d.Group, u.GroupID)))
}

// GET /api/admin/users
func (h *handlers) adminUsersList(c *gin.Context) {
	page := optionalQueryInt(c, "page", 1)
	size := optionalQueryInt(c, "size", 20)
	f := user.ListFilter{
		Keyword: c.Query("keyword"),
		Role:    optionalQueryInt8(c, "role"),
		Status:  optionalQueryInt8(c, "status"),
		GroupID: optionalQueryInt64(c, "group_id"),
		Page:    page, Size: size,
		OrderBy: c.Query("order"),
	}
	items, total, err := h.d.User.List(c.Request.Context(), f)
	if err != nil {
		sendErr(c, err)
		return
	}
	out := make([]gin.H, 0, len(items))
	for _, u := range items {
		out = append(out, userView(u, resolveGroupName(c.Request.Context(), h.d.Group, u.GroupID)))
	}
	c.JSON(200, gin.H{"items": out, "total": total, "page": page, "size": size})
}

// GET /api/admin/users/:id
func (h *handlers) adminUserDetail(c *gin.Context) {
	id, ok := parseInt64Param(c, "id")
	if !ok {
		sendErr(c, errInvalidInt)
		return
	}
	u, err := h.d.User.GetByID(c.Request.Context(), id)
	if err != nil {
		sendErr(c, err)
		return
	}
	if u == nil {
		sendErr(c, apierr.New(apierr.CodeForbidden, "用户不存在"))
		return
	}
	bindings, _ := h.d.OAuthRepo.ListByUser(c.Request.Context(), id)
	out := make([]gin.H, 0, len(bindings))
	for _, b := range bindings {
		out = append(out, gin.H{
			"provider": b.Provider, "uid": b.ProviderUID,
			"bound_at": b.CreatedAt.Format(time.RFC3339Nano),
		})
	}
	c.JSON(200, gin.H{
		"user":     userView(u, resolveGroupName(c.Request.Context(), h.d.Group, u.GroupID)),
		"wallet":   nil,
		"bindings": out,
	})
}

type adminUserPatchReq struct {
	DisplayName *string `json:"display_name"`
	Role        *int8   `json:"role"`
	Status      *int8   `json:"status"`
	GroupID     *int64  `json:"group_id"`
}

// PATCH /api/admin/users/:id
func (h *handlers) adminUserPatch(c *gin.Context) {
	id, ok := parseInt64Param(c, "id")
	if !ok {
		sendErr(c, errInvalidInt)
		return
	}
	actorID := middleware.UserID(c)
	actorRole := middleware.Role(c)
	if actorID == id {
		// 不能修改自己的 role / status,只能改 display_name
	}
	var req adminUserPatchReq
	if err := c.ShouldBindJSON(&req); err != nil {
		sendErr(c, apierr.New(apierr.CodeInvalidParam, "请求格式错误"))
		return
	}
	target, err := h.d.User.GetByID(c.Request.Context(), id)
	if err != nil {
		sendErr(c, err)
		return
	}
	if target == nil {
		sendErr(c, apierr.New(apierr.CodeForbidden, "用户不存在"))
		return
	}
	// 权限校验
	if req.Role != nil {
		if id == actorID {
			sendErr(c, apierr.New(apierr.CodeForbidden, "不能修改自己的角色"))
			return
		}
		if actorRole <= target.Role {
			sendErr(c, apierr.New(apierr.CodeForbidden, "无权修改该用户角色"))
			return
		}
		if actorRole <= *req.Role {
			sendErr(c, apierr.New(apierr.CodeForbidden, "无权设置该角色"))
			return
		}
	}
	if req.Status != nil && id == actorID && *req.Status == user.StatusDisabled {
		sendErr(c, apierr.New(apierr.CodeForbidden, "不能禁用自己"))
		return
	}
	beforeJSON, _ := json.Marshal(map[string]any{
		"display_name": target.DisplayName, "role": target.Role, "status": target.Status, "group_id": target.GroupID,
	})
	updated, err := h.d.User.Update(c.Request.Context(), id, user.UpdateInput{
		DisplayName: req.DisplayName, Role: req.Role, Status: req.Status, GroupID: req.GroupID,
	})
	if err != nil {
		sendErr(c, err)
		return
	}
	afterJSON, _ := json.Marshal(map[string]any{
		"display_name": updated.DisplayName, "role": updated.Role, "status": updated.Status, "group_id": updated.GroupID,
	})
	_ = h.d.Audit.Log(c.Request.Context(), audit.Entry{
		ActorID: &actorID, Action: "user.update", TargetType: "user", TargetID: &id,
		Before: beforeJSON, After: afterJSON, IP: clientIP(c),
	})
	if req.Role != nil && *req.Role != target.Role {
		_ = h.d.Audit.Log(c.Request.Context(), audit.Entry{
			ActorID: &actorID, Action: "user.role_change", TargetType: "user", TargetID: &id,
		})
	}
	if req.Status != nil && *req.Status != target.Status {
		_ = h.d.Audit.Log(c.Request.Context(), audit.Entry{
			ActorID: &actorID, Action: "user.status_change", TargetType: "user", TargetID: &id,
		})
		if *req.Status == user.StatusDisabled {
			_ = h.d.Session.RevokeAllForUser(c.Request.Context(), id)
		}
	}
	c.JSON(200, userView(updated, resolveGroupName(c.Request.Context(), h.d.Group, updated.GroupID)))
}

type adminResetPasswordReq struct {
	NewPassword string `json:"new_password"`
}

// POST /api/admin/users/:id/reset_password
func (h *handlers) adminUserResetPassword(c *gin.Context) {
	id, ok := parseInt64Param(c, "id")
	if !ok {
		sendErr(c, errInvalidInt)
		return
	}
	var req adminResetPasswordReq
	_ = c.ShouldBindJSON(&req)
	actorID := middleware.UserID(c)
	tempIssued := false
	pwd := req.NewPassword
	if pwd == "" {
		pwd = generateTempPassword()
		tempIssued = true
	}
	minLen := h.d.Setting.GetInt(c.Request.Context(), "auth.password.min_length", 8)
	requireMixed := h.d.Setting.GetBool(c.Request.Context(), "auth.password.require_mixed", false)
	if err := password.CheckStrength(pwd, minLen, requireMixed); err != nil {
		sendErr(c, err)
		return
	}
	hash, err := password.Hash(pwd)
	if err != nil {
		sendErr(c, err)
		return
	}
	if err := h.d.User.UpdatePasswordHash(c.Request.Context(), id, hash); err != nil {
		sendErr(c, err)
		return
	}
	_ = h.d.Session.RevokeAllForUser(c.Request.Context(), id)
	_ = h.d.Audit.Log(c.Request.Context(), audit.Entry{
		ActorID: &actorID, Action: "user.reset_password", TargetType: "user", TargetID: &id, IP: clientIP(c),
	})
	resp := gin.H{"ok": true}
	if tempIssued {
		resp["temp_password"] = pwd
	}
	c.JSON(200, resp)
}

// DELETE /api/admin/users/:id
func (h *handlers) adminUserDelete(c *gin.Context) {
	id, ok := parseInt64Param(c, "id")
	if !ok {
		sendErr(c, errInvalidInt)
		return
	}
	actorID := middleware.UserID(c)
	actorRole := middleware.Role(c)
	if id == actorID {
		sendErr(c, apierr.New(apierr.CodeForbidden, "不能删除自己"))
		return
	}
	target, err := h.d.User.GetByID(c.Request.Context(), id)
	if err != nil {
		sendErr(c, err)
		return
	}
	if target == nil {
		sendErr(c, apierr.New(apierr.CodeForbidden, "用户不存在"))
		return
	}
	if actorRole <= target.Role {
		sendErr(c, apierr.New(apierr.CodeForbidden, "无权删除该用户"))
		return
	}
	if err := h.d.User.Delete(c.Request.Context(), id); err != nil {
		sendErr(c, err)
		return
	}
	_ = h.d.Session.RevokeAllForUser(c.Request.Context(), id)
	_ = h.d.Audit.Log(c.Request.Context(), audit.Entry{
		ActorID: &actorID, Action: "user.delete", TargetType: "user", TargetID: &id, IP: clientIP(c),
	})
	c.JSON(200, gin.H{"ok": true})
}
