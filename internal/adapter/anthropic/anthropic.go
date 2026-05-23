// Package anthropic 实现 Anthropic Claude Messages API adapter。
package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ijry/pro-api/internal/adapter"
	"github.com/ijry/pro-api/internal/protocol/ir"
)

const (
	defaultBaseURL      = "https://api.anthropic.com"
	anthropicVersion    = "2023-06-01"
	defaultMaxTokens    = 4096
)

// Anthropic 实现 adapter.Adapter，调用 Anthropic Messages API。
type Anthropic struct {
	client *http.Client
}

// New 构造 Anthropic adapter。
func New() *Anthropic {
	return &Anthropic{
		client: adapter.NewHTTPClient(adapter.ClientConfig{
			Provider:            "anthropic",
			Timeout:             0,
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 32,
			IdleConnTimeout:     90 * time.Second,
		}),
	}
}

func (a *Anthropic) Name() string { return "anthropic" }

func (a *Anthropic) Capabilities() adapter.Capability {
	return adapter.CapChat | adapter.CapStream | adapter.CapVision | adapter.CapToolUse
}

func (a *Anthropic) SupportedModels() []string {
	return []string{
		"claude-3-5-sonnet-20241022",
		"claude-3-5-sonnet-latest",
		"claude-3-5-haiku-20241022",
		"claude-3-opus-20240229",
		"claude-3-sonnet-20240229",
		"claude-3-haiku-20240307",
	}
}

// --- IR → Anthropic translation ---

type anthropicMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"` // string or []contentBlock
}

type contentBlock struct {
	Type      string     `json:"type"`
	Text      string     `json:"text,omitempty"`
	Source    *imgSource `json:"source,omitempty"`
	ID        string     `json:"id,omitempty"`
	Name      string     `json:"name,omitempty"`
	Input     any        `json:"input,omitempty"`
	ToolUseID string     `json:"tool_use_id,omitempty"`
	Content   string     `json:"content,omitempty"`
}

type imgSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

type anthropicTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	InputSchema map[string]any `json:"input_schema"`
}

func buildRequest(req *ir.ChatRequest, stream bool) ([]byte, error) {
	var systemPrompt string
	var msgs []anthropicMessage

	for _, m := range req.Messages {
		if m.Role == ir.RoleSystem {
			// Extract system prompt
			for _, part := range m.Content {
				if part.Type == ir.ContentText {
					if systemPrompt != "" {
						systemPrompt += "\n"
					}
					systemPrompt += part.Text
				}
			}
			continue
		}

		role := m.Role
		if role == ir.RoleTool {
			role = "user"
		}

		var content any
		if len(m.ToolCalls) > 0 {
			// assistant with tool_use
			blocks := make([]contentBlock, 0)
			for _, part := range m.Content {
				if part.Type == ir.ContentText && part.Text != "" {
					blocks = append(blocks, contentBlock{Type: "text", Text: part.Text})
				}
			}
			for _, tc := range m.ToolCalls {
				var input map[string]any
				if tc.Function.Arguments != "" {
					_ = json.Unmarshal([]byte(tc.Function.Arguments), &input)
				}
				blocks = append(blocks, contentBlock{Type: "tool_use", ID: tc.ID, Name: tc.Function.Name, Input: input})
			}
			content = blocks
		} else if m.Role == ir.RoleTool {
			// tool result
			var resultContent string
			for _, part := range m.Content {
				if part.Type == ir.ContentText {
					resultContent = part.Text
				}
			}
			content = []contentBlock{{Type: "tool_result", ToolUseID: m.ToolCallID, Content: resultContent}}
		} else if len(m.Content) == 1 && m.Content[0].Type == ir.ContentText {
			content = m.Content[0].Text
		} else {
			blocks := make([]contentBlock, 0, len(m.Content))
			for _, part := range m.Content {
				switch part.Type {
				case ir.ContentText:
					blocks = append(blocks, contentBlock{Type: "text", Text: part.Text})
				case ir.ContentImageURL:
					blocks = append(blocks, contentBlock{Type: "text", Text: "[image: " + part.ImageURL.URL + "]"})
				}
			}
			content = blocks
		}
		msgs = append(msgs, anthropicMessage{Role: role, Content: content})
	}

	maxTokens := req.MaxTokens
	if maxTokens <= 0 {
		maxTokens = defaultMaxTokens
	}

	body := map[string]any{
		"model":      req.Model,
		"messages":   msgs,
		"max_tokens": maxTokens,
		"stream":     stream,
	}
	if systemPrompt != "" {
		body["system"] = systemPrompt
	}
	if req.Temperature != nil {
		body["temperature"] = *req.Temperature
	}
	if req.TopP != nil {
		body["top_p"] = *req.TopP
	}
	if len(req.Stop) > 0 {
		body["stop_sequences"] = req.Stop
	}
	if len(req.Tools) > 0 {
		tools := make([]anthropicTool, len(req.Tools))
		for i, t := range req.Tools {
			schema := t.Function.Parameters
			if schema == nil {
				schema = map[string]any{"type": "object", "properties": map[string]any{}}
			}
			tools[i] = anthropicTool{Name: t.Function.Name, Description: t.Function.Description, InputSchema: schema}
		}
		body["tools"] = tools
	}

	return json.Marshal(body)
}

