// Package cohere provides a Cohere adapter via Cohere's OpenAI-compatible Compatibility API.
package cohere

import (
	"context"

	"github.com/ijry/pro-api/internal/adapter"
	oadapter "github.com/ijry/pro-api/internal/adapter/openai"
	"github.com/ijry/pro-api/internal/protocol/ir"
)

// supportedModels 列出 Cohere 兼容 API 暴露的常用模型。
// 兼容端点同时支持 chat 与 embeddings;rerank 为 Cohere 原生能力,暂不经此适配器。
var supportedModels = []string{
	"command-a-03-2025",
	"command-r-plus-08-2024",
	"command-r-08-2024",
	"command-r7b-12-2024",
	"embed-v4.0",
	"embed-english-v3.0",
	"embed-multilingual-v3.0",
}

// Adapter 复用 OpenAI 基座,base URL 指向 Cohere Compatibility API。
// 基座会在 base 后追加 /v1,故此处传入 .../compatibility,拼出 .../compatibility/v1。
type Adapter struct{ base *oadapter.OpenAI }

// New 构造 Cohere 适配器。
func New() adapter.Adapter {
	return &Adapter{base: oadapter.New("https://api.cohere.ai/compatibility")}
}

func (a *Adapter) Name() string                     { return "cohere" }
func (a *Adapter) Capabilities() adapter.Capability { return a.base.Capabilities() }
func (a *Adapter) SupportedModels() []string        { return supportedModels }
func (a *Adapter) Chat(ctx context.Context, req *ir.ChatRequest, cred adapter.Credential) (*ir.ChatResponse, error) {
	return a.base.Chat(ctx, req, cred)
}
func (a *Adapter) ChatStream(ctx context.Context, req *ir.ChatRequest, cred adapter.Credential) (adapter.StreamReader, error) {
	return a.base.ChatStream(ctx, req, cred)
}
func (a *Adapter) Embed(ctx context.Context, req *ir.EmbedRequest, cred adapter.Credential) (*ir.EmbedResponse, error) {
	return a.base.Embed(ctx, req, cred)
}
