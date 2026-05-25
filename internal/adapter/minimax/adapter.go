// Package minimax provides a MiniMax adapter (near-OpenAI-compatible, requires GroupID).
package minimax

import (
	"context"

	"github.com/ijry/pro-api/internal/adapter"
	oadapter "github.com/ijry/pro-api/internal/adapter/openai"
	"github.com/ijry/pro-api/internal/protocol/ir"
)

var supportedModels = []string{
	"abab6.5s-chat", "abab6.5-chat",
	"abab5.5s-chat", "abab5.5-chat",
}

// Adapter wraps OpenAI adapter with MiniMax-specific settings.
type Adapter struct{ base *oadapter.OpenAI }

// New returns a MiniMax adapter.
func New() adapter.Adapter {
	return &Adapter{base: oadapter.New("https://api.minimax.chat")}
}

func (a *Adapter) Name() string                    { return "minimax" }
func (a *Adapter) Capabilities() adapter.Capability { return a.base.Capabilities() }
func (a *Adapter) SupportedModels() []string        { return supportedModels }

// injectGroupID extracts group_id from Credential.Extra and appends it to BaseURL.
func injectGroupID(cred adapter.Credential) adapter.Credential {
	if cred.Extra == nil {
		return cred
	}
	groupID := cred.Extra["group_id"]
	if groupID == "" {
		return cred
	}
	// MiniMax requires GroupId as query param on the API URL
	if cred.BaseURL == "" {
		cred.BaseURL = "https://api.minimax.chat/v1?GroupId=" + groupID
	}
	return cred
}

func (a *Adapter) Chat(ctx context.Context, req *ir.ChatRequest, cred adapter.Credential) (*ir.ChatResponse, error) {
	return a.base.Chat(ctx, req, injectGroupID(cred))
}
func (a *Adapter) ChatStream(ctx context.Context, req *ir.ChatRequest, cred adapter.Credential) (adapter.StreamReader, error) {
	return a.base.ChatStream(ctx, req, injectGroupID(cred))
}
func (a *Adapter) Embed(ctx context.Context, req *ir.EmbedRequest, cred adapter.Credential) (*ir.EmbedResponse, error) {
	return a.base.Embed(ctx, req, injectGroupID(cred))
}
