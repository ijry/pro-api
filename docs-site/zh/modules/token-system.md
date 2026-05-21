---
title: API 令牌
outline: deep
---

# API 令牌使用指南

> 令牌(Token)是客户端调用 proapi 代理 API 时的身份凭证。每个用户可创建多个令牌,每个令牌可独立设额度、IP 白名单、模型白名单、RPM/TPM、过期时间。

## 令牌字符串

- **前缀**:`pa-`(proapi 缩写)
- **总长**:51 字符(`pa-` + 48 字符 base64url 随机)
- **示例**:`pa-AbcDefGhIjKlMnOpQrStUvWxYz0123456789abcdefghIjKlMn`

:::danger 令牌只展示一次
创建或重置令牌后,**完整明文只在那一次展示**。数据库存的是 `sha256(token)`,后台仅展示前缀 `pa-AbcDefGh****` —— 丢了只能重置,不能找回。
:::

## 创建令牌(用户前台)

`用户前台 → 令牌 → 新建`,填写以下字段:

| 字段 | 必填 | 说明 |
|---|---|---|
| Name | Y | 令牌别名,用于自己识别 |
| 额度上限(quota_limit) | N | 该令牌可用 quota 上限;0 = 不限(用 user 余额) |
| 模型白名单 | N | 精确名或通配 `gpt-4*`;空 = 全部允许 |
| IP CIDR 白名单 | N | `1.2.3.4/32` / `10.0.0.0/8`;空 = 不限 |
| RPM 限额 | N | 0 = 继承 user |
| TPM 限额 | N | 0 = 继承 user |
| 有效期 | N | NULL = 永不过期 |

> 截图占位:用户前台令牌创建表单

## 使用令牌

### Bearer 头

```bash
curl https://api.proapi.com/v1/chat/completions \
  -H "Authorization: Bearer pa-xxx" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o-mini",
    "messages": [{"role": "user", "content": "Hello"}]
  }'
```

### Python OpenAI SDK

```python
from openai import OpenAI

client = OpenAI(
    api_key="pa-xxx",
    base_url="https://api.proapi.com/v1",
)

resp = client.chat.completions.create(
    model="gpt-4o-mini",
    messages=[{"role": "user", "content": "Hello"}],
)
print(resp.choices[0].message.content)
```

### Node.js OpenAI SDK

```javascript
import OpenAI from "openai";

const openai = new OpenAI({
    apiKey: "pa-xxx",
    baseURL: "https://api.proapi.com/v1",
});

const resp = await openai.chat.completions.create({
    model: "gpt-4o-mini",
    messages: [{ role: "user", content: "Hello" }],
});
console.log(resp.choices[0].message.content);
```

## 令牌的额度

```
实际可用 = min(user.wallet.balance, token.quota_limit - token.consumed)
```

- 用户钱包余额是"账户上限"
- 令牌的 `quota_limit` 是"令牌内的子上限"(0 = 不限)
- 任意一个用尽都会 402

适用场景:把项目 A 的 token quota_limit 设 100k,项目 B 设 200k,共享同一个 user wallet,既能分账又不互抢。

## 模型白名单

- 精确名:`gpt-4o` / `claude-3.5-sonnet`
- 通配:`gpt-4*`(匹配 `gpt-4`、`gpt-4o`、`gpt-4-turbo` 等)
- 多项:数组,任意命中即放行
- 空数组 / NULL = 全部允许

调用不在白名单的模型 → `403 Forbidden`,错误 type=`invalid_request_error`。

## IP CIDR 白名单

```
"allow_ips": ["1.2.3.4/32", "10.0.0.0/8", "2001:db8::/32"]
```

- IPv4 / IPv6 都支持
- 多个网段任意命中即放行
- 空 = 不限制

调用源 IP 不命中 → `403 Forbidden`,错误 type=`ip_not_allowed`。

:::tip 反代场景
proapi 用 `X-Real-IP` / `X-Forwarded-For` 取真实 IP,要求反代正确透传。详见 [反向代理](../deployment/reverse-proxy.md)。
:::

## RPM / TPM 自定义限额

- 默认 `0` = 继承 user 默认(`system_settings.ratelimit.user_default_*`)
- `>0` = 覆盖该值,作为该 token 的硬上限

完整限流机制见 [限流策略](../architecture/ratelimit.md)。

## 过期时间

- `expires_at = NULL` → 永不过期
- 有值时:过期后 `status` 自动置 `expired`,API 调用返回 401

## 令牌操作

| 操作 | 行为 |
|---|---|
| **重置 key** | 生成新明文 + 新 sha256;旧 key 立即失效 |
| **禁用** | `status = 1`,API 立即拒绝(下次 cache 失效后) |
| **删除** | 软删(`deleted_at`),DB 保留以便审计;明文已不可恢复 |

## 安全建议

- **不要 commit 到公网仓库**(GitGuardian / GitHub Secret Scanning 会扫到 `pa-` 前缀)
- 给客户端用的 token 配 **IP 白名单 + 模型白名单**,缩小爆炸半径
- 临时项目用 token 设 **过期时间**,自动失效
- 长期 token **定期 rotate**(M2 加自动提醒)
- 高敏感场景考虑用 token 的 `rpm_limit` / `tpm_limit` 做防爬限速

## 关键要点

- token 只在创建/重置时**完整明文展示一次**,后续仅展示 prefix
- 数据库存 sha256 hash,**无法反查原文** —— 找不回只能重置
- token 限流优先级:token.rpm_limit > user.rpm_limit > system default
- 用户被禁后,该用户的所有 token 立即失效(无须单独操作)
