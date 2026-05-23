package billing

import (
	"fmt"

	"github.com/ijry/pro-api/internal/app"
	"github.com/ijry/pro-api/internal/wallet"
)

// WireBilling 装配 Biller 并注入到 app.Biller。
func WireBilling(a *app.Application) error {
	ws, ok := a.WalletStore.(wallet.Store)
	if !ok {
		return fmt.Errorf("WireBilling: app.WalletStore not wallet.Store")
	}

	var usageInc UsageIncrementer
	if ui, ok := a.TokenStore.(UsageIncrementer); ok {
		usageInc = ui
	}

	b, err := New(Config{
		DB:      a.DB,
		Cache:   a.Cache,
		Log:     a.Log,
		Clock:   a.Clock,
		IDGen:   a.IDGen,
		Setting: a.Setting,
		Audit:   a.Audit,
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
