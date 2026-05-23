package gemini

import (
	"context"
	"encoding/json"
	"io"

	"github.com/ijry/pro-api/internal/adapter"
	"github.com/ijry/pro-api/internal/protocol/ir"
)

// geminiStream reads Gemini SSE stream (JSON lines sent as SSE data events).
type geminiStream struct {
	reader       *adapter.SSEReader
	body         io.ReadCloser
	inputTokens  int
	outputTokens int
}

func newGeminiStream(body io.ReadCloser) *geminiStream {
	return &geminiStream{
		reader: adapter.NewSSEReader(body),
		body:   body,
	}
}

func (s *geminiStream) Close() error { return s.body.Close() }

func (s *geminiStream) Next(ctx context.Context) (*ir.ChatChunk, error) {
	for {
		select {
		case <-ctx.Done():
			return nil, adapter.ClassifyNetErr(ctx.Err())
		default:
		}

		ev, err := s.reader.Next()
		if err != nil {
			if err == io.EOF {
				return nil, io.EOF
			}
			return nil, adapter.ClassifyNetErr(err)
		}
		if ev.Data == "" {
			continue
		}

		var raw geminiResponse
		if err := json.Unmarshal([]byte(ev.Data), &raw); err != nil {
			continue
		}

		s.inputTokens = raw.UsageMetadata.PromptTokenCount
		s.outputTokens = raw.UsageMetadata.CandidatesTokenCount

		chunk := &ir.ChatChunk{}
		if len(raw.Candidates) > 0 {
			cand := raw.Candidates[0]
			for _, p := range cand.Content.Parts {
				if p.Text != "" {
					chunk.Delta.Content += p.Text
				}
			}
			chunk.FinishReason = mapFinishReason(cand.FinishReason)
		}

		if raw.UsageMetadata.TotalTokenCount > 0 {
			chunk.Usage = &ir.Usage{
				PromptTokens:     s.inputTokens,
				CompletionTokens: s.outputTokens,
				TotalTokens:      raw.UsageMetadata.TotalTokenCount,
			}
		}

		if chunk.Delta.Content == "" && chunk.FinishReason == "" && chunk.Usage == nil {
			continue
		}

		if chunk.FinishReason == ir.FinishStop || chunk.FinishReason != "" && chunk.Delta.Content == "" {
			return chunk, nil
		}
		return chunk, nil
	}
}
