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

func TestOpenAI_StartCallbackStub(t *testing.T) {
	o := oauth.NewOpenAI(oauth.Config{TokenURL: "http://x", ClientID: "x"})
	_, _, err := o.Start(context.Background(), 1)
	require.ErrorIs(t, err, oauth.ErrNotImplemented)
	_, err = o.Callback(context.Background(), "s", "c")
	require.ErrorIs(t, err, oauth.ErrNotImplemented)
}
