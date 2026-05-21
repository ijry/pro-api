---
title: 架构总览
outline: deep
---

# 架构总览

> proapi 是一个"协议网关":接收 OpenAI 格式的 API 请求,挑选最合适的上游渠道,做协议转换、计费、限流、日志,然后把上游结果(可能是流式)透传给客户端。
>
> 本页给你一个 10 分钟看完的整体视图;具体子系统看各架构子页。

## 一张图看 proapi

```
                                  ┌─────────────────────────────┐
   Client                         │  proapi                     │
   (OpenAI SDK,                   │  ┌─────────────────────┐    │
    curl, app)                    │  │ /v1/chat/completions│    │
       │                          │  └──────────┬──────────┘    │
       │  POST /v1/chat            │             │               │
       │   Authorization: pa-xxx   │             ▼               │
       ├──────────────────────────▶│  ┌────────────────────┐    │
       │                          │  │  middleware chain  │    │       Upstream
       │                          │  │  - tokenAuth       │    │       ┌────────────┐
       │                          │  │  - rateLimit       │    │       │  OpenAI    │
       │                          │  │  - billing.Reserve │    │       │  Anthropic │
       │                          │  └─────────┬──────────┘    │       │  Gemini    │
       │                          │             │               │       │  DeepSeek  │
       │                          │             ▼               │       │  …         │
       │                          │  ┌────────────────────┐    │       └────────────┘
       │                          │  │ channel selector + │ ──▶│ adapter ─▶ HTTP
       │                          │  │  breaker           │    │
       │                          │  └─────────┬──────────┘    │
       │  SSE stream / JSON       │             │               │
       │◀─────────────────────────┤             ▼               │
       │                          │  ┌────────────────────┐    │
       │                          │  │ billing.Commit +   │    │
       │                          │  │ log.AsyncWrite     │    │
       │                          │  └────────────────────┘    │
       │                          └─────────────────────────────┘
```

## 核心模块

每个模块都有独立子页深入,本页只做名词解释 + 入口链接。

- **协议入口**(`internal/protocol/openai`):OpenAI 风格 HTTP 处理,负责解析请求 / 编码响应(尤其是 SSE)。
- **适配器层**(`internal/adapter/*`):一个 provider 一个包,把内部 IR 转成厂商专属 HTTP。见 [适配器层](./adapter-layer.md)。
- **渠道层**(`internal/channel/{selector,breaker}`):多渠道选择 + 熔断。见 [渠道调度与熔断](./channel-scheduling.md)。
- **计费层**(`internal/billing` + `pricing` + `wallet`):预扣 / 提交 / 退款 / 账本。见 [计费机制](./billing.md)。
- **限流层**(`internal/ratelimit`):4 维度 Redis 滑动窗口。见 [限流策略](./ratelimit.md)。
- **日志层**(`internal/log`):请求日志 / 错误日志 / 用量统计聚合(async writer)。
- **认证层**(`internal/auth` + `token`):session + token,中间件层完成鉴权。
- **前端**(`web/admin` 后台 / `web/user` 用户前台):Vue 3 + Vite + shadcn-vue 风格 UI。
- **文档站**(`docs-site`):VitePress 静态站,可与 proapi 同实例服务,也可独立部署。

## 数据流(代理一次请求)

一次成功的 `chat/completions` 流程(简化):

1. **接入**:Gin 接到 `POST /v1/chat/completions`,middleware chain 起作用。
2. **鉴权**:`tokenAuth` 从 `Authorization: Bearer pa-xxx` 解出 token,查 cache 拿 user。
3. **限流**:`rateLimit` 并行检查 user/token/ip/model 4 个维度,任一超限直接 429。
4. **预估**:`pricing.EstimateMax(model, in_tokens, max_out_tokens)` 算出最大 quota。
5. **预扣**:`billing.Reserve` 走 Lua 在 Redis 原子扣 wallet,余额不足 402。
6. **选渠道**:`channel.Selector.Pick(user, model)` 按 priority+weight 算法挑一个 active channel。
7. **调上游**:`adapter.Chat` / `ChatStream` 用 IR 调 provider HTTP。
8. **回流**:流式时,逐 chunk 通过 OpenAI SSE encoder 推给客户端;非流式攒齐再返。
9. **提交**:拿到 usage 后 `billing.Commit` Lua 原子写实际消耗、退还差额。
10. **落账**:`log.AsyncWrite` 入队,后台 writer 批量写 `request_logs` + `ledger_entries`。

