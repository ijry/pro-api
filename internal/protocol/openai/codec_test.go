package openai_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/ijry/pro-api/internal/protocol/ir"
	"github.com/ijry/pro-api/internal/protocol/openai"
	"github.com/ijry/pro-api/pkg/apierr"
)

func TestDecodeChat_BasicText(t *testing.T) {
	body := `{"model":"gpt-4o","messages":[{"role":"user","content":"hi"}]}`
	r, err := openai.DecodeChat(strings.NewReader(body), openai.DecodeOptions{})
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if r.Model != "gpt-4o" || len(r.Messages) != 1 {
		t.Fatalf("unexpected req: %+v", r)
	}
	if r.Messages[0].Content[0].Text != "hi" {
		t.Fatalf("text mismatch: %q", r.Messages[0].Content[0].Text)
	}
}

func TestDecodeChat_ContentAsArray(t *testing.T) {
	body := `{"model":"gpt-4o","messages":[{"role":"user","content":[
        {"type":"text","text":"hi"},
        {"type":"image_url","image_url":{"url":"https://x/y.png","detail":"high"}}
    ]}]}`
	r, err := openai.DecodeChat(strings.NewReader(body), openai.DecodeOptions{})
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	parts := r.Messages[0].Content
	if len(parts) != 2 || parts[1].ImageURL.URL != "https://x/y.png" {
		t.Fatalf("unexpected parts: %+v", parts)
	}
}

func TestDecodeChat_MaxCompletionTokensOverridesMaxTokens(t *testing.T) {
	body := `{"model":"o1","messages":[{"role":"user","content":"x"}],"max_tokens":10,"max_completion_tokens":42}`
	r, err := openai.DecodeChat(strings.NewReader(body), openai.DecodeOptions{})
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if r.MaxTokens != 42 {
		t.Fatalf("expected 42 (override), got %d", r.MaxTokens)
	}
}

func TestDecodeChat_StopAsString(t *testing.T) {
	body := `{"model":"gpt-4o","messages":[{"role":"user","content":"x"}],"stop":"end"}`
	r, err := openai.DecodeChat(strings.NewReader(body), openai.DecodeOptions{})
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(r.Stop) != 1 || r.Stop[0] != "end" {
		t.Fatalf("stop mismatch: %v", r.Stop)
	}
}

func TestDecodeChat_StopAsArray(t *testing.T) {
	body := `{"model":"gpt-4o","messages":[{"role":"user","content":"x"}],"stop":["a","b"]}`
	r, err := openai.DecodeChat(strings.NewReader(body), openai.DecodeOptions{})
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(r.Stop) != 2 || r.Stop[1] != "b" {
		t.Fatalf("stop mismatch: %v", r.Stop)
	}
}

func TestDecodeChat_ToolsAndToolChoice(t *testing.T) {
	body := `{"model":"gpt-4o","messages":[{"role":"user","content":"x"}],
        "tools":[{"type":"function","function":{"name":"f","description":"d","parameters":{"k":1}}}],
        "tool_choice":"auto"}`
	r, err := openai.DecodeChat(strings.NewReader(body), openai.DecodeOptions{})
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(r.Tools) != 1 || r.Tools[0].Function.Name != "f" {
		t.Fatalf("tools mismatch: %+v", r.Tools)
	}
	if s, _ := r.ToolChoice.(string); s != "auto" {
		t.Fatalf("tool_choice expected 'auto', got %v", r.ToolChoice)
	}
}

func TestDecodeChat_AssistantToolCalls(t *testing.T) {
	body := `{"model":"gpt-4o","messages":[
        {"role":"assistant","tool_calls":[{"id":"c1","type":"function","function":{"name":"echo","arguments":"{}"}}]}
    ]}`
	r, err := openai.DecodeChat(strings.NewReader(body), openai.DecodeOptions{})
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(r.Messages[0].ToolCalls) != 1 || r.Messages[0].ToolCalls[0].ID != "c1" {
		t.Fatalf("tool_calls mismatch: %+v", r.Messages[0].ToolCalls)
	}
}

func TestDecodeChat_ToolMessageRoleAndCallID(t *testing.T) {
	body := `{"model":"gpt-4o","messages":[{"role":"tool","tool_call_id":"c1","content":"out"}]}`
	r, err := openai.DecodeChat(strings.NewReader(body), openai.DecodeOptions{})
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if r.Messages[0].Role != ir.RoleTool || r.Messages[0].ToolCallID != "c1" {
		t.Fatalf("tool message mismatch: %+v", r.Messages[0])
	}
}

func TestDecodeChat_UnknownFieldStrict(t *testing.T) {
	body := `{"model":"gpt-4o","messages":[{"role":"user","content":"x"}],"weird_field":1}`
	_, err := openai.DecodeChat(strings.NewReader(body), openai.DecodeOptions{})
	if err == nil {
		t.Fatalf("expected error in strict mode")
	}
	var ae *apierr.Error
	if !errors.As(err, &ae) || ae.Code != apierr.CodeInvalidParam {
		t.Fatalf("expected CodeInvalidParam, got %v", err)
	}
}

