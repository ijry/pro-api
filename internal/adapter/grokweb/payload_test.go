package grokweb

import (
	"strings"
	"testing"

	"github.com/ijry/pro-api/internal/adapter"
	"github.com/ijry/pro-api/internal/protocol/ir"
)

func TestLookupModel(t *testing.T) {
	got, ok := lookupModel("grok-4.1-fast")
	if !ok {
		t.Fatalf("expected model mapping")
	}
	if got.ModelName != "grok-4-1-thinking-1129" || got.ModelMode != "MODEL_MODE_FAST" {
		t.Fatalf("mapping = %#v", got)
	}
}

func TestLookupModelAcceptsCatalogPrefix(t *testing.T) {
	got, ok := lookupModel("grok-web/grok-4")
	if !ok {
		t.Fatalf("expected prefixed model mapping")
	}
	if got.ModelName != "grok-4" || got.ModelMode != "MODEL_MODE_GROK_4" {
		t.Fatalf("mapping = %#v", got)
	}
}

func TestBuildSSOCookie(t *testing.T) {
	got := buildSSOCookie("sso=abc123")
	if got != "sso=abc123; sso-rw=abc123" {
		t.Fatalf("cookie = %q", got)
	}
}

func TestBuildHeaders(t *testing.T) {
	h := buildHeaders(adapter.Credential{APIKey: "abc123"})
	if h.Get("Cookie") != "sso=abc123; sso-rw=abc123" {
		t.Fatalf("cookie = %q", h.Get("Cookie"))
	}
	if h.Get("Origin") != "https://grok.com" {
		t.Fatalf("origin = %q", h.Get("Origin"))
	}
	if h.Get("x-xai-request-id") == "" {
		t.Fatalf("missing request id")
	}
	if h.Get("x-statsig-id") == "" {
		t.Fatalf("missing statsig id")
	}
}

func TestBuildPayloadFlattensMessagesAndModel(t *testing.T) {
	temp := 0.7
	topP := 0.9
	body, err := buildPayload(&ir.ChatRequest{
		Model:       "grok-4-thinking",
		Temperature: &temp,
		TopP:        &topP,
		Messages: []ir.Message{
			{Role: ir.RoleSystem, Content: []ir.ContentPart{{Type: ir.ContentText, Text: "be direct"}}},
			{Role: ir.RoleUser, Content: []ir.ContentPart{{Type: ir.ContentText, Text: "hello"}}},
			{Role: ir.RoleAssistant, Content: []ir.ContentPart{{Type: ir.ContentText, Text: "hi"}}},
			{Role: ir.RoleUser, Content: []ir.ContentPart{{Type: ir.ContentText, Text: "continue"}}},
		},
	})
	if err != nil {
		t.Fatalf("buildPayload error: %v", err)
	}
	if body["modelName"] != "grok-4" || body["modelMode"] != "MODEL_MODE_GROK_4_THINKING" {
		t.Fatalf("model fields = %#v / %#v", body["modelName"], body["modelMode"])
	}
	msg, _ := body["message"].(string)
	for _, part := range []string{"system: be direct", "user: hello", "assistant: hi", "continue"} {
		if !strings.Contains(msg, part) {
			t.Fatalf("message %q missing %q", msg, part)
		}
	}
	meta := body["responseMetadata"].(map[string]any)
	override := meta["modelConfigOverride"].(map[string]any)
	if override["temperature"] != temp || override["topP"] != topP {
		t.Fatalf("override = %#v", override)
	}
}

func TestBuildPayloadRejectsUnknownModel(t *testing.T) {
	_, err := buildPayload(&ir.ChatRequest{Model: "unknown"})
	if err == nil {
		t.Fatalf("expected unknown model error")
	}
}
