---
title: Docker Compose 部署
outline: deep
---

# Docker Compose 部署

> 把 proapi + DB + Redis + Nginx 一把 compose 起,适合**单机或小集群**快速部署。

## 完整 compose 例(MySQL 版)

```yaml
version: "3.9"

services:
  proapi:
    image: ghcr.io/proapi/proapi:vX.Y.Z
    container_name: proapi
    restart: unless-stopped
    depends_on:
      mysql:
        condition: service_healthy
      redis:
        condition: service_healthy
    environment:
      TZ: Asia/Shanghai
      PROAPI_MASTER_KEY: ${PROAPI_MASTER_KEY}
      PROAPI_DATABASE_DSN: "proapi:pass@tcp(mysql:3306)/proapi?parseTime=true&loc=Local&charset=utf8mb4"
      PROAPI_REDIS_ADDR: "redis:6379"
    volumes:
      - ./config.yaml:/app/config.yaml:ro
    ports:
      - "8080:8080"

  mysql:
    image: mysql:8.0
    restart: unless-stopped
    environment:
      MYSQL_ROOT_PASSWORD: rootpw
      MYSQL_DATABASE: proapi
      MYSQL_USER: proapi
      MYSQL_PASSWORD: pass
    volumes:
      - mysql-data:/var/lib/mysql
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 5s
      timeout: 3s
      retries: 20

  redis:
    image: redis:7-alpine
    restart: unless-stopped
    command: ["redis-server", "--maxmemory", "256mb", "--maxmemory-policy", "allkeys-lru"]
    volumes:
      - redis-data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 20

  nginx:
    image: nginx:alpine
    restart: unless-stopped
    depends_on:
      - proapi
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./certs:/etc/nginx/certs:ro
    ports:
      - "80:80"
      - "443:443"

volumes:
  mysql-data:
  redis-data:
```

## .env 文件

把 master_key 通过 `.env` 注入,不要硬编码进 compose:

```bash
# .env
PROAPI_MASTER_KEY=base64-encoded-32-bytes-from-openssl-rand
```

生成 master_key:

```bash
openssl rand -base64 32 > .master_key
echo "PROAPI_MASTER_KEY=$(cat .master_key)" >> .env
```

:::danger 别 commit .env
`.env` 含 master_key,**绝不能进 git**。`.gitignore` 默认已忽略,确认你的部署 repo 也是如此。
:::

## 启动

```bash
docker compose up -d
docker compose logs -f proapi
```

第一次启动会:

1. mysql / redis 起来 + healthcheck 过
2. proapi 容器入口跑 `migrate-then-serve`,**自动执行所有迁移**(创建表 + seed 默认数据)
3. 监听 `:8080`,nginx 反代 `:443`

访问 `https://your.host/admin` 注册第一个超管账户。

## 备份

```bash
# 1. DB(每天定期跑)
docker compose exec mysql \
  mysqldump -uroot -prootpw proapi > backup-$(date +%F).sql

# 2. Redis(可选,作为缓存可重建)
docker compose exec redis redis-cli --rdb /data/dump.rdb
```

把 backup 上传到对象存储(S3 / OSS / 本地异地),保留 30-90 天。

恢复:

```bash
docker compose exec -T mysql mysql -uroot -prootpw proapi < backup-2026-05-22.sql
docker compose restart proapi
```

## 升级

```bash
# 1. 拉新镜像
docker compose pull proapi

# 2. 备份 DB(强制)
docker compose exec mysql mysqldump -uroot -prootpw proapi > backup-pre-upgrade.sql

# 3. 重启 proapi(入口会自动跑迁移)
docker compose up -d proapi

# 4. 验证
curl http://localhost/healthz
```

详见 [升级指南](../guide/upgrade.md)。

## 多节点扩展

单机 compose 不适合超过 1k QPS。扩展方式:

1. **proapi 横向扩展**:K8s Deployment 多副本,共享 DB / Redis,前面挂 Service / Ingress 做 LB
2. **DB / Redis 用托管服务**:RDS / Aurora / ElastiCache
3. **node_id 唯一**:每副本 `PROAPI_NODE_ID` 不同(K8s 可用 statefulset 或 init container 生成)

详见 [反向代理](./reverse-proxy.md)。

## PG 版本 compose

把 mysql 段换成 postgres:

```yaml
  postgres:
    image: postgres:14-alpine
    restart: unless-stopped
    environment:
      POSTGRES_DB: proapi
      POSTGRES_USER: proapi
      POSTGRES_PASSWORD: pass
    volumes:
      - pg-data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U proapi -d proapi"]
      interval: 5s
      timeout: 3s
      retries: 20
```

DSN 改成:

```yaml
PROAPI_DATABASE_DSN: "postgres://proapi:pass@postgres:5432/proapi?sslmode=disable"
```

且 proapi 的 `config.yaml` 里 `database.driver: postgres`。

## 关键要点

- `depends_on + condition: service_healthy` 确保 DB / Redis ready 才启动 proapi,避免启动 race
- master_key 通过 `.env` 注入,**不要直接写 compose**
- redis 加 `maxmemory + LRU` 防 OOM(默认无上限,在大流量场景可能吃完宿主机内存)
- 第一次启动 proapi 会**自动跑迁移**(镜像入口 `migrate-then-serve`)
