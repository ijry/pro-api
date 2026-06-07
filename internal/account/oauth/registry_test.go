package oauth_test

import (
	"context"
	"testing"
	"time"

	"github.com/ijry/pro-api/internal/account"
	"github.com/ijry/pro-api/internal/account/oauth"
	"github.com/stretchr/testify/require"
)

type stubProvider struct {
	cred    *account.AccountCred
	err     error
	authURL string
}

func (s *stubProvider) AuthCodeURL(state, challenge string) string {
	return s.authURL + "?state=" + state + "&code_challenge=" + challenge
}
func (s *stubProvider) ExchangeCode(_ context.Context, _, _ string) (*account.AccountCred, error) {
	return s.cred, s.err
}
func (s *stubProvider) ExchangeRefreshToken(_ context.Context, _ string) (*account.AccountCred, error) {
	return s.cred, s.err
}

func TestRegistry_DispatchesByProvider(t *testing.T) {
	a := &stubProvider{cred: &account.AccountCred{AccessToken: "a-at"}}
	o := &stubProvider{cred: &account.AccountCred{AccessToken: "o-at"}}
	flow := oauth.NewFlow(map[string]oauth.Provider{"anthropic": a, "openai": o}, oauth.NewMemStateStore())

	got, err := flow.ExchangeRefreshToken(context.Background(), "anthropic", "rt")
	require.NoError(t, err)
	require.Equal(t, "a-at", got.AccessToken)

	got, err = flow.ExchangeRefreshToken(context.Background(), "openai", "rt")
	require.NoError(t, err)
	require.Equal(t, "o-at", got.AccessToken)
}

func TestRegistry_UnknownProvider(t *testing.T) {
	flow := oauth.NewFlow(map[string]oauth.Provider{}, oauth.NewMemStateStore())

	_, err := flow.ExchangeRefreshToken(context.Background(), "bogus", "rt")
	require.Error(t, err)

	_, _, err = flow.Start(context.Background(), "bogus", 1)
	require.Error(t, err)
}

func TestRegistry_StartCallbackRoundTrip(t *testing.T) {
	o := &stubProvider{
		cred: &account.AccountCred{
			AccessToken:  "at",
			RefreshToken: "rt",
			ExpiresAt:    time.Now().Add(time.Hour),
		},
		authURL: "https://auth.example/authorize",
	}
	flow := oauth.NewFlow(map[string]oauth.Provider{"openai": o}, oauth.NewMemStateStore())

	authURL, state, err := flow.Start(context.Background(), "openai", 42)
	require.NoError(t, err)
	require.NotEmpty(t, state)
	require.Contains(t, authURL, "state="+state)
	require.Contains(t, authURL, "code_challenge=")

	acc, err := flow.Callback(context.Background(), state, "auth-code")
	require.NoError(t, err)
	require.Equal(t, int64(42), acc.ChannelID)
	require.Equal(t, "openai", acc.Provider)
	require.Equal(t, "oauth", acc.CredType)
	require.Equal(t, "at", acc.Cred.AccessToken)
	require.EqualValues(t, 1, acc.RefreshTokenValid)
	require.NotNil(t, acc.AccessTokenExpiresAt)

	// state 一次性:第二次回调应失败
	_, err = flow.Callback(context.Background(), state, "auth-code")
	require.ErrorIs(t, err, oauth.ErrStateNotFound)
}

func TestRegistry_CallbackUnknownState(t *testing.T) {
	flow := oauth.NewFlow(map[string]oauth.Provider{}, oauth.NewMemStateStore())
	_, err := flow.Callback(context.Background(), "never-saved", "c")
	require.ErrorIs(t, err, oauth.ErrStateNotFound)
}
