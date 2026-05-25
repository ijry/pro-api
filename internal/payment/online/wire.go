package online

import (
	"github.com/ijry/pro-api/internal/app"
	"github.com/ijry/pro-api/internal/payment"
	payprovider "github.com/ijry/pro-api/internal/payment/provider"
)

// Wire assembles the online.Service and sets it into app.PaymentSvc.(payment.Holder).Online.
func Wire(a *app.Application, providers []payprovider.Provider, wallet WalletCredit, invite InviteRebate) *Service {
	svc := NewService(Deps{
		Repo:      NewRepository(a.DB),
		Providers: providers,
		Wallet:    wallet,
		Invite:    invite,
		IDGen:     a.IDGen,
		Clock:     a.Clock,
		Log:       a.Log,
	})
	h := payment.HolderFrom(a.PaymentSvc)
	h.Online = svc
	a.PaymentSvc = h
	return svc
}
