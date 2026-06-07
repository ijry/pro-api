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

func TestOpenAI_ExchangeRefreshToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/oauth/token", r.URL.Path)
		require.NoError(t, r.ParseForm())
		require.Equal(t, "refresh_token", r.FormValue("grant_type"))
		require.Equal(t, "rt_yy", r.FormValue("refresh_token"))
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "openai-at",
			"refresh_token": "openai-new-rt",
			"id_token":      "eyJabc...",
			"expires_in":    1800,
			"token_type":    "Bearer",
		})
	}))
	defer srv.Close()

	o := oauth.NewOpenAI(oauth.Config{TokenURL: srv.URL + "/oauth/token", ClientID: "cli-o"})
	cred, err := o.ExchangeRefreshToken(context.Background(), "rt_yy")
	require.NoError(t, err)
	require.Equal(t, "openai-at", cred.AccessToken)
	require.Equal(t, "openai-new-rt", cred.RefreshToken)
	require.Equal(t, "eyJabc...", cred.IDToken)
	require.False(t, cred.ExpiresAt.IsZero())
}

func TestOpenAI_Refresh_BadResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	o := oauth.NewOpenAI(oauth.Config{TokenURL: srv.URL, ClientID: "x"})
	_, err := o.ExchangeRefreshToken(context.Background(), "rt_x")
	require.Error(t, err)
}

func TestOpenAI_AuthCodeURL(t *testing.T) {
	o := oauth.NewOpenAI(oauth.Config{
		AuthURL:     "https://auth.openai.test/authorize",
		ClientID:    "cli-o",
		RedirectURI: "https://app.test/cb",
		Scopes:      []string{"openid", "profile"},
	})
	u := o.AuthCodeURL("st8", "chal8")
	require.Contains(t, u, "https://auth.openai.test/authorize?")
	require.Contains(t, u, "response_type=code")
	require.Contains(t, u, "client_id=cli-o")
	require.Contains(t, u, "code_challenge=chal8")
	require.Contains(t, u, "code_challenge_method=S256")
	require.Contains(t, u, "state=st8")
	require.Contains(t, u, "redirect_uri=https%3A%2F%2Fapp.test%2Fcb")
	require.Contains(t, u, "scope=openid+profile")
}

func TestOpenAI_ExchangeCode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseForm())
		require.Equal(t, "authorization_code", r.FormValue("grant_type"))
		require.Equal(t, "the-code", r.FormValue("code"))
		require.Equal(t, "the-verifier", r.FormValue("code_verifier"))
		require.Equal(t, "https://app.test/cb", r.FormValue("redirect_uri"))
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "openai-at",
			"refresh_token": "openai-rt",
			"id_token":      "idt",
			"expires_in":    1800,
			"token_type":    "Bearer",
		})
	}))
	defer srv.Close()

	o := oauth.NewOpenAI(oauth.Config{TokenURL: srv.URL, ClientID: "x", RedirectURI: "https://app.test/cb"})
	cred, err := o.ExchangeCode(context.Background(), "the-code", "the-verifier")
	require.NoError(t, err)
	require.Equal(t, "openai-at", cred.AccessToken)
	require.Equal(t, "idt", cred.IDToken)
	require.False(t, cred.ExpiresAt.IsZero())
}
