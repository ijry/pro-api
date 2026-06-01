package account_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/ijry/pro-api/internal/account"
	"github.com/stretchr/testify/require"
)

type stubProviderProbe struct{ h http.Header }

func (s stubProviderProbe) Probe(context.Context, account.AccountCred) (http.Header, error) {
	return s.h, nil
}

type stubQuotaTracker struct {
	extractCalled bool
	updateCalled  bool
}

func (s *stubQuotaTracker) ExtractFromResponse(_ string, _ http.Header) *account.QuotaSnapshot {
	s.extractCalled = true
	return &account.QuotaSnapshot{}
}

func (s *stubQuotaTracker) UpdateAccount(context.Context, int64, *account.QuotaSnapshot) error {
	s.updateCalled = true
	return nil
}

func TestProbe_RunCallsTracker(t *testing.T) {
	repo := newFakeRepo()
	tk := &stubQuotaTracker{}
	h := http.Header{"X": []string{"1"}}
	p := account.NewProbe(repo, tk, map[string]account.ProviderProbe{
		"anthropic": stubProviderProbe{h: h},
	})
	a := &account.Account{Provider: "anthropic"}
	require.NoError(t, repo.Create(context.Background(), a))
	require.NoError(t, p.Run(context.Background(), a))
	require.True(t, tk.extractCalled)
	require.True(t, tk.updateCalled)

	// event "probed" appended on success
	require.Len(t, repo.events, 1)
	require.Equal(t, "probed", repo.events[0].eventType)
}

func TestProbe_UnknownProviderReturnsErr(t *testing.T) {
	repo := newFakeRepo()
	p := account.NewProbe(repo, &stubQuotaTracker{}, map[string]account.ProviderProbe{})
	a := &account.Account{Provider: "bogus"}
	require.NoError(t, repo.Create(context.Background(), a))
	err := p.Run(context.Background(), a)
	require.Error(t, err)
}
