package anthropic

import (
	"context"
	"encoding/json"
	"io"

	"github.com/ijry/pro-api/internal/adapter"
	"github.com/ijry/pro-api/internal/protocol/ir"
)

type anthropicStream struct {
	reader *adapter.SSEReader
	body   io.ReadCloser
	// accumulate usage from message_delta
	inputTokens  int
	outputTokens int
}

func newAnthropicStream(body io.ReadCloser) *anthropicStream {
	return &anthropicStream{
		reader: adapter.NewSSEReader(body),
		body:   body,
	}
}

func (s *anthropicStream) Close() error { return s.body.Close() }

func (s *anthropicStream) Next(ctx context.Context) (*ir.ChatChunk, error) {
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

		switch ev.Event {
		case "content_block_delta":
			chunk, err := s.decodeDelta(ev.Data)
			if err != nil {
				continue
			}
			return chunk, nil

		case "message_delta":
			// Usage and stop reason
			var raw struct {
				Delta struct {
					StopReason string `json:"stop_reason"`
				} `json:"delta"`
				Usage struct {
					OutputTokens int `json:"output_tokens"`
				} `json:"usage"`
			}
			if err := json.Unmarshal([]byte(ev.Data), &raw); err == nil {
				s.outputTokens = raw.Usage.OutputTokens
				if raw.Delta.StopReason != "" {
					chunk := &ir.ChatChunk{
						FinishReason: mapStopReason(raw.Delta.StopReason),
						Usage: &ir.Usage{
							PromptTokens:     s.inputTokens,
							CompletionTokens: s.outputTokens,
							TotalTokens:      s.inputTokens + s.outputTokens,
						},
					}
					return chunk, nil
				}
			}
			continue

		case "message_start":
			var raw struct {
				Message struct {
					Usage struct {
						InputTokens int `json:"input_tokens"`
					} `json:"usage"`
				} `json:"message"`
			}
			if err := json.Unmarshal([]byte(ev.Data), &raw); err == nil {
				s.inputTokens = raw.Message.Usage.InputTokens
			}
			continue

		case "message_stop":
			return nil, io.EOF

		case "error":
			return nil, adapter.ClassifyHTTPStatus(500, []byte(ev.Data))

		default:
			continue
		}
	}
}

func (s *anthropicStream) decodeDelta(data string) (*ir.ChatChunk, error) {
	var raw struct {
		Index int `json:"index"`
		Delta struct {
			Type        string `json:"type"`
			Text        string `json:"text"`
			PartialJSON string `json:"partial_json"`
		} `json:"delta"`
	}
	if err := json.Unmarshal([]byte(data), &raw); err != nil {
		return nil, err
	}
	chunk := &ir.ChatChunk{}
	switch raw.Delta.Type {
	case "text_delta":
		chunk.Delta.Content = raw.Delta.Text
	case "input_json_delta":
		chunk.Delta.ToolCalls = []ir.ToolCallDelta{{
			Index: raw.Index,
			Function: ir.ToolCallFunctionDelta{Arguments: raw.Delta.PartialJSON},
		}}
	}
	return chunk, nil
}
