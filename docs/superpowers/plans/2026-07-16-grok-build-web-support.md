# Grok Build / Grok Web Support Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `grok-build` and `grok-web` providers that support chat and streaming through the existing relay/channel/account architecture.

**Architecture:** `grok-build` is a thin OpenAI-compatible adapter for xAI's official API. `grok-web` is a focused Grok Web adapter that converts IR chat requests into Grok Web REST payloads and converts Grok's line-delimited JSON stream back into IR responses/chunks. Both providers register in `adapterreg`; probes register in account wiring; model seeds expose both provider families.

**Tech Stack:** Go, `net/http`, `httptest`, existing `internal/adapter`, `internal/protocol/ir`, `internal/account/probe`, SQL migrations for MySQL/Postgres.

## Global Constraints

- Provider names are exactly `grok-build` and `grok-web`.
- `grok-build` default base URL is `https://api.x.ai`.
- `grok-web` default base URL is `https://grok.com`.
- `grok-web` sends `POST /rest/app-chat/conversations/new`.
- `grok-web` credentials use `Cookie: sso=<token>; sso-rw=<token>` built from `adapter.Credential.APIKey` or account credentials resolved into that field.
- Only chat and stream are in scope; images, video, audio, embeddings, and file uploads are out of scope.
- Do not add new global runtime config fields for Grok bases; use provider defaults and existing channel `BaseURL` override.
- Keep existing provider behavior unchanged.

---

## File Structure

- `internal/adapter/grokbuild/adapter.go`: official xAI OpenAI-compatible adapter, capability-limited to chat/stream.
- `internal/adapter/grokbuild/adapter_test.go`: verifies provider metadata, restricted capabilities, request path, and unsupported embeddings.
- `internal/adapter/grokweb/model.go`: Grok Web model mapping table and lookup function.
- `internal/adapter/grokweb/payload.go`: SSO cookie builder, message flattener, Grok Web payload builder, browser-like header builder.
- `internal/adapter/grokweb/adapter.go`: HTTP request execution, non-stream collection, and stream reader.
- `internal/adapter/grokweb/*_test.go`: focused tests for mapping, payload, response collection, streaming, and error classification.
- `internal/account/probe/grok_build.go`: official API probe using `GET /v1/models`.
- `internal/account/probe/grok_web.go`: Web probe using `GET /rest/rate-limits`.
- `internal/account/probe/grok_build_test.go` and `internal/account/probe/grok_web_test.go`: probe header/path/error tests.
- `internal/account/wire/wire.go`: register probes under `grok-build` and `grok-web`.
- `internal/adapterreg/wire.go`: import and register the two adapters.
- `migrations/mysql/000023_seed_model_catalogs.up.sql` and `.down.sql`: append/remove Grok seed models.
- `migrations/postgres/000023_seed_model_catalogs.up.sql` and `.down.sql`: append/remove Grok seed models.
- `docs-site/zh/architecture/adapter-layer.md`, `docs-site/en/architecture/adapter-layer.md`, `README.md`, `README_zh.md`: document supported Grok providers and credential format.

---

### Task 1: Add `grok-build` Official API Adapter

**Files:**
- Create: `internal/adapter/grokbuild/adapter.go`
- Create: `internal/adapter/grokbuild/adapter_test.go`

**Interfaces:**
- Consumes: `openai.New(baseURL string) *openai.OpenAI`, `adapter.Adapter`, `ir.ChatRequest`, `adapter.Credential`
- Produces: `func New() adapter.Adapter`, `func NewWithBase(baseURL string) adapter.Adapter`

- [ ] **Step 1: Write failing tests for `grok-build`**

Create `internal/adapter/grokbuild/adapter_test.go`:

```go
package grokbuild

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ijry/pro-api/internal/adapter"
	"github.com/ijry/pro-api/internal/protocol/ir"
)

func TestAdapterMetadata(t *testing.T) {
	a := New()
	if a.Name() != "grok-build" {
		t.Fatalf("Name() = %q", a.Name())
	}
	caps := a.Capabilities()
	if !caps.Has(adapter.CapChat) || !caps.Has(adapter.CapStream) {
		t.Fatalf("missing chat/stream caps: %v", caps)
	}
	if caps.Has(adapter.CapEmbedding) || caps.Has(adapter.CapImage) || caps.Has(adapter.CapTTS) || caps.Has(adapter.CapSTT) {
		t.Fatalf("unexpected non-chat caps: %v", caps)
	}
	models := a.SupportedModels()
	want := []string{"grok-4", "grok-3", "grok-3-mini", "grok-3-mini-fast"}
	if len(models) != len(want) {
		t.Fatalf("models len = %d, want %d: %v", len(models), len(want), models)
	}
	for i := range want {
		if models[i] != want[i] {
			t.Fatalf("models[%d] = %q, want %q", i, models[i], want[i])
		}
	}
}

func TestChatUsesXAIOpenAICompatiblePath(t *testing.T) {
	var gotPath string
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if body["model"] != "grok-4" {
			t.Fatalf("model = %#v", body["model"])
		}
		_, _ = w.Write([]byte(`{"id":"chatcmpl-1","model":"grok-4","choices":[{"index":0,"message":{"role":"assistant","content":"pong"},"finish_reason":"stop"}],"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`))
	}))
	defer srv.Close()

	a := NewWithBase(srv.URL)
	resp, err := a.Chat(context.Background(), &ir.ChatRequest{
		Model: "grok-4",
		Messages: []ir.Message{{
			Role:    ir.RoleUser,
			Content: []ir.ContentPart{{Type: ir.ContentText, Text: "ping"}},
		}},
	}, adapter.Credential{APIKey: "xai-key"})
	if err != nil {
		t.Fatalf("Chat error: %v", err)
	}
	if gotPath != "/v1/chat/completions" {
		t.Fatalf("path = %q", gotPath)
	}
	if gotAuth != "Bearer xai-key" {
		t.Fatalf("auth = %q", gotAuth)
	}
	if resp.Choices[0].Message.Content[0].Text != "pong" {
		t.Fatalf("content = %#v", resp.Choices[0].Message.Content)
	}
}

func TestEmbedUnsupported(t *testing.T) {
	_, err := New().Embed(context.Background(), &ir.EmbedRequest{Model: "grok-4"}, adapter.Credential{APIKey: "x"})
	if err == nil {
		t.Fatalf("expected unsupported embedding error")
	}
}
```

- [ ] **Step 2: Run the focused tests and verify they fail**

Run: `go test ./internal/adapter/grokbuild`

Expected: FAIL because package `internal/adapter/grokbuild` does not exist.

- [ ] **Step 3: Implement the `grok-build` adapter**

Create `internal/adapter/grokbuild/adapter.go`:

```go
// Package grokbuild provides xAI Grok Build adapter support via the official OpenAI-compatible API.
package grokbuild

