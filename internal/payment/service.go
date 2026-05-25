// Package payment 聚合手动充值 (manual) 与兑换码 (redeem) 两个子服务。
//
// 子模块各自定义了 mini wallet interface(WalletCredit),
// 避免与 M1-06 wallet 包形成强耦合;wire 时由 main.go 注入真正的 wallet.Store。
//
// Holder 是一个跨子包共享的容器类型,放在 application 的 PaymentSvc 字段。
// manual.Wire 与 redeem.Wire 分别填入各自的子服务字段,互不冲突。
package payment

// Holder 是 payment 子服务的共享容器,由 application 的 PaymentSvc 字段持有。
//
// 字段类型用 any:避免本包 import 子包(import cycle)。
// 调用方(handler / 子包自己的 ServiceFrom)用类型断言取回具体类型。
type Holder struct {
	// Manual: *manual.Service(manual 包自己负责装配,通过类型断言取出)。
	Manual any
	// Redeem: *redeem.Service。
	Redeem any
	// Online: *online.Service (online payments: Stripe/Alipay/WechatPay).
	Online any
}

// HolderFrom 是个工具:从一个 any(通常是 app.PaymentSvc)取回 Holder;
// 如果还没被装配,返一个空 Holder(调用方应负责把它回写)。
//
// 用法:
//
//	h := payment.HolderFrom(app.PaymentSvc)
//	h.Manual = manualSvc
//	app.PaymentSvc = h
func HolderFrom(v any) *Holder {
	if h, ok := v.(*Holder); ok && h != nil {
		return h
	}
	return &Holder{}
}