// --- Anthropic → IR translation ---

func decodeResponse(data []byte) (*ir.ChatResponse, error) {
	var raw struct {
		ID      string `json:"id"`
		Model   string `json:"model"`
		Content []struct {
			Type  string `json:"type"`
			Text  string `json:"text"`
			ID    string `json:"id"`
			Name  string `json:"name"`
			Input any    `json:"input"`
		} `json:"content"`
		StopReason string `json:"stop_reason"`
		Usage      struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("anthropic: decode response: %w", err)
	}

	msg := ir.Message{Role: ir.RoleAssistant}
	for _, block := range raw.Content {
		switch block.Type {
		case "text":
			msg.Content = append(msg.Content, ir.ContentPart{Type: ir.ContentText, Text: block.Text})
		case "tool_use":
			argsJSON, _ := json.Marshal(block.Input)
			msg.ToolCalls = append(msg.ToolCalls, ir.ToolCall{
				ID:   block.ID,
				Type: "function",
				Function: ir.ToolCallFunction{Name: block.Name, Arguments: string(argsJSON)},
			})
		}
	}

	finishReason := mapStopReason(raw.StopReason)
	return &ir.ChatResponse{
		ID:    raw.ID,
		Model: raw.Model,
		Choices: []ir.Choice{{
			Index:        0,
			Message:      msg,
			FinishReason: finishReason,
		}},
		Usage: ir.Usage{
			PromptTokens:     raw.Usage.InputTokens,
			CompletionTokens: raw.Usage.OutputTokens,
			TotalTokens:      raw.Usage.InputTokens + raw.Usage.OutputTokens,
		},
	}, nil
}

func mapStopReason(r string) string {
	switch r {
	case "end_turn":
		return ir.FinishStop
	case "max_tokens":
		return ir.FinishLength
	case "tool_use":
		return ir.FinishToolCalls
	default:
		return r
	}
}

// --- HTTP methods ---

func (a *Anthropic) sendRequest(ctx context.Context, cred adapter.Credential, body []byte, stream bool) (*http.Response, error) {
	base := defaultBaseURL
	if cred.BaseURL != "" {
		base = cred.BaseURL
	}
	url := base + "/v1/messages"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, adapter.ClassifyNetErr(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", cred.APIKey)
	req.Header.Set("anthropic-version", anthropicVersion)
	if stream {
		req.Header.Set("Accept", "text/event-stream")
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, adapter.ClassifyNetErr(err)
	}
	return resp, nil
}

func (a *Anthropic) Chat(ctx context.Context, req *ir.ChatRequest, cred adapter.Credential) (*ir.ChatResponse, error) {
	body, err := buildRequest(req, false)
	if err != nil {
		return nil, err
	}
	resp, err := a.sendRequest(ctx, cred, body, false)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, adapter.ClassifyNetErr(err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, adapter.ClassifyHTTPStatus(resp.StatusCode, respBody)
	}
	return decodeResponse(respBody)
}

func (a *Anthropic) ChatStream(ctx context.Context, req *ir.ChatRequest, cred adapter.Credential) (adapter.StreamReader, error) {
	body, err := buildRequest(req, true)
	if err != nil {
		return nil, err
	}
	resp, err := a.sendRequest(ctx, cred, body, true)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, adapter.ClassifyHTTPStatus(resp.StatusCode, b)
	}
	return newAnthropicStream(resp.Body), nil
}

func (a *Anthropic) Embed(_ context.Context, _ *ir.EmbedRequest, _ adapter.Credential) (*ir.EmbedResponse, error) {
	return nil, fmt.Errorf("anthropic: embedding not supported")
}