import (
	"context"
	"fmt"

	"github.com/ijry/pro-api/internal/adapter"
	oadapter "github.com/ijry/pro-api/internal/adapter/openai"
	"github.com/ijry/pro-api/internal/protocol/ir"
)

const defaultBaseURL = "https://api.x.ai"

var supportedModels = []string{
	"grok-4",
	"grok-3",
	"grok-3-mini",
	"grok-3-mini-fast",
}

// Adapter wraps the OpenAI-compatible adapter with xAI-specific identity and capability limits.
type Adapter struct {
	base *oadapter.OpenAI
}

// New returns the default xAI official API adapter.
func New() adapter.Adapter {
	return NewWithBase(defaultBaseURL)
}

// NewWithBase returns an adapter that uses baseURL before appending /v1 paths.
func NewWithBase(baseURL string) adapter.Adapter {
	return &Adapter{base: oadapter.New(baseURL)}
}

func (a *Adapter) Name() string {
	return "grok-build"
}

func (a *Adapter) Capabilities() adapter.Capability {
	return adapter.CapChat | adapter.CapStream
}

func (a *Adapter) SupportedModels() []string {
	return supportedModels
}

func (a *Adapter) Chat(ctx context.Context, req *ir.ChatRequest, cred adapter.Credential) (*ir.ChatResponse, error) {
	return a.base.Chat(ctx, req, cred)
}

func (a *Adapter) ChatStream(ctx context.Context, req *ir.ChatRequest, cred adapter.Credential) (adapter.StreamReader, error) {
	return a.base.ChatStream(ctx, req, cred)
}

func (a *Adapter) Embed(context.Context, *ir.EmbedRequest, adapter.Credential) (*ir.EmbedResponse, error) {
	return nil, fmt.Errorf("grok-build: embedding not supported")
}
```

- [ ] **Step 4: Run the focused tests and verify they pass**

Run: `go test ./internal/adapter/grokbuild`

Expected: PASS.

- [ ] **Step 5: Commit Task 1**

```bash
git add internal/adapter/grokbuild
git commit -m "feat(adapter): add grok build provider"
```

---

### Task 2: Add `grok-web` Model Mapping and Payload Helpers

**Files:**
- Create: `internal/adapter/grokweb/model.go`
- Create: `internal/adapter/grokweb/payload.go`
- Create: `internal/adapter/grokweb/payload_test.go`

**Interfaces:**
- Consumes: `ir.ChatRequest`, `ir.Message`, `adapter.Credential`
- Produces: `func lookupModel(clientModel string) (modelSpec, bool)`, `func buildSSOCookie(token string) string`, `func buildPayload(req *ir.ChatRequest) (map[string]any, error)`, `func buildHeaders(cred adapter.Credential) http.Header`

- [ ] **Step 1: Write failing payload/helper tests**

Create `internal/adapter/grokweb/payload_test.go`:

```go
package grokweb

import (
	"strings"
	"testing"

	"github.com/ijry/pro-api/internal/adapter"
	"github.com/ijry/pro-api/internal/protocol/ir"
)

func TestLookupModel(t *testing.T) {
	got, ok := lookupModel("grok-4.1-fast")
	if !ok {
		t.Fatalf("expected model mapping")
	}
	if got.ModelName != "grok-4-1-thinking-1129" || got.ModelMode != "MODEL_MODE_FAST" {
		t.Fatalf("mapping = %#v", got)
	}
}

func TestLookupModelAcceptsCatalogPrefix(t *testing.T) {
	got, ok := lookupModel("grok-web/grok-4")
	if !ok {
		t.Fatalf("expected prefixed model mapping")
	}
	if got.ModelName != "grok-4" || got.ModelMode != "MODEL_MODE_GROK_4" {
		t.Fatalf("mapping = %#v", got)
	}
}

func TestBuildSSOCookie(t *testing.T) {
	got := buildSSOCookie("sso=abc123")
	if got != "sso=abc123; sso-rw=abc123" {
		t.Fatalf("cookie = %q", got)
	}
}

func TestBuildHeaders(t *testing.T) {
	h := buildHeaders(adapter.Credential{APIKey: "abc123"})
	if h.Get("Cookie") != "sso=abc123; sso-rw=abc123" {
		t.Fatalf("cookie = %q", h.Get("Cookie"))
	}
	if h.Get("Origin") != "https://grok.com" {
		t.Fatalf("origin = %q", h.Get("Origin"))
	}
	if h.Get("x-xai-request-id") == "" {
		t.Fatalf("missing request id")
	}
	if h.Get("x-statsig-id") == "" {
		t.Fatalf("missing statsig id")
	}
}

func TestBuildPayloadFlattensMessagesAndModel(t *testing.T) {
	temp := 0.7
	topP := 0.9
	body, err := buildPayload(&ir.ChatRequest{
		Model:       "grok-4-thinking",
		Temperature: &temp,
		TopP:        &topP,
		Messages: []ir.Message{
			{Role: ir.RoleSystem, Content: []ir.ContentPart{{Type: ir.ContentText, Text: "be direct"}}},
			{Role: ir.RoleUser, Content: []ir.ContentPart{{Type: ir.ContentText, Text: "hello"}}},
			{Role: ir.RoleAssistant, Content: []ir.ContentPart{{Type: ir.ContentText, Text: "hi"}}},
			{Role: ir.RoleUser, Content: []ir.ContentPart{{Type: ir.ContentText, Text: "continue"}}},
		},
	})
	if err != nil {
		t.Fatalf("buildPayload error: %v", err)
	}
	if body["modelName"] != "grok-4" || body["modelMode"] != "MODEL_MODE_GROK_4_THINKING" {
		t.Fatalf("model fields = %#v / %#v", body["modelName"], body["modelMode"])
	}
	msg, _ := body["message"].(string)
	for _, part := range []string{"system: be direct", "user: hello", "assistant: hi", "continue"} {
		if !strings.Contains(msg, part) {
			t.Fatalf("message %q missing %q", msg, part)
		}
	}
	meta := body["responseMetadata"].(map[string]any)
	override := meta["modelConfigOverride"].(map[string]any)
	if override["temperature"] != temp || override["topP"] != topP {
		t.Fatalf("override = %#v", override)
	}
}

