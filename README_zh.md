# proapi

一站式大模型 API 中转网关 · 三协议互转 · 20+ 上游 · 企业级 SSO · 完整计费体系

[English](./README.md) · [文档站](./docs-site)

## 当前状态

🚧 **开发中 · 约 M2 阶段**(M0 / M1 已完成,M2 大部分就绪,M3 未开始)。本地可跑通「注册 → 建令牌 → 三协议调用 → 计费 → 看日志 → 充值」全链路。

**已就绪**

- **三协议入口与互转**:OpenAI `/v1`、Anthropic `/v1/messages`、Gemini `/v1beta`,任意入口可路由到任意上游
- **20 个上游适配器**:openai / azure / anthropic / gemini / deepseek / moonshot / 智谱 / 通义 / 豆包 / groq / grok-build / grok-web / mistral / 零一 / openrouter / huggingface / minimax / 腾讯混元 / Cohere / 讯飞星火
- Grok 支持:`grok-build` 使用 xAI API Key;`grok-web` 使用 Grok Web SSO token,存放在渠道或账号的 API Key 字段。
- **多模态**:对话(含流式)· 图像生成 · TTS · STT · Embeddings
- **计费**:Redis Lua 预扣 / 提交 / 退款 · 模型倍率 · 分组倍率
- **渠道**:CRUD + 优先级 + 权重 + 模型映射 + 熔断;账号池
- **限流**:用户 / 令牌 / IP / 模型 / 分组 多维度
- **鉴权**:邮箱密码 + 邮箱验证码 + Session;OAuth 六家(GitHub / Google / 微信 / 飞书 / 钉钉 / Discord)
- **支付**:Stripe / 支付宝 / 微信 + 手动充值 + 兑换码
- **邀请返佣** · **公告** · **系统设置** · **审计**
- **前端**:后台(Naive UI 全页面)+ 用户前台(含 Playground / 模型广场 / 邀请)+ 文档站;中英双语

**进行中 / 未完成**

- 异步任务系统(asynq)与 Midjourney / Suno
- 账号池 OAuth 拉号(PKCE 流程)
- M3 企业版:SSO(OIDC / SAML / LDAP / CAS)、部门预算、审计可视化、OpenTelemetry

路线图见 [`docs/superpowers/2026-05-21-proapi-总体路线图.md`](./docs/superpowers/2026-05-21-proapi-总体路线图.md)。

## 5 分钟跑起来

### 前置

- Go 1.22+
- Node.js 20+ 与 pnpm 9
- Docker(用于起 MySQL / PostgreSQL / Redis)。podman 用户参见下方"podman 适配"。

### 步骤

```bash
git clone https://github.com/ijry/pro-api.git
cd pro-api
make install-tools          # 装 golangci-lint / pnpm / migrate 等
make docker-up              # 起依赖
export PROAPI_MASTER_KEY=$(openssl rand -base64 32)
make dev                    # 并发起后端 + admin + user + docs
```

访问:

- 后端健康检查 → http://127.0.0.1:8080/healthz
- 后台 → http://127.0.0.1:5173
- 用户前台 → http://127.0.0.1:5174
- 文档站 → http://127.0.0.1:5175

### podman 适配

无 docker 但有 podman(macOS):

```bash
export DOCKER_HOST="unix://$(podman machine inspect --format '{{.ConnectionInfo.PodmanSocket.Path}}')"
make docker-up DOCKER_COMPOSE="podman compose"
```

需要 `docker-compose` 二进制(`brew install docker-compose`)与 `~/.docker/config.json` 配 `cliPluginsExtraDirs`。

## 仓库结构

```
proapi/
├── cmd/proapi/         Go 主入口
├── internal/           业务代码
├── pkg/                可被外部 import 的包
├── web/                前端 monorepo(admin / user / shared)
├── docs-site/          VitePress 文档与开源主页
├── migrations/         SQL 迁移
├── deploy/             Dockerfile / docker-compose
└── docs/superpowers/   内部研发文档(路线图 / spec / plan)
```

## 开发

```bash
make test               # 全量测试
make lint               # 全量 lint
make build              # 嵌入式生产二进制
```

## 许可证

[MIT](./LICENSE)
