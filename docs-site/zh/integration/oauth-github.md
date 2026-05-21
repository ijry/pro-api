---
title: GitHub OAuth
outline: deep
---

# GitHub OAuth 配置完整步骤

> 本页带你从 0 到 1 在 proapi 上启用 GitHub 登录。

## 1. 在 GitHub 创建 OAuth App

1. 浏览器打开 [GitHub Settings → Developer settings → OAuth Apps](https://github.com/settings/developers)
2. 点 **New OAuth App**
3. 填表:
   - **Application name**:`proapi`(或你的部署名,如 "MyCompany proapi")
   - **Homepage URL**:`https://api.example.com`(你的 proapi 根 URL)
   - **Authorization callback URL**:`https://api.example.com/api/auth/oauth/github/callback`
4. 提交后 GitHub 会给你:
   - **Client ID**(公开值,可见)
   - **Client Secret**(机密值,**只展示一次**,务必立刻复制)

> 截图占位:GitHub OAuth App 创建页

:::warning Client Secret 只展示一次
GitHub 创建 OAuth App 后,Client Secret 在页面上**只显示一次**。关掉页面后再回来只能看 `****`。务必当场复制保存。
:::

## 2. 在 proapi 后台填写凭证

1. 登录 proapi admin 后台(超管账号)
2. **系统设置 → 账号对接 → GitHub**
3. 填入:
   - **Client ID**(明文)
   - **Client Secret**(后端自动用 master_key 加密存)
   - **Redirect URL**:留空 = 用请求时的 Host 自动拼;若有多套部署环境(dev / staging / prod)需要分别填明确值
4. 保存 → 点击 **测试** 按钮 → proapi 调一次 GitHub 的 `GET /user` 验证凭证

或直接调 API:

```bash
curl -X PATCH https://api.example.com/api/admin/settings/auth.github_oauth \
  -H "Cookie: session=...; csrf_token=..." \
  -H "X-CSRF-Token: ..." \
  -H "Content-Type: application/json" \
  -d '{
    "value": {
      "client_id": "Iv1.xxxxxxxx",
      "client_secret": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
      "redirect_url": ""
    }
  }'
```

> 截图占位:proapi 后台 GitHub OAuth 配置页

## 3. 验证

1. 退出 proapi 登录
2. 打开登录页 → 应出现 **使用 GitHub 登录** 按钮
3. 点击 → 跳转到 GitHub 授权页 → 用户授权 → 回到 proapi → 自动登录或注册

第一次登录该 GitHub 账户时:

- 如果该 GitHub 关联的邮箱已在 proapi 注册 → 自动绑定该 user
- 否则 → 创建新 user(role=0,普通用户),自动绑定

## 4. 用户绑定已有账号

如果用户已经用邮箱密码注册了,想加 GitHub 作为第二种登录方式:

1. 用户前台 → **账号设置 → 绑定 GitHub**
2. 点击 → GitHub 授权 → 回到 proapi → 显示 "已绑定"

之后用 GitHub / 邮箱密码都能登到同一个账号。

## 5. 解绑

1. 用户前台 → **账号设置 → 已绑定 → GitHub → 解绑**
2. proapi 校验:**若该账号没有其他登录方式**(如没有密码,只有 GitHub)→ 报错 "请先设置密码" 拒绝解绑
3. 通过校验 → 删除 `oauth_bindings` 记录

## 安全实践

- **Client Secret 不要进 git**:从前台 / 后台填入,而不是写到任何配置文件
- **Redirect URL 严格匹配**:proapi 后端会校验请求里的 `redirect_uri` 必须在白名单
- **启用邮箱验证**:`auth.email_verification_required = true`,即使 GitHub 邮箱已验也建议二次校验
- **限制能注册的 GitHub 用户**(M2 增强):比如要求是某 GitHub Organization 的成员才允许注册

## 常见问题

### "callback url mismatch"

GitHub OAuth App 里的 Authorization callback URL 必须与 proapi 实际请求时的 `redirect_uri`
**严格匹配**,包括 `https://` / 端口 / 末尾斜杠。

**调试**:在 proapi 日志里搜 `oauth.callback_url=`,看实际发出去的 URL 是什么。

### 用户首次 GitHub 登录,proapi 自动创建账户但没邮箱

GitHub 用户在 GitHub 设置里勾了 "Keep my email addresses private" 时,API 返回的 `email` 是 `null` 或 `users.noreply.github.com` 格式。

proapi 处理:

- 仍然创建账户(`email` 用 noreply 格式 placeholder)
- 引导用户在账号设置里补填真实邮箱
- 若 `auth.email_verification_required = true`,补填后才能用 API

### 一个 GitHub 账号能不能绑两个 proapi 账号

**不能**。`oauth_bindings` 表上 `UNIQUE (provider, provider_uid)`,一个 GitHub 账号严格对应一个 proapi 账号。

若需要换绑:先在原 proapi 账号 → 账号设置 → 解绑 GitHub,然后在新 proapi 账号上绑定。

### 解绑后能不能再绑回原来的 proapi 账号

可以。解绑只是删除 `oauth_bindings` 记录,GitHub uid 释放后可重新绑定到任意 proapi 账号。

## 关键要点

- GitHub Client Secret 在 OAuth App 创建后**只展示一次**,务必立即复制
- proapi 把 Client Secret 用 `master_key` 加密存到 `system_settings.auth.github_oauth.client_secret`
- GitHub 邮箱常常是 `users.noreply.github.com`(用户勾了"保护邮箱"时),要做兼容
- 解绑前必须有 fallback 登录方式,否则用户会被永久锁出
