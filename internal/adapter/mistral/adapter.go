// Package mistral provides a Mistral adapter (OpenAI-compatible).
package mistral

import (
	"context"

	"github.com/ijry/pro-api/internal/adapter"
	oadapter "github.com/ijry/pro-api/internal/adapter/openai"
	"github.com/ijry/pro-api/internal/protocol/ir"
)

var supportedModels = []string{
	"mistral-large-latest", "mistral-medium-latest",
	"mistral-small-latest", "open-mistral-7b",
	"open-mixtral-8x7b", "open-mixtral-8x22b",
}

type Adapter struct{ base *oadapter.OpenAI }

func New() adapter.Adapter { return &Adapter{base: oadapter.New("https://api.mistral.ai")} }

func (a *Adapter) Name() string                    { return "mistral" }
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