func TestDecodeChat_UnknownFieldAllowed(t *testing.T) {
	body := `{"model":"gpt-4o","messages":[{"role":"user","content":"x"}],"weird_field":1}`
	_, err := openai.DecodeChat(strings.NewReader(body), openai.DecodeOptions{AllowUnknownFields: true})
	if err != nil {
		t.Fatalf("unexpected error in lax mode: %v", err)
	}
}

func TestDecodeEmbed_StringInput(t *testing.T) {
	body := `{"model":"text-embedding-3-small","input":"hello"}`
	r, err := openai.DecodeEmbed(strings.NewReader(body), openai.DecodeOptions{})
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(r.Input) != 1 || r.Input[0] != "hello" {
		t.Fatalf("input mismatch: %v", r.Input)
	}
}

func TestDecodeEmbed_ArrayInput(t *testing.T) {
	body := `{"model":"text-embedding-3-small","input":["a","b"]}`
	r, err := openai.DecodeEmbed(strings.NewReader(body), openai.DecodeOptions{})
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(r.Input) != 2 || r.Input[1] != "b" {
		t.Fatalf("input mismatch: %v", r.Input)
	}
}

func TestDecodeCompletion_StringPrompt(t *testing.T) {
	body := `{"model":"gpt-3.5-turbo-instruct","prompt":"hello"}`
	r, err := openai.DecodeCompletion(strings.NewReader(body), openai.DecodeOptions{})
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(r.Prompt) != 1 || r.Prompt[0] != "hello" {
		t.Fatalf("prompt: %v", r.Prompt)
	}
}

func TestDecodeCompletion_ArrayPrompt(t *testing.T) {
	body := `{"model":"gpt-3.5-turbo-instruct","prompt":["a","b"]}`
	r, err := openai.DecodeCompletion(strings.NewReader(body), openai.DecodeOptions{})
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(r.Prompt) != 2 {
		t.Fatalf("prompt: %v", r.Prompt)
	}
}

func TestEncodeChat_SingleTextChoice(t *testing.T) {
	resp := &ir.ChatResponse{
		ID:    "chatcmpl-x",
		Model: "gpt-4o",
		Choices: []ir.Choice{{
			Index:        0,
			Message:      ir.Message{Role: ir.RoleAssistant, Content: []ir.ContentPart{{Type: ir.ContentText, Text: "Hi!"}}},
			FinishReason: ir.FinishStop,
		}},
		Usage: ir.Usage{PromptTokens: 5, CompletionTokens: 2, TotalTokens: 7},
	}
	out := openai.EncodeChat(resp)
	if out.Object != "chat.completion" || out.ID != "chatcmpl-x" {
		t.Fatalf("bad envelope: %+v", out)
	}
	if len(out.Choices) != 1 {
		t.Fatalf("expected 1 choice")
	}
	// single-text → content encoded as string ("Hi!")
	if string(out.Choices[0].Message.Content) != `"Hi!"` {
		t.Fatalf("expected JSON string, got %s", string(out.Choices[0].Message.Content))
	}
}

func TestEncodeChat_WithToolCalls(t *testing.T) {
	resp := &ir.ChatResponse{
		ID: "x", Model: "gpt-4o",
		Choices: []ir.Choice{{
			Message: ir.Message{Role: ir.RoleAssistant, ToolCalls: []ir.ToolCall{
				{ID: "c1", Type: "function", Function: ir.ToolCallFunction{Name: "f", Arguments: "{}"}},
			}},
			FinishReason: ir.FinishToolCalls,
		}},
	}
	out := openai.EncodeChat(resp)
	if len(out.Choices[0].Message.ToolCalls) != 1 || out.Choices[0].Message.ToolCalls[0].ID != "c1" {
		t.Fatalf("tool_calls missing: %+v", out.Choices[0].Message.ToolCalls)
	}
}

func TestEncodeEmbed_FloatEncoding(t *testing.T) {
	resp := &ir.EmbedResponse{
		Model: "text-embedding-3-small",
		Data:  []ir.EmbedData{{Index: 0, Embedding: []float32{0.5, -0.5}}},
		Usage: ir.EmbedUsage{PromptTokens: 1, TotalTokens: 1},
	}
	out := openai.EncodeEmbed(resp, "float")
	if len(out.Data) != 1 {
		t.Fatalf("data missing")
	}
	if _, ok := out.Data[0].Embedding.([]float32); !ok {
		t.Fatalf("expected []float32, got %T", out.Data[0].Embedding)
	}
}

func TestEncodeEmbed_Base64Encoding(t *testing.T) {
	resp := &ir.EmbedResponse{
		Model: "text-embedding-3-small",
		Data:  []ir.EmbedData{{Index: 0, Embedding: []float32{1.0}}},
	}
	out := openai.EncodeEmbed(resp, "base64")
	s, ok := out.Data[0].Embedding.(string)
	if !ok || s == "" {
		t.Fatalf("expected base64 string, got %T (%v)", out.Data[0].Embedding, out.Data[0].Embedding)
	}
}
