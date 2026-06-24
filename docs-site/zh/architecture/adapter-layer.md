---
title: 适配器层
outline: deep
---

# 适配器层

> 适配器(Adapter)是连接 pro-api 与上游模型厂商的"翻译官":把内部统一的 IR(Intermediate Representation)请求,翻译成厂商专属的 HTTP 调用,再把响应翻回 IR。

## 为什么要有适配器层

不同上游的 HTTP 协议差异很大:

- 字段名不同:OpenAI 的 `messages[].role` vs Anthropic 的 `messages[].role` + 独立 `system` 顶层字段
- 流式格式不同:OpenAI 用 `data: ...\n\n` SSE;Anthropic 用 `event: ...\ndata: ...`;Gemini 用 chunked JSON
- 字段语义不同:Gemini 的 `parts` vs OpenAI 的 `content`(字符串或数组)
- 错误格式不同

适配器层把这些差异封装,让 channel / billing / log / ratelimit 等业务模块**完全 provider-agnostic** —— 增加一家新厂商不需要改这些模块。

## M1 已支持的 9 家

| Provider | 支持端点 | 流式 | 主要模型 | 备注 |
|---|---|---|---|---|
| openai | chat / completions / embeddings | ✅ | gpt-3.5* / gpt-4* / gpt-4o* / o1* | 标准实现 |
| azure | chat / embeddings | ✅ | 同 OpenAI(部署名映射) | 需要 region + deployment 字典 |
| anthropic | chat | ✅ | claude-3.5-sonnet / claude-3-haiku | M1 仅出向,入口推 M2 |
| gemini | chat / embeddings | ✅ | gemini-1.5-pro / -flash / 2.0 | |
| deepseek | chat | ✅ | deepseek-chat / -reasoner | reasoner 有 `reasoning_tokens` |
| moonshot | chat | ✅ | moonshot-v1-8k / -32k / -128k | |
| zhipu | chat | ✅ | glm-4 / glm-4-flash | |
| qwen | chat / embeddings | ✅ | qwen-turbo / -plus / -max | |
| doubao | chat | ✅ | doubao-pro / -lite | |

更多 provider 在 M2 加入(Groq、Mistral、Cohere、xAI 等)。

## IR(中间表示)

`internal/protocol/ir` 定义了与 provider 无关的请求/响应结构:

```go
type ChatRequest struct {
    Model       string
    Messages    []Message
    MaxTokens   *int
    Temperature *float32
    TopP        *float32
    Stream      bool
    Tools       []Tool
    ToolChoice  *ToolChoice
    // ... 见仓库
}

type Message struct {
    Role    string         // system / user / assistant / tool
    Content []ContentPart  // text / image_url / tool_result ...
    Name    string
    ToolCallID string
}

type ChatResponse struct {
    ID      string
    Model   string
    Choices []Choice
    Usage   Usage
}

type Usage struct {
    InputTokens     int
    OutputTokens    int
    CachedTokens    int    // OpenAI prompt caching 命中
    ReasoningTokens int    // o1 / deepseek-reasoner 专有
}

type ChatChunk struct {  // 流式
    Delta   Delta
    Done    bool
    Usage   *Usage  // 仅最后一个 chunk 可能有
}
```

完整字段见 `internal/protocol/ir/*.go`,本页只截关键部分。

## 协议转换示例

以 "OpenAI 入口请求 → Anthropic 上游" 为例:

```
[OpenAI IR]                              [Anthropic HTTP]
{                                        {
  "model": "claude-3.5-sonnet",     ──▶    "model": "claude-3-5-sonnet-20241022",
  "messages": [                            "system": "You are helpful.",
    {"role":"system", "content":"..."}, ─▶ "messages": [
    {"role":"user",   "content":"Hi"}    ──▶   {"role":"user", "content":"Hi"}
  ],                                       ],
  "max_tokens": 1024                       "max_tokens": 1024
}                                        }
```

完整字段映射表见 `internal/adapter/anthropic/chat.go` 与对应单测 `chat_test.go`(testdata fixture)。

## 流式响应处理

`adapter.ChatStream` 返回的不是 `ChatResponse`,而是一个 `StreamReader`:

```go
type StreamReader interface {
    Next() (*ChatChunk, error)   // io.EOF 表示流结束
    Close() error                 // 必须 Close 释放底层 HTTP body
}
```

relay 层逐 chunk 拿到 IR `ChatChunk`,通过 OpenAI SSE encoder 转成 `data: {"choices":[{"delta":...}]}\n\n`,flush 给客户端。

:::tip TTFT 指标
pro-api 记录每个流式请求的 **TTFT**(Time To First Token):从 adapter.ChatStream 返回到第一个 chunk 出现的时间。
后台日志页可按 TTFT 排序排查慢请求。
:::

## 错误归类

适配器把上游各种错误统一映射成 `pkg/apierr` 的标准 code:

| 上游情况 | 归类 | pro-api code |
|---|---|---|
| HTTP 5xx | 上游服务器错 | `CodeUpstreamError` (60001) |
| 网络超时 / connect refused | 上游超时 | `CodeUpstreamTimeout` (60003) |
| HTTP 429 | 上游限流 | `CodeUpstreamRateLimit` (60004) |
| HTTP 401/403 | 上游凭证错 | `CodeUpstreamAuth` (60005) |
| HTTP 404 model | 模型不存在 | `CodeModelNotFound` (60002) |
| HTTP 400 其他 | 透传(转 InvalidRequest) | `CodeInvalidRequest` (20001) |

熔断器只对 `CodeUpstreamError` / `CodeUpstreamTimeout` 计数(渠道侧问题);
`CodeUpstreamAuth` 也计入(说明凭证失效)。**4xx 客户端错不计** —— 那是用户的锅。

## 扩展性

新增一家 provider 是 pro-api 二次开发**最常见的需求**。详细步骤见 [新增上游适配器](../development/adapter-guide.md)。

简要清单:

1. 在 `internal/adapter/<name>/` 新建包
2. 实现 `adapter.Adapter` 接口(`Chat` / `ChatStream` / `Embed` 等按能力)
3. 用 testdata fixture + httptest 写单测
4. 在 `internal/adapter/registry.go` 注册
5. seed 模型字典(`migrations/.../seed`)
6. 更新本页支持列表

## 关键要点

- M1 入口只有 OpenAI 协议,**出向** 9 家;Anthropic / Gemini 入口要等 M2。
- Azure OpenAI 较特殊:`client_model` → `deployment` 映射存在渠道 `extra` 字段。
- 流式响应有 `TTFT` 指标,日志里专门记。
- 上游返回 `usage` 时**优先用上游的**,本地 tokenizer 仅作 fallback(`pricing.tiktoken` / `pricing.approximate`)。
