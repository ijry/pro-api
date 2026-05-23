// Package gemini 实现 Google Gemini generateContent API adapter。
package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ijry/pro-api/internal/adapter"
	"github.com/ijry/pro-api/internal/protocol/ir"
)

const defaultBaseURL = "https://generativelanguage.googleapis.com"

// Gemini 实现 adapter.Adapter，调用 Google Gemini API。
type Gemini struct {
	client *http.Client
}

// New 构造 Gemini adapter。
func New() *Gemini {
	return &Gemini{
		client: adapter.NewHTTPClient(adapter.ClientConfig{
			Provider:            "gemini",
			Timeout:             0,
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 32,
			IdleConnTimeout:     90 * time.Second,
		}),
	}
}

func (a *Gemini) Name() string { return "gemini" }

func (a *Gemini) Capabilities() adapter.Capability {
	return adapter.CapChat | adapter.CapStream | adapter.CapVision | adapter.CapToolUse | adapter.CapEmbedding
}

func (a *Gemini) SupportedModels() []string {
	return []string{
		"gemini-2.0-flash-exp",
		"gemini-1.5-pro",
		"gemini-1.5-flash",
		"gemini-1.0-pro",
		"text-embedding-004",
	}
}

// --- IR → Gemini translation ---

type geminiContent struct {
	Role  string       `json:"role"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text         string          `json:"text,omitempty"`
	FunctionCall *geminiFuncCall `json:"functionCall,omitempty"`
	FunctionResp *geminiFuncResp `json:"functionResponse,omitempty"`
}

type geminiFuncCall struct {
	Name string `json:"name"`
	Args any    `json:"args,omitempty"`
}

type geminiFuncResp struct {
	Name     string `json:"name"`
	Response any    `json:"response"`
}

type geminiTool struct {
	FunctionDeclarations []geminiFuncDecl `json:"functionDeclarations"`
}

type geminiFuncDecl struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters,omitempty"`
}

func buildRequest(req *ir.ChatRequest, stream bool) ([]byte, string, error) {
	var systemInstruction *geminiContent
	var contents []geminiContent

	for _, m := range req.Messages {
		if m.Role == ir.RoleSystem {
			var parts []geminiPart
			for _, p := range m.Content {
				if p.Type == ir.ContentText {
					parts = append(parts, geminiPart{Text: p.Text})
				}
			}
			systemInstruction = &geminiContent{Parts: parts}
			continue
		}

		role := "user"
		if m.Role == ir.RoleAssistant {
			role = "model"
		}

		var parts []geminiPart
		if len(m.ToolCalls) > 0 {
			for _, tc := range m.ToolCalls {
				var args map[string]any
				if tc.Function.Arguments != "" {
					_ = json.Unmarshal([]byte(tc.Function.Arguments), &args)
				}
				parts = append(parts, geminiPart{FunctionCall: &geminiFuncCall{Name: tc.Function.Name, Args: args}})
			}
		} else if m.Role == ir.RoleTool {
			var result any
			for _, p := range m.Content {
				if p.Type == ir.ContentText {
					result = map[string]any{"output": p.Text}
				}
			}
			parts = append(parts, geminiPart{FunctionResp: &geminiFuncResp{Name: m.Name, Response: result}})
			role = "user"
		} else {
			for _, p := range m.Content {
				if p.Type == ir.ContentText {
					parts = append(parts, geminiPart{Text: p.Text})
				}
			}
		}
		contents = append(contents, geminiContent{Role: role, Parts: parts})
	}

	body := map[string]any{
		"contents": contents,
	}
	if systemInstruction != nil {
		body["systemInstruction"] = systemInstruction
	}

	genCfg := map[string]any{}
	if req.MaxTokens > 0 {
		genCfg["maxOutputTokens"] = req.MaxTokens
	}
	if req.Temperature != nil {
		genCfg["temperature"] = *req.Temperature
	}
	if req.TopP != nil {
		genCfg["topP"] = *req.TopP
	}
	if len(req.Stop) > 0 {
		genCfg["stopSequences"] = req.Stop
	}
	if len(genCfg) > 0 {
		body["generationConfig"] = genCfg
	}

	if len(req.Tools) > 0 {
		decls := make([]geminiFuncDecl, len(req.Tools))
		for i, t := range req.Tools {
			decls[i] = geminiFuncDecl{Name: t.Function.Name, Description: t.Function.Description, Parameters: t.Function.Parameters}
		}
		body["tools"] = []geminiTool{{FunctionDeclarations: decls}}
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, "", err
	}

	method := "generateContent"
	if stream {
		method = "streamGenerateContent"
	}
	return data, method, nil
}

