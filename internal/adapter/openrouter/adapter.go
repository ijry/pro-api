// Package openrouter provides an OpenRouter adapter (OpenAI-compatible + special headers).
package openrouter

import (
	"context"

	"github.com/ijry/pro-api/internal/adapter"
	oadapter "github.com/ijry/pro-api/internal/adapter/openai"
	"github.com/ijry/pro-api/internal/protocol/ir"
)

var supportedModels = []string{
	"openai/gpt-4o", "anthropic/claude-3.5-sonnet",
	"google/gemini-pro", "meta-llama/llama-3.1-70b-instruct",
}

type Adapter struct{ base *oadapter.OpenAI }

func New() adapter.Adapter {
	return &Adapter{base: oadapter.New("https://openrouter.ai/api")}
}

func (a *Adapter) Name() string                    { return "openrouter" }
func (a *Adapter) Capabilities() adapter.Capability { return a.base.Capabilities() }
func (a *Adapter) SupportedModels() []string        { return supportedModels }

// addHeaders injects OpenRouter-specific headers via Credential.Extra.
func addHeaders(cred adapter.Credential) adapter.Credential {
	if cred.Extra == nil {
		cred.Extra = map[string]string{}
	}
	if cred.Extra["HTTP-Referer"] == "" {
		cred.Extra["HTTP-Referer"] = "https://pro-api.app"
	}
	if cred.Extra["X-Title"] == "" {
		cred.Extra["X-Title"] = "pro-api"
	}
	return cred
}

func (a *Adapter) Chat(ctx context.Context, req *ir.ChatRequest, cred adapter.Credential) (*ir.ChatResponse, error) {
	return a.base.Chat(ctx, req, addHeaders(cred))
}
func (a *Adapter) ChatStream(ctx context.Context, req *ir.ChatRequest, cred adapter.Credential) (adapter.StreamReader, error) {
	return a.base.ChatStream(ctx, req, addHeaders(cred))
}
func (a *Adapter) Embed(ctx context.Context, req *ir.EmbedRequest, cred adapter.Credential) (*ir.EmbedResponse, error) {
	return a.base.Embed(ctx, req, addHeaders(cred))
}
