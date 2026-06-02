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

func TestAnthropic_Probe_ReturnsHeaders(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Contains(t, r.URL.Path, "count_tokens")
		require.NotEmpty(t, r.Header.Get("Authorization"))
		w.Header().Set("anthropic-ratelimit-tokens-remaining", "9500")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"input_tokens":3}`))
	}))
	defer srv.Close()

	p := probe.NewAnthropic(srv.URL)
	h, err := p.Probe(context.Background(), account.AccountCred{AccessToken: "at-x"})
	require.NoError(t, err)
	require.Equal(t, "9500", h.Get("anthropic-ratelimit-tokens-remaining"))
}

func TestAnthropic_Non2xxReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"invalid_api_key"}`))
	}))
	t.Cleanup(srv.Close)

	p := probe.NewAnthropic(srv.URL)
	_, err := p.Probe(context.Background(), account.AccountCred{AccessToken: "at-bad"})
	require.Error(t, err)
}
