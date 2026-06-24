---
title: 配置
outline: deep
---

# 配置

> pro-api 的配置分为两层:
>
> - **启动配置**:`config.yaml` + 环境变量,只在进程启动时读一次。
> - **运行时配置**:数据库表 `system_settings`,后台可改、Redis Pub/Sub 集群内 60 秒内热生效。

## 配置加载优先级

```
环境变量(最高) > config.yaml > 默认值
```

启动时按以下顺序查找 yaml:

1. `./config.yaml`(当前工作目录)
2. `/etc/proapi/config.yaml`

无论哪个被命中,环境变量都会覆盖同名字段。

## 启动配置(config.yaml)

### node_id

集群中每个实例的 Snowflake 节点号,取值 `0-1023`。

- 单实例:可不填,默认 0。
- 多实例:**必须每个实例不同**,否则 ID 会冲突。
- env:`PROAPI_NODE_ID`

### master_key

**必填**。用于 AES-256-GCM 加密渠道凭证、OAuth client_secret 等敏感字段。

- 格式:base64 编码的 32 字节随机串
- 生成:`openssl rand -base64 32`
- env:`PROAPI_MASTER_KEY`(优先级高于 yaml,推荐用 env)

:::danger 数据丢失风险
master_key 一旦丢失,所有用它加密的数据都不可解密。请在密码管理器或 HSM 中**永久备份**。换 key = 所有渠道凭证重填。
:::

### server

```yaml
server:
  addr: ":8080"
  read_timeout: 30s
  write_timeout: 600s     # 流式响应需要长 write timeout
  shutdown_timeout: 30s
```

### log

```yaml
log:
  level: info             # debug / info / warn / error
  format: json            # json / text
  access_log_enabled: true
```

### database

```yaml
database:
  driver: postgres        # mysql / postgres
  dsn: "postgres://user:pass@localhost:5432/proapi?sslmode=disable"
  max_open: 50
  max_idle: 10
  max_lifetime: 1h
```

MySQL DSN 示例:

```
user:pass@tcp(127.0.0.1:3306)/proapi?parseTime=true&loc=Local&charset=utf8mb4
```

### redis

```yaml
redis:
  addr: "127.0.0.1:6379"
  db: 0
  password: ""
  pool_size: 50
```

### 完整 config.example.yaml

```yaml
node_id: 0
master_key: ""                    # 用 PROAPI_MASTER_KEY env 注入

server:
  addr: ":8080"
  read_timeout: 30s
  write_timeout: 600s
  shutdown_timeout: 30s

log:
  level: info
  format: json
  access_log_enabled: true

database:
  driver: postgres
  dsn: "postgres://user:pass@localhost:5432/proapi?sslmode=disable"
  max_open: 50
  max_idle: 10
  max_lifetime: 1h

redis:
  addr: "127.0.0.1:6379"
  db: 0
  password: ""
  pool_size: 50
```

## 运行时配置(system_settings 表)

这些 key 存在数据库 `system_settings` 表里,后台 **系统设置** 页可改。改完通过 Redis Pub/Sub 广播,**所有节点 60 秒内本地缓存失效并重新拉取**,无需重启。

按命名空间分组:

### auth.*

| key | 默认值 | 说明 |
|---|---|---|
| `auth.allow_register` | true | 是否允许公网注册 |
| `auth.email_verification_required` | false | 注册后是否需要邮箱验证 |
| `auth.github_oauth.client_id` | "" | GitHub OAuth Client ID |
| `auth.github_oauth.client_secret` | "" | GitHub OAuth Client Secret(自动加密存) |
| `auth.github_oauth.redirect_url` | "" | 留空 = 用当前 host |
| `auth.password.min_length` | 8 | 密码最小长度 |
| `auth.password.require_mixed` | false | 是否要求大小写 + 数字 |

### session.*

| key | 默认值 | 说明 |
|---|---|---|
| `session.ttl_days` | 30 | session 有效期 |
| `session.sliding` | true | 是否滑动延期 |

### token.*

| key | 默认值 | 说明 |
|---|---|---|
| `token.default_rpm_limit` | 0 | 0 = 用 user 默认 |
| `token.default_tpm_limit` | 0 | 同上 |
| `token.prefix_show_len` | 8 | 后台展示 `pa-xxxxxx**...` 的前缀长度 |