// --- Gemini → IR translation ---

type geminiResponse struct {
	Candidates []struct {
		Content       geminiContent `json:"content"`
		FinishReason  string        `json:"finishReason"`
		Index         int           `json:"index"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
		TotalTokenCount      int `json:"totalTokenCount"`
	} `json:"usageMetadata"`
}

func decodeResponse(data []byte) (*ir.ChatResponse, error) {
	var raw geminiResponse
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("gemini: decode response: %w", err)
	}

	resp := &ir.ChatResponse{
		Usage: ir.Usage{
			PromptTokens:     raw.UsageMetadata.PromptTokenCount,
			CompletionTokens: raw.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      raw.UsageMetadata.TotalTokenCount,
		},
	}

	for i, cand := range raw.Candidates {
		msg := ir.Message{Role: ir.RoleAssistant}
		for _, p := range cand.Content.Parts {
			if p.Text != "" {
				msg.Content = append(msg.Content, ir.ContentPart{Type: ir.ContentText, Text: p.Text})
			}
			if p.FunctionCall != nil {
				argsJSON, _ := json.Marshal(p.FunctionCall.Args)
				msg.ToolCalls = append(msg.ToolCalls, ir.ToolCall{
					ID:   fmt.Sprintf("call_%d", i),
					Type: "function",
					Function: ir.ToolCallFunction{Name: p.FunctionCall.Name, Arguments: string(argsJSON)},
				})
			}
		}
		resp.Choices = append(resp.Choices, ir.Choice{
			Index:        cand.Index,
			Message:      msg,
			FinishReason: mapFinishReason(cand.FinishReason),
		})
	}
	return resp, nil
}

func mapFinishReason(r string) string {
	switch strings.ToUpper(r) {
	case "STOP":
		return ir.FinishStop
	case "MAX_TOKENS":
		return ir.FinishLength
	default:
		return strings.ToLower(r)
	}
}

// --- HTTP methods ---

func buildURL(base, model, method, apiKey string) string {
	if base == "" {
		base = defaultBaseURL
	}
	return fmt.Sprintf("%s/v1beta/models/%s:%s?key=%s", base, model, method, apiKey)
}

func (a *Gemini) Chat(ctx context.Context, req *ir.ChatRequest, cred adapter.Credential) (*ir.ChatResponse, error) {
	body, method, err := buildRequest(req, false)
	if err != nil {
		return nil, err
	}
	url := buildURL(cred.BaseURL, req.Model, method, cred.APIKey)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, adapter.ClassifyNetErr(err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(httpReq)
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
	return decodeResponse(respBody)
}

func (a *Gemini) ChatStream(ctx context.Context, req *ir.ChatRequest, cred adapter.Credential) (adapter.StreamReader, error) {
	body, method, err := buildRequest(req, true)
	if err != nil {
		return nil, err
	}
	// Gemini streaming uses alt=sse parameter
	url := buildURL(cred.BaseURL, req.Model, method, cred.APIKey) + "&alt=sse"

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, adapter.ClassifyNetErr(err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, adapter.ClassifyNetErr(err)
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, adapter.ClassifyHTTPStatus(resp.StatusCode, b)
	}
	return newGeminiStream(resp.Body), nil
}

func (a *Gemini) Embed(ctx context.Context, req *ir.EmbedRequest, cred adapter.Credential) (*ir.EmbedResponse, error) {
	base := cred.BaseURL
	if base == "" {
		base = defaultBaseURL
	}
	url := fmt.Sprintf("%s/v1beta/models/%s:embedContent?key=%s", base, req.Model, cred.APIKey)

	body, err := json.Marshal(map[string]any{
		"model":   "models/" + req.Model,
		"content": map[string]any{"parts": []map[string]any{{"text": strings.Join(req.Input, " ")}}},
	})
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, adapter.ClassifyNetErr(err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(httpReq)
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

	var raw struct {
		Embedding struct {
			Values []float32 `json:"values"`
		} `json:"embedding"`
	}
	if err := json.Unmarshal(respBody, &raw); err != nil {
		return nil, fmt.Errorf("gemini: decode embed: %w", err)
	}
	return &ir.EmbedResponse{
		Model: req.Model,
		Data:  []ir.EmbedData{{Index: 0, Embedding: raw.Embedding.Values}},
	}, nil
}
