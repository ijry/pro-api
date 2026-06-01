package oauth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ijry/pro-api/internal/account"
	"github.com/ijry/pro-api/internal/account/oauth"
	"github.com/stretchr/testify/require"
)

type stubProvider struct {
	cred *account.AccountCred
	err  error
}

func (s *stubProvider) Start(_ context.Context, _ int64) (string, string, error) {
	return "", "", oauth.ErrNotImplemented
}
func (s *stubProvider) Callback(_ context.Context, _, _ string) (*account.Account, error) {
	return nil, oauth.ErrNotImplemented
}
func (s *stubProvider) ExchangeRefreshToken(_ context.Context, _ string) (*account.AccountCred, error) {
	return s.cred, s.err
}

func TestRegistry_DispatchesByProvider(t *testing.T) {
	a := &stubProvider{cred: &account.AccountCred{AccessToken: "a-at"}}
	o := &stubProvider{cred: &account.AccountCred{AccessToken: "o-at"}}
	flow := oauth.NewFlow(map[string]oauth.Provider{"anthropic": a, "openai": o})

	got, err := flow.ExchangeRefreshToken(context.Background(), "anthropic", "rt")
	require.NoError(t, err)
	require.Equal(t, "a-at", got.AccessToken)

	got, err = flow.ExchangeRefreshToken(context.Background(), "openai", "rt")
	require.NoError(t, err)
	require.Equal(t, "o-at", got.AccessToken)
}

func TestRegistry_UnknownProvider(t *testing.T) {
	flow := oauth.NewFlow(map[string]oauth.Provider{})
	_, err := flow.ExchangeRefreshToken(context.Background(), "bogus", "rt")
	require.Error(t, err)

	_, _, err = flow.Start(context.Background(), "bogus", 1)
	require.Error(t, err)
}

func TestRegistry_CallbackStub(t *testing.T) {
	flow := oauth.NewFlow(map[string]oauth.Provider{})
	_, err := flow.Callback(context.Background(), "s", "c")
	require.True(t, errors.Is(err, oauth.ErrNotImplemented))
}
