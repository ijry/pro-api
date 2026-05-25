// Package openai 实现 OpenAI 原生 adapter。
package openai

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

const defaultBaseURL = "https://api.openai.com"

// OpenAI 实现 adapter.Adapter，调用 OpenAI 原生 REST API。
type OpenAI struct {
	name    string
	baseURL string
	client  *http.Client
}

// New 构造 OpenAI adapter。
// baseURL 为空时使用默认 https://api.openai.com。
func New(baseURL string) *OpenAI {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	return &OpenAI{
		name:    "openai",
		baseURL: baseURL,
		client: adapter.NewHTTPClient(adapter.ClientConfig{
			Provider:            "openai",
			Timeout:             0, // 流式不设端到端超时
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 32,
			IdleConnTimeout:     90 * time.Second,
		}),
	}
}

func (a *OpenAI) Name() string { return a.name }

func (a *OpenAI) Capabilities() adapter.Capability {
	return adapter.CapChat | adapter.CapStream | adapter.CapCompletion |
		adapter.CapEmbedding | adapter.CapVision | adapter.CapToolUse |
		adapter.CapImage | adapter.CapTTS | adapter.CapSTT
}

func (a *OpenAI) SupportedModels() []string {
	return []string{
		"gpt-4o",
		"gpt-4o-mini",
		"gpt-4-turbo",
		"gpt-4",
		"gpt-3.5-turbo",
		"o1",
		"o1-mini",
		"o3",
		"o3-mini",
		"text-embedding-3-large",
		"text-embedding-3-small",
		"text-embedding-ada-002",
	}
}

// Chat 发送非流式 chat completion 请求。
func (a *OpenAI) Chat(ctx context.Context, req *ir.ChatRequest, cred adapter.Credential) (*ir.ChatResponse, error) {
	url := a.baseURL + "/v1/chat/completions"
	if cred.BaseURL != "" {
		url = cred.BaseURL + "/v1/chat/completions"
	}
	return ChatWithClient(ctx, a.client, req, cred, url, "bearer")
}

// ChatStream 发送流式 chat completion 请求，返回 StreamReader。
func (a *OpenAI) ChatStream(ctx context.Context, req *ir.ChatRequest, cred adapter.Credential) (adapter.StreamReader, error) {
	url := a.baseURL + "/v1/chat/completions"
	if cred.BaseURL != "" {
		url = cred.BaseURL + "/v1/chat/completions"
	}
	return ChatStreamWithClient(ctx, a.client, req, cred, url, "bearer")
}

// Embed 发送 embedding 请求。
func (a *OpenAI) Embed(ctx context.Context, req *ir.EmbedRequest, cred adapter.Credential) (*ir.EmbedResponse, error) {
	url := a.baseURL + "/v1/embeddings"
	if cred.BaseURL != "" {
		url = cred.BaseURL + "/v1/embeddings"
	}
	return EmbedWithClient(ctx, a.client, req, cred, url, "bearer")
}

