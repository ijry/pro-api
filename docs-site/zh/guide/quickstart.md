# 5 分钟跑起来

> M0 阶段为骨架,以下命令在 M1 起完整可用。

## 环境要求

- Go 1.22+
- Node.js 20+ 与 pnpm 9
- Docker(可选,用于启动 MySQL / PostgreSQL / Redis)

## 起依赖

```bash
make docker-up
```

## 起后端

```bash
export PROAPI_MASTER_KEY=$(openssl rand -base64 32)
make dev-backend
```

访问 [http://127.0.0.1:8080/healthz](http://127.0.0.1:8080/healthz),应返回 `{"status":"ok",...}`。

## 起前端

```bash
make dev-admin      # 后台 → http://127.0.0.1:5173
make dev-user       # 前台 → http://127.0.0.1:5174
make dev-docs       # 文档站 → http://127.0.0.1:5175
```

## 一键全起

```bash
make dev
```
