package public

import (
	"errors"

	"github.com/ijry/pro-api/internal/app"
	"github.com/ijry/pro-api/internal/notice"
)

// WirePublicNotice 构造公开公告 NoticeHandler。
func WirePublicNotice(a *app.Application) (*NoticeHandler, error) {
	if a == nil {
		return nil, errors.New("public: app is nil")
	}
	svc := notice.ServiceFrom(a)
	if svc == nil {
		return nil, errors.New("public: notice service not wired")
	}
	return NewNoticeHandler(svc), nil
}
