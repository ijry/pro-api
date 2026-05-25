---
title: 5 分钟跑起来
outline: deep
---

# 5 分钟跑起来

> 💡 **想直接看看?** <a :href="$withBase('/admin-demo/index.html')" target="_blank">管理后台演示</a> · <a :href="$withBase('/user-demo/index.html')" target="_blank">用户中心演示</a>(纯前端 mock,无需部署后端)

> 本页面向第一次接触 proapi 的用户:从 0 到调通第一个请求,目标耗时 5 分钟。
>
> M0 阶段为骨架,完整体验自 M1 起可用;部分上游能力(Anthropic / Gemini 入口)在 M2 才完整。

## 环境要求

proapi 提供两种推荐路径,二选一即可:

- **推荐路径(零运维)**:Docker / Docker Compose。需要 `docker` 与 `docker compose` 可用。
- **源码 / 二进制路径**:适合想要二开或不便装 Docker 的环境。需要 Go 1.22+、Node.js 20+、pnpm 9。

无论哪条路径,proapi 自身**不内嵌** MySQL / PostgreSQL / Redis,你需要单独起这些依赖。

## 起依赖

仓库提供了 `docker-compose.dev.yml`,一行命令拉起本地依赖(MySQL + PostgreSQL + Redis 都齐):

```bash
make docker-up
```

:::tip MySQL 还是 PostgreSQL?
proapi 同时支持 MySQL ≥ 8 与 PostgreSQL ≥ 14。若你没有强偏好,**路线图推荐 PostgreSQL** —— 它对 JSON 字段与 partial index 更友好。
:::

## 起后端

```bash
export PROAPI_MASTER_KEY=$(openssl rand -base64 32)
make dev-backend
```

访问 [http://127.0.0.1:8080/healthz](http://127.0.0.1:8080/healthz),应返回 `{"status":"ok",...}`。

:::danger PROAPI_MASTER_KEY 必须在生产环境固定
`PROAPI_MASTER_KEY` 用于加密渠道凭证。**一旦丢失,所有渠道凭证都将无法解密。** 上线前请把它备份到密码管理器或 HSM,不要每次启动重新生成。
:::

## 起前端

proapi 前端拆为三个独立 Vite 工程,可分别启动:

```bash
make dev-admin      # 后台 → http://127.0.0.1:5173
make dev-user       # 前台 → http://127.0.0.1:5174
make dev-docs       # 文档站 → http://127.0.0.1:5175
```

## 一键全起

```bash
make dev
```

后端、admin、user、docs 同时跑起来,适合本地开发。

## 首次登录与初始化

服务首次启动时,proapi 会自动 seed 三个默认用户分组:

- `default`(group_ratio = 1.0)
- `vip`(group_ratio = 0.8)
- `svip`(group_ratio = 0.6)

之后访问 `http://127.0.0.1:5173`(admin 前台)注册账户。

:::warning 第一个注册的用户自动成为超级管理员
"首个注册即超管"的策略只在用户表为空时生效。**生产部署后第一个注册的账户务必是运维自己**,否则会被陌生用户抢到 root 权限。
:::

用户前台地址是 `http://127.0.0.1:5174`,普通用户从这里登录,管理自己的令牌与消费。

## 调通第一个 API 请求

需要先在 admin 后台 **系统设置 → 渠道管理** 配一条上游渠道(填 provider + api_key),
然后到用户前台 **令牌 → 新建** 拿一个 `pa-xxx` 令牌。详细步骤见
[渠道管理](../modules/channel-management.md) 与 [API 令牌](../modules/token-system.md)。

拿到令牌后用 curl 调一次:

```bash
curl http://127.0.0.1:8080/v1/chat/completions \
  -H "Authorization: Bearer pa-xxx" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o-mini",
    "messages": [{"role": "user", "content": "Hello"}]
  }'
```

或用 Python OpenAI SDK:

```python
from openai import OpenAI

client = OpenAI(
    api_key="pa-xxx",
    base_url="http://127.0.0.1:8080/v1",
)
resp = client.chat.completions.create(
    model="gpt-4o-mini",
    messages=[{"role": "user", "content": "Hello"}],
)
print(resp.choices[0].message.content)
```

## 下一步去哪儿

- 想做生产部署?读 [安装](./installation.md) 与 [Docker Compose 部署](../deployment/docker-compose.md)
- 想理解配置项?读 [配置](./configuration.md)
- 想用 OpenAI SDK 的更多能力(流式 / 函数调用 / Vision)?读 [OpenAI 兼容协议](../api/openai-compat.md)
- 想知道你的钱怎么算?读 [计费如何工作](../billing-guide/how-billing-works.md)
