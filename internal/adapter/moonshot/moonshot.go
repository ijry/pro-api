// Package moonshot 实现 Moonshot（Kimi）API adapter（OpenAI 兼容）。
package moonshot

import (
	"context"
	"net/http"
	"time"

	"github.com/ijry/pro-api/internal/adapter"
	oadapter "github.com/ijry/pro-api/internal/adapter/openai"
	"github.com/ijry/pro-api/internal/protocol/ir"
)

const defaultBaseURL = "https://api.moonshot.cn"

// Moonshot 实现 adapter.Adapter，调用 Moonshot API（OpenAI 兼容）。
type Moonshot struct {
	client *http.Client
}

// New 构造 Moonshot adapter。
func New() *Moonshot {
	return &Moonshot{
		client: adapter.NewHTTPClient(adapter.ClientConfig{
			Provider:            "moonshot",
			Timeout:             0,
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 32,
			IdleConnTimeout:     90 * time.Second,
		}),
	}
}

func (a *Moonshot) Name() string { return "moonshot" }

func (a *Moonshot) Capabilities() adapter.Capability {
	return adapter.CapChat | adapter.CapStream | adapter.CapToolUse
}

func (a *Moonshot) SupportedModels() []string {
	return []string{
		"moonshot-v1-8k",
		"moonshot-v1-32k",
		"moonshot-v1-128k",
	}
}

func baseURL(cred adapter.Credential) string {
	if cred.BaseURL != "" {
		return cred.BaseURL
	}
	return defaultBaseURL
}

func (a *Moonshot) Chat(ctx context.Context, req *ir.ChatRequest, cred adapter.Credential) (*ir.ChatResponse, error) {
	return oadapter.ChatWithClient(ctx, a.client, req, cred, baseURL(cred)+"/v1/chat/completions", "bearer")
}

func (a *Moonshot) ChatStream(ctx context.Context, req *ir.ChatRequest, cred adapter.Credential) (adapter.StreamReader, error) {
	return oadapter.ChatStreamWithClient(ctx, a.client, req, cred, baseURL(cred)+"/v1/chat/completions", "bearer")
}

func (a *Moonshot) Embed(ctx context.Context, req *ir.EmbedRequest, cred adapter.Credential) (*ir.EmbedResponse, error) {
	return oadapter.EmbedWithClient(ctx, a.client, req, cred, baseURL(cred)+"/v1/embeddings", "bearer")
}
