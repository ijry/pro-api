package openai

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/ijry/pro-api/internal/protocol/ir"
	"github.com/ijry/pro-api/pkg/apierr"
)

// DecodeOptions 控制解码行为。
type DecodeOptions struct {
	// AllowUnknownFields=false 时,JSON 中出现未知字段会被拒绝。
	AllowUnknownFields bool
}

// DecodeChat 把 HTTP body → ir.ChatRequest。
//
// 默认严格模式(DisallowUnknownFields);客户端发未知字段会被拒绝。
// 由 setting "openai.allow_unknown_fields" 控制是否放宽。
func DecodeChat(r io.Reader, opt DecodeOptions) (*ir.ChatRequest, error) {
	var dto ChatRequestDTO
	dec := json.NewDecoder(r)
	if !opt.AllowUnknownFields {
		dec.DisallowUnknownFields()
	}
	if err := dec.Decode(&dto); err != nil {
		return nil, apierr.New(apierr.CodeInvalidParam, "invalid JSON: "+err.Error())
	}
	req := &ir.ChatRequest{
		Model:       dto.Model,
		Temperature: dto.Temperature,
		TopP:        dto.TopP,
		Stream:      dto.Stream,
		User:        dto.User,
		Seed:        dto.Seed,
		ExtraParams: map[string]any{},
	}
	// max_completion_tokens 优先(o1/o3 系列);否则用 max_tokens
	if dto.MaxCompletionTokens != nil {
		req.MaxTokens = *dto.MaxCompletionTokens
	} else if dto.MaxTokens != nil {
		req.MaxTokens = *dto.MaxTokens
	}
	// stop: string 或 []string
	if len(dto.Stop) > 0 {
		var s string
		if err := json.Unmarshal(dto.Stop, &s); err == nil {
			req.Stop = []string{s}
		} else {
			var arr []string
			if err := json.Unmarshal(dto.Stop, &arr); err != nil {
				return nil, apierr.New(apierr.CodeInvalidParam, "stop must be string or []string")
			}
			req.Stop = arr
		}
	}
	// messages
	for i, mDTO := range dto.Messages {
		m := ir.Message{
			Role:       mDTO.Role,
			Name:       mDTO.Name,
			ToolCallID: mDTO.ToolCallID,
		}
		if len(mDTO.Content) > 0 {
			var s string
			if err := json.Unmarshal(mDTO.Content, &s); err == nil {
				m.Content = []ir.ContentPart{{Type: ir.ContentText, Text: s}}
			} else {
				var parts []ContentPartDTO
				if err := json.Unmarshal(mDTO.Content, &parts); err != nil {
					return nil, apierr.New(apierr.CodeInvalidParam,
						fmt.Sprintf("messages[%d].content must be string or array", i))
				}
				for _, p := range parts {
					cp := ir.ContentPart{Type: p.Type, Text: p.Text}
					if p.ImageURL != nil {
						cp.ImageURL = ir.ImageURL{URL: p.ImageURL.URL, Detail: p.ImageURL.Detail}
					}
					m.Content = append(m.Content, cp)
				}
			}
		}
		for _, tc := range mDTO.ToolCalls {
			m.ToolCalls = append(m.ToolCalls, ir.ToolCall{
				ID:   tc.ID,
				Type: tc.Type,
				Function: ir.ToolCallFunction{
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				},
			})
		}
		req.Messages = append(req.Messages, m)
	}
	// tools
	for _, t := range dto.Tools {
		req.Tools = append(req.Tools, ir.Tool{
			Type: t.Type,
			Function: ir.FunctionTool{
				Name:        t.Function.Name,
				Description: t.Function.Description,
				Parameters:  t.Function.Parameters,
			},
		})
	}
	// tool_choice: string | object
	if len(dto.ToolChoice) > 0 {
		var any any
		if err := json.Unmarshal(dto.ToolChoice, &any); err == nil {
			req.ToolChoice = any
		}
	}
	if dto.ResponseFormat != nil {
		req.ResponseFormat = &ir.ResponseFormat{
			Type:   dto.ResponseFormat.Type,
			Schema: dto.ResponseFormat.JSONSchema,
		}
	}
	return req, nil
}

// DecodeEmbed 把 HTTP body → ir.EmbedRequest。
//
// input 接受 string / []string。
func DecodeEmbed(r io.Reader, opt DecodeOptions) (*ir.EmbedRequest, error) {
	var dto EmbedRequestDTO
	dec := json.NewDecoder(r)
	if !opt.AllowUnknownFields {
		dec.DisallowUnknownFields()
	}
	if err := dec.Decode(&dto); err != nil {
		return nil, apierr.New(apierr.CodeInvalidParam, "invalid JSON: "+err.Error())
	}
	req := &ir.EmbedRequest{
		Model:          dto.Model,
		EncodingFormat: dto.EncodingFormat,
		User:           dto.User,
		Dimensions:     dto.Dimensions,
		ExtraParams:    map[string]any{},
	}
	if len(dto.Input) > 0 {
		var s string
		if err := json.Unmarshal(dto.Input, &s); err == nil {
			req.Input = []string{s}
		} else {
			var arr []string
			if err := json.Unmarshal(dto.Input, &arr); err != nil {
				return nil, apierr.New(apierr.CodeInvalidParam, "input must be string or []string")
			}
			req.Input = arr
		}
	}
	return req, nil
}

// DecodeCompletion 把 HTTP body → ir.CompletionRequest。
//
// prompt 接受 string / []string。
func DecodeCompletion(r io.Reader, opt DecodeOptions) (*ir.CompletionRequest, error) {
	var dto CompletionRequestDTO
	dec := json.NewDecoder(r)
	if !opt.AllowUnknownFields {
		dec.DisallowUnknownFields()
	}
	if err := dec.Decode(&dto); err != nil {
		return nil, apierr.New(apierr.CodeInvalidParam, "invalid JSON: "+err.Error())
	}
	req := &ir.CompletionRequest{
		Model:       dto.Model,
		Temperature: dto.Temperature,
		TopP:        dto.TopP,
		Stream:      dto.Stream,
		User:        dto.User,
		Suffix:      dto.Suffix,
		Echo:        dto.Echo,
		Logprobs:    dto.Logprobs,
		ExtraParams: map[string]any{},
	}
	if dto.MaxTokens != nil {
		req.MaxTokens = *dto.MaxTokens
	}
	if len(dto.Prompt) > 0 {
		var s string
		if err := json.Unmarshal(dto.Prompt, &s); err == nil {
			req.Prompt = []string{s}
		} else {
			var arr []string
			if err := json.Unmarshal(dto.Prompt, &arr); err != nil {
				return nil, apierr.New(apierr.CodeInvalidParam, "prompt must be string or []string")
			}
			req.Prompt = arr
		}
	}
	if len(dto.Stop) > 0 {
		var s string
		if err := json.Unmarshal(dto.Stop, &s); err == nil {
			req.Stop = []string{s}
		} else {
			var arr []string
			if err := json.Unmarshal(dto.Stop, &arr); err == nil {
				req.Stop = arr
			}
		}
	}
	return req, nil
}
