package authhttp

import (
	"github.com/gin-gonic/gin"
	"github.com/ijry/pro-api/internal/server/middleware"
	"github.com/ijry/pro-api/internal/user"
	"github.com/ijry/pro-api/pkg/apierr"
)

// GET /api/user/profile
func (h *handlers) userProfile(c *gin.Context) {
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

type profilePatchReq struct {
	DisplayName *string `json:"display_name"`
	Avatar      *string `json:"avatar"`
}

// PATCH /api/user/profile
func (h *handlers) userProfilePatch(c *gin.Context) {
	uid := middleware.UserID(c)
	var req profilePatchReq
	if err := c.ShouldBindJSON(&req); err != nil {
		sendErr(c, apierr.New(apierr.CodeInvalidParam, "请求格式错误"))
		return
	}
	u, err := h.d.User.Update(c.Request.Context(), uid, user.UpdateInput{
		DisplayName: req.DisplayName,
		Avatar:      req.Avatar,
	})
	if err != nil {
		sendErr(c, err)
		return
	}
	c.JSON(200, userView(u, resolveGroupName(c.Request.Context(), h.d.Group, u.GroupID)))
}

type userPasswordReq struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// POST /api/user/password
func (h *handlers) userPassword(c *gin.Context) {
	uid := middleware.UserID(c)
	sid := middleware.SessionID(c)
	var req userPasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		sendErr(c, apierr.New(apierr.CodeInvalidParam, "请求格式错误"))
		return
	}
	if err := h.d.Auth.ChangePassword(c.Request.Context(), uid, sid, req.OldPassword, req.NewPassword); err != nil {
		sendErr(c, err)
		return
	}
	c.JSON(200, gin.H{"ok": true})
}
