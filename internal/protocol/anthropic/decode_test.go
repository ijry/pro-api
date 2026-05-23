package anthropic

import (
	"testing"

	"github.com/ijry/pro-api/internal/protocol/ir"
)

// Ensure ir package is used
var _ ir.ChatRequest

func TestDecode_BasicTextMessage(t *testing.T) {
	body := []byte(`{
		"model":"claude-3-5-sonnet-20241022",
		"max_tokens":1024,
		"messages":[{"role":"user","content":"Hello"}]
	}`)
	req, err := DecodeRequest(body)
	if err != nil {
		t.Fatal(err)
	}
	if req.Model != "claude-3-5-sonnet-20241022" {
		t.Errorf("model: got %q", req.Model)
	}
	if req.MaxTokens != 1024 {
		t.Errorf("max_tokens: got %d", req.MaxTokens)
	}
	if len(req.Messages) != 1 || req.Messages[0].Role != "user" {
		t.Errorf("messages: %+v", req.Messages)
	}
}

func TestDecode_SystemPrompt(t *testing.T) {
	body := []byte(`{
		"model":"claude-3-5-sonnet-20241022",
		"max_tokens":512,
		"system":"You are helpful.",
		"messages":[{"role":"user","content":"Hi"}]
	}`)
	req, err := DecodeRequest(body)
	if err != nil {
		t.Fatal(err)
	}
	if len(req.Messages) != 2 {
		t.Fatalf("want 2 messages (system+user), got %d", len(req.Messages))
	}
	if req.Messages[0].Role != "system" {
		t.Errorf("first message should be system, got %s", req.Messages[0].Role)
	}
	if len(req.Messages[0].Content) == 0 || req.Messages[0].Content[0].Text != "You are helpful." {
		t.Errorf("system content wrong: %+v", req.Messages[0].Content)
	}
}

func TestDecode_MultiPartContent(t *testing.T) {
	body := []byte(`{
		"model":"claude-3-5-sonnet-20241022",
		"max_tokens":512,
		"messages":[{
			"role":"user",
			"content":[
				{"type":"text","text":"What is this?"},
				{"type":"image","source":{"type":"url","url":"https://example.com/img.png"}}
			]
		}]
	}`)
	req, err := DecodeRequest(body)
	if err != nil {
		t.Fatal(err)
	}
	if len(req.Messages[0].Content) != 2 {
		t.Fatalf("want 2 parts, got %d", len(req.Messages[0].Content))
	}
	if req.Messages[0].Content[1].Type != "image_url" {
		t.Errorf("second part should be image_url, got %s", req.Messages[0].Content[1].Type)
	}
}

func TestDecode_ToolUse(t *testing.T) {
	body := []byte(`{
		"model":"claude-3-5-sonnet-20241022",
		"max_tokens":512,
		"tools":[{"name":"get_weather","description":"Get weather","input_schema":{"type":"object","properties":{"city":{"type":"string"}}}}],
		"messages":[{"role":"user","content":"What's the weather?"}]
	}`)
	req, err := DecodeRequest(body)
	if err != nil {
		t.Fatal(err)
	}
	if len(req.Tools) != 1 || req.Tools[0].Function.Name != "get_weather" {
		t.Errorf("tools: %+v", req.Tools)
	}
}

func TestDecode_Stream(t *testing.T) {
	body := []byte(`{"model":"claude-3-5-sonnet-20241022","max_tokens":512,"stream":true,"messages":[{"role":"user","content":"Hi"}]}`)
	req, err := DecodeRequest(body)
	if err != nil {
		t.Fatal(err)
	}
	if !req.Stream {
		t.Error("stream should be true")
	}
}
