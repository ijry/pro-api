package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/ijry/pro-api/internal/adapter"
	"github.com/ijry/pro-api/internal/protocol/ir"
)

// sseStreamReader wraps an SSE stream body and implements adapter.StreamReader.
type sseStreamReader struct {
	reader *adapter.SSEReader
	body   io.ReadCloser
}

func newSSEStreamReader(body io.ReadCloser) *sseStreamReader {
	return &sseStreamReader{
		reader: adapter.NewSSEReader(body),
		body:   body,
	}
}

func (r *sseStreamReader) Close() error { return r.body.Close() }

func (r *sseStreamReader) Next(ctx context.Context) (*ir.ChatChunk, error) {
	for {
		select {
		case <-ctx.Done():
			return nil, adapter.ClassifyNetErr(ctx.Err())
		default:
		}
		ev, err := r.reader.Next()
		if err != nil {
			if err == io.EOF {
				return nil, io.EOF
			}
			return nil, adapter.ClassifyNetErr(err)
		}
		if ev.Data == "[DONE]" {
			return nil, io.EOF
		}
		chunk, err := decodeChunk([]byte(ev.Data))
		if err != nil {
			continue // skip malformed chunks
		}
		return chunk, nil
	}
}

func decodeChunk(data []byte) (*ir.ChatChunk, error) {
	var raw struct {
		ID      string `json:"id"`
		Model   string `json:"model"`
		Choices []struct {
			Index int `json:"index"`
			Delta struct {
				Role      string `json:"role"`
				Content   string `json:"content"`
				ToolCalls []struct {
					Index    int    `json:"index"`
					ID       string `json:"id"`
					Type     string `json:"type"`
					Function struct {
						Name      string `json:"name"`
						Arguments string `json:"arguments"`
					} `json:"function"`
				} `json:"tool_calls"`
			} `json:"delta"`
			FinishReason *string `json:"finish_reason"`
		} `json:"choices"`
		Usage *struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
			PromptTokensDetails *struct {
				CachedTokens int `json:"cached_tokens"`
			} `json:"prompt_tokens_details"`
			CompletionTokensDetails *struct {
				ReasoningTokens int `json:"reasoning_tokens"`
			} `json:"completion_tokens_details"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("openai: decode chunk: %w", err)
	}
	chunk := &ir.ChatChunk{ID: raw.ID, Model: raw.Model}
	if raw.Usage != nil {
		chunk.Usage = &ir.Usage{
			PromptTokens:     raw.Usage.PromptTokens,
			CompletionTokens: raw.Usage.CompletionTokens,
			TotalTokens:      raw.Usage.TotalTokens,
		}
		if raw.Usage.PromptTokensDetails != nil {
			chunk.Usage.CachedTokens = raw.Usage.PromptTokensDetails.CachedTokens
		}
		if raw.Usage.CompletionTokensDetails != nil {
			chunk.Usage.ReasoningTokens = raw.Usage.CompletionTokensDetails.ReasoningTokens
		}
	}
	// Take first choice's delta (IR uses single delta per chunk)
	if len(raw.Choices) > 0 {
		c := raw.Choices[0]
		chunk.Delta = ir.Delta{
			Role:    c.Delta.Role,
			Content: c.Delta.Content,
		}
		for _, tc := range c.Delta.ToolCalls {
			chunk.Delta.ToolCalls = append(chunk.Delta.ToolCalls, ir.ToolCallDelta{
				Index: tc.Index,
				ID:    tc.ID,
				Function: ir.ToolCallFunctionDelta{
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				},
			})
		}
		if c.FinishReason != nil {
			chunk.FinishReason = *c.FinishReason
		}
	}
	return chunk, nil
}
