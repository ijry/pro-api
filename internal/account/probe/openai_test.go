package probe_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ijry/pro-api/internal/account"
	"github.com/ijry/pro-api/internal/account/probe"
	"github.com/stretchr/testify/require"
)

// TestOpenAI_Probe_PrefersAPIKey 验证探测选凭证与 relay.resolveCred 一致:
// 同时有 APIKey 与 AccessToken(典型 Codex OAuth 账号)时优先用 APIKey,
// 否则会出现"探测测 access_token、转发用 api_key"的错配。
func TestOpenAI_Probe_PrefersAPIKey(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1/models", r.URL.Path)
		require.Equal(t, "Bearer sk-proj-K", r.Header.Get("Authorization"))
		w.Header().Set("x-ratelimit-remaining-tokens", "5000")
		w.WriteHeader(200)
	}))
	defer srv.Close()

	p := probe.NewOpenAI(srv.URL)
	h, err := p.Probe(context.Background(), account.AccountCred{APIKey: "sk-proj-K", AccessToken: "at-X"})
	require.NoError(t, err)
	require.Equal(t, "5000", h.Get("x-ratelimit-remaining-tokens"))
}

func TestOpenAI_Probe_FallsBackToAccessToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "Bearer at-X", r.Header.Get("Authorization"))
		w.WriteHeader(200)
	}))
	defer srv.Close()

	p := probe.NewOpenAI(srv.URL)
	_, err := p.Probe(context.Background(), account.AccountCred{AccessToken: "at-X"})
	require.NoError(t, err)
}

func TestOpenAI_Non2xxReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":{"message":"invalid api key"}}`))
	}))
	t.Cleanup(srv.Close)

	p := probe.NewOpenAI(srv.URL)
	_, err := p.Probe(context.Background(), account.AccountCred{APIKey: "sk-bad"})
	require.Error(t, err)
}