func TestBuildPayloadRejectsUnknownModel(t *testing.T) {
	_, err := buildPayload(&ir.ChatRequest{Model: "unknown"})
	if err == nil {
		t.Fatalf("expected unknown model error")
	}
}
```

- [ ] **Step 2: Run tests and verify they fail**

Run: `go test ./internal/adapter/grokweb`

Expected: FAIL because package `internal/adapter/grokweb` does not exist.

- [ ] **Step 3: Implement model mapping**

Create `internal/adapter/grokweb/model.go`:

```go
package grokweb

import "strings"

type modelSpec struct {
	ModelName string
	ModelMode string
}

var supportedModels = []string{
	"grok-3",
	"grok-3-mini",
	"grok-3-thinking",
	"grok-4",
	"grok-4-mini",
	"grok-4-thinking",
	"grok-4-heavy",
	"grok-4.1-mini",
	"grok-4.1-fast",
	"grok-4.1-expert",
	"grok-4.1-thinking",
}

var modelMap = map[string]modelSpec{
	"grok-3":            {ModelName: "grok-3", ModelMode: "MODEL_MODE_GROK_3"},
	"grok-3-mini":       {ModelName: "grok-3", ModelMode: "MODEL_MODE_GROK_3_MINI_THINKING"},
	"grok-3-thinking":   {ModelName: "grok-3", ModelMode: "MODEL_MODE_GROK_3_THINKING"},
	"grok-4":            {ModelName: "grok-4", ModelMode: "MODEL_MODE_GROK_4"},
	"grok-4-mini":       {ModelName: "grok-4-mini", ModelMode: "MODEL_MODE_GROK_4_MINI_THINKING"},
	"grok-4-thinking":   {ModelName: "grok-4", ModelMode: "MODEL_MODE_GROK_4_THINKING"},
	"grok-4-heavy":      {ModelName: "grok-4", ModelMode: "MODEL_MODE_HEAVY"},
	"grok-4.1-mini":     {ModelName: "grok-4-1-thinking-1129", ModelMode: "MODEL_MODE_GROK_4_1_MINI_THINKING"},
	"grok-4.1-fast":     {ModelName: "grok-4-1-thinking-1129", ModelMode: "MODEL_MODE_FAST"},
	"grok-4.1-expert":   {ModelName: "grok-4-1-thinking-1129", ModelMode: "MODEL_MODE_EXPERT"},
	"grok-4.1-thinking": {ModelName: "grok-4-1-thinking-1129", ModelMode: "MODEL_MODE_GROK_4_1_THINKING"},
}

func lookupModel(clientModel string) (modelSpec, bool) {
	clientModel = strings.TrimPrefix(clientModel, "grok-web/")
	spec, ok := modelMap[clientModel]
	return spec, ok
}
```

- [ ] **Step 4: Implement payload and header helpers**

Create `internal/adapter/grokweb/payload.go`:

```go
package grokweb

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

	"github.com/ijry/pro-api/internal/adapter"
	"github.com/ijry/pro-api/internal/protocol/ir"
)

const defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"

func buildSSOCookie(token string) string {
	token = strings.TrimSpace(token)
	token = strings.TrimPrefix(token, "sso=")
	return "sso=" + token + "; sso-rw=" + token
}

func buildHeaders(cred adapter.Credential) http.Header {
	h := http.Header{}
	h.Set("Accept", "*/*")
	h.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	h.Set("Content-Type", "application/json")
	h.Set("Origin", "https://grok.com")
	h.Set("Referer", "https://grok.com/")
	h.Set("Sec-Fetch-Dest", "empty")
	h.Set("Sec-Fetch-Mode", "cors")
	h.Set("Sec-Fetch-Site", "same-origin")
	h.Set("User-Agent", defaultUserAgent)
	h.Set("Cookie", buildSSOCookie(cred.APIKey))
	h.Set("x-statsig-id", randomID(16))
	h.Set("x-xai-request-id", randomID(16))
	return h
}

func randomID(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return strings.Repeat("0", n*2)
	}
	return hex.EncodeToString(b)
}

func buildPayload(req *ir.ChatRequest) (map[string]any, error) {
	spec, ok := lookupModel(req.Model)
	if !ok {
		return nil, fmt.Errorf("grok-web: unsupported model %q", req.Model)
	}
	message := flattenMessages(req.Messages)
	body := map[string]any{
		"deviceEnvInfo": map[string]any{
			"darkModeEnabled": false,
			"devicePixelRatio": 2,
			"screenWidth": 2056,
			"screenHeight": 1329,
			"viewportWidth": 2056,
			"viewportHeight": 1083,
		},
		"disableMemory": false,
		"disableSearch": false,
		"disableSelfHarmShortCircuit": false,
		"disableTextFollowUps": false,
		"enableImageGeneration": false,
		"enableImageStreaming": false,
		"enableSideBySide": true,
		"fileAttachments": []string{},
		"forceConcise": false,
		"forceSideBySide": false,
		"imageAttachments": []string{},
		"imageGenerationCount": 0,
		"isAsyncChat": false,
		"isReasoning": false,
		"message": message,
		"modelMode": spec.ModelMode,
		"modelName": spec.ModelName,
		"responseMetadata": map[string]any{
			"requestModelDetails": map[string]any{"modelId": spec.ModelName},
		},
		"returnImageBytes": false,
		"returnRawGrokInXaiRequest": false,
		"sendFinalMetadata": true,
		"temporary": true,
		"toolOverrides": map[string]any{},
	}
	if req.Temperature != nil || req.TopP != nil {
		override := map[string]any{}
		if req.Temperature != nil {
			override["temperature"] = *req.Temperature
		}
		if req.TopP != nil {
			override["topP"] = *req.TopP
		}
		body["responseMetadata"].(map[string]any)["modelConfigOverride"] = override
	}
	return body, nil
}

func flattenMessages(messages []ir.Message) string {
	var parts []string
	lastUser := -1
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == ir.RoleUser {
			lastUser = i
			break
		}
	}
	for i, msg := range messages {
		text := flattenContent(msg.Content)
		if strings.TrimSpace(text) == "" {
			continue
		}
		if i == lastUser {
			parts = append(parts, text)
			continue
		}
		role := msg.Role
		if role == "" {
			role = ir.RoleUser
		}
		parts = append(parts, role+": "+text)
	}
	return strings.Join(parts, "\n\n")
}

