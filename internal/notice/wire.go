package notice

import (
	"errors"

	"github.com/ijry/pro-api/internal/app"
)

// WireNotice 装配 notice 服务并挂到 application 容器的 NoticeSvc 字段。
//
// 依赖 M1-01 提供:app.DB / app.Cache / app.IDGen / app.Clock / app.Audit / app.Log。
// 调用方应在装配链中(cmd/proapi/main.go)调用此函数。M1-02 LoginHandler 可通过
// notice.ServiceFrom(app) 取回 Service 并调用 UnreadCountForUser(ctx, uid)。
func WireNotice(a *app.Application) error {
	if a == nil {
		return errors.New("notice: app is nil")
	}
	if a.DB == nil {
		return errors.New("notice: app.DB is nil")
	}
	if a.Cache == nil {
		return errors.New("notice: app.Cache is nil")
	}
	if a.IDGen == nil {
		return errors.New("notice: app.IDGen is nil")
	}
	repo := NewRepo(a.DB)
	reader := NewReader(a.Cache)
	svc := NewService(Config{
		Repo:   repo,
		Reader: reader,
		IDGen:  a.IDGen,
		Clock:  a.Clock,
		Audit:  a.Audit,
		Log:    a.Log,
	})
	a.NoticeSvc = svc
	return nil
}

// ServiceFrom 从 app.NoticeSvc 取回 Service;若装配未完成或类型不符,返回 nil。
//
// 跨包消费者(M1-02 登录 handler、M1-09 后台 wire、本包 handler)
// 都用此函数取 Service,避免暴露内部容器字段名。
func ServiceFrom(a *app.Application) Service {
	if a == nil {
		return nil
	}
	if svc, ok := a.NoticeSvc.(Service); ok {
		return svc
	}
	return nil
}
