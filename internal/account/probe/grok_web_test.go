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
