package ir_test

import (
	"errors"
	"testing"

	"github.com/ijry/pro-api/internal/protocol/ir"
	"github.com/ijry/pro-api/pkg/apierr"
)

func ptr[T any](v T) *T { return &v }

func mustErr(t *testing.T, err error, want apierr.Code) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error code %d, got nil", want)
	}
	var ae *apierr.Error
	if !errors.As(err, &ae) {
		t.Fatalf("expected *apierr.Error, got %T", err)
	}
	if ae.Code != want {
		t.Fatalf("expected code %d, got %d (%s)", want, ae.Code, ae.Message)
	}
}

func TestValidate_NilRequest(t *testing.T) {
	mustErr(t, ir.Validate(nil), apierr.CodeMissingParam)
}

func TestValidate_EmptyModel(t *testing.T) {
	mustErr(t, ir.Validate(&ir.ChatRequest{}), apierr.CodeMissingParam)
}

func TestValidate_EmptyMessages(t *testing.T) {
	mustErr(t, ir.Validate(&ir.ChatRequest{Model: "gpt-4o"}), apierr.CodeMissingParam)
}

func TestValidate_InvalidRole(t *testing.T) {
	r := &ir.ChatRequest{
		Model: "gpt-4o",
		Messages: []ir.Message{
			{Role: "alien", Content: []ir.ContentPart{{Type: ir.ContentText, Text: "hi"}}},
		},
	}
	mustErr(t, ir.Validate(r), apierr.CodeInvalidParam)
}

func TestValidate_ToolMessageWithoutCallID(t *testing.T) {
	r := &ir.ChatRequest{
		Model: "gpt-4o",
		Messages: []ir.Message{
			{Role: ir.RoleTool, Content: []ir.ContentPart{{Type: ir.ContentText, Text: "x"}}},
		},
	}
	mustErr(t, ir.Validate(r), apierr.CodeInvalidParam)
}

func TestValidate_ToolCallMissingID(t *testing.T) {
	r := &ir.ChatRequest{
		Model: "gpt-4o",
		Messages: []ir.Message{
			{
				Role:      ir.RoleAssistant,
				ToolCalls: []ir.ToolCall{{Type: "function", Function: ir.ToolCallFunction{Name: ""}}},
			},
		},
	}
	mustErr(t, ir.Validate(r), apierr.CodeInvalidParam)
}

func TestValidate_ContentPartEmptyType(t *testing.T) {
	r := &ir.ChatRequest{
		Model: "gpt-4o",
		Messages: []ir.Message{
			{Role: ir.RoleUser, Content: []ir.ContentPart{{Type: "", Text: "x"}}},
		},
	}
	mustErr(t, ir.Validate(r), apierr.CodeInvalidParam)
}

func TestValidate_NegativeMaxTokens(t *testing.T) {
	r := &ir.ChatRequest{
		Model: "gpt-4o", MaxTokens: -1,
		Messages: []ir.Message{{Role: ir.RoleUser, Content: []ir.ContentPart{{Type: ir.ContentText, Text: "x"}}}},
	}
	mustErr(t, ir.Validate(r), apierr.CodeInvalidParam)
}

func TestValidate_TemperatureOutOfRange(t *testing.T) {
	r := &ir.ChatRequest{
		Model: "gpt-4o", Temperature: ptr(3.0),
		Messages: []ir.Message{{Role: ir.RoleUser, Content: []ir.ContentPart{{Type: ir.ContentText, Text: "x"}}}},
	}
	mustErr(t, ir.Validate(r), apierr.CodeInvalidParam)
}

func TestValidate_TopPOutOfRange(t *testing.T) {
	r := &ir.ChatRequest{
		Model: "gpt-4o", TopP: ptr(1.5),
		Messages: []ir.Message{{Role: ir.RoleUser, Content: []ir.ContentPart{{Type: ir.ContentText, Text: "x"}}}},
	}
	mustErr(t, ir.Validate(r), apierr.CodeInvalidParam)
}

func TestValidate_TooManyStops(t *testing.T) {
	r := &ir.ChatRequest{
		Model: "gpt-4o",
		Stop:  []string{"a", "b", "c", "d", "e"},
		Messages: []ir.Message{
			{Role: ir.RoleUser, Content: []ir.ContentPart{{Type: ir.ContentText, Text: "x"}}},
		},
	}
	mustErr(t, ir.Validate(r), apierr.CodeInvalidParam)
}

func TestValidate_HappyPath(t *testing.T) {
	r := &ir.ChatRequest{
		Model:       "gpt-4o",
		MaxTokens:   100,
		Temperature: ptr(0.5),
		TopP:        ptr(0.9),
		Stop:        []string{"end"},
		Messages: []ir.Message{
			{Role: ir.RoleSystem, Content: []ir.ContentPart{{Type: ir.ContentText, Text: "you are helpful"}}},
			{Role: ir.RoleUser, Content: []ir.ContentPart{{Type: ir.ContentText, Text: "hi"}}},
			{Role: ir.RoleAssistant, ToolCalls: []ir.ToolCall{
				{ID: "call_1", Type: "function", Function: ir.ToolCallFunction{Name: "echo", Arguments: "{}"}},
			}},
			{Role: ir.RoleTool, ToolCallID: "call_1", Content: []ir.ContentPart{{Type: ir.ContentText, Text: "ok"}}},
		},
	}
	if err := ir.Validate(r); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestValidateEmbed_NilRequest(t *testing.T) {
	mustErr(t, ir.ValidateEmbed(nil), apierr.CodeMissingParam)
}

func TestValidateEmbed_EmptyInput(t *testing.T) {
	mustErr(t, ir.ValidateEmbed(&ir.EmbedRequest{Model: "text-embedding-3-small"}), apierr.CodeMissingParam)
}

func TestValidateEmbed_BlankItem(t *testing.T) {
	mustErr(t, ir.ValidateEmbed(&ir.EmbedRequest{Model: "text-embedding-3-small", Input: []string{"a", ""}}), apierr.CodeInvalidParam)
}

func TestValidateEmbed_HappyPath(t *testing.T) {
	if err := ir.ValidateEmbed(&ir.EmbedRequest{Model: "text-embedding-3-small", Input: []string{"hello"}}); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestCompletionToChat_JoinsPrompts(t *testing.T) {
	r := &ir.CompletionRequest{
		Model:  "gpt-3.5-turbo-instruct",
		Prompt: []string{"hello", "world"},
	}
	chat := r.ToChat()
	if chat.Model != r.Model {
		t.Fatalf("model mismatch: %s != %s", chat.Model, r.Model)
	}
	if len(chat.Messages) != 1 || chat.Messages[0].Role != ir.RoleUser {
		t.Fatalf("expected 1 user message")
	}
	if chat.Messages[0].Content[0].Text != "hello\nworld" {
		t.Fatalf("expected joined prompt, got %q", chat.Messages[0].Content[0].Text)
	}
}
