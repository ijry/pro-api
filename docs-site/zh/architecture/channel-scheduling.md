---
title: 渠道调度与熔断
outline: deep
---

# 渠道调度与熔断

> 当多家上游都能跑同一个模型时(如 4 个 OpenAI 渠道 / 3 个 Anthropic 渠道),proapi 用"优先级 + 加权随机"挑渠道,失败时自动重试别的渠道,连续失败会熔断让该渠道休息。

## 整体流程

```
[请求进入]
      │
      ▼
[Selector.Pick(user, model)]
      │
      │   ─ 候选筛选(active + 兼容 + 未熔断)
      │   ─ 优先级分桶
      │   ─ 加权随机抽样
      │
      ▼
[Adapter.Chat(channel)]
      │
      ├── 成功 ─▶ Reporter.OnSuccess(channel) ─▶ Breaker reset
      │
      └── 失败 ─▶ Reporter.OnFailure(channel, err)
                  │
                  ├── err 计入熔断 → Breaker.Increment
                  │       │
                  │       └── 达阈值 → state=OPEN → PUBLISH 集群
                  │
                  ├── 加入 excluded
                  └── retry 剩余 ? Pick 下一个 : 返回错误
```

## 选渠道算法

### 候选筛选

只保留满足以下条件的渠道:

- `status = 0`(active)
- 熔断状态 `CLOSED` 或 `HALF_OPEN`
- 不在本次重试的 `excluded` 列表里
- `provider` + `model` 在该渠道的模型映射里
- `tags` 与用户分组的 require_tags 匹配(M2 加,M1 简化)

### 优先级选择

先把候选按 `priority` **降序**分桶,取 **最高 priority** 的所有渠道作为 pool。

> `priority` 越大越优先,默认 0。

### 加权随机

在 pool 内按 `weight` 字段做加权随机抽样:

```go
// 简化伪码
pool = highest_priority(candidates)
total = sum(c.weight for c in pool)
r = rand_uniform(0, total)
for c in pool:
    if r < c.weight: return c
    r -= c.weight
```

### 算法示例

3 个 OpenAI 渠道:

| 渠道 | priority | weight |
|---|---|---|
| openai-main-1 | 10 | 3 |
| openai-main-2 | 10 | 1 |
| openai-backup | 5  | 1 |

第一次选:pool = {main-1, main-2}(priority=10);
按 3:1 抽样 → main-1 概率 75%,main-2 概率 25%;
两者都失败后才会落到 backup(下一轮 priority=5 的桶)。

## 故障转移

- M1 默认 `channel.selector.max_retries = 3`(可改)
- 失败渠道加入 `excluded`,**本次请求**不会再选
- 全部尝试完仍失败 → 返回上游错误(透传错误码)+ 提示用户

:::tip 重试与用户超时
重试会显著拉长用户感知的延迟。建议:
- `client_timeout > max_retries × per_call_timeout`
- 或在 token 层禁用重试(M2 加 token 级 `no_retry` 字段)
:::

## 熔断器

### 状态机

```
        ┌──────── connection failures > threshold ────────┐
        │                                                  ▼
   ┌────────┐                                       ┌──────────┐
   │ CLOSED │                                       │   OPEN   │
   │ (正常) │                                       │ (隔离中) │
   └────────┘                                       └──────────┘
        ▲                                                  │
        │                                                  │ cool_down 到期
        │ probe success                                    ▼
        │                                          ┌─────────────┐
        └──────────────────────────────────────────│  HALF_OPEN  │
                                                   │  (半开探针) │
                                                   └─────────────┘
                                                          │
                                                          │ probe failure
                                                          ▼
                                                  (回到 OPEN,cool_down 加倍)
```

### 触发条件

- **计入失败**:5xx / 网络错误 / 超时 / 凭证错(60005)
- **不计入失败**:4xx 客户端错(那是用户的锅,不是渠道的)

### 默认参数(可在 system_settings 改)

| key | 默认 | 说明 |
|---|---|---|
| `channel.breaker.window_seconds` | 60 | 统计窗口 |
| `channel.breaker.fail_threshold` | 5 | 窗口内失败数阈值 |
| `channel.breaker.cool_down_seconds` | 30 | OPEN → HALF_OPEN 时长 |

### 半开探针

`HALF_OPEN` 状态下,**同一时刻只放 1 个请求**做探针:

- 探针成功 → 立即转 `CLOSED`,完全恢复
- 探针失败 → 立即重新 `OPEN`,`cool_down` 翻倍(指数退避,最大 5 分钟)

实现上用 Redis SETNX + TTL 控制"只放一个"。

## 集群广播

熔断状态变化在集群内同步:

```
某节点 Breaker.Trip(channel_id, state=OPEN)
      │
      ▼
PUBLISH proapi:channel:breaker {id: 1234, state: "open", ts: ...}
      │
      ▼
所有节点 SUBSCRIBE → 更新本地缓存
```

这样 `Selector.Pick` 不需要每次查 Redis 拿熔断状态,本地缓存即可。

## 后台运维

- **渠道管理列表**:实时显示 `breaker_state` 列(closed / open / half_open)
- **手动操作**:
  - "禁用" → 等价 status=1,**强制下线**(优先于熔断)
  - "启用" → status=0
  - "重置熔断" → 强制回 CLOSED,清零失败计数
- **测试连通性**:后台 "测试" 按钮,后端调一次最小请求验证凭证(默认用 `gpt-3.5-turbo` 占位模型,见 [渠道管理](../modules/channel-management.md))

## 监控指标

Prometheus 暴露:

```
# 渠道选择计数器
proapi_channel_select_total{provider="openai", channel_id="1001", result="success|fail|breaker_open"}

# 熔断状态 gauge
proapi_channel_breaker_state{channel_id="1001"} 0|1|2  # 0=closed 1=open 2=half_open

# 重试次数 histogram
proapi_channel_retry_count_bucket{le="0|1|2|3|+Inf"}
```

## 关键要点

- 熔断会让那个渠道**完全不被选**,后台展示要醒目(让运维知道哪条线挂了)。
- 半开期只放 1 个探针,所以恢复有滞后(< 1 秒级)。
- `priority` 与 `weight` 语义不同:
  - **priority 越大越优先**,不同优先级**互不混选**
  - **weight 在同一优先级内**做加权随机
- 重试导致总延迟变长,与超时配合调:`total_wait ≤ user_timeout`。
