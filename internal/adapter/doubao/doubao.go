// Package doubao 实现字节豆包（火山引擎）API adapter（OpenAI 兼容）。
package doubao

import (
	"context"
	"net/http"
	"time"

	"github.com/ijry/pro-api/internal/adapter"
	oadapter "github.com/ijry/pro-api/internal/adapter/openai"
	"github.com/ijry/pro-api/internal/protocol/ir"
)

const defaultBaseURL = "https://ark.cn-beijing.volces.com/api/v3"

// Doubao 实现 adapter.Adapter，调用字节豆包（火山引擎）API。
type Doubao struct {
	client *http.Client
}

// New 构造 Doubao adapter。
func New() *Doubao {
	return &Doubao{
		client: adapter.NewHTTPClient(adapter.ClientConfig{
			Provider:            "doubao",
			Timeout:             0,
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 32,
			IdleConnTimeout:     90 * time.Second,
		}),
	}
}

func (a *Doubao) Name() string { return "doubao" }

func (a *Doubao) Capabilities() adapter.Capability {
	return adapter.CapChat | adapter.CapStream | adapter.CapToolUse | adapter.CapEmbedding
}

func (a *Doubao) SupportedModels() []string {
	return []string{
		"doubao-pro-32k",
		"doubao-pro-128k",
		"doubao-lite-32k",
		"doubao-lite-128k",
		"doubao-embedding",
	}
}

func baseURL(cred adapter.Credential) string {
	if cred.BaseURL != "" {
		return cred.BaseURL
	}
	return defaultBaseURL
}

func (a *Doubao) Chat(ctx context.Context, req *ir.ChatRequest, cred adapter.Credential) (*ir.ChatResponse, error) {
	return oadapter.ChatWithClient(ctx, a.client, req, cred, baseURL(cred)+"/chat/completions", "bearer")
}

func (a *Doubao) ChatStream(ctx context.Context, req *ir.ChatRequest, cred adapter.Credential) (adapter.StreamReader, error) {
	return oadapter.ChatStreamWithClient(ctx, a.client, req, cred, baseURL(cred)+"/chat/completions", "bearer")
}

func (a *Doubao) Embed(ctx context.Context, req *ir.EmbedRequest, cred adapter.Credential) (*ir.EmbedResponse, error) {
	return oadapter.EmbedWithClient(ctx, a.client, req, cred, baseURL(cred)+"/embeddings", "bearer")
}
