package openai_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/ijry/pro-api/internal/protocol/ir"
	"github.com/ijry/pro-api/internal/protocol/openai"
	"github.com/ijry/pro-api/pkg/apierr"
)

func TestStreamWriter_WritesData(t *testing.T) {
	var buf bytes.Buffer
	called := 0
	sw := openai.NewStreamWriter(&buf, func() { called++ })
	if err := sw.WriteChunk(&ir.ChatChunk{
		ID: "x", Model: "gpt-4o", Delta: ir.Delta{Content: "hi"},
	}); err != nil {
		t.Fatalf("write: %v", err)
	}
	out := buf.String()
	if !strings.HasPrefix(out, "data: ") || !strings.HasSuffix(out, "\n\n") {
		t.Fatalf("bad SSE framing: %q", out)
	}
	if called != 1 {
		t.Fatalf("expected flush called once, got %d", called)
	}
}

func TestStreamWriter_WriteDone(t *testing.T) {
	var buf bytes.Buffer
	sw := openai.NewStreamWriter(&buf, nil)
	if err := sw.WriteDone(); err != nil {
		t.Fatalf("write: %v", err)
	}
	if buf.String() != "data: [DONE]\n\n" {
		t.Fatalf("bad DONE: %q", buf.String())
	}
}

func TestStreamWriter_WriteError_EnvelopeAndDone(t *testing.T) {
	var buf bytes.Buffer
	sw := openai.NewStreamWriter(&buf, nil)
	if err := sw.WriteError(apierr.New(apierr.CodeUpstreamRateLimit, "slow down"), "rate_limit_exceeded", "rate_limit_exceeded"); err != nil {
		t.Fatalf("write: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `"message":"slow down"`) {
		t.Fatalf("missing error envelope: %s", out)
	}
	if !strings.HasSuffix(out, "data: [DONE]\n\n") {
		t.Fatalf("missing DONE: %s", out)
	}
}

func TestStreamWriter_ToolCallDelta(t *testing.T) {
	var buf bytes.Buffer
	sw := openai.NewStreamWriter(&buf, nil)
	chunk := &ir.ChatChunk{
		ID: "x", Model: "gpt-4o",
		Delta: ir.Delta{ToolCalls: []ir.ToolCallDelta{
			{Index: 0, ID: "c1", Function: ir.ToolCallFunctionDelta{Name: "f", Arguments: `{"a":1}`}},
		}},
	}
	if err := sw.WriteChunk(chunk); err != nil {
		t.Fatalf("write: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `"tool_calls"`) || !strings.Contains(out, `"name":"f"`) {
		t.Fatalf("missing tool_calls: %s", out)
	}
}

func TestStreamWriter_UsageInLastChunk(t *testing.T) {
	var buf bytes.Buffer
	sw := openai.NewStreamWriter(&buf, nil)
	chunk := &ir.ChatChunk{
		ID: "x", Model: "gpt-4o", FinishReason: ir.FinishStop,
		Usage: &ir.Usage{PromptTokens: 3, CompletionTokens: 5, TotalTokens: 8},
	}
	if err := sw.WriteChunk(chunk); err != nil {
		t.Fatalf("write: %v", err)
	}
	out := buf.String()
	// extract JSON portion
	data := strings.TrimPrefix(strings.TrimSuffix(out, "\n\n"), "data: ")
	var dto map[string]any
	if err := json.Unmarshal([]byte(data), &dto); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	usage, ok := dto["usage"].(map[string]any)
	if !ok {
		t.Fatalf("usage missing: %v", dto)
	}
	if v, _ := usage["total_tokens"].(float64); v != 8 {
		t.Fatalf("total_tokens expected 8, got %v", v)
	}
}
