package admin

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/ijry/pro-api/internal/account"
	"github.com/ijry/pro-api/internal/app"
	"github.com/ijry/pro-api/internal/notice"
)

// WireAdminSetting 构造 SettingHandler 并返回(由调用方挂路由)。
//
// actorOf 提供从 gin.Context 解出操作者 user id 的函数;若 nil 用 0 占位。
// mailer 为可选,nil 时 test_smtp 走 stub。
func WireAdminSetting(a *app.Application, mailer Mailer, actorOf func(*gin.Context) int64) (*SettingHandler, error) {
	if a == nil {
		return nil, errors.New("admin: app is nil")
	}
	if a.Setting == nil {
		return nil, errors.New("admin: app.Setting is nil")
	}
	if a.Crypto == nil {
		return nil, errors.New("admin: app.Crypto is nil")
	}
	return NewSettingHandler(a.Setting, a.Crypto, mailer, a.Audit, actorOf), nil
}

// WireAdminNotice 构造 NoticeHandler 并返回。
//
// 依赖 app.NoticeSvc 已被 notice.WireNotice 装配。
func WireAdminNotice(a *app.Application, actorOf func(*gin.Context) int64) (*NoticeHandler, error) {
	if a == nil {
		return nil, errors.New("admin: app is nil")
	}
	svc := notice.ServiceFrom(a)
	if svc == nil {
		return nil, errors.New("admin: notice service not wired")
	}
	return NewNoticeHandler(svc, actorOf), nil
}

// WireAdminAccount 构造 AccountHandler 并返回。
//
// 依赖 app.AccountSvc 已被 accountwire.WireAccount 装配为 *account.Facade。
// 未装配时返回 error,由调用方决定是否致命(M1 可视作 skip)。
func WireAdminAccount(a *app.Application, actorOf func(*gin.Context) int64) (*AccountHandler, error) {
	if a == nil {
		return nil, errors.New("admin: app is nil")
	}
	facade := account.FacadeFrom(a)
	if facade == nil {
		return nil, errors.New("admin: account facade not wired")
	}
	return NewAccountHandler(facade, a.Audit, actorOf), nil
}
