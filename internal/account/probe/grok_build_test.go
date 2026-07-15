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
