// Package grokweb provides Grok Web reverse-adapter support.
package grokweb

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ijry/pro-api/internal/adapter"
	"github.com/ijry/pro-api/internal/protocol/ir"
)

const defaultBaseURL = "https://grok.com"

// Adapter implements the Grok Web line-delimited JSON chat API.
type Adapter struct {
	baseURL string
	client  *http.Client
}

// New returns the default Grok Web adapter.
func New() adapter.Adapter {
	return NewWithBase(defaultBaseURL)
}

// NewWithBase returns an adapter using baseURL before appending Grok Web paths.
func NewWithBase(baseURL string) adapter.Adapter {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	return &Adapter{
		baseURL: strings.TrimRight(baseURL, "/"),
		client: adapter.NewHTTPClient(adapter.ClientConfig{
			Provider:            "grok-web",
			Timeout:             0,
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 32,
			IdleConnTimeout:     90 * time.Second,
		}),
	}
}

func (a *Adapter) Name() string { return "grok-web" }

func (a *Adapter) Capabilities() adapter.Capability {
	return adapter.CapChat | adapter.CapStream
}

func (a *Adapter) SupportedModels() []string { return supportedModels }

func (a *Adapter) Chat(ctx context.Context, req *ir.ChatRequest, cred adapter.Credential) (*ir.ChatResponse, error) {
	resp, err := a.do(ctx, req, cred)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return collectResponse(resp.Body, req.Model)
}

func (a *Adapter) ChatStream(ctx context.Context, req *ir.ChatRequest, cred adapter.Credential) (adapter.StreamReader, error) {
	resp, err := a.do(ctx, req, cred)
	if err != nil {
		return nil, err
	}
	return &streamReader{
		scanner: bufio.NewScanner(resp.Body),
		body:    resp.Body,
		model:   req.Model,
	}, nil
}

func (a *Adapter) Embed(context.Context, *ir.EmbedRequest, adapter.Credential) (*ir.EmbedResponse, error) {
	return nil, fmt.Errorf("grok-web: embedding not supported")
}

func (a *Adapter) do(ctx context.Context, req *ir.ChatRequest, cred adapter.Credential) (*http.Response, error) {
	body, err := buildPayload(req)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	baseURL := a.baseURL
	if cred.BaseURL != "" {
		baseURL = strings.TrimRight(cred.BaseURL, "/")
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/rest/app-chat/conversations/new", bytes.NewReader(data))
	if err != nil {
		return nil, adapter.ClassifyNetErr(err)
	}
	for k, vals := range buildHeaders(cred) {
		for _, v := range vals {
			httpReq.Header.Add(k, v)
		}
	}

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, adapter.ClassifyNetErr(err)
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, adapter.ClassifyHTTPStatus(resp.StatusCode, b)
	}
	return resp, nil
}

func collectResponse(body io.Reader, model string) (*ir.ChatResponse, error) {
	scanner := bufio.NewScanner(body)
	var id string
	var fingerprint string
	var final string
	var tokens strings.Builder
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		ev, ok := parseLine(line)
		if !ok {
			continue
		}
		if ev.HasModelResponse {
			if ev.ResponseID != "" {
				id = ev.ResponseID
			}
			if ev.Message != "" {
				final = ev.Message
			}
			if ev.ModelHash != "" {
				fingerprint = ev.ModelHash
			}
			continue
		}
		if ev.ResponseID != "" && id == "" {
			id = ev.ResponseID
		}
		if ev.HasToken {
			tokens.WriteString(ev.Token)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, adapter.ClassifyNetErr(err)
	}
	if final == "" {
		final = tokens.String()
	}
	if id == "" {
		id = "grok-web"
	}
	return &ir.ChatResponse{
		ID:                id,
		Model:             model,
		SystemFingerprint: fingerprint,
		Choices: []ir.Choice{{
			Index: 0,
			Message: ir.Message{
				Role:    ir.RoleAssistant,
				Content: []ir.ContentPart{{Type: ir.ContentText, Text: final}},
			},
			FinishReason: "stop",
		}},
		Usage: ir.Usage{},
	}, nil
}

type lineEvent struct {
	Token            string
	HasToken         bool
	ResponseID       string
	Message          string
	ModelHash        string
	HasModelResponse bool
}

func parseLine(line string) (lineEvent, bool) {
	var raw struct {
		Result struct {
			Response struct {
				Token         *string `json:"token"`
				ResponseID    string  `json:"responseId"`
				ModelResponse *struct {
					ResponseID string `json:"responseId"`
					Message    string `json:"message"`
					Metadata   struct {
						LLMInfo struct {
							ModelHash string `json:"modelHash"`
						} `json:"llm_info"`
					} `json:"metadata"`
				} `json:"modelResponse"`
			} `json:"response"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(line), &raw); err != nil {
		return lineEvent{}, false
	}
	resp := raw.Result.Response
	ev := lineEvent{ResponseID: resp.ResponseID}
	if resp.Token != nil {
		ev.Token = *resp.Token
		ev.HasToken = true
	}
	if resp.ModelResponse != nil {
		ev.HasModelResponse = true
		ev.ResponseID = resp.ModelResponse.ResponseID
		ev.Message = resp.ModelResponse.Message
		ev.ModelHash = resp.ModelResponse.Metadata.LLMInfo.ModelHash
	}
	return ev, true
}

type streamReader struct {
	scanner *bufio.Scanner
	body    io.ReadCloser
	model   string
	id      string
	done    bool
	closed  bool
}

func (r *streamReader) Next(ctx context.Context) (*ir.ChatChunk, error) {
	if r.done {
		return nil, io.EOF
	}
	for {
		select {
		case <-ctx.Done():
			return nil, adapter.ClassifyNetErr(ctx.Err())
		default:
		}
		if !r.scanner.Scan() {
			if err := r.scanner.Err(); err != nil {
				return nil, adapter.ClassifyNetErr(err)
			}
			r.done = true
			return nil, io.EOF
		}
		line := strings.TrimSpace(r.scanner.Text())
		if line == "" {
			continue
		}
		ev, ok := parseLine(line)
		if !ok || !ev.HasToken || ev.Token == "" {
			continue
		}
		if ev.ResponseID != "" {
			r.id = ev.ResponseID
		}
		if r.id == "" {
			r.id = "grok-web"
		}
		return &ir.ChatChunk{
			ID:    r.id,
			Model: r.model,
			Delta: ir.Delta{
				Role:    ir.RoleAssistant,
				Content: ev.Token,
			},
		}, nil
	}
}

func (r *streamReader) Close() error {
	if r.closed {
		return nil
	}
	r.closed = true
	if r.body == nil {
		return nil
	}
	err := r.body.Close()
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	return nil
}
