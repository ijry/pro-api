---
title: 二次开发总览
outline: deep
---

# 二次开发总览

> pro-api 设计上把"协议适配""支付""OAuth""推送"等关键扩展点都接口化。二次开发者可以在不动主仓库的前提下新增厂商或定制业务。

## 仓库布局

```
proapi/
├── cmd/proapi/                主入口(单二进制)
├── internal/                  业务代码(只在仓库内可见)
│   ├── adapter/               厂商适配器(新增厂商看这里)
│   │   ├── openai/
│   │   ├── anthropic/
│   │   ├── gemini/
│   │   ├── deepseek/
│   │   ├── moonshot/
│   │   ├── zhipu/
│   │   ├── qwen/
│   │   ├── doubao/
│   │   ├── azure/
│   │   └── registry.go        provider 注册中心
│   ├── auth/                  邮箱密码 / 验证码 / OAuth
│   ├── billing/               预扣 / 提交 / ledger
│   ├── channel/               选渠道 + 熔断
│   ├── pricing/               倍率匹配 + tokenize
│   ├── ratelimit/             4 维度滑动窗口
│   ├── log/                   request / error / usage 日志
│   ├── payment/               (M2 实现)支付方式
│   ├── protocol/openai/       OpenAI 入口
│   ├── server/                Gin server + middleware + routes
│   └── ...
├── pkg/                       可对外暴露的工具包
│   ├── apierr/                错误码 + i18n
│   ├── snowflake/             ID 生成
│   ├── crypto/                AES-256-GCM 工具
│   └── ...
├── web/admin/                 后台前端(Vue 3)
├── web/user/                  用户前台(Vue 3)
├── docs-site/                 文档站(VitePress)
├── migrations/                SQL 迁移(golang-migrate)
└── docs/superpowers/          内部设计稿(spec)
```

## 开发环境

完整搭建步骤见 [安装 → 源码构建](../guide/installation.md#方式三:源码构建)。

最小命令:

```bash
git clone https://github.com/ijry/pro-api.git
cd pro-api
make install-tools       # golangci-lint / golang-migrate
make docker-up           # 起 MySQL / PG / Redis
export PROAPI_MASTER_KEY=$(openssl rand -base64 32)
make dev                 # 起后端 + admin + user + docs
```

## 你能扩展什么

| 扩展点 | 文档 | 难度 | 是否需改主仓库 |
|---|---|---|---|
| 新增上游厂商适配器 | [adapter-guide](./adapter-guide.md) | 中 | 是 |
| 新增支付方式 | [payment-guide](./payment-guide.md) | 中(M2 实现) | 是 |
| 新增 OAuth provider | 类比 adapter-guide,M2 完整文档 | 易 | 是 |
| 改前端 UI | web/admin / web/user 各自 README | 易 | 否(fork 即可) |
| 新增中间件 | `internal/server/middleware/` | 易 | 是 |
| 自定义 Prometheus 指标 | Prometheus Go client | 易 | 是 |
| 自定义日志 schema | 看 `internal/log` 现有结构 | 中 | 是 |

## 测试与 CI

```bash
make test         # 单测 + 集成测(dockertest)
make lint         # go vet / golangci-lint / eslint
make test-coverage
```

CI 流程见 `.github/workflows/`。提交 PR 必须全部 pass(详见 [贡献指南](./contributing.md))。

## 调试技巧

- **使用 dlv**:`dlv debug ./cmd/proapi -- serve`
- **改 log level**:启动时 `PROAPI_LOG_LEVEL=debug`,会输出每次上游 HTTP 的字段映射
- **看 request_logs / error_logs**:admin 后台 → 日志,按 `X-Request-ID` 搜
- **端到端串 trace**:客户端发请求时主动带 `X-Request-ID: my-trace-001`,pro-api 会全链路用这个 ID

## 贡献

完整流程见 [贡献指南](./contributing.md)。简而言之:

1. fork → 新建 feat/xxx 分支
2. 实现 + 单测
3. `make test && make lint` 全过
4. 提 PR + 描述

## 关键要点

- 接口化的扩展点必须用 Go interface,**不要在 internal 包之间直接 import 具体 struct**
- 新增适配器是 pro-api 二开**最常见的需求**,所以 [adapter-guide](./adapter-guide.md) 写得最详细
- 新增支付方式 M1 没实例,stub 先放
- 任何扩展都要写测试 + 同步更新对应文档站页(`docs-site/zh/*`)