func flattenContent(parts []ir.ContentPart) string {
	var out []string
	for _, p := range parts {
		switch p.Type {
		case ir.ContentText:
			if strings.TrimSpace(p.Text) != "" {
				out = append(out, p.Text)
			}
		case ir.ContentImageURL:
			if p.ImageURL.URL != "" {
				out = append(out, "[image: "+p.ImageURL.URL+"]")
			}
		}
	}
	return strings.Join(out, "\n")
}
```

- [ ] **Step 5: Run payload/helper tests and verify they pass**

Run: `go test ./internal/adapter/grokweb -run 'Test(Build|Lookup)'`

Expected: PASS.

- [ ] **Step 6: Commit Task 2**

```bash
git add internal/adapter/grokweb/model.go internal/adapter/grokweb/payload.go internal/adapter/grokweb/payload_test.go
git commit -m "feat(adapter): add grok web payload mapping"
```

---

### Task 3: Implement `grok-web` Adapter and Stream Reader

**Files:**
- Create: `internal/adapter/grokweb/adapter.go`
- Create: `internal/adapter/grokweb/adapter_test.go`

**Interfaces:**
- Consumes: `buildPayload`, `buildHeaders`, `lookupModel`, `adapter.ClassifyHTTPStatus`, `adapter.ClassifyNetErr`
- Produces: `func New() adapter.Adapter`, `func NewWithBase(baseURL string) adapter.Adapter`, `type streamReader struct`

- [ ] **Step 1: Write failing adapter tests**

Create `internal/adapter/grokweb/adapter_test.go`:

```go
package grokweb

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ijry/pro-api/internal/adapter"
	"github.com/ijry/pro-api/internal/protocol/ir"
	"github.com/ijry/pro-api/pkg/apierr"
)

func chatReq(stream bool) *ir.ChatRequest {
	return &ir.ChatRequest{
		Model:  "grok-4",
		Stream: stream,
		Messages: []ir.Message{{
			Role:    ir.RoleUser,
			Content: []ir.ContentPart{{Type: ir.ContentText, Text: "ping"}},
		}},
	}
}

func TestAdapterMetadata(t *testing.T) {
	a := New()
	if a.Name() != "grok-web" {
		t.Fatalf("Name() = %q", a.Name())
	}
	caps := a.Capabilities()
	if !caps.Has(adapter.CapChat) || !caps.Has(adapter.CapStream) {
		t.Fatalf("missing chat/stream caps: %v", caps)
	}
	if caps.Has(adapter.CapEmbedding) || caps.Has(adapter.CapImage) {
		t.Fatalf("unexpected caps: %v", caps)
	}
}

func TestChatCollectsFinalModelResponse(t *testing.T) {
	var gotPath string
	var gotCookie string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotCookie = r.Header.Get("Cookie")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"result":{"response":{"token":"ignored partial"}}}` + "\n"))
		_, _ = w.Write([]byte(`{"result":{"response":{"modelResponse":{"responseId":"resp-1","message":"final answer","metadata":{"llm_info":{"modelHash":"hash-1"}}}}}}` + "\n"))
	}))
	defer srv.Close()

	resp, err := NewWithBase(srv.URL).Chat(context.Background(), chatReq(false), adapter.Credential{APIKey: "sso-token"})
	if err != nil {
		t.Fatalf("Chat error: %v", err)
	}
	if gotPath != "/rest/app-chat/conversations/new" {
		t.Fatalf("path = %q", gotPath)
	}
	if gotCookie != "sso=sso-token; sso-rw=sso-token" {
		t.Fatalf("cookie = %q", gotCookie)
	}
	if resp.ID != "resp-1" || resp.SystemFingerprint != "hash-1" {
		t.Fatalf("resp metadata = %#v", resp)
	}
	if resp.Choices[0].Message.Content[0].Text != "final answer" {
		t.Fatalf("content = %#v", resp.Choices[0].Message.Content)
	}
}

func TestChatStreamSkipsBadLinesAndReturnsTokens(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flusher, _ := w.(http.Flusher)
		_, _ = w.Write([]byte("\n"))
		_, _ = w.Write([]byte("{bad json}\n"))
		_, _ = w.Write([]byte(`{"result":{"response":{"responseId":"resp-2","token":"hello"}}}` + "\n"))
		_, _ = w.Write([]byte(`{"result":{"response":{"token":" world"}}}` + "\n"))
		if flusher != nil {
			flusher.Flush()
		}
	}))
	defer srv.Close()

	reader, err := NewWithBase(srv.URL).ChatStream(context.Background(), chatReq(true), adapter.Credential{APIKey: "abc"})
	if err != nil {
		t.Fatalf("ChatStream error: %v", err)
	}
	defer reader.Close()

	first, err := reader.Next(context.Background())
	if err != nil {
		t.Fatalf("first chunk: %v", err)
	}
	if first.ID != "resp-2" || first.Delta.Content != "hello" {
		t.Fatalf("first = %#v", first)
	}
	second, err := reader.Next(context.Background())
	if err != nil {
		t.Fatalf("second chunk: %v", err)
	}
	if second.Delta.Content != " world" {
		t.Fatalf("second = %#v", second)
	}
	_, err = reader.Next(context.Background())
	if !errors.Is(err, io.EOF) {
		t.Fatalf("EOF err = %v", err)
	}
}

func TestHTTPErrorClassified(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "rate limited", http.StatusTooManyRequests)
	}))
	defer srv.Close()

	_, err := NewWithBase(srv.URL).Chat(context.Background(), chatReq(false), adapter.Credential{APIKey: "abc"})
	if err == nil {
		t.Fatalf("expected error")
	}
	var ae *apierr.Error
	if !errors.As(err, &ae) || ae.Code != apierr.CodeUpstreamRateLimit {
		t.Fatalf("err = %#v", err)
	}
}

func TestEmbedUnsupported(t *testing.T) {
	_, err := New().Embed(context.Background(), &ir.EmbedRequest{Model: "grok-4"}, adapter.Credential{APIKey: "abc"})
	if err == nil || !strings.Contains(err.Error(), "embedding not supported") {
		t.Fatalf("err = %v", err)
	}
}
```

- [ ] **Step 2: Run tests and verify they fail**

Run: `go test ./internal/adapter/grokweb`

Expected: FAIL with undefined `New`, `NewWithBase`, or adapter methods.

- [ ] **Step 3: Implement `grok-web` adapter**

Create `internal/adapter/grokweb/adapter.go`:

