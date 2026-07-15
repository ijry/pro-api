package grokbuild

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ijry/pro-api/internal/adapter"
	"github.com/ijry/pro-api/internal/protocol/ir"
)

func TestAdapterMetadata(t *testing.T) {
	a := New()
	if a.Name() != "grok-build" {
		t.Fatalf("Name() = %q", a.Name())
	}
	caps := a.Capabilities()
	if !caps.Has(adapter.CapChat) || !caps.Has(adapter.CapStream) {
		t.Fatalf("missing chat/stream caps: %v", caps)
	}
	if caps.Has(adapter.CapEmbedding) || caps.Has(adapter.CapImage) || caps.Has(adapter.CapTTS) || caps.Has(adapter.CapSTT) {
		t.Fatalf("unexpected non-chat caps: %v", caps)
	}
	models := a.SupportedModels()
	want := []string{"grok-4", "grok-3", "grok-3-mini", "grok-3-mini-fast"}
	if len(models) != len(want) {
		t.Fatalf("models len = %d, want %d: %v", len(models), len(want), models)
	}
	for i := range want {
		if models[i] != want[i] {
			t.Fatalf("models[%d] = %q, want %q", i, models[i], want[i])
		}
	}
}

func TestChatUsesXAIOpenAICompatiblePath(t *testing.T) {
	var gotPath string
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if body["model"] != "grok-4" {
			t.Fatalf("model = %#v", body["model"])
		}
		_, _ = w.Write([]byte(`{"id":"chatcmpl-1","model":"grok-4","choices":[{"index":0,"message":{"role":"assistant","content":"pong"},"finish_reason":"stop"}],"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`))
	}))
	defer srv.Close()

	a := NewWithBase(srv.URL)
	resp, err := a.Chat(context.Background(), &ir.ChatRequest{
		Model: "grok-4",
		Messages: []ir.Message{{
			Role:    ir.RoleUser,
			Content: []ir.ContentPart{{Type: ir.ContentText, Text: "ping"}},
		}},
	}, adapter.Credential{APIKey: "xai-key"})
	if err != nil {
		t.Fatalf("Chat error: %v", err)
	}
	if gotPath != "/v1/chat/completions" {
		t.Fatalf("path = %q", gotPath)
	}
	if gotAuth != "Bearer xai-key" {
		t.Fatalf("auth = %q", gotAuth)
	}
	if resp.Choices[0].Message.Content[0].Text != "pong" {
		t.Fatalf("content = %#v", resp.Choices[0].Message.Content)
	}
}

func TestEmbedUnsupported(t *testing.T) {
	_, err := New().Embed(context.Background(), &ir.EmbedRequest{Model: "grok-4"}, adapter.Credential{APIKey: "x"})
	if err == nil {
		t.Fatalf("expected unsupported embedding error")
	}
}
