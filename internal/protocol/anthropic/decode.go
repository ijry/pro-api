// Package anthropic 处理 Anthropic Messages API 的入口协议编解码。
//
// 注意：这是入口解码（客户端发过来的 Anthropic 格式 → IR）。
// 与 internal/adapter/anthropic（后端适配器，把 IR 发给 Claude API）是两个不同的包。
package anthropic

import (
	"encoding/json"
	"fmt"

	"github.com/ijry/pro-api/internal/protocol/ir"
)

type anthropicRequest struct {
	Model         string          `json:"model"`
	MaxTokens     int             `json:"max_tokens"`
	System        string          `json:"system,omitempty"`
	Messages      []anthropicMsg  `json:"messages"`
	Stream        bool            `json:"stream,omitempty"`
	Tools         []anthropicTool `json:"tools,omitempty"`
	ToolChoice    any             `json:"tool_choice,omitempty"`
	Temperature   *float64        `json:"temperature,omitempty"`
	TopP          *float64        `json:"top_p,omitempty"`
	StopSequences []string        `json:"stop_sequences,omitempty"`
}

type anthropicMsg struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

type anthropicBlock struct {
	Type      string          `json:"type"`
	Text      string          `json:"text,omitempty"`
	Source    *anthropicSrc   `json:"source,omitempty"`
	ID        string          `json:"id,omitempty"`
	Name      string          `json:"name,omitempty"`
	Input     map[string]any  `json:"input,omitempty"`
	ToolUseID string          `json:"tool_use_id,omitempty"`
	Content   json.RawMessage `json:"content,omitempty"`
}

type anthropicSrc struct {
	Type      string `json:"type"`
	URL       string `json:"url,omitempty"`
	MediaType string `json:"media_type,omitempty"`
	Data      string `json:"data,omitempty"`
}

type anthropicTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	InputSchema map[string]any `json:"input_schema"`
}

// DecodeRequest 把 Anthropic Messages API 请求体解析为 IR。
func DecodeRequest(body []byte) (*ir.ChatRequest, error) {
	var ar anthropicRequest
	if err := json.Unmarshal(body, &ar); err != nil {
		return nil, fmt.Errorf("anthropic decode: %w", err)
	}

	req := &ir.ChatRequest{
		Model:     ar.Model,
		MaxTokens: ar.MaxTokens,
		Stream:    ar.Stream,
	}
	if ar.Temperature != nil {
		req.Temperature = ar.Temperature
	}
	if ar.TopP != nil {
		req.TopP = ar.TopP
	}
	if len(ar.StopSequences) > 0 {
		req.Stop = ar.StopSequences
	}

	if ar.System != "" {
		req.Messages = append(req.Messages, ir.Message{
			Role:    "system",
			Content: []ir.ContentPart{{Type: "text", Text: ar.System}},
		})
	}

	for _, m := range ar.Messages {
		irMsg, err := convertMsg(m)
		if err != nil {
			return nil, err
		}
		req.Messages = append(req.Messages, irMsg)
	}

	for _, t := range ar.Tools {
		req.Tools = append(req.Tools, ir.Tool{
			Type: "function",
			Function: ir.FunctionTool{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.InputSchema,
			},
		})
	}
	if ar.ToolChoice != nil {
		req.ToolChoice = ar.ToolChoice
	}

	return req, nil
}

func convertMsg(m anthropicMsg) (ir.Message, error) {
	irMsg := ir.Message{Role: m.Role}

	var strContent string
	if err := json.Unmarshal(m.Content, &strContent); err == nil {
		irMsg.Content = []ir.ContentPart{{Type: "text", Text: strContent}}
		return irMsg, nil
	}

	var blocks []anthropicBlock
	if err := json.Unmarshal(m.Content, &blocks); err != nil {
		return irMsg, fmt.Errorf("anthropic decode: parse content: %w", err)
	}

	for _, b := range blocks {
		switch b.Type {
		case "text":
			irMsg.Content = append(irMsg.Content, ir.ContentPart{Type: "text", Text: b.Text})
		case "image":
			part := ir.ContentPart{Type: "image_url"}
			if b.Source != nil {
				if b.Source.Type == "url" {
					part.ImageURL = ir.ImageURL{URL: b.Source.URL}
				} else {
					part.ImageURL = ir.ImageURL{
						URL: "data:" + b.Source.MediaType + ";base64," + b.Source.Data,
					}
				}
			}
			irMsg.Content = append(irMsg.Content, part)
		case "tool_use":
			inputJSON, _ := json.Marshal(b.Input)
			irMsg.ToolCalls = append(irMsg.ToolCalls, ir.ToolCall{
				ID:   b.ID,
				Type: "function",
				Function: ir.ToolCallFunction{
					Name:      b.Name,
					Arguments: string(inputJSON),
				},
			})
		case "tool_result":
			var resultText string
			if err := json.Unmarshal(b.Content, &resultText); err == nil {
				irMsg.Content = append(irMsg.Content, ir.ContentPart{Type: "text", Text: resultText})
			}
			irMsg.ToolCallID = b.ToolUseID
			irMsg.Role = "tool"
		}
	}
	return irMsg, nil
}