整体客户端可观察到的延迟 ≈ TTFT (上游) + 几毫秒 proapi 开销。

## 数据流(后台改配置)

`system_settings` 变更走如下路径:

```
admin UI ──▶ PATCH /api/admin/settings/:key
                │
                ▼
       repo.UpsertSetting()   (DB UPSERT)
                │
                ▼
       redis PUBLISH proapi:settings { key, version }
                │
                ▼
   ┌──────────┴──────────┐
   │                     │
   ▼                     ▼
 node-1 invalidate     node-N invalidate
   local cache           local cache
   (LRU + lazy reload from DB)
```

所有节点在 60 秒内反映新值,无需重启。

## 技术栈速查

| 层 | 选型 |
|---|---|
| 语言 | Go 1.22(后端) + Vue 3 / TypeScript(前端) |
| Web 框架 | Gin |
| ORM | GORM |
| DB | MySQL ≥ 8 / PostgreSQL ≥ 14 |
| 缓存 | Redis ≥ 6 |
| 配置 | Viper |
| 日志 | zap |
| 前端构建 | Vite + pnpm workspace |
| UI 组件 | shadcn-vue 风格 |
| 文档站 | VitePress |
| 监控 | Prometheus(M3 加 OpenTelemetry) |
| 测试 | go test + testify + dockertest(集成) |

## 部署形态

- **单实例**:`proapi` + MySQL/PG + Redis,适合 < 100 QPS 的内部场景。
- **多实例水平扩展**:多个 `proapi` 共享 DB / Redis,前面挂反代做负载均衡。
  - 唯一约束:每个实例 `node_id` 不能撞。
  - Redis 单点可承 5w+ TPS,M1 范围内通常够用。
- **托管依赖**:DB 与 Redis 用云托管(RDS / ElastiCache),proapi 容器化跑 K8s。

详细方案见 [Docker Compose 部署](../deployment/docker-compose.md)。

## 与同类项目的差异点

| 维度 | proapi | one-api / new-api |
|---|---|---|
| 协议互转 | M1 出向 9 家 + M2 三入口互转 | 多家 + OpenAI 入口 |
| 计费 | 预扣 + Lua 原子 + append-only ledger | 后扣 |
| 限流 | 4 维度滑动窗口 | 较粗 |
| 企业 SSO | M3 OIDC/SAML/LDAP/CAS | 无 |
| 前端 | Vue3 + 现代组件库 | 旧版 |
| 文档站 | VitePress 独立站,持续维护 | 较薄 |

M1 阶段差距仍在补;M2 之后会拉开。

## 范围提示(M1 vs M2/M3)

| 能力 | M1 | M2 | M3 |
|---|---|---|---|
| OpenAI 入口 | ✅ | | |
| Anthropic / Gemini 入口 | | ✅ | |
| 9 家适配器(出向) | ✅ | | |
| 18+ 适配器 | | ✅ | |
| GitHub OAuth | ✅ | | |
| Google / 微信 / 飞书 等 OAuth | | ✅ | |
| OIDC / SAML / LDAP / CAS | | | ✅ |
| 手动充值 / 兑换码 | ✅ | | |
| Stripe / 支付宝 / 微信支付 | | ✅ | |
| 部门 / 预算 / 配额 | | | ✅ |

> 本文档站描述的能力,文末或正文都会标 milestone。请先看清范围再做架构决策。
