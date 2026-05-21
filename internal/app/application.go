// Package app 装配应用级容器与启动顺序。
package app

import (
	"context"

	"github.com/proapi/proapi/internal/app/config"
	"github.com/proapi/proapi/internal/audit"
	"github.com/proapi/proapi/internal/setting"
	"github.com/proapi/proapi/internal/util/clock"
	"github.com/proapi/proapi/internal/util/crypto"
	"github.com/proapi/proapi/internal/util/idgen"
	"github.com/proapi/proapi/internal/util/tokenize"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Application 是应用全局容器。基础设施层字段在 SetupBasic 里填好;
// 业务层(token/wallet/...)由各模块 spec 的 wire 函数后续填到 any 字段。
type Application struct {
	Config   *config.Config
	Log      *zap.Logger
	DB       *gorm.DB
	Cache    *redis.Client
	Clock    clock.Clock
	IDGen    *idgen.Generator
	Crypto   *crypto.AESGCM
	Setting  setting.Store
	Audit    audit.Logger
	Tokenize *tokenize.Registry

	// 业务层占位(由各业务 spec 的 Wire 函数填入,本 spec 不绑定具体类型)
	UserSvc     any
	GroupSvc    any
	TokenStore  any
	WalletStore any
	ChannelSvc  any
	PricingSvc  any
	Biller      any
	Limiter     any
	LogStore    any
	AuthSvc     any
	NoticeSvc   any
	PaymentSvc  any
	AdapterReg  any
	Relay       any

	cls closers
}

// AddCloser 注册关停回调。
func (a *Application) AddCloser(name string, fn func() error) { a.cls.Add(name, fn) }

// Shutdown 调用所有 closer(LIFO)。
func (a *Application) Shutdown(ctx context.Context) error {
	_ = ctx
	return a.cls.Run()
}
