package oauth_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ijry/pro-api/internal/account/oauth"
	"github.com/stretchr/testify/require"
)

func TestAnthropic_ExchangeRefreshToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/oauth/token", r.URL.Path)
		require.NoError(t, r.ParseForm())
		require.Equal(t, "refresh_token", r.FormValue("grant_type"))
		require.Equal(t, "rt-XX", r.FormValue("refresh_token"))
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "new-at",
			"refresh_token": "new-rt",
			"expires_in":    3600,
			"token_type":    "Bearer",
		})
	}))
	defer srv.Close()

	a := oauth.NewAnthropic(oauth.Config{TokenURL: srv.URL + "/oauth/token", ClientID: "cli-x"})
	cred, err := a.ExchangeRefreshToken(context.Background(), "rt-XX")
	require.NoError(t, err)
	require.Equal(t, "new-at", cred.AccessToken)
	require.Equal(t, "new-rt", cred.RefreshToken)
	require.False(t, cred.ExpiresAt.IsZero())
}

func TestAnthropic_Refresh_InvalidGrant(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid_grant"}`))
	}))
	defer srv.Close()

	a := oauth.NewAnthropic(oauth.Config{TokenURL: srv.URL, ClientID: "cli-x"})
	_, err := a.ExchangeRefreshToken(context.Background(), "rt-X")
	require.Error(t, err)
}

func TestAnthropic_AuthCodeURL(t *testing.T) {
	a := oauth.NewAnthropic(oauth.Config{
		AuthURL:     "https://claude.test/authorize",
		ClientID:    "cli-x",
		RedirectURI: "https://app.test/cb",
		Scopes:      []string{"org:create_api_key"},
	})
	u := a.AuthCodeURL("s1", "c1")
	require.Contains(t, u, "https://claude.test/authorize?")
	require.Contains(t, u, "response_type=code")
	require.Contains(t, u, "client_id=cli-x")
	require.Contains(t, u, "code_challenge=c1")
	require.Contains(t, u, "code_challenge_method=S256")
	require.Contains(t, u, "state=s1")
}

func TestAnthropic_ExchangeCode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseForm())
		require.Equal(t, "authorization_code", r.FormValue("grant_type"))
		require.Equal(t, "ac", r.FormValue("code"))
		require.Equal(t, "ver", r.FormValue("code_verifier"))
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "an-at",
			"refresh_token": "an-rt",
			"expires_in":    3600,
			"token_type":    "Bearer",
			"scope":         "org:create_api_key",
		})
	}))
	defer srv.Close()

	a := oauth.NewAnthropic(oauth.Config{TokenURL: srv.URL, ClientID: "cli-x", RedirectURI: "https://app.test/cb"})
	cred, err := a.ExchangeCode(context.Background(), "ac", "ver")
	require.NoError(t, err)
	require.Equal(t, "an-at", cred.AccessToken)
	require.Equal(t, "org:create_api_key", cred.Scope)
	require.False(t, cred.ExpiresAt.IsZero())
}
