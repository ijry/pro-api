package channel

import (
	"context"
	"fmt"
	"time"

	"github.com/ijry/pro-api/internal/app"
	"github.com/ijry/pro-api/internal/token"
	"github.com/ijry/pro-api/internal/util/clock"
	"go.uber.org/zap"
)

// Facade 是 channel 对外暴露的门面,持有 Selector / Breaker / Repo。
type Facade struct {
	Repo     *repo
	Cache    *channelCache
	Breaker  Breaker
	Selector Selector
	Log      *zap.Logger
}

// WireChannel 装配渠道选择器 / 熔断器 / 本地缓存到 app.ChannelSvc。
// 返回的 ChannelSvc 类型为 *Facade(可通过 FacadeFrom 取回)。
func WireChannel(ctx context.Context, a *app.Application) error {
	if a.DB == nil {
		return fmt.Errorf("channel: app.DB is nil")
	}
	if a.Cache == nil {
		return fmt.Errorf("channel: app.Cache is nil")
	}

	clk := a.Clock
	if clk == nil {
		clk = clock.Real
	}

	r := newRepo(a.DB, a.Crypto, a.IDGen, clk)
	c := newChannelCache(r, a.Cache, a.Log)

	loaded, failed, err := c.Warmup(ctx)
	if err != nil {
		a.Log.Warn("channel: warmup error", zap.Error(err))
	}
	a.Log.Sugar().Infof("channel: warmed %d channels (%d decrypt-failed)", loaded, len(failed))
	c.StartPubSub(ctx)

	b := newBreaker(a.Cache, clk, a.Log)
	b.Start(ctx)

	sel := newSelector(c, b, time.Now().UnixNano())

	facade := &Facade{
		Repo:     r,
		Cache:    c,
		Breaker:  b,
		Selector: sel,
		Log:      a.Log,
	}

	a.ChannelSvc = facade
	a.AddCloser("channel_cache", c.Close)
	a.AddCloser("channel_breaker", b.Close)
	return nil
}

// FacadeFrom extracts the channel *Facade from app.ChannelSvc.
func FacadeFrom(a *app.Application) *Facade {
	if f, ok := a.ChannelSvc.(*Facade); ok {
		return f
	}
	return nil
}

// ActiveModels implements token.ModelLister for /v1/models.
func (f *Facade) ActiveModels(ctx context.Context) []string {
	if f == nil || f.Cache == nil {
		return nil
	}
	return f.Cache.ActiveModels(token.GroupIDFromContext(ctx))
}

// ModelInfo implements token.ModelLister. Channel mappings do not currently
// persist created/owned_by metadata, so the handler applies protocol defaults.
func (f *Facade) ModelInfo(_ string) (token.ModelMeta, bool) {
	return token.ModelMeta{}, false
}
