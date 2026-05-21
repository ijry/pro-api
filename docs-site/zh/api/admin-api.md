---
title: 管理 API 索引
outline: deep
---

# 管理 API 索引(M1)

> 这是给运维 / 自动化脚本用的**路由索引**。完整字段定义见各模块 spec / GORM model;本页仅列路由 + 简介。
>
> 完整 OpenAPI 文档(含每个字段的 schema、example、错误码)将在 M2 自动生成并挂 Swagger UI。

## 鉴权

管理 API 使用 **Session cookie + CSRF 双重 cookie** 鉴权:

1. 调 `POST /api/auth/login` 拿到 `session` cookie + `csrf_token` cookie
2. 后续请求带:
   - `Cookie: session=...; csrf_token=...`
   - `X-CSRF-Token: <同 csrf_token 值>`

只有 `role = 3`(超级管理员)的 session 能访问 `/api/admin/*`。
M2 会区分 `role = 2`(租户管理员)能做什么。

## 路由清单

### 用户管理

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/api/admin/users` | 列表(分页 + 按 email/role/status/group_id 筛选) |
| GET | `/api/admin/users/:id` | 用户详情 |
| PATCH | `/api/admin/users/:id` | 改分组 / 状态 / 角色 |
| POST | `/api/admin/users/:id/reset-password` | 重置密码(发邮件或返回临时密码) |
| POST | `/api/admin/users/:id/revoke-sessions` | 强制下线所有设备 |

### 用户分组

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/api/admin/groups` | 列表 |
| POST | `/api/admin/groups` | 新建(name / ratio) |
| PATCH | `/api/admin/groups/:id` | 改 |
| DELETE | `/api/admin/groups/:id` | 删(已绑定用户的不能删) |

### 令牌(全局,跨用户)

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/api/admin/tokens` | 列表(可按 `user_id` 筛) |
| DELETE | `/api/admin/tokens/:id` | 删除(等价强制吊销) |

### 渠道

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/api/admin/channels` | 列表 |
| POST | `/api/admin/channels` | 新增 |
| GET | `/api/admin/channels/:id` | 详情 |
| PATCH | `/api/admin/channels/:id` | 改(凭证自动重新加密) |
| DELETE | `/api/admin/channels/:id` | 删 |
| POST | `/api/admin/channels/:id/test` | 测试连通性 |
| POST | `/api/admin/channels/:id/reset-breaker` | 重置熔断 |

### 模型字典

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/api/admin/models` | 列表 |
| POST | `/api/admin/models` | 新增 |
| PATCH | `/api/admin/models/:name` | 改(`name` 作为 PK) |
| DELETE | `/api/admin/models/:name` | 删 |

### 定价规则

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/api/admin/pricing-rules` | 列表(可按 scope 筛) |
| POST | `/api/admin/pricing-rules` | 新增 |
| PATCH | `/api/admin/pricing-rules/:id` | 改 |
| DELETE | `/api/admin/pricing-rules/:id` | 删 |

### 充值审批

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/api/admin/manual-recharges` | 列表(默认 `status=pending`) |
| POST | `/api/admin/manual-recharges/:id/approve` | 通过(填实际入账金额) |
| POST | `/api/admin/manual-recharges/:id/reject` | 驳回 |

### 兑换码

| 方法 | 路径 | 说明 |
|---|---|---|
| POST | `/api/admin/redeem-codes/batch` | 批量生成,返 CSV |
| GET | `/api/admin/redeem-codes` | 列表(按 `batch_no` 筛) |
| POST | `/api/admin/redeem-codes/:id/disable` | 单码禁用 |
| POST | `/api/admin/redeem-codes/batch/:batch_no/disable` | 整批禁用 |

### 日志

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/api/admin/logs/requests` | 请求日志(分页 + 多维筛:user/model/channel/status/range) |
| GET | `/api/admin/logs/errors` | 错误日志 |
| GET | `/api/admin/logs/usage-stats` | 聚合统计(by day / model / user) |

### 公告

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/api/admin/notices` | 列表 |
| POST | `/api/admin/notices` | 新建(草稿) |
| PATCH | `/api/admin/notices/:id` | 改 |
| DELETE | `/api/admin/notices/:id` | 删 |
| POST | `/api/admin/notices/:id/publish` | 发布 |
| POST | `/api/admin/notices/:id/archive` | 下架 |

### 系统设置

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/api/admin/settings` | 全量读 |
| GET | `/api/admin/settings/:key` | 单 key 读 |
| PATCH | `/api/admin/settings/:key` | 单 key 改(自动 Pub/Sub 通知集群) |

### 审计

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/api/admin/audit-logs` | 审计日志查询(谁、什么时间、做了什么) |

## 请求 / 响应风格

- Content-Type: `application/json`
- 统一错误格式见 [API 概览](./overview.md#错误响应格式-管理---用户-api)
- 删除用 `DELETE` 方法时通常返回 `204 No Content`;部分软删返 `200` + body

## 速率限制

管理 API **不受** 代理 API 限流,但有保护性的单 IP 60 RPM(防爬 / 防暴力探测)。

## 完整 OpenAPI 文档

M2 自动生成 + Swagger UI(`/api/admin/swagger`);M1 暂无,以本页为索引,字段定义看具体模块 spec 或 GORM model 定义。

## 关键要点

- 本页**只列路由**,不写每个字段(M1 阶段过长,放 M2 OpenAPI)
- **CSRF 双 cookie**模式:Cookie 里有 `csrf_token`,请求头里要带 `X-CSRF-Token` 同值
- 凭证字段(`channels.credentials`、`settings.auth.github_oauth.client_secret`)写入时自动加密,**读取时不会返回明文**(只返脱敏值)