```go
// Package grokweb provides Grok Web SSO-cookie adapter support.
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

// Adapter calls Grok Web's app-chat endpoint and translates line-delimited JSON to IR.
type Adapter struct {
	baseURL string
	client  *http.Client
}

// New returns a Grok Web adapter with the default upstream base.
func New() adapter.Adapter {
	return NewWithBase(defaultBaseURL)
}

// NewWithBase returns a Grok Web adapter using baseURL for tests or channel overrides.
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

func (a *Adapter) Name() string {
	return "grok-web"
}

func (a *Adapter) Capabilities() adapter.Capability {
	return adapter.CapChat | adapter.CapStream
}

func (a *Adapter) SupportedModels() []string {
	return supportedModels
}

func (a *Adapter) Chat(ctx context.Context, req *ir.ChatRequest, cred adapter.Credential) (*ir.ChatResponse, error) {
	resp, err := a.send(ctx, req, cred)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	collector := newCollector(req.Model)
	scanner := newLineScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		event, ok := parseLine(line)
		if !ok {
			continue
		}
		collector.accept(event)
	}
	if err := scanner.Err(); err != nil {
		return nil, adapter.ClassifyNetErr(err)
	}
	return collector.response(), nil
}

func (a *Adapter) ChatStream(ctx context.Context, req *ir.ChatRequest, cred adapter.Credential) (adapter.StreamReader, error) {
	resp, err := a.send(ctx, req, cred)
	if err != nil {
		return nil, err
	}
	return &streamReader{body: resp.Body, scanner: newLineScanner(resp.Body), model: req.Model}, nil
}

func (a *Adapter) Embed(context.Context, *ir.EmbedRequest, adapter.Credential) (*ir.EmbedResponse, error) {
	return nil, fmt.Errorf("grok-web: embedding not supported")
}

func (a *Adapter) send(ctx context.Context, req *ir.ChatRequest, cred adapter.Credential) (*http.Response, error) {
	body, err := buildPayload(req)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("grok-web: encode request: %w", err)
	}
	base := a.baseURL
	if cred.BaseURL != "" {
		base = strings.TrimRight(cred.BaseURL, "/")
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/rest/app-chat/conversations/new", bytes.NewReader(data))
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

type lineScanner interface {
	Scan() bool
	Text() string
	Err() error
}

func newLineScanner(r io.Reader) *bufio.Scanner {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
	return scanner
}

type responseEvent struct {
	ResponseID string
	Token      string
	Message    string
	ModelHash  string
	HasToken   bool
	HasFinal   bool
}

func parseLine(line string) (responseEvent, bool) {
	var raw struct {
		Result struct {
			Response struct {
				ResponseID string `json:"responseId"`
				Token      *string `json:"token"`
				LLMInfo    struct {
					ModelHash string `json:"modelHash"`
				} `json:"llmInfo"`
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
		return responseEvent{}, false
	}
	resp := raw.Result.Response
	ev := responseEvent{ResponseID: resp.ResponseID, ModelHash: resp.LLMInfo.ModelHash}
	if resp.Token != nil {
		ev.Token = *resp.Token
		ev.HasToken = true
	}
	if resp.ModelResponse != nil {
		ev.HasFinal = true
		ev.Message = resp.ModelResponse.Message
		if resp.ModelResponse.ResponseID != "" {
			ev.ResponseID = resp.ModelResponse.ResponseID
		}
		if resp.ModelResponse.Metadata.LLMInfo.ModelHash != "" {
			ev.ModelHash = resp.ModelResponse.Metadata.LLMInfo.ModelHash
		}
	}
	if !ev.HasToken && !ev.HasFinal && ev.ResponseID == "" && ev.ModelHash == "" {
		return responseEvent{}, false
	}
	return ev, true
}

type collector struct {
	id          string
	model       string
	fingerprint string
	content     strings.Builder
	final       string
}

func newCollector(model string) *collector {
	return &collector{model: model}
}

func (c *collector) accept(ev responseEvent) {
	if ev.ResponseID != "" {
		c.id = ev.ResponseID
	}
	if ev.ModelHash != "" {
		c.fingerprint = ev.ModelHash
	}
	if ev.HasFinal {
		c.final = ev.Message
		return
	}
	if ev.HasToken {
		c.content.WriteString(ev.Token)
	}
}

func (c *collector) response() *ir.ChatResponse {
	content := c.final
	if content == "" {
		content = c.content.String()
	}
	if c.id == "" {
		c.id = "grok-web"
	}
	return &ir.ChatResponse{
		ID:                c.id,
		Model:             c.model,
		SystemFingerprint: c.fingerprint,
		Choices: []ir.Choice{{
			Index: 0,
			Message: ir.Message{
				Role:    ir.RoleAssistant,
				Content: []ir.ContentPart{{Type: ir.ContentText, Text: content}},
			},
			FinishReason: ir.FinishStop,
		}},
		Usage: ir.Usage{},
	}
}

type streamReader struct {
	body    io.ReadCloser
	scanner lineScanner
	model   string
	id      string
	closed  bool
	done    bool
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
```

- [ ] **Step 4: Run focused adapter tests**

Run: `go test ./internal/adapter/grokweb`

Expected: PASS.

- [ ] **Step 5: Run adapter package tests**

Run: `go test ./internal/adapter/...`

Expected: PASS.

- [ ] **Step 6: Commit Task 3**

```bash
git add internal/adapter/grokweb
git commit -m "feat(adapter): add grok web provider"
```

---

### Task 4: Add Grok Account Probes and Wiring

**Files:**
- Create: `internal/account/probe/grok_build.go`
- Create: `internal/account/probe/grok_build_test.go`
- Create: `internal/account/probe/grok_web.go`
- Create: `internal/account/probe/grok_web_test.go`
- Modify: `internal/account/wire/wire.go`

**Interfaces:**
- Consumes: `account.ProviderProbe`, `account.AccountCred`, `probe.NewOpenAI`, `probe.NewAnthropic`
- Produces: `func NewGrokBuild(base string) *GrokBuild`, `func NewGrokWeb(base string) *GrokWeb`

- [ ] **Step 1: Write failing probe tests**

Create `internal/account/probe/grok_build_test.go`:

```go
package probe_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ijry/pro-api/internal/account"
	"github.com/ijry/pro-api/internal/account/probe"
)

func TestGrokBuildProbeUsesBearerAndModelsPath(t *testing.T) {
	var gotPath string
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer srv.Close()

	p := probe.NewGrokBuild(srv.URL)
	if _, err := p.Probe(context.Background(), account.AccountCred{APIKey: "xai-key"}); err != nil {
		t.Fatalf("Probe error: %v", err)
	}
	if gotPath != "/v1/models" {
		t.Fatalf("path = %q", gotPath)
	}
	if gotAuth != "Bearer xai-key" {
		t.Fatalf("auth = %q", gotAuth)
	}
}

func TestGrokBuildProbeReturnsErrorOnBadStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "no", http.StatusUnauthorized)
	}))
	defer srv.Close()

	_, err := probe.NewGrokBuild(srv.URL).Probe(context.Background(), account.AccountCred{APIKey: "bad"})
	if err == nil {
		t.Fatalf("expected error")
	}
}
```

