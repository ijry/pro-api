---
title: 常见问题
outline: deep
---

# 常见问题(FAQ)

## 1. proapi 与 one-api / new-api 有什么区别

简短对比(M1 完整版):

| 维度 | proapi | one-api / new-api |
|---|---|---|
| 协议互转 | M1 出向 9 家 + M2 三入口互转 | 多家 + OpenAI 入口 |
| 计费 | 预扣 Lua 原子 + append-only ledger | 后扣 |
| 限流 | 4 维度滑动窗口 | 较粗 |
| 企业 SSO | M3 OIDC / SAML / LDAP / CAS | 无 |
| 前端 | Vue 3 + 现代组件库 | 旧版 |
| 文档 | VitePress 独立站持续维护 | 较薄 |

M1 阶段差距在补;M2/M3 后会拉开。

## 2. 是否支持 Anthropic / Gemini 入口

**M1 仅 OpenAI 入口**。Anthropic / Gemini 入口在 M2 加。**出向**(后端调用)9 家都支持。

也就是说:你的客户端目前用 OpenAI SDK 调 proapi,proapi 可以路由到 Anthropic / Gemini 等上游;但客户端不能用 Anthropic SDK 直接调 proapi。

## 3. 多个用户共享一个 GitHub 账号能行吗

**不能**。`oauth_bindings` 表上 `UNIQUE (provider, provider_uid)`,一个 GitHub 账户严格对应一个 proapi 账户。

若需要换绑:先在原 proapi 账号 → 账号设置 → 解绑 GitHub,然后在新账号上绑定。

## 4. 调用突然 402 Insufficient Quota

按以下顺序排查:

1. **钱包余额**:用户前台 → 仪表盘,看 `wallet.balance` 是否 ≤ 0
2. **令牌额度**:令牌可能设了 `quota_limit`,已用完
3. **流式预扣**:长流式请求会按 `max_tokens` 估算预扣,可能比想象多。等流末退差额或调小 `max_tokens`
4. **倍率改动**:管理员可能近期调高了 ratio

## 5. 如何提高单令牌的 RPM / TPM

- **令牌级别**:用户前台 → 令牌 → 编辑 → 设 `rpm_limit` / `tpm_limit`(`0 = 用 user 默认`,`> 0 = 覆盖`)
- **用户级别**:由管理员在系统设置 `ratelimit.user_default_rpm` / `_tpm` 改
- **系统级别**:同上,影响所有用户的默认值

完整限流机制见 [限流策略](./architecture/ratelimit.md)。

## 6. 渠道熔断后多久会自动恢复

默认 `cool_down_seconds = 30` 秒。

熔断后:

- 30s 后进入 HALF_OPEN 状态,放 1 个探针
- 探针成功 → 立即转 CLOSED 完全恢复
- 探针失败 → 重新 OPEN,cool_down 加倍(指数退避,上限 5 分钟)

也可以后台 → 渠道管理 → 手动 **重置熔断**。

## 7. master_key 丢了怎么办

**灾难场景**。所有用 master_key 加密的数据(渠道凭证 / OAuth client_secret 等)都将**无法解密**,等于全部要重填。

:::danger 强烈建议
**永远把 master_key 备份到密码管理器或 HSM**。换 key 没有"自动迁移"机制,旧数据全部作废。
:::

## 8. 支持多实例集群部署吗

支持。前提:

- **同一份 `master_key`**(否则加密数据不能跨实例)
- **同一个 DB / Redis**(数据共享)
- **每个实例 `node_id` 唯一**(否则 Snowflake ID 冲突)

详细方案见 [Docker Compose 部署](./deployment/docker-compose.md) 与 [反向代理](./deployment/reverse-proxy.md)。

## 9. 默认有哪些用户分组

seed 阶段会写入 3 个:

| 分组 | group_ratio | 用途 |
|---|---|---|
| `default` | 1.0 | 默认所有新注册用户 |
| `vip` | 0.8 | 8 折,运维手动晋升 |
| `svip` | 0.6 | 6 折,高价值用户 |

可改名、加分组、改 ratio。如果不需要分组,把所有用户挂 default 即可,管理面板会少几列。

## 10. 文档站可以离线部署吗

可以。`pnpm --filter docs-site build` 后,`docs-site/.vitepress/dist/` 是纯静态文件,丢任意 web server(Nginx / Caddy / GitHub Pages / S3 静态网站)即可。

也可以与 proapi 同实例服务 —— 默认 `proapi` 进程会把 `docs-site/dist` 挂到 `/docs` 路径(M2 完整支持)。

## 11. 流式响应被 Nginx 截断

`proxy_read_timeout` 默认 60s,长流式响应会被截断。改成 600s+:

```nginx
proxy_read_timeout 600s;
proxy_send_timeout 600s;
proxy_buffering off;
```

完整反代配置见 [反向代理](./deployment/reverse-proxy.md)。

## 12. 怎么对接监控

proapi 在 `:8080/metrics` 暴露 Prometheus 指标。生产建议:

- Prometheus 抓 `/metrics`(配 `allow` 限制源 IP)
- Grafana 展示官方 dashboard(M2 发布)
- 告警关键指标:`proapi_channel_breaker_state` / `proapi_billing_reserve_fail_total` / `proapi_ratelimit_deny_total`

## 13. 如何更换上游模型 ratio

后台 → **系统设置 → 模型字典 → 编辑某模型** → 改 `default_*_ratio`。

更精细的覆盖(按分组 / 按渠道 / 按用户组+模型)用 `pricing_rules`,详见 [定价与倍率](./modules/pricing.md)。

## 关键要点

- 上述问题覆盖最高频的运维 / 用户 / 部署疑问
- 每个问题答案 1-3 段,避免冗长
- 链回相关 docs 章节,本页作为 FAQ "hub"
- 找不到答案?[在 GitHub Discussions 提问](https://github.com/ijry/pro-api/discussions)
