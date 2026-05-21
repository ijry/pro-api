---
title: 安装
outline: deep
---

# 安装

> proapi 提供三种安装方式:**Docker(推荐)**、**预编译二进制**、**源码构建**。本页介绍各自的步骤、适用场景与故障排查。

## 选哪种?

- **零运维 / 想最快上线**:选 Docker。镜像里已封装迁移与启动入口。
- **轻量 VPS 不想装 Docker / 想绑定 systemd**:选预编译二进制。
- **打算二次开发 / 想跑非稳定版**:选源码构建。

## 方式一:Docker 一键安装

### 拉取镜像

```bash
docker pull ghcr.io/proapi/proapi:latest
```

**标签策略**:

| 标签 | 含义 | 推荐场景 |
|---|---|---|
| `vX.Y.Z` | 精确版本 | 生产 |
| `vX.Y` | minor 滚动(自动接收 patch) | 测试环境 |
| `vX` | major 滚动 | 体验 |
| `latest` | 最新稳定 | 不推荐生产 |
| `edge` | 夜间 main | 仅做尝鲜 |

### 准备配置文件

参考仓库根下 `config.example.yaml`,拷一份到 `./config.yaml`。
最关键的几个字段:

- `database.driver` / `database.dsn` —— 指向你的 MySQL/PG
- `redis.addr` —— 指向你的 Redis
- `master_key` —— 通过环境变量 `PROAPI_MASTER_KEY` 注入,**不要直接写到文件**

### 启动

```bash
docker run -d \
  --name proapi \
  -p 8080:8080 \
  -v $(pwd)/config.yaml:/app/config.yaml:ro \
  -e PROAPI_MASTER_KEY=$(openssl rand -base64 32) \
  --restart unless-stopped \
  ghcr.io/proapi/proapi:latest
```

:::tip 容器入口
Docker 镜像的 entrypoint 默认是 `migrate-then-serve` —— 启动前自动执行 `proapi migrate up`,
再起 HTTP 服务。多实例集群里只在一个 leader 节点跑 migrate(用 `--no-migrate` 跳过)。
:::

### 验证

```bash
curl http://localhost:8080/healthz
```

期望响应:

```json
{ "status": "ok", "version": "vX.Y.Z", "build_time": "2026-..." }
```

## 方式二:预编译二进制

### 下载

到 [GitHub Releases](https://github.com/ijry/pro-api/releases) 下载对应平台的归档。

| 平台 | 归档名 |
|---|---|
| Linux x86_64 | `proapi_linux_amd64.tar.gz` |
| Linux ARM64 | `proapi_linux_arm64.tar.gz` |
| macOS Intel | `proapi_darwin_amd64.tar.gz` |
| macOS Apple Silicon | `proapi_darwin_arm64.tar.gz` |
| Windows | `proapi_windows_amd64.zip` |

```bash
wget https://github.com/ijry/pro-api/releases/download/vX.Y.Z/proapi_linux_amd64.tar.gz
sha256sum -c proapi_linux_amd64.tar.gz.sha256
tar -xzf proapi_linux_amd64.tar.gz
chmod +x proapi
```

### 准备依赖

- MySQL ≥ 8.0 或 PostgreSQL ≥ 14
- Redis ≥ 6
- 一个空数据库,授权 proapi 账号可建表

### 配置

把 `config.yaml` 放到 `./config.yaml` 或 `/etc/proapi/config.yaml`,proapi 启动时自动读取。

### 跑迁移

```bash
./proapi migrate up
```

### 启动

直接前台运行:

```bash
PROAPI_MASTER_KEY=$(cat /etc/proapi/master.key) ./proapi serve
```

或注册为 systemd unit(示例 `/etc/systemd/system/proapi.service`):

```ini
[Unit]
Description=proapi gateway
After=network.target

[Service]
Type=simple
User=proapi
WorkingDirectory=/opt/proapi
EnvironmentFile=/etc/proapi/env
ExecStart=/opt/proapi/proapi serve
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

`/etc/proapi/env` 内容:

```bash
PROAPI_MASTER_KEY=...base64-32-bytes...
PROAPI_DATABASE_DSN=...
PROAPI_REDIS_ADDR=...
```

## 方式三:源码构建

### 环境要求

- Go 1.22+
- Node.js 20+ / pnpm 9
- make / git

### 构建步骤

```bash
git clone https://github.com/ijry/pro-api.git
cd pro-api
make install-tools       # 拉 golangci-lint / golang-migrate 等
make build               # 输出 ./bin/proapi
make web-build           # 构建 admin + user
```

### 部署

把 `./bin/proapi` 与 `web/admin/dist`、`web/user/dist` 一并部署到服务器,启动方式同方式二。

## 验证安装

### `/healthz` 探活

返回:

```json
{
  "status": "ok",
  "version": "vX.Y.Z",
  "build_time": "2026-05-22T10:00:00Z",
  "node_id": 0
}
```

### `/metrics` Prometheus 端点

默认在 `:8080/metrics` 开放。生产环境建议加 IP 白名单,见
[反向代理](../deployment/reverse-proxy.md)。

### 后台首次登录

打开 `http://<your-host>/admin`,注册第一个账户。它会自动晋升为超级管理员,
之后所有运维操作都从这里开始。

## 常见故障

- **`Error: PROAPI_MASTER_KEY required`** —— 未设置环境变量。见 [配置](./configuration.md#master-key)。
- **`Error: redis: dial tcp ...: connection refused`** —— Redis 未启动或 host/port 错。
- **`Error: failed to verify migration version`** —— migration 状态错乱,见 [升级指南](./upgrade.md)。
- **Docker 镜像拉不下来** —— 国内网络可改用 GHCR 镜像加速或自己 build。
- **ARM Mac 拉到的镜像跑不起来** —— 显式指定 `--platform linux/arm64`。

:::warning 任何方式都要先 migrate 再 serve
Docker 镜像入口已封装为 `migrate-then-serve`;裸二进制 / 源码构建需要你**手动**先跑 `./proapi migrate up`,否则启动时会报"missing migrations"。
:::
