package anthropic

import (
	"encoding/json"
	"fmt"

	"github.com/ijry/pro-api/internal/protocol/ir"
)

// SSEEvent は一条 Anthropic SSE イベント。
type SSEEvent struct {
	Event string
	Data  []byte
}

// EncodeResponse は IR 响应转为 Anthropic Messages API 响应 JSON。
func EncodeResponse(resp *ir.ChatResponse) ([]byte, error) {
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("anthropic encode: no choices")
	}
	choice := resp.Choices[0]

	type textBlock struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	type toolUseBlock struct {
		Type  string         `json:"type"`
		ID    string         `json:"id"`
		Name  string         `json:"name"`
		Input map[string]any `json:"input"`
	}

	var content []any
	for _, part := range choice.Message.Content {
		if part.Type == "text" {
			content = append(content, textBlock{Type: "text", Text: part.Text})
		}
	}
	for _, tc := range choice.Message.ToolCalls {
		var input map[string]any
		_ = json.Unmarshal([]byte(tc.Function.Arguments), &input)
		content = append(content, toolUseBlock{
			Type:  "tool_use",
			ID:    tc.ID,
			Name:  tc.Function.Name,
			Input: input,
		})
	}
	if len(content) == 0 {
		content = []any{textBlock{Type: "text", Text: ""}}
	}

	stopReason := mapStopReason(choice.FinishReason)

	out := map[string]any{
		"id":            resp.ID,
		"type":          "message",
		"role":          "assistant",
		"model":         resp.Model,
		"content":       content,
		"stop_reason":   stopReason,
		"stop_sequence": nil,
		"usage": map[string]int{
			"input_tokens":  resp.Usage.PromptTokens,
			"output_tokens": resp.Usage.CompletionTokens,
		},
	}
	return json.Marshal(out)
}

// EncodeChunk は IR chunk 转为 Anthropic SSE events。
func EncodeChunk(chunk *ir.ChatChunk, index int, isStop bool) ([]SSEEvent, error) {
	var events []SSEEvent

	if chunk.Delta.Content != "" {
		data, _ := json.Marshal(map[string]any{
			"type":  "content_block_delta",
			"index": index,
			"delta": map[string]string{"type": "text_delta", "text": chunk.Delta.Content},
		})
		events = append(events, SSEEvent{Event: "content_block_delta", Data: data})
	}

	if isStop || chunk.FinishReason != "" {
		stopReason := mapStopReason(chunk.FinishReason)
		if stopReason == "" {
			stopReason = "end_turn"
		}

		deltaData, _ := json.Marshal(map[string]any{
			"type":  "message_delta",
			"delta": map[string]string{"type": "message_delta", "stop_reason": stopReason},
			"usage": buildUsage(chunk.Usage),
		})
		events = append(events, SSEEvent{Event: "message_delta", Data: deltaData})

		stopData, _ := json.Marshal(map[string]string{"type": "message_stop"})
		events = append(events, SSEEvent{Event: "message_stop", Data: stopData})
	}

	return events, nil
}

func mapStopReason(r string) string {
	switch r {
	case "stop", "":
		return "end_turn"
	case "length":
		return "max_tokens"
	case "tool_calls", "tool_use":
		return "tool_use"
	default:
		return r
	}
}

func buildUsage(u *ir.Usage) map[string]int {
	if u == nil {
		return map[string]int{"output_tokens": 0}
	}
	return map[string]int{
		"input_tokens":  u.PromptTokens,
		"output_tokens": u.CompletionTokens,
	}
}
