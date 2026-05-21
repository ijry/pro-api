---
title: 新增上游适配器
outline: deep
---

# 如何新增一个上游适配器

> 以"新增 mistral 适配器"为例,完整讲清楚步骤。其他 provider 类比。

## 适配器接口契约

`internal/adapter/adapter.go` 定义了核心接口:

```go
type Adapter interface {
    Name() string
    Capabilities() Capability
    SupportedModels() []string

    Chat(ctx context.Context, req *ir.ChatRequest, cred *Credentials) (*ir.ChatResponse, error)
    ChatStream(ctx context.Context, req *ir.ChatRequest, cred *Credentials) (ir.StreamReader, error)
    Embed(ctx context.Context, req *ir.EmbedRequest, cred *Credentials) (*ir.EmbedResponse, error)
}

type Capability uint32
const (
    CapChat Capability = 1 << iota
    CapStream
    CapEmbed
    CapVision
    CapToolUse
    CapReasoning
)
```

完整定义见仓库,本页只截关键签名。详细 IR 字段见 [适配器层](../architecture/adapter-layer.md)。

## 步骤一:新建包

```bash
mkdir -p internal/adapter/mistral/testdata
touch internal/adapter/mistral/{adapter,chat,stream,embed}.go
touch internal/adapter/mistral/{adapter_test,chat_test,stream_test}.go
```

## 步骤二:实现 Adapter

```go
package mistral

import (
    "context"
    "net/http"
    "time"

    "go.uber.org/zap"
    "github.com/ijry/pro-api/internal/adapter"
)

type Mistral struct {
    httpClient *http.Client
    log        *zap.Logger
}

func New(log *zap.Logger) *Mistral {
    return &Mistral{
        httpClient: &http.Client{Timeout: 5 * time.Minute},
        log:        log,
    }
}

func (m *Mistral) Name() string { return "mistral" }

func (m *Mistral) Capabilities() adapter.Capability {
    return adapter.CapChat | adapter.CapStream
}

func (m *Mistral) SupportedModels() []string {
    return []string{
        "mistral-large-latest",
        "mistral-small-latest",
        "codestral-latest",
    }
}
```

## 步骤三:实现 Chat(非流式)

```go
package mistral

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"

    "github.com/ijry/pro-api/internal/adapter"
    "github.com/ijry/pro-api/internal/protocol/ir"
    "github.com/ijry/pro-api/pkg/apierr"
)

func (m *Mistral) Chat(ctx context.Context, req *ir.ChatRequest, cred *adapter.Credentials) (*ir.ChatResponse, error) {
    // 1. IR → Mistral 请求体
    body := map[string]any{
        "model":    req.Model,
        "messages": toMistralMessages(req.Messages),
    }
    if req.MaxTokens != nil {
        body["max_tokens"] = *req.MaxTokens
    }
    if req.Temperature != nil {
        body["temperature"] = *req.Temperature
    }

    raw, _ := json.Marshal(body)

    // 2. 调上游 HTTP
    httpReq, _ := http.NewRequestWithContext(ctx, "POST",
        "https://api.mistral.ai/v1/chat/completions", bytes.NewReader(raw))
    httpReq.Header.Set("Content-Type", "application/json")
    httpReq.Header.Set("Authorization", "Bearer "+cred.APIKey)

    httpResp, err := m.httpClient.Do(httpReq)
    if err != nil {
        return nil, classifyError(err)
    }
    defer httpResp.Body.Close()

    if httpResp.StatusCode >= 500 {
        return nil, apierr.New(apierr.CodeUpstreamError,
            fmt.Sprintf("mistral upstream %d", httpResp.StatusCode))
    }

    raw, _ = io.ReadAll(httpResp.Body)

    // 3. Mistral 响应 → IR
    var mResp mistralChatResponse
    if err := json.Unmarshal(raw, &mResp); err != nil {
        return nil, apierr.Wrap(apierr.CodeUpstreamError, err)
    }
    return toIRResponse(&mResp), nil
}
```

`toMistralMessages` / `toIRResponse` 是字段映射函数,定义在同包内。

## 步骤四:实现 ChatStream

```go
func (m *Mistral) ChatStream(ctx context.Context, req *ir.ChatRequest, cred *adapter.Credentials) (ir.StreamReader, error) {
    body := map[string]any{
        "model":    req.Model,
        "messages": toMistralMessages(req.Messages),
        "stream":   true,
    }
    raw, _ := json.Marshal(body)

    httpReq, _ := http.NewRequestWithContext(ctx, "POST",
        "https://api.mistral.ai/v1/chat/completions", bytes.NewReader(raw))
    httpReq.Header.Set("Content-Type", "application/json")
    httpReq.Header.Set("Authorization", "Bearer "+cred.APIKey)
    httpReq.Header.Set("Accept", "text/event-stream")

    httpResp, err := m.httpClient.Do(httpReq)
    if err != nil {
        return nil, classifyError(err)
    }

    return newStreamReader(httpResp.Body, m.log), nil
}

// streamReader 实现 ir.StreamReader 接口
type streamReader struct {
    body   io.ReadCloser
    scanner *bufio.Scanner
    log    *zap.Logger
}

func (r *streamReader) Next() (*ir.ChatChunk, error) {
    for r.scanner.Scan() {
        line := r.scanner.Bytes()
        if !bytes.HasPrefix(line, []byte("data: ")) {
            continue
        }
        payload := bytes.TrimPrefix(line, []byte("data: "))
        if bytes.Equal(payload, []byte("[DONE]")) {
            return &ir.ChatChunk{Done: true}, nil
        }
        var event mistralStreamEvent
        if err := json.Unmarshal(payload, &event); err != nil {
            continue
        }
        return toIRChunk(&event), nil
    }
    return nil, io.EOF
}

func (r *streamReader) Close() error {
    return r.body.Close()  // 必须释放底层连接
}
```

