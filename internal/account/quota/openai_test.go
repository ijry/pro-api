package quota_test

import (
	"net/http"
	"testing"

	"github.com/ijry/pro-api/internal/account/quota"
	"github.com/stretchr/testify/require"
)

func TestOpenAI_ExtractFromResponse(t *testing.T) {
	p := quota.NewOpenAI()
	h := http.Header{}
	h.Set("x-ratelimit-limit-tokens", "12000")
	h.Set("x-ratelimit-remaining-tokens", "8200")
	h.Set("x-ratelimit-reset-tokens", "1h")
	h.Set("x-codex-week-limit", "1500000")
	h.Set("x-codex-week-remaining", "980000")

	snap := p.Extract(h)
	require.NotNil(t, snap)
	require.Equal(t, int64(12000), *snap.Quota5h.Total)
	require.Equal(t, int64(8200), *snap.Quota5h.Remaining)
	require.NotNil(t, snap.Quota5h.ResetAt)
}

func TestOpenAI_EmptyReturnsNil(t *testing.T) {
	p := quota.NewOpenAI()
	require.Nil(t, p.Extract(http.Header{}))
}
