---
title: 升级
outline: deep
---

# 升级指南

> pro-api 遵循 [语义化版本 SemVer 2.0](https://semver.org/lang/zh-CN/)。本页讲清楚升级前需要准备什么、按什么顺序操作、出错怎么回滚。

## 版本号规则

`MAJOR.MINOR.PATCH`

- **MAJOR**:不兼容变更(API/数据库/配置)
- **MINOR**:向后兼容的新增
- **PATCH**:向后兼容的 bug 修复

:::warning M1 期间所有版本 < 1.0
0.x 阶段 API 仍可能有 breaking change,**每次升级前务必读 CHANGELOG 的 "Breaking changes" 段**。1.0 之后才会严格遵守 SemVer。
:::

## 升级步骤

### 通用流程

1. **备份数据库**(强制 —— 是否回滚都需要它兜底)
2. 查看 [CHANGELOG](/zh/changelog) 的 Breaking changes 段
3. 拉新代码 / 新镜像
4. 跑 `proapi migrate up`(增量,可重入)
5. 重启服务
6. `/healthz` 检查 + 跑一次代理请求验证

### Docker 升级

```bash
# 拉新镜像
docker pull ghcr.io/ijry/pro-api:vX.Y.Z

# 停旧
docker stop proapi
docker rm proapi

# 起新(配置与 volume 不变)
docker run -d \
  --name proapi \
  -p 8080:8080 \
  -v $(pwd)/config.yaml:/app/config.yaml:ro \
  -e PROAPI_MASTER_KEY=$(cat .master_key) \
  --restart unless-stopped \
  ghcr.io/ijry/pro-api:vX.Y.Z
```

Docker 镜像入口已封装 `migrate-then-serve`,会在启动时自动跑迁移。

### 二进制升级

```bash
# 1. 备份
mysqldump proapi > backup-$(date +%F).sql

# 2. 停服务
systemctl stop proapi

# 3. 替换二进制
wget https://github.com/ijry/pro-api/releases/download/vX.Y.Z/proapi_linux_amd64.tar.gz
tar -xzf proapi_linux_amd64.tar.gz -C /opt/proapi/
chmod +x /opt/proapi/proapi

# 4. 跑迁移
sudo -u proapi /opt/proapi/proapi migrate up

# 5. 起服务
systemctl start proapi

# 6. 验证
curl http://127.0.0.1:8080/healthz
```

### 源码升级

```bash
git fetch --tags
git checkout vX.Y.Z
make build
# 同二进制升级的步骤 2-6
```

## 数据库迁移

- pro-api 用 [golang-migrate](https://github.com/golang-migrate/migrate) 管理迁移。
- 迁移文件在 `migrations/` 下,按 `NNNNNN_description.{up,down}.sql` 命名。
- 增量执行,**可重入**(`schema_migrations` 表跟踪当前版本)。
- 失败时,migrate 会标记 `dirty=true`,**必须人工排查** —— 不能直接重跑。

回滚一步:

```bash
./proapi migrate down --steps 1
```

:::tip 破坏性 migration
大多数 migration 可以下移,但部分(如分区表拆分、列删除)不可逆。CHANGELOG 会专门标注。
:::

## 多实例滚动升级

1. 先在一个节点跑 `proapi migrate up`(其他节点用 `--no-migrate` 或镜像入口的 `MIGRATE=false` 跳过)。
2. 逐个 drain & 替换实例:
   - 反向代理把该实例从 upstream 摘掉
   - 等 in-flight 请求完成(典型 < 60s,流式可能更长)
   - 重启该实例
   - 加回 upstream
3. 全部替换完,再确认 `/healthz` 都返回新版本。

:::warning 同 master_key + 同 DB + 唯一 node_id
滚动升级期间新旧版本会**短时间共存**,所以:
- master_key 必须保持一致
- 数据库 / Redis 必须共享
- node_id 不能撞(同一时刻不能有两个 node_id = 0 的实例)
:::

## 不兼容变更

完整不兼容变更见 [CHANGELOG](/zh/changelog) 的 Breaking 段。常见类型:

- 配置 key 重命名 / 删除
- API 响应字段调整
- 数据库 schema breaking(列类型变更等)

每个 breaking change 都会附 "migration recipe"。

## 回滚

```bash
# Docker
docker stop proapi
docker run -d ... ghcr.io/ijry/pro-api:vOLD.Y.Z

# 二进制
systemctl stop proapi
cp /opt/proapi/proapi.bak /opt/proapi/proapi
# 如果有破坏性 migration,需要先 ./proapi migrate down --steps N
systemctl start proapi
```

:::danger 回滚前要确认 migration 兼容
若新版本跑了不可逆 migration(如删了列),回到旧版本会启动失败。**升级前一定要保留 DB 备份**,出问题用备份恢复。
:::

## 当前阶段说明

M1 阶段 pro-api 仍在 0.x,版本节奏:

- 每个 milestone 完成发一个 minor(如 0.1.0 / 0.2.0)
- Bug fix 发 patch(0.1.1 / 0.1.2)
- 1.0 计划在 M3 后发,届时正式承诺 API 稳定性
