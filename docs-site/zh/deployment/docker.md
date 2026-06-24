---
title: Docker 部署
outline: deep
---

# Docker 部署

> 单节点 Docker 部署最简流程。多节点见 [Docker Compose 部署](./docker-compose.md)。

## 镜像

- **官方**:`ghcr.io/ijry/pro-api:vX.Y.Z`
- **国内镜像源**:占位(M1 GA 时确定)

### 标签策略

| 标签 | 含义 | 推荐场景 |
|---|---|---|
| `vX.Y.Z` | 精确版本 | **生产** |
| `vX.Y` | minor 滚动(接收 patch) | 测试环境 |
| `vX` | major 滚动 | 体验 |
| `latest` | 最新稳定 | 不推荐生产(可能突然跳大版本) |
| `edge` | 夜间 main | 仅尝鲜 |

生产环境**强烈推荐 pin 到 `vX.Y.Z`**,升级时显式改版本。

## 准备依赖

pro-api 不内嵌 DB / Redis,需要单独起。如果只是测试,可以用 docker-compose:

```yaml
# deps.yaml
version: "3.9"
services:
  postgres:
    image: postgres:14-alpine
    environment:
      POSTGRES_DB: proapi
      POSTGRES_USER: proapi
      POSTGRES_PASSWORD: pass
    ports: ["5432:5432"]
  redis:
    image: redis:7-alpine
    ports: ["6379:6379"]
```

```bash
docker compose -f deps.yaml up -d
```

## 配置文件

参考仓库根 `configs/proapi.example.yaml`，拷贝为 `./config.yaml`:

```yaml
node_id: 0

server:
  addr: ":8080"

log:
  level: info
  format: json

database:
  driver: postgres
  dsn: "postgres://proapi:pass@host.docker.internal:5432/proapi?sslmode=disable"

redis:
  addr: "host.docker.internal:6379"
```

> `host.docker.internal` 让容器访问宿主机端口(Linux 上可能需要 `--add-host=host.docker.internal:host-gateway`)。生产环境应该用真实 DNS / IP。

## 启动命令

```bash
docker run -d \
  --name proapi \
  -p 8080:8080 \
  -v $(pwd)/config.yaml:/app/config.yaml:ro \
  -e PROAPI_MASTER_KEY=$(cat .master_key) \
  -e TZ=Asia/Shanghai \
  --restart unless-stopped \
  ghcr.io/ijry/pro-api:vX.Y.Z
```

参数说明:

- `-v ...:ro`:配置只读挂载,防容器内改
- `-e PROAPI_MASTER_KEY=...`:加密 key,**必填**
- `-e TZ=Asia/Shanghai`:容器内时区(影响日志时间戳,默认 UTC)
- `--restart unless-stopped`:守护重启

## 数据持久化

- **DB / Redis**:用宿主机 volume 或托管服务
- **pro-api 容器本身**:**完全无状态**,随时可删 / 重建
- **日志**:默认写 stdout,用 `docker logs` 或挂日志收集器

## 资源建议

| 量级 | pro-api CPU | pro-api 内存 | DB 磁盘 |
|---|---|---|---|
| 小(< 10 QPS) | 0.5 core | 256 MB | 5 GB |
| 中(10-100 QPS) | 1 core | 512 MB | 50 GB |
| 大(> 100 QPS) | 2+ core × 多实例 | 1 GB / 实例 | 500 GB+ |

pro-api 是 IO-bound,CPU 通常不是瓶颈;内存随 RPS 微涨(大头在 HTTP body + 连接池)。

## 日志收集

pro-api 默认 JSON 结构化日志到 stdout。常见方式:

- `docker logs proapi -f`(开发查看)
- Docker `--log-driver=fluentd` / `json-file` + 收集到 Loki / ELK
- K8s 场景:DaemonSet 跑 fluent-bit

字段:`timestamp` / `level` / `msg` / `request_id` / `user_id` / `model` / `latency_ms` / ...

## 升级

详见 [升级指南](../guide/upgrade.md)。Docker 场景:

```bash
docker pull ghcr.io/ijry/pro-api:vX.Y.Z
docker stop proapi && docker rm proapi
docker run -d ... ghcr.io/ijry/pro-api:vX.Y.Z  # 同启动命令
```

## 多架构

| 平台 | 标签后缀 / 自动选择 |
|---|---|
| `linux/amd64` | 默认 |
| `linux/arm64` | M1/M2 Mac、ARM 服务器 |

docker 默认会选择匹配宿主机架构的 manifest。若强制:

```bash
docker run --platform linux/arm64 ghcr.io/ijry/pro-api:vX.Y.Z
```

## 故障排查

| 现象 | 排查 |
|---|---|
| 容器起不来,exit code 1 | `docker logs proapi`,通常是配置错(`master_key` / `database.dsn`) |
| 起来后 `/healthz` 超时 | 检查 `host.docker.internal` / 网络是否能到 DB |
| 日志报 `failed to verify migration version` | 之前升级 migration 卡 `dirty=true`,人工处理 |
| 容器内时间不对 | 没加 `-e TZ=...`,默认 UTC |
| 镜像拉不下来 | 国内网络问题,换镜像源或自建 registry |
| 流式响应被反代截断 | 见 [反向代理](./reverse-proxy.md),关 buffering |

## 关键要点

- 镜像 entrypoint 默认是 `migrate-then-serve`;高可用场景里**只在 leader 节点跑 migrate**,其他节点用 `--no-migrate` 或环境变量 `MIGRATE=false`
- volume 挂载 `config.yaml` 用 `:ro` 防容器内改
- 容器内时区默认 UTC,需要本地化日志时用 `-e TZ=Asia/Shanghai`
- pro-api **不内嵌** DB / Redis,要单独起
