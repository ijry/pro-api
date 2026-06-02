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

func TestAnthropic_StartCallbackStub(t *testing.T) {
	a := oauth.NewAnthropic(oauth.Config{TokenURL: "http://x", ClientID: "x"})
	_, _, err := a.Start(context.Background(), 1)
	require.ErrorIs(t, err, oauth.ErrNotImplemented)
	_, err = a.Callback(context.Background(), "s", "c")
	require.ErrorIs(t, err, oauth.ErrNotImplemented)
}
