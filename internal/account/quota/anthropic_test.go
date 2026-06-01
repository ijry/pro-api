package quota_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/ijry/pro-api/internal/account/quota"
	"github.com/stretchr/testify/require"
)

func TestAnthropic_ExtractFromResponse(t *testing.T) {
	p := quota.NewAnthropic()
	h := http.Header{}
	h.Set("anthropic-ratelimit-tokens-limit", "12000")
	h.Set("anthropic-ratelimit-tokens-remaining", "8200")
	h.Set("anthropic-ratelimit-tokens-reset", "2026-05-27T18:00:00Z")
	h.Set("anthropic-ratelimit-tokens-week-limit", "1500000")
	h.Set("anthropic-ratelimit-tokens-week-remaining", "980000")
	h.Set("anthropic-ratelimit-tokens-week-reset", "2026-06-02T00:00:00Z")

	snap := p.Extract(h)
	require.NotNil(t, snap)
	require.NotNil(t, snap.Quota5h.Total)
	require.Equal(t, int64(12000), *snap.Quota5h.Total)
	require.Equal(t, int64(8200), *snap.Quota5h.Remaining)
	require.Equal(t, int64(1500000), *snap.QuotaWeek.Total)
	require.WithinDuration(t,
		time.Date(2026, 5, 27, 18, 0, 0, 0, time.UTC),
		*snap.Quota5h.ResetAt, time.Second)
}

func TestAnthropic_PartialHeadersKeepsExisting(t *testing.T) {
	p := quota.NewAnthropic()
	h := http.Header{}
	h.Set("anthropic-ratelimit-tokens-remaining", "5000")
	snap := p.Extract(h)
	require.NotNil(t, snap)
	require.NotNil(t, snap.Quota5h.Remaining)
	require.Nil(t, snap.Quota5h.Total, "total absent → field stays nil")
}
