// Package grokbuild provides xAI Grok Build adapter support via the official OpenAI-compatible API.
package grokbuild

import (
	"context"
	"fmt"

	"github.com/ijry/pro-api/internal/adapter"
	oadapter "github.com/ijry/pro-api/internal/adapter/openai"
	"github.com/ijry/pro-api/internal/protocol/ir"
)

const defaultBaseURL = "https://api.x.ai"

var supportedModels = []string{
	"grok-4",
	"grok-3",
	"grok-3-mini",
	"grok-3-mini-fast",
}

// Adapter wraps the OpenAI-compatible adapter with xAI-specific identity and capability limits.
type Adapter struct {
	base *oadapter.OpenAI
}

// New returns the default xAI official API adapter.
func New() adapter.Adapter {
	return NewWithBase(defaultBaseURL)
}

// NewWithBase returns an adapter that uses baseURL before appending /v1 paths.
func NewWithBase(baseURL string) adapter.Adapter {
	return &Adapter{base: oadapter.New(baseURL)}
}

func (a *Adapter) Name() string {
	return "grok-build"
}

func (a *Adapter) Capabilities() adapter.Capability {
	return adapter.CapChat | adapter.CapStream
}

func (a *Adapter) SupportedModels() []string {
	return supportedModels
}

func (a *Adapter) Chat(ctx context.Context, req *ir.ChatRequest, cred adapter.Credential) (*ir.ChatResponse, error) {
	return a.base.Chat(ctx, req, cred)
}

func (a *Adapter) ChatStream(ctx context.Context, req *ir.ChatRequest, cred adapter.Credential) (adapter.StreamReader, error) {
	return a.base.ChatStream(ctx, req, cred)
}

func (a *Adapter) Embed(context.Context, *ir.EmbedRequest, adapter.Credential) (*ir.EmbedResponse, error) {
	return nil, fmt.Errorf("grok-build: embedding not supported")
}
