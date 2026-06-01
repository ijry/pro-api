package account

import (
	"context"
	"net/http"
	"time"

	"github.com/ijry/pro-api/internal/channel"
)

// Facade 是 account 模块对外暴露的服务门面,聚合 Repo / Selector / Breaker /
// Refresher / Probe / Importer / OAuthFlow / QuotaTracker。
// 它实现了 relay.AccountFacade(隐式接口),由 WireAccount 装配后挂到
// app.Application.AccountSvc 上,可通过 FacadeFrom 取回。
type Facade struct {
	Repo       Repo
	Selector   Selector
	Breaker    Breaker
	Refresher  Refresher
	Probe      Probe
	Importer   Importer
	OAuth      OAuthFlow
	QuotaTrack QuotaTracker
}

// Select 转发到内部 Selector,满足 relay.AccountFacade.Select。
func (f *Facade) Select(ctx context.Context, ch *channel.Channel, hint SelectHint) (*Account, error) {
	return f.Selector.Select(ctx, ch, hint)
}

// ReportSuccess 转发到内部 Selector,满足 relay.AccountFacade.ReportSuccess。
func (f *Facade) ReportSuccess(accountID int64, latency time.Duration) {
	f.Selector.ReportSuccess(accountID, latency)
}

// ReportFailure 转发到内部 Selector,满足 relay.AccountFacade.ReportFailure。
func (f *Facade) ReportFailure(accountID int64, err error, headers http.Header) {
	f.Selector.ReportFailure(accountID, err, headers)
}
