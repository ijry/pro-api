---
title: 用户体系
outline: deep
---

# 用户体系

> proapi 的账号体系覆盖"普通用户 / 租户管理员 / 超级管理员"三种角色,支持邮箱密码、邮箱验证码、GitHub OAuth 注册,通过用户分组关联消费倍率。

## 角色定义

| 角色 | role 值 | 典型权限 |
|---|---|---|
| 普通用户 | 0 | 用户前台:管理自己的令牌、查看消费日志、申请充值、兑换兑换码 |
| 租户管理员 | 2 | 同超管,但不可改 cluster-wide 系统设置(M2 才严格区分) |
| 超级管理员 | 3 | 后台一切权限:用户管理 / 渠道 / 定价 / 系统设置 / 审计 |

## 注册方式

proapi 支持 3 种注册方式,可在 `system_settings.auth.*` 控制开关。

### 邮箱密码注册

```
用户填 email + password
   │
   ▼
auth.email_verification_required ?
   │
   ├── true → 发送验证码到邮箱 → 用户输验证码 → 创建账户(status=2 待验证 → 0 正常)
   │
   └── false → 直接创建账户(status=0)
   │
   ▼
若用户表为空 → role=3(超级管理员)
否则        → role=0(普通用户)
```

### 邮箱验证码

无密码登录:用户填邮箱 → proapi 发 6 位验证码 → 用户输入 → 命中则自动登录或注册。

适用于"不想记密码"的轻量场景。

### GitHub OAuth

OAuth 2.0 标准流程。详细配置见 [GitHub OAuth 配置](../integration/oauth-github.md)。

## 用户分组

分组的核心作用是**绑定 group_ratio**:

| 分组 | group_ratio | 用途 |
|---|---|---|
| default | 1.0 | 默认所有新用户 |
| vip | 0.8 | 充值或邀请晋升 |
| svip | 0.6 | 高价值用户 |

后台可改名、加分组、改 ratio。每个用户**隶属一个分组**,改分组立即生效(影响后续计费)。

> 完整定价规则见 [定价与倍率](./pricing.md)。

## 用户状态

| status | 含义 | 影响 |
|---|---|---|
| 0 | 正常 | 可登录、可调 API |
| 1 | 禁用 | 不可登录、API 立即拒绝(401) |
| 2 | 待邮箱验证 | 可登录但部分功能锁,验证后转 0 |

## Session 机制

- **存储**:Redis 主存(热)+ DB mirror(冷,可审计)
- **格式**:不可猜测的 opaque token(非 JWT)
- **TTL**:默认 30 天(`session.ttl_days`)
- **滑动延期**:`session.sliding=true` 时,每次访问刷新 TTL
- **强制下线**:管理员可一键撤销某用户所有 session

:::tip 为什么不用 JWT
JWT 一旦签发,服务端无法主动作废(除非维护黑名单)。proapi 用 opaque session,改密码 / 禁用账号可立即下线所有设备 —— 更适合企业场景。
:::

## API 调用与用户的关系

```
API 调用走 token,不走 session。
token 关联 user_id。
user 被禁用 → 所有 token 立即失效(token middleware 每次检查 user.status)。
user 改分组 → 后续请求按新 group_ratio 计费。
user 改邮箱 → 不影响 token。
```

## 用户密码

- **哈希**:bcrypt,cost=10
- **强度策略**:
  - `auth.password.min_length`(默认 8)
  - `auth.password.require_mixed`(大小写 + 数字,默认 off)
- **找回**:M1 通过 "邮箱验证码 → 设置新密码" 流程;M2 增加更多方式

## 后台操作

管理员在 admin 后台 → 用户管理 可:

- **列表**:按 email / role / status / 分组筛选;分页
- **改分组**:`PATCH /api/admin/users/:id` body=`{group_id}`
- **改状态**:同上,body=`{status}`(禁用立即生效)
- **重置密码**:`POST /api/admin/users/:id/reset-password`,生成临时密码或发邮件
- **强制下线**:`POST /api/admin/users/:id/revoke-sessions`,撤销所有 active session
- **查看消费**:跳转消费日志页,预筛选 `user_id`

完整 API 见 [管理 API 索引](../api/admin-api.md#用户管理)。

## 关键要点

- **"首个注册即超管"** 策略只在用户表为空时生效。**生产部署后第一个一定要是运维自己**,否则会被陌生用户抢权。
- session 是后端化的(opaque token,非 JWT),改密码可立即下线全部设备。
- 用户禁用后,Redis 里的 active session 也会被批量撤销(`SessionStore.RevokeAllForUser`)。
- 同一 user 可以同时绑邮箱密码 + GitHub,任一方式都能登录。