// ChatWithClient 是可复用的 chat 实现，供其他 adapter（如 Azure）使用。
// authScheme: "bearer" → Authorization: Bearer <key>; "api-key" → api-key: <key>
func ChatWithClient(ctx context.Context, client *http.Client, req *ir.ChatRequest, cred adapter.Credential, url, authScheme string) (*ir.ChatResponse, error) {
	body, err := BuildChatRequest(req, false)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, adapter.ClassifyNetErr(err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	setAuthHeader(httpReq, cred.APIKey, authScheme)

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, adapter.ClassifyNetErr(err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, adapter.ClassifyNetErr(err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, adapter.ClassifyHTTPStatus(resp.StatusCode, respBody)
	}
	return DecodeChatResponse(respBody)
}

// ChatStreamWithClient 是可复用的流式 chat 实现。
func ChatStreamWithClient(ctx context.Context, client *http.Client, req *ir.ChatRequest, cred adapter.Credential, url, authScheme string) (adapter.StreamReader, error) {
	body, err := BuildChatRequest(req, true)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, adapter.ClassifyNetErr(err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	setAuthHeader(httpReq, cred.APIKey, authScheme)

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, adapter.ClassifyNetErr(err)
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, adapter.ClassifyHTTPStatus(resp.StatusCode, b)
	}
	return newSSEStreamReader(resp.Body), nil
}

// EmbedWithClient 是可复用的 embed 实现。
func EmbedWithClient(ctx context.Context, client *http.Client, req *ir.EmbedRequest, cred adapter.Credential, url, authScheme string) (*ir.EmbedResponse, error) {
	body, err := BuildEmbedRequest(req)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, adapter.ClassifyNetErr(err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	setAuthHeader(httpReq, cred.APIKey, authScheme)

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, adapter.ClassifyNetErr(err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, adapter.ClassifyNetErr(err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, adapter.ClassifyHTTPStatus(resp.StatusCode, respBody)
	}
	return DecodeEmbedResponse(respBody)
}

func setAuthHeader(r *http.Request, key, scheme string) {
	if scheme == "api-key" {
		r.Header.Set("api-key", key)
	} else {
		r.Header.Set("Authorization", "Bearer "+key)
	}
}

// BuildChatRequest 构造 OpenAI chat completion 请求体（导出，供其他 adapter 复用）。
func BuildChatRequest(req *ir.ChatRequest, stream bool) ([]byte, error) {
	type msg struct {
		Role       string `json:"role"`
		Content    any    `json:"content"`
		Name       string `json:"name,omitempty"`
		ToolCallID string `json:"tool_call_id,omitempty"`
		ToolCalls  any    `json:"tool_calls,omitempty"`
	}
	type toolFunc struct {
		Name        string         `json:"name"`
		Description string         `json:"description,omitempty"`
		Parameters  map[string]any `json:"parameters,omitempty"`
	}
	type tool struct {
		Type     string   `json:"type"`
		Function toolFunc `json:"function"`
	}
	type tcFunc struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	}
	type toolCall struct {
		ID       string `json:"id"`
		Type     string `json:"type"`
		Function tcFunc `json:"function"`
	}
	type respFmt struct {
		Type string `json:"type"`
	}

	msgs := make([]msg, 0, len(req.Messages))
	for _, m := range req.Messages {
		var content any
		if len(m.Content) == 1 && m.Content[0].Type == ir.ContentText {
			content = m.Content[0].Text
		} else if len(m.Content) > 0 {
			content = m.Content
		} else {
			content = ""
		}
		var tcs any
		if len(m.ToolCalls) > 0 {
			calls := make([]toolCall, len(m.ToolCalls))
			for i, tc := range m.ToolCalls {
				calls[i] = toolCall{ID: tc.ID, Type: tc.Type, Function: tcFunc{Name: tc.Function.Name, Arguments: tc.Function.Arguments}}
			}
			tcs = calls
		}
		msgs = append(msgs, msg{Role: m.Role, Content: content, Name: m.Name, ToolCallID: m.ToolCallID, ToolCalls: tcs})
	}

	bodyMap := map[string]any{
		"model":    req.Model,
		"messages": msgs,
		"stream":   stream,
	}
	if req.MaxTokens > 0 {
		bodyMap["max_tokens"] = req.MaxTokens
	}
	if req.Temperature != nil {
		bodyMap["temperature"] = *req.Temperature
	}
	if req.TopP != nil {
		bodyMap["top_p"] = *req.TopP
	}
	if len(req.Stop) > 0 {
		bodyMap["stop"] = req.Stop
	}
	if req.Seed != nil {
		bodyMap["seed"] = *req.Seed
	}
	if req.User != "" {
		bodyMap["user"] = req.User
	}
	if len(req.Tools) > 0 {
		tools := make([]tool, len(req.Tools))
		for i, t := range req.Tools {
			tools[i] = tool{Type: t.Type, Function: toolFunc{Name: t.Function.Name, Description: t.Function.Description, Parameters: t.Function.Parameters}}
		}
		bodyMap["tools"] = tools
	}
	if req.ToolChoice != nil {
		bodyMap["tool_choice"] = req.ToolChoice
	}
	if req.ResponseFormat != nil {
		bodyMap["response_format"] = respFmt{Type: req.ResponseFormat.Type}
	}
	if stream {
		bodyMap["stream_options"] = map[string]any{"include_usage": true}
	}
	return json.Marshal(bodyMap)
}

// BuildEmbedRequest 构造 embedding 请求体（导出）。
func BuildEmbedRequest(req *ir.EmbedRequest) ([]byte, error) {
	body := map[string]any{
		"model": req.Model,
		"input": req.Input,
	}
	if req.EncodingFormat != "" {
		body["encoding_format"] = req.EncodingFormat
	}
	if req.Dimensions != nil {
		body["dimensions"] = *req.Dimensions
	}
	if req.User != "" {
		body["user"] = req.User
	}
	return json.Marshal(body)
}

// DecodeChatResponse 解码非流式响应（导出）。
func DecodeChatResponse(data []byte) (*ir.ChatResponse, error) {
	var raw struct {
		ID                string `json:"id"`
		Model             string `json:"model"`
		SystemFingerprint string `json:"system_fingerprint"`
		Choices           []struct {
			Index   int `json:"index"`
			Message struct {
				Role      string `json:"role"`
				Content   string `json:"content"`
				ToolCalls []struct {
					ID       string `json:"id"`
					Type     string `json:"type"`
					Function struct {
						Name      string `json:"name"`
						Arguments string `json:"arguments"`
					} `json:"function"`
				} `json:"tool_calls"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage struct {
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
		return nil, fmt.Errorf("openai: decode response: %w", err)
	}
	resp := &ir.ChatResponse{
		ID:                raw.ID,
		Model:             raw.Model,
		SystemFingerprint: raw.SystemFingerprint,
		Usage: ir.Usage{
			PromptTokens:     raw.Usage.PromptTokens,
			CompletionTokens: raw.Usage.CompletionTokens,
			TotalTokens:      raw.Usage.TotalTokens,
		},
	}
	if raw.Usage.PromptTokensDetails != nil {
		resp.Usage.CachedTokens = raw.Usage.PromptTokensDetails.CachedTokens
	}
	if raw.Usage.CompletionTokensDetails != nil {
		resp.Usage.ReasoningTokens = raw.Usage.CompletionTokensDetails.ReasoningTokens
	}
	for _, c := range raw.Choices {
		msg := ir.Message{
			Role:    c.Message.Role,
			Content: []ir.ContentPart{{Type: ir.ContentText, Text: c.Message.Content}},
		}
		for _, tc := range c.Message.ToolCalls {
			msg.ToolCalls = append(msg.ToolCalls, ir.ToolCall{
				ID:   tc.ID,
				Type: tc.Type,
				Function: ir.ToolCallFunction{Name: tc.Function.Name, Arguments: tc.Function.Arguments},
			})
		}
		resp.Choices = append(resp.Choices, ir.Choice{
			Index:        c.Index,
			Message:      msg,
			FinishReason: c.FinishReason,
		})
	}
	return resp, nil
}

// DecodeEmbedResponse 解码 embedding 响应（导出）。
func DecodeEmbedResponse(data []byte) (*ir.EmbedResponse, error) {
	var raw struct {
		Model string `json:"model"`
		Data  []struct {
			Index     int       `json:"index"`
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
		Usage struct {
			PromptTokens int `json:"prompt_tokens"`
			TotalTokens  int `json:"total_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("openai: decode embed response: %w", err)
	}
	resp := &ir.EmbedResponse{
		Model: raw.Model,
		Usage: ir.EmbedUsage{
			PromptTokens: raw.Usage.PromptTokens,
			TotalTokens:  raw.Usage.TotalTokens,
		},
	}
	for _, d := range raw.Data {
		resp.Data = append(resp.Data, ir.EmbedData{
			Index:     d.Index,
			Embedding: d.Embedding,
		})
	}
	return resp, nil
}
