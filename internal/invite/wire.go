package invite

import "github.com/ijry/pro-api/internal/app"

// Wire assembles the invite.Service.
func Wire(a *app.Application, wallet WalletCredit) *Service {
	return NewService(Deps{
		Repo:    NewRepository(a.DB),
		DB:      a.DB,
		Wallet:  wallet,
		Setting: a.Setting,
		IDGen:   a.IDGen,
		Clock:   a.Clock,
		Log:     a.Log,
	})
}