Create `internal/account/probe/grok_web_test.go`:

```go
package probe_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ijry/pro-api/internal/account"
	"github.com/ijry/pro-api/internal/account/probe"
)

func TestGrokWebProbeUsesSSOCookieAndRateLimitsPath(t *testing.T) {
	var gotPath string
	var gotCookie string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotCookie = r.Header.Get("Cookie")
		_, _ = w.Write([]byte(`{"remainingQueries":80}`))
	}))
	defer srv.Close()

	p := probe.NewGrokWeb(srv.URL)
	if _, err := p.Probe(context.Background(), account.AccountCred{APIKey: "sso-token"}); err != nil {
		t.Fatalf("Probe error: %v", err)
	}
	if gotPath != "/rest/rate-limits" {
		t.Fatalf("path = %q", gotPath)
	}
	if gotCookie != "sso=sso-token; sso-rw=sso-token" {
		t.Fatalf("cookie = %q", gotCookie)
	}
}

func TestGrokWebProbeFallsBackToAccessToken(t *testing.T) {
	var gotCookie string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotCookie = r.Header.Get("Cookie")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	p := probe.NewGrokWeb(srv.URL)
	if _, err := p.Probe(context.Background(), account.AccountCred{AccessToken: "at-token"}); err != nil {
		t.Fatalf("Probe error: %v", err)
	}
	if gotCookie != "sso=at-token; sso-rw=at-token" {
		t.Fatalf("cookie = %q", gotCookie)
	}
}
```

- [ ] **Step 2: Run probe tests and verify they fail**

Run: `go test ./internal/account/probe -run 'Grok(Build|Web)'`

Expected: FAIL because `NewGrokBuild` and `NewGrokWeb` are undefined.

- [ ] **Step 3: Implement `grok-build` probe**

Create `internal/account/probe/grok_build.go`:

```go
package probe

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ijry/pro-api/internal/account"
)

const grokBuildDefaultBase = "https://api.x.ai"

// GrokBuild probes xAI official API credentials using GET /v1/models.
type GrokBuild struct {
	base   string
	client *http.Client
}

func NewGrokBuild(base string) *GrokBuild {
	if base == "" {
		base = grokBuildDefaultBase
	}
	return &GrokBuild{base: strings.TrimRight(base, "/"), client: &http.Client{Timeout: 3 * time.Second}}
}

func (g *GrokBuild) Probe(ctx context.Context, cred account.AccountCred) (http.Header, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, g.base+"/v1/models", nil)
	if err != nil {
		return nil, err
	}
	tok := cred.APIKey
	if tok == "" {
		tok = cred.AccessToken
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode >= 400 {
		return resp.Header, fmt.Errorf("grok-build probe: status %d", resp.StatusCode)
	}
	return resp.Header, nil
}
```

- [ ] **Step 4: Implement `grok-web` probe**

Create `internal/account/probe/grok_web.go`:

```go
package probe

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ijry/pro-api/internal/account"
)

const grokWebDefaultBase = "https://grok.com"

// GrokWeb probes Grok Web SSO credentials using GET /rest/rate-limits.
type GrokWeb struct {
	base   string
	client *http.Client
}

func NewGrokWeb(base string) *GrokWeb {
	if base == "" {
		base = grokWebDefaultBase
	}
	return &GrokWeb{base: strings.TrimRight(base, "/"), client: &http.Client{Timeout: 3 * time.Second}}
}

func (g *GrokWeb) Probe(ctx context.Context, cred account.AccountCred) (http.Header, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, g.base+"/rest/rate-limits", nil)
	if err != nil {
		return nil, err
	}
	tok := cred.APIKey
	if tok == "" {
		tok = cred.AccessToken
	}
	tok = strings.TrimPrefix(tok, "sso=")
	req.Header.Set("Cookie", "sso="+tok+"; sso-rw="+tok)
	req.Header.Set("Origin", "https://grok.com")
	req.Header.Set("Referer", "https://grok.com/")
	resp, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode >= 400 {
		return resp.Header, fmt.Errorf("grok-web probe: status %d", resp.StatusCode)
	}
	return resp.Header, nil
}
```

- [ ] **Step 5: Register probes in account wiring**

Modify `internal/account/wire/wire.go` inside the `account.NewProbe` providers map:

```go
pr := account.NewProbe(repo, tracker, br, map[string]account.ProviderProbe{
	"anthropic":  probe.NewAnthropic(anthropicProbeBase),
	"openai":     probe.NewOpenAI(openaiProbeBase),
	"grok-build": probe.NewGrokBuild(""),
	"grok-web":   probe.NewGrokWeb(""),
}, probeCfg, a.Log)
```

- [ ] **Step 6: Run account probe tests**

Run: `go test ./internal/account/probe ./internal/account`

Expected: PASS.

- [ ] **Step 7: Commit Task 4**

```bash
git add internal/account/probe/grok_build.go internal/account/probe/grok_build_test.go internal/account/probe/grok_web.go internal/account/probe/grok_web_test.go internal/account/wire/wire.go
git commit -m "feat(account): add grok account probes"
```

---

### Task 5: Register Adapters, Seed Models, and Update Docs

**Files:**
- Modify: `internal/adapterreg/wire.go`
- Modify: `migrations/mysql/000023_seed_model_catalogs.up.sql`
- Modify: `migrations/mysql/000023_seed_model_catalogs.down.sql`
- Modify: `migrations/postgres/000023_seed_model_catalogs.up.sql`
- Modify: `migrations/postgres/000023_seed_model_catalogs.down.sql`
- Modify: `docs-site/zh/architecture/adapter-layer.md`
- Modify: `docs-site/en/architecture/adapter-layer.md`
- Modify: `README.md`
- Modify: `README_zh.md`

**Interfaces:**
- Consumes: `grokbuild.New() adapter.Adapter`, `grokweb.New() adapter.Adapter`
- Produces: registered provider names `grok-build` and `grok-web`; seeded model IDs `33..47`

- [ ] **Step 1: Add adapter registration**

Modify `internal/adapterreg/wire.go` imports:

```go
	"github.com/ijry/pro-api/internal/adapter/grokbuild"
	"github.com/ijry/pro-api/internal/adapter/grokweb"
```

Add registration after `groq.New()`:

```go
	reg.Register(groq.New())
	reg.Register(grokbuild.New())
	reg.Register(grokweb.New())
	reg.Register(mistral.New())
```

