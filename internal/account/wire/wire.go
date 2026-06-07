// Package accountwire 装配账号池子系统到 app.Application。
// 拆为独立包是为了避免 internal/account 与 internal/account/oauth(及
// probe/quota)之间的 import 循环——子包 import account 包以拿 Account/Cred 类型,
// 而装配代码反过来需要 import 子包来构造各 provider 实例,故必须放在独立包里。
package accountwire

import (
	"context"
	"errors"
	"time"

	"github.com/ijry/pro-api/internal/account"
	"github.com/ijry/pro-api/internal/account/oauth"
	"github.com/ijry/pro-api/internal/account/probe"
	"github.com/ijry/pro-api/internal/account/quota"
	"github.com/ijry/pro-api/internal/app"
	"github.com/ijry/pro-api/internal/util/clock"
)

// WireAccount 装配账号池子系统(Repo / Selector / Breaker / Refresher /
// Probe / Importer / OAuthFlow / QuotaTracker),挂到 a.AccountSvc。
//
// 副作用:启动 Breaker.Run 与 Refresher.Run 两个后台 goroutine,
// 通过 AddCloser 注册关停回调。
func WireAccount(ctx context.Context, a *app.Application) error {
	if a == nil {
		return errors.New("account: app is nil")
	}
	if a.DB == nil {
		return errors.New("account: app.DB is nil")
	}
	if a.Cache == nil {
		return errors.New("account: app.Cache is nil")
	}
	clk := a.Clock
	if clk == nil {
		clk = clock.Real
	}
	repo := account.NewRepository(a.DB, a.Crypto, a.IDGen, clk, a.Log)

	var anthropicCfg, openaiCfg oauth.Config
	if a.Config != nil {
		anthropicCfg = oauth.Config{
			TokenURL:    a.Config.Account.OAuthAnthropicTokenURL,
			AuthURL:     a.Config.Account.OAuthAnthropicAuthURL,
			ClientID:    a.Config.Account.OAuthAnthropicClientID,
			RedirectURI: a.Config.Account.OAuthAnthropicRedirectURI,
			Scopes:      a.Config.Account.OAuthAnthropicScopes,
		}
		openaiCfg = oauth.Config{
			TokenURL:    a.Config.Account.OAuthOpenAITokenURL,
			AuthURL:     a.Config.Account.OAuthOpenAIAuthURL,
			ClientID:    a.Config.Account.OAuthOpenAIClientID,
			RedirectURI: a.Config.Account.OAuthOpenAIRedirectURI,
			Scopes:      a.Config.Account.OAuthOpenAIScopes,
		}
	}
	oa := oauth.NewFlow(map[string]oauth.Provider{
		"anthropic": oauth.NewAnthropic(anthropicCfg),
		"openai":    oauth.NewOpenAI(openaiCfg),
	}, oauth.NewRedisStateStore(a.Cache))

	tracker := quota.NewTracker(repo)

	var anthropicProbeBase, openaiProbeBase string
	if a.Config != nil {
		anthropicProbeBase = a.Config.Account.AnthropicProbeBase
		openaiProbeBase = a.Config.Account.OpenAIProbeBase
	}
	pr := account.NewProbe(repo, tracker, map[string]account.ProviderProbe{
		"anthropic": probe.NewAnthropic(anthropicProbeBase),
		"openai":    probe.NewOpenAI(openaiProbeBase),
	})

	br := account.NewBreaker(repo, a.Cache, a.Log)
	go func() { _ = br.Run(ctx) }()

	sel := account.NewSelector(repo, br, time.Now().UnixNano())
	rf := account.NewRefresher(repo, a.Cache, oa, a.Log)
	go func() { _ = rf.Run(ctx) }()

	imp := account.NewDefaultImporter(oa)

	a.AccountSvc = &account.Facade{
		Repo:       repo,
		Selector:   sel,
		Breaker:    br,
		Refresher:  rf,
		Probe:      pr,
		Importer:   imp,
		OAuth:      oa,
		QuotaTrack: tracker,
	}
	a.AddCloser("account_breaker", br.Close)
	a.AddCloser("account_refresher", rf.Close)
	return nil
}