## 步骤五:错误归类

```go
func classifyError(err error) error {
    if err == nil {
        return nil
    }
    var netErr net.Error
    if errors.As(err, &netErr) && netErr.Timeout() {
        return apierr.Wrap(apierr.CodeUpstreamTimeout, err)
    }
    // ... 其他归类
    return apierr.Wrap(apierr.CodeUpstreamError, err)
}
```

HTTP 状态码归类:5xx → `CodeUpstreamError`,429 → `CodeUpstreamRateLimit`,401/403 → `CodeUpstreamAuth`。

## 步骤六:写测试

`testdata/` 放 fixture:

```
testdata/
├── request_chat_basic.json         IR ChatRequest 示例
├── response_chat_basic.json        Mistral 响应
└── response_chat_stream.sse        Mistral SSE 原始流
```

测试用 `httptest.Server` 把 mistral 上游 mock 出来:

```go
func TestChat_Basic(t *testing.T) {
    fixture, _ := os.ReadFile("testdata/response_chat_basic.json")
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.Write(fixture)
    }))
    defer srv.Close()

    m := New(zap.NewNop())
    m.httpClient = srv.Client()
    // 覆盖 base URL(可注入或反射,设计上推荐 New() 接 baseURL 参数)

    req := loadIRRequest(t, "testdata/request_chat_basic.json")
    resp, err := m.Chat(context.Background(), req, &adapter.Credentials{APIKey: "test"})
    require.NoError(t, err)
    assert.Equal(t, "mistral-large-latest", resp.Model)
    assert.Equal(t, 10, resp.Usage.InputTokens)
}
```

## 步骤七:注册到 Registry

`internal/adapter/registry.go`:

```go
func DefaultRegistry(log *zap.Logger) *Registry {
    r := New()
    r.Register("openai", openai.New(log))
    r.Register("anthropic", anthropic.New(log))
    // ...
    r.Register("mistral", mistral.New(log))  // 新增这行
    return r
}
```

## 步骤八:种子模型字典

在 seed migration 或新增 migration 里 INSERT 几个 mistral 模型:

```sql
INSERT INTO model_catalogs (name, provider, capability_flags,
    default_input_ratio, default_output_ratio)
VALUES
    ('mistral-large-latest',  'mistral', 3, 1.0, 3.0),
    ('mistral-small-latest',  'mistral', 3, 0.1, 0.3),
    ('codestral-latest',      'mistral', 3, 0.1, 0.3);
```

ratio 数值按"上游公开价 USD / proapi base 价"对照,默认 `base_quota_per_dollar = 500000` 时 `ratio × 2 = $/M tokens`。

## 步骤九:文档与 CHANGELOG

- 在 `docs-site/zh/architecture/adapter-layer.md` 的支持列表加 mistral 行
- 在 `docs-site/zh/billing-guide/model-pricing.md` 加 Mistral 章节
- 在仓库根 `CHANGELOG.md` 的 Unreleased 段加 `Added: mistral adapter`

## 步骤十:PR + Code Review

详见 [贡献指南](./contributing.md)。

## 进阶话题

- **Vision 支持**:在 `Chat` 中识别 `content[].type == "image_url"`,转 mistral 的图像输入(若上游支持)
- **Tool use**:在 `Chat` 中识别 `tools` / `tool_choice`,转 mistral 的 function calling 协议
- **Reasoning**:若上游返回 `reasoning_tokens`,填到 `Usage.ReasoningTokens`,proapi 自动按 `reasoning_ratio` 计费
- **Prompt caching**:若上游有 caching 命中信息,填到 `Usage.CachedTokens`

## 调试技巧

- 用真 mistral key 跑集成测:`go test ./internal/adapter/mistral -tags integration`
- 用 mitmproxy 抓 HTTP 包对比 IR 字段映射是否正确
- 流式响应建议 chunk-by-chunk 打 debug 日志,定位"丢字 / 多字 / 顺序错"问题

## 关键要点

- **`StreamReader.Close()` 必须释放底层 HTTP body**,不然连接泄漏
- 上游若在流末发 usage(像 OpenAI `stream_options.include_usage`),要在 `Next` 里 capture 进 `Usage`
- adapter 不依赖任何业务包(billing / channel / log / ratelimit),保持纯净 —— 只依赖 `pkg/apierr` 与 `internal/protocol/ir`
- testdata fixture 文件**提交进 git**,便于回归