Update the comment above `WireAdapters` to:

```go
// WireAdapters 向 Registry 注册所有 20 家 adapter，并注册 tokenizers。
```

- [ ] **Step 2: Add adapter registration tests**

Create `internal/adapterreg/wire_test.go`:

```go
package adapterreg_test

import (
	"testing"

	"github.com/ijry/pro-api/internal/adapter"
	"github.com/ijry/pro-api/internal/adapterreg"
)

func TestWireAdaptersRegistersGrokProviders(t *testing.T) {
	reg := adapter.NewRegistry()
	adapterreg.WireAdapters(reg, nil)
	for _, name := range []string{"grok-build", "grok-web"} {
		a, ok := reg.Get(name)
		if !ok {
			t.Fatalf("missing adapter %q; names=%v", name, reg.Names())
		}
		if a.Name() != name {
			t.Fatalf("adapter name = %q, want %q", a.Name(), name)
		}
	}
}
```

- [ ] **Step 3: Add Postgres seed rows**

Modify `migrations/postgres/000023_seed_model_catalogs.up.sql` by adding rows `33..47` before `ON CONFLICT` and changing row `32` to end with a comma:

```sql
-- Grok Build
(33, 'grok-4',                     'chat',  '["chat","stream"]'::jsonb,                    0.0,   0.0,  NULL, NULL,  131072, 0, 'grok-build', NOW(), NOW()),
(34, 'grok-3',                     'chat',  '["chat","stream"]'::jsonb,                    0.0,   0.0,  NULL, NULL,  131072, 0, 'grok-build', NOW(), NOW()),
(35, 'grok-3-mini',                'chat',  '["chat","stream"]'::jsonb,                    0.0,   0.0,  NULL, NULL,  131072, 0, 'grok-build', NOW(), NOW()),
(36, 'grok-3-mini-fast',           'chat',  '["chat","stream"]'::jsonb,                    0.0,   0.0,  NULL, NULL,  131072, 0, 'grok-build', NOW(), NOW()),
-- Grok Web
(37, 'grok-web/grok-3',            'chat',  '["chat","stream"]'::jsonb,                    0.0,   0.0,  NULL, NULL,  131072, 0, 'grok-web',   NOW(), NOW()),
(38, 'grok-web/grok-3-mini',       'chat',  '["chat","stream"]'::jsonb,                    0.0,   0.0,  NULL, NULL,  131072, 0, 'grok-web',   NOW(), NOW()),
(39, 'grok-web/grok-3-thinking',   'chat',  '["chat","stream"]'::jsonb,                    0.0,   0.0,  NULL, NULL,  131072, 0, 'grok-web',   NOW(), NOW()),
(40, 'grok-web/grok-4',            'chat',  '["chat","stream"]'::jsonb,                    0.0,   0.0,  NULL, NULL,  131072, 0, 'grok-web',   NOW(), NOW()),
(41, 'grok-web/grok-4-mini',       'chat',  '["chat","stream"]'::jsonb,                    0.0,   0.0,  NULL, NULL,  131072, 0, 'grok-web',   NOW(), NOW()),
(42, 'grok-web/grok-4-thinking',   'chat',  '["chat","stream"]'::jsonb,                    0.0,   0.0,  NULL, NULL,  131072, 0, 'grok-web',   NOW(), NOW()),
(43, 'grok-web/grok-4-heavy',      'chat',  '["chat","stream"]'::jsonb,                    0.0,   0.0,  NULL, NULL,  131072, 0, 'grok-web',   NOW(), NOW()),
(44, 'grok-web/grok-4.1-mini',     'chat',  '["chat","stream"]'::jsonb,                    0.0,   0.0,  NULL, NULL,  131072, 0, 'grok-web',   NOW(), NOW()),
(45, 'grok-web/grok-4.1-fast',     'chat',  '["chat","stream"]'::jsonb,                    0.0,   0.0,  NULL, NULL,  131072, 0, 'grok-web',   NOW(), NOW()),
(46, 'grok-web/grok-4.1-expert',   'chat',  '["chat","stream"]'::jsonb,                    0.0,   0.0,  NULL, NULL,  131072, 0, 'grok-web',   NOW(), NOW()),
(47, 'grok-web/grok-4.1-thinking', 'chat',  '["chat","stream"]'::jsonb,                    0.0,   0.0,  NULL, NULL,  131072, 0, 'grok-web',   NOW(), NOW())
```

Use `grok-web/` prefixes for Web model catalog rows to avoid unique-name collisions with `grok-build`. Task 2's `lookupModel` strips that prefix, so direct requests using catalog names such as `grok-web/grok-4` work without requiring channel model mapping.

- [ ] **Step 4: Add MySQL seed rows**

Modify `migrations/mysql/000023_seed_model_catalogs.up.sql` by adding rows `33..47` before `ON DUPLICATE KEY UPDATE` and changing row `32` to end with a comma:

