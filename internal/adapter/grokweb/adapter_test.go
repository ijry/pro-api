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