### pricing.*

| key | 默认值 | 说明 |
|---|---|---|
| `pricing.base_quota_per_dollar` | 500000 | 1 USD 折算多少 quota |
| `pricing.exchange_rate_cny_per_usd` | 7.2 | CNY/USD,影响人民币充值入账 |

### channel.*

| key | 默认值 | 说明 |
|---|---|---|
| `channel.selector.mode` | priority+weight | 选渠道算法 |
| `channel.selector.max_retries` | 3 | 故障转移次数 |
| `channel.breaker.window_seconds` | 60 | 熔断统计窗口 |
| `channel.breaker.fail_threshold` | 5 | 窗口内失败阈值 |
| `channel.breaker.cool_down_seconds` | 30 | OPEN → HALF_OPEN 时长 |

### ratelimit.*

| key | 默认值 | 说明 |
|---|---|---|
| `ratelimit.enabled` | true | 全局开关(故障应急可关) |
| `ratelimit.user_default_rpm` | 60 | 用户默认 RPM |
| `ratelimit.user_default_tpm` | 100000 | 用户默认 TPM |
| `ratelimit.ip_rpm` | 600 | 单 IP /24 的 RPM |
| `ratelimit.model_default_rpm` | 0 | 0 = 不限 |
| `ratelimit.model_default_tpm` | 0 | 同上 |

### log.*

| key | 默认值 | 说明 |
|---|---|---|
| `log.retain_request_days` | 30 | 请求日志保留天数 |
| `log.retain_error_days` | 90 | 错误日志保留天数 |
| `log.flush_interval_ms` | 500 | 异步落库 flush 间隔 |

### notice.*

| key | 默认值 | 说明 |
|---|---|---|
| `notice.show_max` | 5 | 用户前台最多展示几条公告 |

### billing.*

| key | 默认值 | 说明 |
|---|---|---|
| `billing.reserve_ttl_seconds` | 600 | reservation 自动过期回收 |
| `billing.reconcile_interval_seconds` | 30 | reconcile job 频率 |

### manual_recharge.*

| key | 默认值 | 说明 |
|---|---|---|
| `manual_recharge.enabled` | true | 是否开放手动充值申请 |
| `manual_recharge.min_amount_cny` | 100 | 单次最小金额(元) |
| `manual_recharge.max_amount_cny` | 100000 | 单次最大金额(元) |

### redeem.*

| key | 默认值 | 说明 |
|---|---|---|
| `redeem.default_expires_days` | 365 | 兑换码默认有效期 |

## 通过 API 改配置

```bash
curl -X PATCH https://api.example.com/api/admin/settings/ratelimit.user_default_rpm \
  -H "Cookie: session=..." \
  -H "X-CSRF-Token: ..." \
  -d '{ "value": 120 }'
```

完整 API 见 [管理 API 索引](../api/admin-api.md#系统设置)。

## 环境变量速查表

| 用途 | env 名 | 等价 yaml 字段 |
|---|---|---|
| 加密 master key | `PROAPI_MASTER_KEY` | `master_key` |
| 节点号 | `PROAPI_NODE_ID` | `node_id` |
| DB DSN | `PROAPI_DATABASE_DSN` | `database.dsn` |
| Redis 地址 | `PROAPI_REDIS_ADDR` | `redis.addr` |
| Redis 密码 | `PROAPI_REDIS_PASSWORD` | `redis.password` |
| 日志级别 | `PROAPI_LOG_LEVEL` | `log.level` |
| 监听地址 | `PROAPI_SERVER_ADDR` | `server.addr` |

## 关键要点

- 启动配置**不读** `system_settings`(它本身需要 db 才能读)—— 所以 `database` / `redis` 必须在 yaml 或 env。
- 运行时配置变更通过 Redis Pub/Sub 在集群内 60s 内传播,**不需要重启**。
- 写敏感字段(如 `auth.github_oauth.client_secret`)时,服务端会自动用 master_key 加密存,后台展示为 `*****`。
- 改 `master_key` 等同重新部署:旧加密数据全部不可解密。
