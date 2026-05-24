// Package huggingface provides a HuggingFace Inference adapter (OpenAI-compatible).
package huggingface

import (
	"context"

	"github.com/ijry/pro-api/internal/adapter"
	oadapter "github.com/ijry/pro-api/internal/adapter/openai"
	"github.com/ijry/pro-api/internal/protocol/ir"
)

var supportedModels = []string{
	"meta-llama/Meta-Llama-3.1-8B-Instruct",
	"meta-llama/Meta-Llama-3.1-70B-Instruct",
	"mistralai/Mixtral-8x7B-Instruct-v0.1",
}

type Adapter struct{ base *oadapter.OpenAI }

func New() adapter.Adapter {
	return &Adapter{base: oadapter.New("https://api-inference.huggingface.co")}
}

func (a *Adapter) Name() string                    { return "huggingface" }
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
