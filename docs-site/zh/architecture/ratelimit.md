---
title: 限流策略
outline: deep
---

# 限流策略

> pro-api 用 Redis 滑动窗口算法实现 **4 个维度** 的限流,所有维度并行检查,任一维度超限即拒绝。

## 4 个维度

| 维度 | Redis key 命名 | 默认值 | 谁配 |
|---|---|---|---|
| user RPM | `ratelimit:user:{id}:rpm` | 60 | `system_settings.ratelimit.user_default_rpm` |
| user TPM | `ratelimit:user:{id}:tpm` | 100000 | `system_settings.ratelimit.user_default_tpm` |
| token RPM | `ratelimit:token:{id}:rpm` | 继承 user | `api_tokens.rpm_limit` |
| token TPM | `ratelimit:token:{id}:tpm` | 继承 user | `api_tokens.tpm_limit` |
| ip RPM | `ratelimit:ip:{ip24}:rpm` | 600 | `system_settings.ratelimit.ip_rpm` |
| model RPM | `ratelimit:model:{model}:rpm` | 0 (不限) | `system_settings.ratelimit.model_default_rpm` |
| model TPM | `ratelimit:model:{model}:tpm` | 0 (不限) | `system_settings.ratelimit.model_default_tpm` |

> "user" 与 "token" 是两个独立维度 —— 同一 user 多个 token 都会受 user 维度限,但每个 token 也有自己的 RPM/TPM。

## 滑动窗口算法

每个限流 key 对应一个 Redis `ZSET`:

```
key = ratelimit:user:1001:rpm
window = 60 秒

每次请求:
  ZADD key now_ms  member="now_ms-rand"
  ZREMRANGEBYSCORE key 0 (now_ms - window_ms)
  count = ZCARD key
  if count > limit: reject
```

TPM 维度的 `cost > 1`(等于 token 数),用多个 ZADD 拼出 cost。

整套操作在一个 Lua 脚本里原子完成,避免 race。

## RPM vs TPM

| 维度 | cost | 例子 |
|---|---|---|
| **RPM**(每分钟请求数) | 1 / 请求 | RPM=60 → 1 分钟最多 60 次调用 |
| **TPM**(每分钟 token 数) | `input + max_out` / 请求 | TPM=100000 → 1 分钟最多 10w token 流量 |

TPM 在请求**进入时**预扣(`input_tokens + max_out_tokens` 估算值),
**不会**因为实际消费少了而退还配额 —— 这是有意的,避免被恶意构造大 `max_tokens` 占额度的情况下绕过限流。

## 优先级

```
token-level limit > user-level limit
ip-level   limit:  独立维度,不与 user/token 互替代
model-level limit: 独立维度
```

两者都生效时取**更严**的(`min(token_limit, user_limit)`)。

## 触发限流后

```
HTTP/1.1 429 Too Many Requests
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1716396000
Retry-After: 23

{
  "error": {
    "message": "rate limit exceeded: user RPM",
    "type": "rate_limit_exceeded",
    "code": "rate_limit_exceeded"
  }
}
```

字段 `type` 与 OpenAI 兼容,主流 SDK(如 OpenAI Python `RateLimitError`)能识别并自动重试。

## 全局开关

```
system_settings.ratelimit.enabled = false
```

可以在故障应急时关闭**所有限流**(放行所有请求)。一般用于:

- 限流 Lua 脚本上线后发现 bug
- Redis 慢导致大量 false positive
- 大促 / 压测场景

## IP 维度的取值

- **IPv4**:取 `/24` 截断(如 `1.2.3.45` → `ratelimit:ip:1.2.3:rpm`)
  - 目的:减少误伤同公司 NAT 出口的多用户
  - 副作用:同办公网用户共享配额,需要时调高 `ip_rpm`
- **IPv6**:取完整地址(IPv6 地址空间大,不需要聚合)

:::warning 反代必须透传 X-Real-IP
pro-api 用 `X-Real-IP` / `X-Forwarded-For` 取真实 IP。如果反代没透传,**所有请求会被算到同一个反代 IP**,触发误限流。详见 [反向代理](../deployment/reverse-proxy.md)。
:::

## 关键运维

- **Redis 故障**:限流模块设计为 **fail-open**(放行)—— Redis 不可用时不阻断业务。
- **阈值热调**:`PATCH /api/admin/settings/ratelimit.*` 改完 60s 内集群生效,**不需要重启**。
- **观察**:Prometheus 指标 `proapi_ratelimit_check_total{dimension, result=allow|deny}`。
- **绕过 / 提额**:为大客户单独建 token,把 token 的 `rpm_limit` / `tpm_limit` 调高(绕开 user 默认)。

## 当前阶段(M1)

- 单 Redis 实例可承 5w+ ZADD QPS,M1 单点足够。
- M2 会引入**本地预扣 + 异步上报**模式,减少 Redis 压力。
- M3 引入 Redis Cluster 模式 + 按 user_id 哈希分片。
