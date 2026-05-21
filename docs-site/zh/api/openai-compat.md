---
title: OpenAI 兼容协议
outline: deep
---

# OpenAI 兼容协议接入

> proapi 的代理 API 100% 兼容 OpenAI v1 协议。任何能调 OpenAI 的客户端,改 `base_url` 与 `api_key` 就能用。

## 准备工作

1. 在 proapi **用户前台 → 令牌** 创建一个 `pa-xxx` 令牌(见 [API 令牌](../modules/token-system.md))
2. proapi **admin 后台 → 渠道管理** 至少配一个跑通的渠道(见 [渠道管理](../modules/channel-management.md))
3. 知道 proapi 的 base_url(本地是 `http://127.0.0.1:8080/v1`,生产是你的部署域名 + `/v1`)

## chat/completions

### 非流式

#### curl

```bash
curl https://api.example.com/v1/chat/completions \
  -H "Authorization: Bearer pa-xxx" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "messages": [
      {"role": "system", "content": "You are a helpful assistant."},
      {"role": "user", "content": "Hello"}
    ]
  }'
```

#### Python

```python
from openai import OpenAI

client = OpenAI(
    api_key="pa-xxx",
    base_url="https://api.example.com/v1",
)

resp = client.chat.completions.create(
    model="gpt-4o",
    messages=[
        {"role": "system", "content": "You are a helpful assistant."},
        {"role": "user", "content": "Hello"},
    ],
)
print(resp.choices[0].message.content)
```

#### Node.js

```javascript
import OpenAI from "openai";

const openai = new OpenAI({
    apiKey: "pa-xxx",
    baseURL: "https://api.example.com/v1",
});

const resp = await openai.chat.completions.create({
    model: "gpt-4o",
    messages: [
        { role: "system", content: "You are a helpful assistant." },
        { role: "user", content: "Hello" },
    ],
});
console.log(resp.choices[0].message.content);
```

### 流式

#### curl(SSE)

```bash
curl https://api.example.com/v1/chat/completions \
  -N \
  -H "Authorization: Bearer pa-xxx" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "stream": true,
    "messages": [{"role": "user", "content": "Hi"}]
  }'
```

:::tip curl -N
curl 默认开启 line buffering 会让流式响应看起来"卡",加 `-N`(或 `--no-buffer`)关掉。
:::

#### Python

```python
stream = client.chat.completions.create(
    model="gpt-4o",
    messages=[{"role": "user", "content": "Hi"}],
    stream=True,
    stream_options={"include_usage": True},  # 流末附 usage
)
for chunk in stream:
    delta = chunk.choices[0].delta
    if delta.content:
        print(delta.content, end="", flush=True)

# 最后一个 chunk 的 usage 字段会带完整 tokens 统计
```

### 支持的请求字段

| 字段 | 支持 | 备注 |
|---|---|---|
| `model` | ✅ | proapi 通过渠道的模型映射对到上游 |
| `messages` | ✅ | 支持 system / user / assistant / tool 角色 |
| `max_tokens` | ✅ | 影响计费预扣上限 |
| `temperature` / `top_p` | ✅ | |
| `stream` | ✅ | |
| `stop` | ✅ | 字符串或数组 |
| `tools` / `tool_choice` | ✅(M1) | function calling |
| `response_format` | ✅ | `text` / `json_object` / `json_schema` |
| `user` | ✅ | 透传给上游做反 abuse |
| `seed` | ✅ | 透传 |
| `stream_options` | ✅(M1) | `include_usage: true` 时流末附 usage |
| `presence_penalty` / `frequency_penalty` | ✅(透传) | 上游支持时生效 |
| `logit_bias` / `logprobs` | ✅(透传) | 上游支持时生效 |

### Vision

```python
resp = client.chat.completions.create(
    model="gpt-4o",
    messages=[
        {
            "role": "user",
            "content": [
                {"type": "text", "text": "What's in this image?"},
                {
                    "type": "image_url",
                    "image_url": {
                        "url": "https://example.com/cat.jpg"
                    }
                }
            ]
        }
    ],
)
```

### tool_use(function calling)

```python
resp = client.chat.completions.create(
    model="gpt-4o",
    messages=[{"role": "user", "content": "What's the weather in Beijing?"}],
    tools=[{
        "type": "function",
        "function": {
            "name": "get_weather",
            "parameters": {
                "type": "object",
                "properties": {"city": {"type": "string"}},
                "required": ["city"]
            }
        }
    }],
)

tool_calls = resp.choices[0].message.tool_calls
```

## completions(legacy)

```bash
curl https://api.example.com/v1/completions \
  -H "Authorization: Bearer pa-xxx" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-3.5-turbo-instruct",
    "prompt": "Say hello",
    "max_tokens": 30
  }'
```

> 新代码请用 `chat/completions`,legacy 接口仅为兼容老 SDK。

## embeddings

```python
resp = client.embeddings.create(
    model="text-embedding-3-small",
    input=["hello", "world"],
)
for d in resp.data:
    print(d.embedding[:5])
```

支持 batch(input 传数组)、不同 dimensions(`dimensions` 字段,仅部分模型)。

## models

```bash
curl https://api.example.com/v1/models \
  -H "Authorization: Bearer pa-xxx"
```

返回该令牌实际可用的模型 —— 经过:

- 用户分组允许的模型
- 令牌的模型白名单
- 后台 active 渠道支持的模型

三者**交集**。

## 错误对照

| OpenAI `error.type` | proapi `code` | 含义 |
|---|---|---|
| `invalid_request_error` / `invalid_api_key` | 20002 | 令牌无效 |
| `insufficient_quota` | 40004 | 余额不足 |
| `rate_limit_exceeded` | 50001-50004 | 限流(具体维度看 message) |
| `api_error` | 60001 | 上游错误(5xx) |
| `invalid_request_error` / `model_not_found` | 60002 | 模型不存在 |

OpenAI 官方 SDK 的 `RateLimitError` / `AuthenticationError` 等异常类型会被正确触发。

## 与原版 OpenAI 的差异

- **模型名**:proapi 允许任意命名,通过渠道的模型映射对到上游;`/v1/models` 返回的就是用户能用的名字。
- **usage 字段**:proapi 始终返回(即使是流式响应,通过 `stream_options.include_usage=true`)。
- **速率限制**:proapi 用自己的 4 维度限流,与 OpenAI 的全局账户级限流不同(见 [限流策略](../architecture/ratelimit.md))。
- **缓存命中**:OpenAI 的 prompt caching 由上游识别,proapi 通过 `usage.prompt_tokens_details.cached_tokens` 透传(并按 `cached_ratio` 计费)。

## 调试技巧

- **加 `X-Request-ID`**:工单时附上,帮助运维定位
- **看请求日志**:admin 后台 → 日志 → 请求日志,按 `X-Request-ID` 搜
- **dev 模式**:本地启动时 `log.level=debug`,会输出每次上游 HTTP 的字段映射

## 关键要点

- `base_url` 末尾 **要 `/v1`**(SDK 内部不再加)
- curl 流式响应要 `-N` 或 `--no-buffer`
- Vision / tool_use 在 M1 是"透传给上游" —— 上游不支持就报错,proapi 不做能力补齐
- `stream_options.include_usage=true` 推荐总是带上,流式响应能拿到准确 usage
