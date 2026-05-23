// Package deepseek 实现 DeepSeek API adapter（OpenAI 兼容）。
package deepseek

import (
	"context"
	"net/http"
	"time"

	"github.com/ijry/pro-api/internal/adapter"
	oadapter "github.com/ijry/pro-api/internal/adapter/openai"
	"github.com/ijry/pro-api/internal/protocol/ir"
)

const defaultBaseURL = "https://api.deepseek.com"

// DeepSeek 实现 adapter.Adapter，调用 DeepSeek API（OpenAI 兼容）。
type DeepSeek struct {
	client *http.Client
}

// New 构造 DeepSeek adapter。
func New() *DeepSeek {
	return &DeepSeek{
		client: adapter.NewHTTPClient(adapter.ClientConfig{
			Provider:            "deepseek",
			Timeout:             0,
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 32,
			IdleConnTimeout:     90 * time.Second,
		}),
	}
}

func (a *DeepSeek) Name() string { return "deepseek" }

func (a *DeepSeek) Capabilities() adapter.Capability {
	return adapter.CapChat | adapter.CapStream | adapter.CapToolUse | adapter.CapReasoning
}

func (a *DeepSeek) SupportedModels() []string {
	return []string{
		"deepseek-chat",
		"deepseek-reasoner",
		"deepseek-coder",
	}
}

func baseURL(cred adapter.Credential) string {
	if cred.BaseURL != "" {
		return cred.BaseURL
	}
	return defaultBaseURL
}

func (a *DeepSeek) Chat(ctx context.Context, req *ir.ChatRequest, cred adapter.Credential) (*ir.ChatResponse, error) {
	return oadapter.ChatWithClient(ctx, a.client, req, cred, baseURL(cred)+"/v1/chat/completions", "bearer")
}

func (a *DeepSeek) ChatStream(ctx context.Context, req *ir.ChatRequest, cred adapter.Credential) (adapter.StreamReader, error) {
	return oadapter.ChatStreamWithClient(ctx, a.client, req, cred, baseURL(cred)+"/v1/chat/completions", "bearer")
}

func (a *DeepSeek) Embed(ctx context.Context, req *ir.EmbedRequest, cred adapter.Credential) (*ir.EmbedResponse, error) {
	return oadapter.EmbedWithClient(ctx, a.client, req, cred, baseURL(cred)+"/v1/embeddings", "bearer")
}
