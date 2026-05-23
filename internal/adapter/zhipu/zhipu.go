// Package zhipu 实现智谱 GLM API adapter。
// 认证：API Key 格式为 "id.secret"，使用 HS256 JWT 签名（TTL 15 分钟）。
package zhipu

import (
	"context"
	"net/http"
	"time"

	"github.com/ijry/pro-api/internal/adapter"
	oadapter "github.com/ijry/pro-api/internal/adapter/openai"
	"github.com/ijry/pro-api/internal/protocol/ir"
)

const defaultBaseURL = "https://open.bigmodel.cn/api/paas/v4"

// Zhipu 实现 adapter.Adapter，调用智谱 GLM API。
type Zhipu struct {
	client *http.Client
}

// New 构造 Zhipu adapter。
func New() *Zhipu {
	return &Zhipu{
		client: adapter.NewHTTPClient(adapter.ClientConfig{
			Provider:            "zhipu",
			Timeout:             0,
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 32,
			IdleConnTimeout:     90 * time.Second,
		}),
	}
}

func (a *Zhipu) Name() string { return "zhipu" }

func (a *Zhipu) Capabilities() adapter.Capability {
	return adapter.CapChat | adapter.CapStream | adapter.CapToolUse | adapter.CapEmbedding | adapter.CapVision
}

func (a *Zhipu) SupportedModels() []string {
	return []string{
		"glm-4-plus",
		"glm-4-air",
		"glm-4-flash",
		"glm-4",
		"glm-3-turbo",
		"embedding-3",
		"embedding-2",
	}
}

func baseURL(cred adapter.Credential) string {
	if cred.BaseURL != "" {
		return cred.BaseURL
	}
	return defaultBaseURL
}

// makeCredWithJWT 生成携带 JWT token 的 Credential。
// 智谱认证 API Key 格式："{id}.{secret}"，JWT 有效期 15 分钟。
func makeCredWithJWT(cred adapter.Credential) adapter.Credential {
	token, err := GenerateJWT(cred.APIKey, 15*time.Minute)
	if err != nil {
		// 如果 JWT 生成失败，原样传递 apikey（让上游报 401）
		return cred
	}
	return adapter.Credential{
		APIKey:  token,
		BaseURL: cred.BaseURL,
		Extra:   cred.Extra,
	}
}

func (a *Zhipu) Chat(ctx context.Context, req *ir.ChatRequest, cred adapter.Credential) (*ir.ChatResponse, error) {
	return oadapter.ChatWithClient(ctx, a.client, req, makeCredWithJWT(cred), baseURL(cred)+"/chat/completions", "bearer")
}

func (a *Zhipu) ChatStream(ctx context.Context, req *ir.ChatRequest, cred adapter.Credential) (adapter.StreamReader, error) {
	return oadapter.ChatStreamWithClient(ctx, a.client, req, makeCredWithJWT(cred), baseURL(cred)+"/chat/completions", "bearer")
}

func (a *Zhipu) Embed(ctx context.Context, req *ir.EmbedRequest, cred adapter.Credential) (*ir.EmbedResponse, error) {
	return oadapter.EmbedWithClient(ctx, a.client, req, makeCredWithJWT(cred), baseURL(cred)+"/embeddings", "bearer")
}
