package manual

import (
	"errors"
	"fmt"

	"github.com/ijry/pro-api/internal/app"
	"github.com/ijry/pro-api/internal/payment"
)

// Wire 装配 manual 服务并挂到 application 容器的 PaymentSvc.Manual。
//
// 装配后 app.PaymentSvc 类型为 *payment.Holder;redeem.Wire 与本函数
// 共享同一个 Holder 实例(通过 payment.HolderFrom 复用),调用顺序无关。
//
// 依赖:
//   - app.DB / app.Setting / app.Audit / app.IDGen / app.Clock / app.Log
//   - app.WalletStore — 必须已被 M1-06 WireWallet 装配,且实现 WalletCredit
//     接口(Credit(ctx, userID, amount, refType, refID, desc) error)。
func Wire(a *app.Application) error {
	if a == nil {
		return errors.New("payment.manual: app is nil")
	}
	if a.DB == nil {
		return errors.New("payment.manual: app.DB is nil")
	}
	if a.IDGen == nil {
		return errors.New("payment.manual: app.IDGen is nil")
	}
	if a.Setting == nil {
		return errors.New("payment.manual: app.Setting is nil")
	}
	if a.WalletStore == nil {
		return errors.New("payment.manual: app.WalletStore is nil (call WireWallet first)")
	}
	wallet, ok := a.WalletStore.(WalletCredit)
	if !ok {
		return fmt.Errorf("payment.manual: app.WalletStore (type %T) does not satisfy WalletCredit", a.WalletStore)
	}

	svc := New(Config{
		DB:      a.DB,
		Setting: a.Setting,
		Wallet:  wallet,
		Audit:   a.Audit,
		IDGen:   a.IDGen,
		Clock:   a.Clock,
		Log:     a.Log,
	})

	h := payment.HolderFrom(a.PaymentSvc)
	h.Manual = svc
	a.PaymentSvc = h
	return nil
}

// ServiceFrom 从 app.PaymentSvc 取回 manual.Service。装配未完成或类型不符返 nil。
func ServiceFrom(a *app.Application) Service {
	if a == nil {
		return nil
	}
	h, ok := a.PaymentSvc.(*payment.Holder)
	if !ok || h == nil {
		return nil
	}
	svc, _ := h.Manual.(Service)
	return svc
}
