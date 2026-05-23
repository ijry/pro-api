package anthropic

import (
	"testing"

	"github.com/ijry/pro-api/internal/protocol/ir"
)

func TestEncodeResponse_TextContent(t *testing.T) {
	resp := &ir.ChatResponse{
		ID:    "msg_123",
		Model: "claude-3-5-sonnet-20241022",
		Choices: []ir.Choice{{
			Index:        0,
			FinishReason: "end_turn",
			Message: ir.Message{
				Role:    "assistant",
				Content: []ir.ContentPart{{Type: "text", Text: "Hello!"}},
			},
		}},
		Usage: ir.Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
	}

	b, err := EncodeResponse(resp)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) == 0 {
		t.Fatal("empty response")
	}
	// spot-check JSON fields
	s := string(b)
	if !contains(s, `"type":"message"`) && !contains(s, `"type": "message"`) {
		t.Errorf("missing type=message in: %s", s)
	}
}

func TestEncodeChunk_TextDelta(t *testing.T) {
	chunk := &ir.ChatChunk{
		ID:    "msg_123",
		Model: "claude-3-5-sonnet-20241022",
		Delta: ir.Delta{Content: "Hello"},
	}
	events, err := EncodeChunk(chunk, 0, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) == 0 {
		t.Fatal("expected events")
	}
}

func TestEncodeChunk_StopEvent(t *testing.T) {
	chunk := &ir.ChatChunk{
		ID:           "msg_123",
		FinishReason: "end_turn",
		Usage:        &ir.Usage{PromptTokens: 10, CompletionTokens: 5},
	}
	events, err := EncodeChunk(chunk, 5, true)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, e := range events {
		if e.Event == "message_stop" {
			found = true
		}
	}
	if !found {
		t.Error("expected message_stop event")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
