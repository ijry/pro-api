package user

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/ijry/pro-api/internal/app"
	"github.com/ijry/pro-api/internal/notice"
)

// WireUserNotice 构造用户公告 NoticeHandler。
//
// userOf 提供从 gin.Context 解出 user id 的函数;若 nil 视为未登录(<= 0)。
func WireUserNotice(a *app.Application, userOf func(*gin.Context) int64) (*NoticeHandler, error) {
	if a == nil {
		return nil, errors.New("user: app is nil")
	}
	svc := notice.ServiceFrom(a)
	if svc == nil {
		return nil, errors.New("user: notice service not wired")
	}
	return NewNoticeHandler(svc, userOf), nil
}
