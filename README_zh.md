# proapi

一站式大模型 API 中转网关 · 三协议互转 · 18+ 上游 · 企业级 SSO · 完整计费体系

[English](./README.md) · [文档站](./docs-site)

## 当前状态

🚧 M0 工程脚手架阶段 — 仅骨架可跑,业务功能在 M1 起逐步实现。

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
