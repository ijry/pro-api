package gemini

import (
	"testing"

	"github.com/ijry/pro-api/internal/protocol/ir"
)

var _ ir.ChatRequest

func TestDecode_BasicChat(t *testing.T) {
	body := []byte(`{
		"contents": [
			{"role": "user", "parts": [{"text": "Hello"}]}
		],
		"generationConfig": {"maxOutputTokens": 1024, "temperature": 0.7}
	}`)
	req, err := DecodeRequest(body, "gemini-1.5-flash")
	if err != nil {
		t.Fatal(err)
	}
	if req.Model != "gemini-1.5-flash" {
		t.Errorf("model: %q", req.Model)
	}
	if req.MaxTokens != 1024 {
		t.Errorf("maxTokens: %d", req.MaxTokens)
	}
	if len(req.Messages) != 1 || req.Messages[0].Role != "user" {
		t.Errorf("messages: %+v", req.Messages)
	}
}

func TestDecode_SystemInstruction(t *testing.T) {
	body := []byte(`{
		"system_instruction": {"parts": [{"text": "Be helpful."}]},
		"contents": [{"role": "user", "parts": [{"text": "Hi"}]}]
	}`)
	req, err := DecodeRequest(body, "gemini-1.5-flash")
	if err != nil {
		t.Fatal(err)
	}
	if len(req.Messages) != 2 || req.Messages[0].Role != "system" {
		t.Fatalf("expected system+user messages, got: %+v", req.Messages)
	}
}

func TestDecode_ModelRole(t *testing.T) {
	body := []byte(`{
		"contents": [
			{"role": "user", "parts": [{"text": "Hello"}]},
			{"role": "model", "parts": [{"text": "Hi there!"}]}
		]
	}`)
	req, err := DecodeRequest(body, "gemini-1.5-flash")
	if err != nil {
		t.Fatal(err)
	}
	if req.Messages[1].Role != "assistant" {
		t.Errorf("model role should map to assistant, got %s", req.Messages[1].Role)
	}
}

func TestEncode_Response(t *testing.T) {
	resp := &ir.ChatResponse{
		Choices: []ir.Choice{{
			Message:      ir.Message{Role: "assistant", Content: []ir.ContentPart{{Type: "text", Text: "Hi!"}}},
			FinishReason: "stop",
		}},
		Usage: ir.Usage{PromptTokens: 5, CompletionTokens: 3},
	}
	b, err := EncodeResponse(resp)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) == 0 {
		t.Fatal("empty response")
	}
}
