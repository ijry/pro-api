// Package groq provides a Groq adapter (OpenAI-compatible).
package groq

import (
	"context"

	"github.com/ijry/pro-api/internal/adapter"
	oadapter "github.com/ijry/pro-api/internal/adapter/openai"
	"github.com/ijry/pro-api/internal/protocol/ir"
)

var supportedModels = []string{
	"llama-3.3-70b-versatile", "llama-3.1-70b-versatile",
	"llama-3.1-8b-instant", "mixtral-8x7b-32768", "gemma2-9b-it",
}

// Adapter wraps the OpenAI adapter with Groq-specific settings.
type Adapter struct {
	base *oadapter.OpenAI
}

// New returns a Groq adapter.
func New() adapter.Adapter {
	return &Adapter{base: oadapter.New("https://api.groq.com/openai")}
}

func (a *Adapter) Name() string                                                                        { return "groq" }
func (a *Adapter) Capabilities() adapter.Capability                                                   { return a.base.Capabilities() }
func (a *Adapter) SupportedModels() []string                                                           { return supportedModels }
func (a *Adapter) Chat(ctx context.Context, req *ir.ChatRequest, cred adapter.Credential) (*ir.ChatResponse, error) {
	return a.base.Chat(ctx, req, cred)
}
func (a *Adapter) ChatStream(ctx context.Context, req *ir.ChatRequest, cred adapter.Credential) (adapter.StreamReader, error) {
	return a.base.ChatStream(ctx, req, cred)
}
func (a *Adapter) Embed(ctx context.Context, req *ir.EmbedRequest, cred adapter.Credential) (*ir.EmbedResponse, error) {
	return a.base.Embed(ctx, req, cred)
}
