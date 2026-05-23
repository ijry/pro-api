// Package azure 实现 Azure OpenAI adapter。
// 复用 openai adapter 的编解码，仅覆盖 URL 构造和认证头（api-key）。
package azure

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/ijry/pro-api/internal/adapter"
	oadapter "github.com/ijry/pro-api/internal/adapter/openai"
	"github.com/ijry/pro-api/internal/protocol/ir"
)

const defaultAPIVersion = "2024-06-01"

// Azure 实现 adapter.Adapter，调用 Azure OpenAI REST API。
type Azure struct {
	client *http.Client
}

// New 构造 Azure adapter。
func New() *Azure {
	return &Azure{
		client: adapter.NewHTTPClient(adapter.ClientConfig{
			Provider:            "azure",
			Timeout:             0,
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 32,
			IdleConnTimeout:     90 * time.Second,
		}),
	}
}

func (a *Azure) Name() string { return "azure" }

func (a *Azure) Capabilities() adapter.Capability {
	return adapter.CapChat | adapter.CapStream | adapter.CapCompletion |
		adapter.CapEmbedding | adapter.CapVision | adapter.CapToolUse
}

func (a *Azure) SupportedModels() []string {
	return []string{
		"gpt-4o",
		"gpt-4o-mini",
		"gpt-4-turbo",
		"gpt-4",
		"gpt-3.5-turbo",
		"text-embedding-3-large",
		"text-embedding-3-small",
	}
}

// buildURL 构造 Azure OpenAI 请求 URL。
// cred.BaseURL = "https://<resource>.openai.azure.com"
// cred.Extra["deployment"] = deployment name (fallback = model name)
// cred.Extra["api_version"] = optional api version override
func buildURL(cred adapter.Credential, path string, model string) string {
	deployment := cred.Extra["deployment"]
	if deployment == "" {
		deployment = model
	}
	apiVersion := cred.Extra["api_version"]
	if apiVersion == "" {
		apiVersion = defaultAPIVersion
	}
	base := cred.BaseURL
	return fmt.Sprintf("%s/openai/deployments/%s%s?api-version=%s", base, deployment, path, apiVersion)
}

func (a *Azure) Chat(ctx context.Context, req *ir.ChatRequest, cred adapter.Credential) (*ir.ChatResponse, error) {
	return oadapter.ChatWithClient(ctx, a.client, req, cred, buildURL(cred, "/chat/completions", req.Model), "api-key")
}

func (a *Azure) ChatStream(ctx context.Context, req *ir.ChatRequest, cred adapter.Credential) (adapter.StreamReader, error) {
	return oadapter.ChatStreamWithClient(ctx, a.client, req, cred, buildURL(cred, "/chat/completions", req.Model), "api-key")
}

func (a *Azure) Embed(ctx context.Context, req *ir.EmbedRequest, cred adapter.Credential) (*ir.EmbedResponse, error) {
	return oadapter.EmbedWithClient(ctx, a.client, req, cred, buildURL(cred, "/embeddings", req.Model), "api-key")
}
