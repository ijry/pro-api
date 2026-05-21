---
title: API 概览
outline: deep
---

# API 概览

> proapi 提供三类 API:
>
> - **代理 API**:给客户端用的 OpenAI 兼容代理(`/v1/*`)
> - **管理 API**:给后台 / 运维脚本用的运维接口(`/api/admin/*`)
> - **用户 API**:给用户前台用的自服务接口(`/api/user/*`)

## 三类 API 速查

| 类别 | 前缀 | 鉴权方式 | 典型调用方 |
|---|---|---|---|
| 代理 API | `/v1/*` | `Authorization: Bearer pa-xxx` | 客户端 / SDK |
| 管理 API | `/api/admin/*` | Session cookie + CSRF | 后台前端 / 运维脚本 |
| 用户 API | `/api/user/*` | Session cookie + CSRF | 用户前台 |
| 认证 API | `/api/auth/*` | 无鉴权(登录入口) | 登录注册流程 |
| 公开 API | `/api/public/*` | 无鉴权(只读) | 公告 / 模型清单 |
| 健康 / 监控 | `/healthz`, `/metrics` | 无鉴权 | 探活 / Prometheus |

## 代理 API M1 支持的端点

| 端点 | 协议 | 说明 |
|---|---|---|
| POST `/v1/chat/completions` | OpenAI | 聊天补全(支持流式 `stream: true`) |
| POST `/v1/completions` | OpenAI | 文本补全(legacy) |
| POST `/v1/embeddings` | OpenAI | 向量化 |
| GET `/v1/models` | OpenAI | 模型清单(交集:用户分组 ∩ 令牌白名单 ∩ 渠道支持) |

> Anthropic 入口(`/v1/messages`)与 Gemini 入口在 M2 加入。
> DALL-E / TTS / STT / Midjourney / Suno 等多模态端点也在 M2。

## 请求 / 响应通用约定

### 请求头

| Header | 必填 | 说明 |
|---|---|---|
| `Authorization` | Y | `Bearer pa-xxx` |
| `Content-Type` | Y | `application/json` |
| `Accept` | N | `text/event-stream`(流式)或 `application/json` |
| `X-Request-ID` | N | 客户端可指定;不指定 proapi 自动生成(36 字符 UUID) |

### 响应头

| Header | 说明 |
|---|---|
| `X-Request-ID` | 本次请求的 trace_id —— 工单时附上 |
| `X-RateLimit-Remaining` | 当前维度剩余配额(取最紧的) |
| `X-RateLimit-Reset` | 配额恢复时间戳(秒) |
| `X-Channel-ID` | 实际命中的渠道 ID(可选,`system_settings` 控制是否暴露) |

### 错误响应格式(代理 API,OpenAI 兼容)

```json
{
  "error": {
    "message": "Insufficient quota, please top up",
    "type": "insufficient_quota",
    "code": "insufficient_quota"
  }
}
```

字段与 OpenAI 完全一致,SDK(如 OpenAI Python `APIError` 子类)能正确识别。

### 错误响应格式(管理 / 用户 API)

```json
{
  "code": 40004,
  "message": "余额不足,请充值",
  "details": null
}
```

不是 OpenAI 兼容格式,但更符合内部 RESTful 约定。`code` 见 `pkg/apierr`。

## HTTP 状态码约定

| 状态 | 含义 |
|---|---|
| 200 | 成功 |
| 400 | 参数错 |
| 401 | 未鉴权 / 令牌无效 |
| 402 | 余额不足 |
| 403 | 无权访问 / IP 不在白名单 / 模型不在白名单 |
| 404 | 资源不存在 |
| 409 | 状态冲突(如重复审批) |
| 410 | 资源已失效(如 reservation 已过期) |
| 429 | 限流(任一维度) |
| 500 | proapi 内部错 —— 工单时附 X-Request-ID |
| 503 | 上游不可用 / 渠道全部熔断 |

## 速率限制

代理 API 受 4 维度限流约束,详见 [限流策略](../architecture/ratelimit.md)。

管理 API / 用户 API **不受** 代理 API 限流,但有保护性的单 IP 60 RPM(防爬)。

## 版本化

- **代理 API**:保持 OpenAI 协议兼容,跟随 OpenAI 协议演进,**不另起 proapi 版本**
- **管理 / 用户 API**:按 proapi 版本管理,breaking change 见 [CHANGELOG](/zh/changelog)

## SDK 推荐

直接用 OpenAI 官方 SDK(Python / Node / Go / .NET),改 `base_url` 与 `api_key` 即可。详见
[OpenAI 兼容协议](./openai-compat.md)。

## 关键要点

- `base_url` 末尾要 **带 `/v1`**(SDK 内部不再加 `/v1`)
- 流式请求需要 `stream: true` 且响应 `Content-Type: text/event-stream`
- 错误格式两套并存的原因:代理 API 必须兼容 OpenAI SDK 的解析逻辑;管理 API 不需要,用更符合 RESTful 的格式
- `X-Request-ID` 是工单关键信息,客户端建议都带上
