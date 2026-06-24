---
title: 计费如何工作
outline: deep
---

# 计费如何工作(用户视角)

> 这是给 pro-api 终端用户(API 调用方)看的"我的钱怎么算"。
>
> 系统底层的预扣 / 提交 / ledger 机制见 [计费机制](../architecture/billing.md);倍率配置方式见 [定价与倍率](../modules/pricing.md)。

## 一次请求要花多少钱

公式:

```
钱 = ( input tokens     × input ratio
      + output tokens    × output ratio
      + cached tokens    × cached ratio
      + reasoning tokens × reasoning ratio )
    × 你的分组倍率
```

**简单理解**:你用了多少 token × 这个 token 类型的倍率 × 你的会员倍率。

## 举例

调一次 `gpt-4o`,你是 `default` 分组(`group_ratio = 1.0`):

- 输入 1000 token,`gpt-4o` `input_ratio = 1.25`
- 输出 500 token,`gpt-4o` `output_ratio = 5.0`

```
input  = 1000 × 1.25 = 1250
output = 500  × 5.0  = 2500
total  = (1250 + 2500) × 1.0 = 3750 quota
       ≈ $0.0075
```

> 假定 `base_quota_per_dollar = 500000`(默认);实际以你登录的实例的 **系统设置 → 模型字典** 为准。

## quota 与人民币 / 美元

- **quota**:内部记账单位,整数,精度极高(避免浮点误差)
- **默认换算**:1 美元 = 500000 quota → 1 quota = $0.000002 = 0.0002 美分
- **人民币**:1 元 ≈ 500000 / 7.2 ≈ 69444 quota(`exchange_rate_cny_per_usd = 7.2`)

部署方可自定义 `base_quota_per_dollar`,看后台 **系统设置 → pricing**。

## 看自己花了多少

| 看什么 | 在哪 |
|---|---|
| 单次请求明细 | 用户前台 → **消费日志** |
| 按天 / 按模型 / 按渠道聚合 | 用户前台 → **仪表盘** |
| 钱包余额 | 用户前台 顶部 |
| 倍率快照(每条日志) | 消费日志详情页 — 可追溯 |

> 倍率"快照"意思是:**当时这条请求扣钱时实际用的 ratio 数字**会写到日志里。即使后台改了 ratio,历史日志仍能准确还原"为什么扣这么多"。

## 余额不足怎么办

- HTTP 状态:`402 Payment Required`
- 错误:`insufficient_quota` / "余额不足,请充值"
- **响应里会带 `X-Request-ID`**,工单时附上

充值方式见 [充值方式](../modules/payment.md):

- 手动充值申请(管理员审批入账)
- 兑换码(管理员发的福利码)
- M2 加在线支付(Stripe / 支付宝 / 微信支付)

## 流式响应的特殊性

流式比较特殊,你的钱会经历:

1. **开始时**:按 `max_tokens` 估算上限,**预扣** 这个数 —— 余额不足直接 402
2. **流中**:边读边累计实际消耗
3. **流末**:按**实际消费**提交,自动**退还多扣**的部分
4. **中途断开**:已发送的部分按实际计费,未发送的退回

这就是为什么"我请求了 4096 max_tokens,实际只回了 200 字,但开始时被扣很多" —— 别担心,流末会退。

## 缓存命中(prompt caching)

OpenAI 的 prompt caching:同一长 prompt 在 5-10 分钟内重复发送,命中缓存的部分上游标记为 `cached_tokens`。

- 命中部分用 **更低倍率**(默认 `cached_ratio = 0.5 × input_ratio`,即 5 折)
- 节省的钱直接反映在你的账单上
- 不需要你做任何特殊配置 —— pro-api 透传上游的 caching 信息

## 推理 token(reasoning tokens)

某些模型(`o1` / `o3` / `deepseek-reasoner`)在"思考"时会产生 `reasoning_tokens`:

- **单独计费**,通常等于或高于 output 倍率
- 你看不到具体的思考内容(上游隐藏),但会被计费
- 适合复杂推理任务(数学 / 代码 / 多步规划),但比 `gpt-4o` 等贵很多

## 退款规则

| 触发 | 是否收钱 |
|---|---|
| 上游 5xx / timeout / 网络错误 | **不收**,自动退预扣 |
| 4xx 客户端错(模型不存在、参数错) | **不收** |
| 上游 429 限流 | 重试别的渠道,最终成功才收 |
| 渠道全部熔断 / 全部 retry 用完 | **不收** |
| 上游正常返回 | 按**实际消费**收 |

## 我能控制什么

- **给令牌设额度上限**(`quota_limit`)防止单个项目超支(见 [API 令牌](../modules/token-system.md))
- **用便宜模型**(`gpt-4o-mini` 比 `gpt-4o` 便宜 17 倍)
- **用 cached prompts**(把固定 system prompt 放前面,触发上游 caching)
- **给不同项目用不同令牌**,方便分账与限流

## 我看不到的

- **pro-api 部署方对你的定价规则**(可能加价或打折,看后台 `pricing_rules`)
- **上游真实价格**(对 pro-api 部署方而言)—— 你只看到你的最终价

如果你是 pro-api 的**部署方**,本节相反:你能看到上游真实价 + 自己的定价策略,但用户只看最终价。

## 关键要点

- **预扣 + 提交**模式:别看到 402 就慌,长流式响应会先按 max_tokens 锁定,完成后退差额
- **缓存命中** 是节省钱的利器:把固定 prompt 放前面,反复用同一前缀
- **推理 token** 单独贵,不要对每个请求都用 `o1` 系列
- 倍率有"快照"机制,历史日志可追溯
