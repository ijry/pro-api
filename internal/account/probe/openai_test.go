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

func TestOpenAI_Probe_UsesAccessToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1/models", r.URL.Path)
		require.Equal(t, "Bearer at-X", r.Header.Get("Authorization"))
		w.Header().Set("x-ratelimit-remaining-tokens", "5000")
		w.WriteHeader(200)
	}))
	defer srv.Close()

	p := probe.NewOpenAI(srv.URL)
	h, err := p.Probe(context.Background(), account.AccountCred{AccessToken: "at-X"})
	require.NoError(t, err)
	require.Equal(t, "5000", h.Get("x-ratelimit-remaining-tokens"))
}

func TestOpenAI_Probe_FallsBackToAPIKey(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "Bearer sk-proj-K", r.Header.Get("Authorization"))
		w.WriteHeader(200)
	}))
	defer srv.Close()

	p := probe.NewOpenAI(srv.URL)
	_, err := p.Probe(context.Background(), account.AccountCred{APIKey: "sk-proj-K"})
	require.NoError(t, err)
}
