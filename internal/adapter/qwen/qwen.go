// Package qwen 实现阿里通义千问 Qwen API adapter（DashScope OpenAI 兼容模式）。
package qwen

import (
	"context"
	"net/http"
	"time"

	"github.com/ijry/pro-api/internal/adapter"
	oadapter "github.com/ijry/pro-api/internal/adapter/openai"
	"github.com/ijry/pro-api/internal/protocol/ir"
)

const defaultBaseURL = "https://dashscope.aliyuncs.com/compatible-mode"

// Qwen 实现 adapter.Adapter，调用阿里 DashScope OpenAI 兼容接口。
type Qwen struct {
	client *http.Client
}

// New 构造 Qwen adapter。
func New() *Qwen {
	return &Qwen{
		client: adapter.NewHTTPClient(adapter.ClientConfig{
			Provider:            "qwen",
			Timeout:             0,
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 32,
			IdleConnTimeout:     90 * time.Second,
		}),
	}
}

func (a *Qwen) Name() string { return "qwen" }

func (a *Qwen) Capabilities() adapter.Capability {
	return adapter.CapChat | adapter.CapStream | adapter.CapToolUse | adapter.CapEmbedding
}

func (a *Qwen) SupportedModels() []string {
	return []string{
		"qwen-max",
		"qwen-plus",
		"qwen-turbo",
		"qwen-long",
		"text-embedding-v3",
		"text-embedding-v2",
	}
}

func baseURL(cred adapter.Credential) string {
	if cred.BaseURL != "" {
		return cred.BaseURL
	}
	return defaultBaseURL
}

func (a *Qwen) Chat(ctx context.Context, req *ir.ChatRequest, cred adapter.Credential) (*ir.ChatResponse, error) {
	return oadapter.ChatWithClient(ctx, a.client, req, cred, baseURL(cred)+"/v1/chat/completions", "bearer")
}

func (a *Qwen) ChatStream(ctx context.Context, req *ir.ChatRequest, cred adapter.Credential) (adapter.StreamReader, error) {
	return oadapter.ChatStreamWithClient(ctx, a.client, req, cred, baseURL(cred)+"/v1/chat/completions", "bearer")
}

func (a *Qwen) Embed(ctx context.Context, req *ir.EmbedRequest, cred adapter.Credential) (*ir.EmbedResponse, error) {
	return oadapter.EmbedWithClient(ctx, a.client, req, cred, baseURL(cred)+"/v1/embeddings", "bearer")
}
