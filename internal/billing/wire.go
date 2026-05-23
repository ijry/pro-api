package billing

import (
	"fmt"

	"github.com/ijry/pro-api/internal/app"
	"github.com/ijry/pro-api/internal/audit"
	"github.com/ijry/pro-api/internal/util/clock"
	"github.com/ijry/pro-api/internal/wallet"
)

// WireBilling 装配 Biller 并注入到 app.Biller。
// 需要 app.Wallet / app.Setting / app.IDGen / app.Cache / app.DB / app.Audit / app.Clock 已就绪。
func WireBilling(a *app.Application) error {
	ws, ok := a.WalletStore.(wallet.Store)
	if !ok {
		return fmt.Errorf("WireBilling: app.WalletStore not wallet.Store")
	}

	var usageInc UsageIncrementer
	// 如果 TokenStore 实现了 UsageIncrementer,注入;否则 nil
	if ui, ok := a.TokenStore.(UsageIncrementer); ok {
		usageInc = ui
	}

	clk, ok := a.Clock.(clock.Clock)
	if !ok {
		clk = clock.Real
	}

	b, err := New(Config{
		DB:      a.DB,
		Cache:   a.Cache,
		Log:     a.Log,
		Clock:   clk,
		IDGen:   a.IDGen,
		Setting: a.Setting,
		Audit:   a.Audit.(audit.Logger),
		Wallet:  ws,
		Usage:   usageInc,
	})
	if err != nil {
		return fmt.Errorf("WireBilling: %w", err)
	}
	a.Biller = b
	a.AddCloser("billing", b.Close)
	return nil
}
