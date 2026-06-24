---
title: 新增支付方式
outline: deep
---

# 如何新增一个支付方式

> M1 阶段,pro-api 仅提供 **手动充值** 与 **兑换码** 两种支付方式。Stripe / 支付宝 / 微信支付 / Paddle / PayPal 等在线支付预计在 **M2** 实现。
>
> 本页是 **stub**,待 M2 完整。当前先讲已规划的接口契约,供二次开发者提前准备。

## 已规划的接口(M2 实现)

```go
package payment

import (
    "context"
    "net/http"
)

type Provider interface {
    Name() string

    // CreateOrder 创建支付订单,返回支付 URL 给用户跳转
    CreateOrder(ctx context.Context, userID int64, amount int64, currency string) (*Order, error)

    // HandleWebhook 处理支付平台的 webhook 回调(成功/失败/退款)
    HandleWebhook(ctx context.Context, body []byte, headers http.Header) (*WebhookResult, error)

    // QueryOrder 主动查询订单状态(对账 / 兜底)
    QueryOrder(ctx context.Context, orderID string) (*Order, error)

    // Cancel 取消未支付订单
    Cancel(ctx context.Context, orderID string) error
}

type Order struct {
    ID       string   // proapi 内部订单 ID
    UserID   int64
    Amount   int64    // 最小单位整数(分 / 美分)
    Currency string   // CNY / USD
    PayURL   string   // 用户跳转支付页
    Status   string   // pending / paid / canceled / failed / refunded
    CreatedAt time.Time
    PaidAt    *time.Time
}

type WebhookResult struct {
    OrderID  string
    Status   string
    ProviderRef string  // 第三方平台流水号
}
```

## 已规划的 provider(指引性)

| Provider | 文档优先级 | 备注 |
|---|---|---|
| Stripe | 高 | 海外通用,API 文档完善 |
| 支付宝 | 高 | 国内常见 |
| 微信支付 | 高 | 国内常见 |
| Paddle | 中 | 海外 SaaS 友好 |
| PayPal | 中 | 海外通用 |
| Coinbase Commerce | 低 | 加密货币 |

具体由 M2 spec 定稿。本表仅说明大方向。

## 当前可做的(M1 兜底)

- 用 [手动充值](../modules/payment.md#手动充值申请) 作为兜底:用户线下转账 → 上传凭证 → 管理员审批入账
- 用 [兑换码](../modules/payment.md#兑换码) 批量发放:适合预付费 / 福利场景
- 极特殊场景:**自己写脚本调 admin API 直接 `wallet.Credit`** —— **不推荐生产** 用,因为缺少订单 / 审计上下文

## M2 完整后,本页会变成什么样

待 M2 上线后,本页将更新为:

1. **从 Stripe 起步给出完整步骤**:申请账户 → 创建 webhook → 在 pro-api 后台填写 Public Key / Secret Key
2. **接口实现示例**:CreateOrder / HandleWebhook / QueryOrder 完整 Go 代码
3. **测试方案**:Stripe CLI 本地测 webhook
4. **类比其他 provider**:支付宝 / 微信支付 / Paddle / PayPal

请关注 [CHANGELOG](/zh/changelog) 的 M2 节点。

## 关键要点

- 这是 **stub**,M2 才会有详细内容;放在 docs-site 是给二开者一个"心理预期"
- 当前 M1 阶段**强烈不建议**自己改 `wallet.Credit`(无审计、无幂等性,容易破坏 ledger 一致性)
- 如果迫切需要在 M1 接在线支付,可以在自家**外部系统**完成支付,然后调 admin API 触发手动充值审批流程(自动通过)