```sql
-- Grok Build
(33, 'grok-4',                     'chat',  JSON_ARRAY('chat','stream'),                    0.0,   0.0,  NULL, NULL,  131072, 0, 'grok-build', NOW(3), NOW(3)),
(34, 'grok-3',                     'chat',  JSON_ARRAY('chat','stream'),                    0.0,   0.0,  NULL, NULL,  131072, 0, 'grok-build', NOW(3), NOW(3)),
(35, 'grok-3-mini',                'chat',  JSON_ARRAY('chat','stream'),                    0.0,   0.0,  NULL, NULL,  131072, 0, 'grok-build', NOW(3), NOW(3)),
(36, 'grok-3-mini-fast',           'chat',  JSON_ARRAY('chat','stream'),                    0.0,   0.0,  NULL, NULL,  131072, 0, 'grok-build', NOW(3), NOW(3)),
-- Grok Web
(37, 'grok-web/grok-3',            'chat',  JSON_ARRAY('chat','stream'),                    0.0,   0.0,  NULL, NULL,  131072, 0, 'grok-web',   NOW(3), NOW(3)),
(38, 'grok-web/grok-3-mini',       'chat',  JSON_ARRAY('chat','stream'),                    0.0,   0.0,  NULL, NULL,  131072, 0, 'grok-web',   NOW(3), NOW(3)),
(39, 'grok-web/grok-3-thinking',   'chat',  JSON_ARRAY('chat','stream'),                    0.0,   0.0,  NULL, NULL,  131072, 0, 'grok-web',   NOW(3), NOW(3)),
(40, 'grok-web/grok-4',            'chat',  JSON_ARRAY('chat','stream'),                    0.0,   0.0,  NULL, NULL,  131072, 0, 'grok-web',   NOW(3), NOW(3)),
(41, 'grok-web/grok-4-mini',       'chat',  JSON_ARRAY('chat','stream'),                    0.0,   0.0,  NULL, NULL,  131072, 0, 'grok-web',   NOW(3), NOW(3)),
(42, 'grok-web/grok-4-thinking',   'chat',  JSON_ARRAY('chat','stream'),                    0.0,   0.0,  NULL, NULL,  131072, 0, 'grok-web',   NOW(3), NOW(3)),
(43, 'grok-web/grok-4-heavy',      'chat',  JSON_ARRAY('chat','stream'),                    0.0,   0.0,  NULL, NULL,  131072, 0, 'grok-web',   NOW(3), NOW(3)),
(44, 'grok-web/grok-4.1-mini',     'chat',  JSON_ARRAY('chat','stream'),                    0.0,   0.0,  NULL, NULL,  131072, 0, 'grok-web',   NOW(3), NOW(3)),
(45, 'grok-web/grok-4.1-fast',     'chat',  JSON_ARRAY('chat','stream'),                    0.0,   0.0,  NULL, NULL,  131072, 0, 'grok-web',   NOW(3), NOW(3)),
(46, 'grok-web/grok-4.1-expert',   'chat',  JSON_ARRAY('chat','stream'),                    0.0,   0.0,  NULL, NULL,  131072, 0, 'grok-web',   NOW(3), NOW(3)),
(47, 'grok-web/grok-4.1-thinking', 'chat',  JSON_ARRAY('chat','stream'),                    0.0,   0.0,  NULL, NULL,  131072, 0, 'grok-web',   NOW(3), NOW(3))
```

- [ ] **Step 5: Update migration down files**

Modify both down files to:

```sql
DELETE FROM model_catalogs WHERE id BETWEEN 1 AND 47;
```

- [ ] **Step 6: Update documentation**

In `docs-site/zh/architecture/adapter-layer.md`, replace heading `## M1 已支持的 9 家` with `## 已支持的 20 家`, then add two rows to the provider table:

```markdown
| grok-build | chat | ✅ | grok-4 / grok-3 / grok-3-mini | xAI 官方 OpenAI-compatible API,API Key |
| grok-web | chat | ✅ | grok-4 / grok-4.1-* / grok-3* | Grok Web SSO Cookie 反代 |
```

In `docs-site/en/architecture/adapter-layer.md`, append this section after the existing introductory bullet list:

```markdown
## Grok Providers

- `grok-build`: xAI official OpenAI-compatible API. Store the xAI API key in the channel/account API key field.
- `grok-web`: Grok Web reverse adapter. Store the bare SSO token in the channel/account API key field; the adapter sends it as `sso` and `sso-rw` cookies.
```

In `README.md`, replace the existing `**18 upstream adapters**` bullet with:

```markdown
- **20 upstream adapters**: openai / azure / anthropic / gemini / deepseek / moonshot / zhipu / qwen / doubao / groq / grok-build / grok-web / mistral / yi / openrouter / huggingface / minimax / tencent / cohere / xunfei
```

Then add this note immediately after that bullet:

```markdown
- Grok support: `grok-build` uses xAI API keys; `grok-web` uses a Grok Web SSO token stored in the channel/account API key field.
```

In `README_zh.md`, replace the existing `**18 个上游适配器**` bullet with:

```markdown
- **20 个上游适配器**:openai / azure / anthropic / gemini / deepseek / moonshot / 智谱 / 通义 / 豆包 / groq / grok-build / grok-web / mistral / 零一 / openrouter / huggingface / minimax / 腾讯混元 / Cohere / 讯飞星火
```

Then add this note immediately after that bullet:

```markdown
- Grok 支持:`grok-build` 使用 xAI API Key;`grok-web` 使用 Grok Web SSO token,存放在渠道或账号的 API Key 字段。
```

- [ ] **Step 7: Run registration and adapter tests**

Run: `go test ./internal/adapterreg ./internal/adapter/...`

Expected: PASS.

- [ ] **Step 8: Commit Task 5**

```bash
git add internal/adapterreg/wire.go internal/adapterreg/wire_test.go migrations/mysql/000023_seed_model_catalogs.up.sql migrations/mysql/000023_seed_model_catalogs.down.sql migrations/postgres/000023_seed_model_catalogs.up.sql migrations/postgres/000023_seed_model_catalogs.down.sql docs-site/zh/architecture/adapter-layer.md docs-site/en/architecture/adapter-layer.md README.md README_zh.md
git commit -m "feat(adapter): register grok providers and seed models"
```

---

### Task 6: Full Verification and Cleanup

**Files:**
- Modify only files changed by earlier tasks if test failures require fixes.

**Interfaces:**
- Consumes: all prior task outputs.
- Produces: passing full test run and final commit if cleanup changes are needed.

- [ ] **Step 1: Run focused package tests**

Run:

```powershell
go test ./internal/adapter/grokbuild ./internal/adapter/grokweb ./internal/account/probe ./internal/adapterreg
```

Expected: PASS.

- [ ] **Step 2: Run broader backend tests**

Run:

```powershell
go test ./internal/adapter/... ./internal/account/... ./internal/relay/... ./internal/server/handler/relay/...
```

Expected: PASS.

- [ ] **Step 3: Run full test suite**

Run:

```powershell
go test ./...
```

Expected: PASS.

- [ ] **Step 4: Inspect git diff for accidental scope creep**

Run:

```powershell
git diff --stat HEAD
git diff --name-only HEAD
```

Expected: only Grok adapter, probe, registration, migration, docs, and test files are listed.

- [ ] **Step 5: Commit cleanup if any files changed during verification**

If Step 3 required code or doc fixes, run:

```bash
git add internal docs-site README.md README_zh.md migrations
git commit -m "fix: stabilize grok provider tests"
```

If Step 3 passed without additional changes, skip this commit step.

---

## Self-Review Notes

- Spec coverage: Tasks 1 and 3 implement both adapters; Task 4 implements probes; Task 5 implements registration, model seeds, and docs; Task 6 verifies the suite.
- Placeholder scan: The plan contains no open-ended placeholders. Every code-producing step includes concrete file paths, function names, and code or exact edit text.
- Type consistency: Adapter constructors return `adapter.Adapter`; probe constructors return concrete probe types implementing `account.ProviderProbe`; stream reader implements `adapter.StreamReader`.
- Scope check: Images, video, audio, embeddings, and file upload remain out of scope. `Embed` returns explicit unsupported errors for both Grok providers.
