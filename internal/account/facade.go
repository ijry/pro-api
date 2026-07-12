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

// ReportUsage 在一次调用成功后,按 token 用量扣减 manual 模式账号的手动额度。
// 仅 quota_mode='manual' 的账号会被扣(见 Repo.DeductManualQuota);auto / none
// 模式与未启用账号池(accountID==0)均为 no-op。tokens<=0 不产生副作用。
// 扣减在后台 goroutine 里异步进行,不阻塞转发热路径的返回。
func (f *Facade) ReportUsage(accountID int64, tokens int64) {
	if accountID == 0 || tokens <= 0 || f.Repo == nil {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = f.Repo.DeductManualQuota(ctx, accountID, tokens)
	}()
}
