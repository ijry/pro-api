---
title: 账号对接总览
outline: deep
---

# 账号对接总览

> proapi 的账号体系既支持自带的"邮箱密码 / 邮箱验证码"注册,也支持对接第三方身份系统。M1 提供 GitHub OAuth;M2 加 Google / 微信 / 飞书 / 钉钉 / Discord;M3 加企业 SSO(OIDC / SAML / LDAP / CAS)。

## M1 支持的对接

| 类型 | provider | 状态 | 文档 |
|---|---|---|---|
| OAuth 2.0 | GitHub | ✅ M1 | [oauth-github](./oauth-github.md) |
| OAuth 2.0 | Google | 🚧 M2 | — |
| OAuth 2.0 | 微信 / 飞书 / 钉钉 / Discord | 🚧 M2 | — |
| OIDC | 通用 | 🚧 M3 | — |
| SAML 2.0 | 通用 | 🚧 M3 | — |
| LDAP / Active Directory | 通用 | 🚧 M3 | — |
| CAS | 通用 | 🚧 M3 | — |

## 通用 OAuth 流程

OAuth 2.0 标准的 4 步授权码模式:

```
1. 用户在 proapi 登录页点 "用 X 登录"
        │
        ▼
2. proapi 重定向到 X 的授权页(带 redirect_url, scope, state)
        │
        ▼
3. 用户在 X 授权,X 重定向回 proapi 的 callback,带 ?code=...&state=...
        │
        ▼
4. proapi 用 code 调 X 的 token endpoint 换 access_token
        │
        ▼
5. proapi 用 access_token 调 X 的 user info,拿 (uid, email, avatar, login)
        │
        ▼
6. 查 oauth_bindings 表:
   - 命中已绑定的 user → 登录该 user
   - 未绑定:按 email 查 users
     - 命中 → 自动绑定该 user + 登录
     - 未命中 → 创建新 user + 绑定 + 登录
```

`state` 字段用于防 CSRF,proapi 会校验回调里的 `state` 与 cookie 里存的一致。

## 已有账号绑定第三方

```
已登录用户 → 账号设置 → 绑定 GitHub
        │
        ▼
进入 OAuth 流程
        │
        ▼
回调后写入 oauth_bindings(user_id, provider, provider_uid)
        │
        ▼
显示"已绑定"
```

## 多种登录方式共存

同一个 user 可以同时绑:

- 邮箱密码
- 邮箱验证码(无密码,直接发码)
- GitHub OAuth

任一方式都能登到同一个 user。通过 `oauth_bindings` 表多对一关联。

## 安全建议

- **redirect_url 严格白名单**:在 provider 的 OAuth 应用里把 callback URL 写死
- **client_secret 保密**:后台填入后自动加密存(master_key 加密)
- **启用邮箱验证**:`auth.email_verification_required = true` 防恶意注册
- **禁用注册时**:`auth.allow_register = false`,只有管理员邀请才能创建账号(M2 加邀请功能)

## 路由总览

| 路径 | 说明 |
|---|---|
| `GET /api/auth/oauth/:provider/authorize` | 跳转到 provider 授权页 |
| `GET /api/auth/oauth/:provider/callback` | provider 回调入口 |
| `POST /api/user/me/bindings` | 绑定第三方(已登录) |
| `DELETE /api/user/me/bindings/:provider` | 解绑 |
| `GET /api/user/me/bindings` | 查看自己的绑定列表 |

## 关键要点

- M1 **只 GitHub**;别的 provider 留接口契约,实施在 M2
- **解绑前必须确保用户还有别的登录方式**(否则会被永久锁出)—— proapi 在解绑前会校验
- provider profile **只取必要字段**(uid, email, avatar, login),不存敏感数据
- 同一 GitHub 账号**不能绑多个 proapi 账号**(`oauth_bindings.UNIQUE(provider, provider_uid)`)
