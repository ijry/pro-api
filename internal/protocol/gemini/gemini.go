// Package gemini 处理 Gemini API 的入口协议编解码。
package gemini

import (
	"encoding/json"
	"fmt"

	"github.com/ijry/pro-api/internal/protocol/ir"
)

type geminiRequest struct {
	Contents          []geminiContent   `json:"contents"`
	SystemInstruction *geminiContent    `json:"system_instruction,omitempty"`
	Tools             []geminiTool      `json:"tools,omitempty"`
	GenerationConfig  *generationConfig `json:"generationConfig,omitempty"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text         string            `json:"text,omitempty"`
	InlineData   *geminiInlineData `json:"inline_data,omitempty"`
	FunctionCall *geminiFuncCall   `json:"function_call,omitempty"`
}

type geminiInlineData struct {
	MIMEType string `json:"mime_type"`
	Data     string `json:"data"`
}

type geminiFuncCall struct {
	Name string         `json:"name"`
	Args map[string]any `json:"args"`
}

type geminiTool struct {
	FunctionDeclarations []geminiFuncDecl `json:"function_declarations"`
}

type geminiFuncDecl struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

type generationConfig struct {
	MaxOutputTokens int      `json:"maxOutputTokens,omitempty"`
	Temperature     float64  `json:"temperature,omitempty"`
	TopP            float64  `json:"topP,omitempty"`
	TopK            int      `json:"topK,omitempty"`
	StopSequences   []string `json:"stopSequences,omitempty"`
}

// DecodeRequest 把 Gemini generateContent 请求体解析为 IR。
// model 由 URL 路径参数传入。
func DecodeRequest(body []byte, model string) (*ir.ChatRequest, error) {
	var gr geminiRequest
	if err := json.Unmarshal(body, &gr); err != nil {
		return nil, fmt.Errorf("gemini decode: %w", err)
	}

	req := &ir.ChatRequest{Model: model}

	if gr.GenerationConfig != nil {
		if gr.GenerationConfig.MaxOutputTokens > 0 {
			req.MaxTokens = gr.GenerationConfig.MaxOutputTokens
		}
		if gr.GenerationConfig.Temperature > 0 {
			t := gr.GenerationConfig.Temperature
			req.Temperature = &t
		}
		if gr.GenerationConfig.TopP > 0 {
			p := gr.GenerationConfig.TopP
			req.TopP = &p
		}
		req.Stop = gr.GenerationConfig.StopSequences
	}

	if gr.SystemInstruction != nil && len(gr.SystemInstruction.Parts) > 0 {
		req.Messages = append(req.Messages, ir.Message{
			Role:    "system",
			Content: []ir.ContentPart{{Type: "text", Text: gr.SystemInstruction.Parts[0].Text}},
		})
	}

	for _, c := range gr.Contents {
		role := c.Role
		if role == "model" {
			role = "assistant"
		}
		msg := ir.Message{Role: role}
		for _, p := range c.Parts {
			if p.Text != "" {
				msg.Content = append(msg.Content, ir.ContentPart{Type: "text", Text: p.Text})
			}
			if p.InlineData != nil {
				msg.Content = append(msg.Content, ir.ContentPart{
					Type:     "image_url",
					ImageURL: ir.ImageURL{URL: "data:" + p.InlineData.MIMEType + ";base64," + p.InlineData.Data},
				})
			}
			if p.FunctionCall != nil {
				argsJSON, _ := json.Marshal(p.FunctionCall.Args)
				msg.ToolCalls = append(msg.ToolCalls, ir.ToolCall{
					Type: "function",
					Function: ir.ToolCallFunction{
						Name:      p.FunctionCall.Name,
						Arguments: string(argsJSON),
					},
				})
			}
		}
		req.Messages = append(req.Messages, msg)
	}

	for _, t := range gr.Tools {
		for _, fd := range t.FunctionDeclarations {
			req.Tools = append(req.Tools, ir.Tool{
				Type: "function",
				Function: ir.FunctionTool{
					Name:        fd.Name,
					Description: fd.Description,
					Parameters:  fd.Parameters,
				},
			})
		}
	}

	return req, nil
}

// SSEEvent は Gemini SSE 流事件。
type SSEEvent struct {
	Data []byte
}

// EncodeResponse は IR 响应转为 Gemini generateContent 响应 JSON。
func EncodeResponse(resp *ir.ChatResponse) ([]byte, error) {
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("gemini encode: no choices")
	}
	choice := resp.Choices[0]

	var parts []map[string]any
	for _, p := range choice.Message.Content {
		if p.Type == "text" {
			parts = append(parts, map[string]any{"text": p.Text})
		}
	}
	for _, tc := range choice.Message.ToolCalls {
		var args map[string]any
		_ = json.Unmarshal([]byte(tc.Function.Arguments), &args)
		parts = append(parts, map[string]any{
			"functionCall": map[string]any{"name": tc.Function.Name, "args": args},
		})
	}
	if len(parts) == 0 {
		parts = []map[string]any{{"text": ""}}
	}

	finishReason := mapFinishReason(choice.FinishReason)

	out := map[string]any{
		"candidates": []map[string]any{
			{
				"content":      map[string]any{"role": "model", "parts": parts},
				"finishReason": finishReason,
				"index":        0,
			},
		},
		"usageMetadata": map[string]int{
			"promptTokenCount":     resp.Usage.PromptTokens,
			"candidatesTokenCount": resp.Usage.CompletionTokens,
			"totalTokenCount":      resp.Usage.TotalTokens,
		},
	}
	return json.Marshal(out)
}

// EncodeChunk は IR chunk 转为 Gemini SSE 流事件。
func EncodeChunk(chunk *ir.ChatChunk, isStop bool) ([]SSEEvent, error) {
	var parts []map[string]any
	if chunk.Delta.Content != "" {
		parts = append(parts, map[string]any{"text": chunk.Delta.Content})
	}

	finishReason := ""
	if isStop || chunk.FinishReason != "" {
		finishReason = mapFinishReason(chunk.FinishReason)
		if finishReason == "" {
			finishReason = "STOP"
		}
	}

	candidate := map[string]any{
		"content": map[string]any{"role": "model", "parts": parts},
		"index":   0,
	}
	if finishReason != "" {
		candidate["finishReason"] = finishReason
	}
	out := map[string]any{"candidates": []any{candidate}}
	if isStop && chunk.Usage != nil {
		out["usageMetadata"] = map[string]int{
			"promptTokenCount":     chunk.Usage.PromptTokens,
			"candidatesTokenCount": chunk.Usage.CompletionTokens,
		}
	}
	data, err := json.Marshal(out)
	if err != nil {
		return nil, err
	}
	return []SSEEvent{{Data: data}}, nil
}

func mapFinishReason(r string) string {
	switch r {
	case "stop", "":
		return "STOP"
	case "length":
		return "MAX_TOKENS"
	case "tool_calls":
		return "STOP"
	default:
		return "STOP"
	}
}
