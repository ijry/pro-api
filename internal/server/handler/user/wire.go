package user

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/ijry/pro-api/internal/app"
	"github.com/ijry/pro-api/internal/channel"
	"github.com/ijry/pro-api/internal/invite"
	"github.com/ijry/pro-api/internal/notice"
	"github.com/ijry/pro-api/internal/relay"
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

// WirePlayground 构造 PlaygroundHandler。
//
// 依赖 relay.Service、channel.Facade 和 setting.Store 均从 app.Application 取。
func WirePlayground(a *app.Application) (*PlaygroundHandler, error) {
	if a == nil {
		return nil, errors.New("user: app is nil")
	}
	relaySvc, ok := a.Relay.(*relay.Service)
	if !ok || relaySvc == nil {
		return nil, errors.New("user: relay service not wired")
	}
	facade := channel.FacadeFrom(a)
	if facade == nil {
		return nil, errors.New("user: channel facade not wired")
	}
	return NewPlaygroundHandler(PlaygroundDeps{
		Relay:   relaySvc,
		Channel: facade,
		Setting: a.Setting,
		Log:     a.Log,
	}), nil
}

// WireInvite constructs an InviteHandler for the user API group.
//
// wallet=nil is safe: HTTP handlers only call read methods (GetSummary,
// ListInvitees, ListRecords). OnOrderPaid — the only method that uses the
// wallet — is called exclusively from the payment service which wires its
// own invite.Service instance with a real wallet.
func WireInvite(a *app.Application, userOf func(*gin.Context) int64) (*InviteHandler, error) {
	if a == nil {
		return nil, errors.New("user: app is nil")
	}
	svc := invite.Wire(a, nil)
	return NewInviteHandler(svc, userOf), nil
}
