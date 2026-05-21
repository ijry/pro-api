---
title: 计费机制
outline: deep
---

# 计费机制

> proapi 的计费基于 **预扣 + 提交**(Reserve + Commit)模式。所有钱包变动通过 Redis Lua 脚本原子完成,辅以 DB 账本(`ledger_entries`)做最终一致对账与审计。

## 关键概念

| 术语 | 含义 |
|---|---|
| **quota** | 内部记账单位,整数,无小数,精度极高 |
| **wallet** | 用户的钱包余额,Redis 热数据 + DB cold mirror |
| **reservation** | 预扣(临时锁定的 quota),有 TTL,过期自动 refund |
| **ratio** | 倍率,float,某个 token 维度的乘数 |
| **ledger** | append-only 账本,所有 quota 变动一行一条 |

## 计费单位

```
1 美元 = base_quota_per_dollar quota
        默认 500000
        → 1 quota = 0.002 美分(0.000002 USD)
```

`base_quota_per_dollar` 是运行时配置(`system_settings.pricing.base_quota_per_dollar`),改它会**影响所有用户**,慎重。

ratio 是"每 token 多少 quota"的乘数。例:

- gpt-4o input ratio = 1.25 → 输入 1k tokens 消耗 `1000 × 1.25 = 1250 quota = $0.0025`
- gpt-4o output ratio = 5.0 → 输出 1k tokens 消耗 `1000 × 5.0 = 5000 quota = $0.01`

价格表见 [模型价格表](../billing-guide/model-pricing.md)。

## 预扣 / 提交流程

```
[请求进入]
    │
    │  1. tokenize 入参 → input_tokens
    │  2. pricing.EstimateMax(model, in_tokens, max_out_tokens) → est_quota
    │
    ▼
[Reserve.lua]                     原子:wallet.balance -= est_quota
                                       记 reservation(req_id, est_quota, ttl=600s)
    │
    │   余额不足 → 返回 402
    │   余额够 → 继续
    │
    ▼
[选渠道 + 调上游]
    │
    │   失败 → Refund.lua(整笔退预扣)→ 返回错误
    │   成功 → 拿到 actual usage
    │
    ▼
[Commit.lua]                      原子:
                                       wallet.consumed += actual_quota
                                       wallet.balance  += (est - actual)  -- 退差额
                                       reservation 标记 committed
    │
    ▼
[async ledger writer]             批量写 ledger_entries(append-only)
[async log writer]                批量写 request_logs(含倍率快照)
```

## 倍率匹配顺序

`pricing.Match(user_group, model, channel)` 按从精确到通用的优先级:

1. **channel_model_mappings** 该渠道该模型的 ratio 覆盖(仅作用于该渠道)
2. **pricing_rules** scope=`group_model`(用户分组 + 模型)
3. **pricing_rules** scope=`model`(全局对该模型)
4. **pricing_rules** scope=`group`(全局对该分组)
5. **pricing_rules** scope=`global`(系统全局)
6. **model_catalogs.default_*_ratio**(模型字典默认值)
7. 兜底 `1.0`

`group_ratio` 是独立一层,匹配到 user_group 就生效,与上面 4 个 token ratio **相乘**。

## 流式响应如何计费

流式比较特殊:开始时还不知道 output 长度。

1. **开始时**:按 `max_tokens`(请求里的字段,默认 4096)估算上限,`Reserve` 这个数。
2. **流中**:边读边累计 `output_tokens`(从 chunk.delta 估算)。
3. **流末**:上游通常会在最后一个 chunk 带完整 usage(`stream_options.include_usage=true`);若没有就用本地累计。
4. **Commit**:按实际 `output_tokens` 算 quota,自动退多扣的差额。

中途客户端断开时:`relay` 检测到 `ctx.Done()`,按已读的部分 `Commit`。

## 退款规则

退款只有这几种情况:

| 触发 | 行为 |
|---|---|
| 选渠道失败 / 全部 retry 用完 | `Refund.lua` 整笔退 |
| 上游 5xx / timeout / 网络错误 | `Refund.lua` 整笔退 |
| 模型不存在 / 凭证失效(60002 / 60005) | `Refund.lua` 整笔退(`channel` 计入熔断) |
| 上游 4xx 客户端错(请求字段错) | `Refund.lua` 整笔退(用户的锅,但不收钱) |
| 上游正常返回 | `Commit.lua` 按实际扣,退差额 |

:::info M1 不实现"主动退款"
管理员目前**无法**通过后台对历史请求做主动退款。M2 增加该能力,会走专门的 ledger 调账接口。
:::

## 对账与回收

- **Reservation TTL**:默认 10 分钟(`billing.reserve_ttl_seconds = 600`)。
- **Reconcile job**:每 30 秒扫一次过期未 commit 的 reservation,自动 `Refund.lua`。

这是个兜底机制:防止程序崩溃 / 网络丢包导致钱被永久锁住。

## 一致性保证

- **Redis Lua 原子** → 单 Redis 实例内所有钱包操作串行化,无 race。
- **DB ledger append-only** → 完整审计轨迹,**禁止 UPDATE/DELETE**;改账户用反向 ledger 抵消。
- **Wallet.balance 双写**:Redis 是源头(实时),DB 是 cold mirror(异步对账)。
- **极端故障**(Redis 数据丢失):可从 DB ledger 重放 quota 余额(M2 提供工具)。

:::warning Redis Cluster 模式注意
若用 Redis Cluster,`wallet:{user_id}` 与 `reservation:{user_id}:{req_id}` 必须用 hashtag 绑到同一槽位,否则 Lua 多 key 操作会拒绝。
:::

## 多用户并发

- 单 wallet 高并发 → Lua 在 Redis 单线程下序列化,无并发问题。
- 不同 wallet 完全并行,无相互锁。
- Redis 单实例足以承载 5w+ TPS,M1 不需要分片。

## 关键要点

- **预扣**是为了避免长流式响应中余额耗尽却没法停止 —— 提前锁定上限。
- Lua 脚本必须**幂等**:`Commit.lua` 重复调用应返回错误而不是再扣一次。
- ledger append-only 是审计基础,**任何变动都新增一行**,不要 UPDATE。
- 跨实例时,Redis Cluster 需要 hashtag 绑同一槽。

更详细的用户视角见 [计费如何工作](../billing-guide/how-billing-works.md);定价配置见 [定价与倍率](../modules/pricing.md)。
