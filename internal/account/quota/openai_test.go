package quota_test

import (
	"net/http"
	"testing"
	"time"

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
	require.WithinDuration(t, time.Now().UTC().Add(time.Hour), *snap.Quota5h.ResetAt, 2*time.Second)
	require.Equal(t, int64(1500000), *snap.QuotaWeek.Total)
	require.Equal(t, int64(980000), *snap.QuotaWeek.Remaining)
}

func TestOpenAI_EmptyReturnsNil(t *testing.T) {
	p := quota.NewOpenAI()
	require.Nil(t, p.Extract(http.Header{}))
}

func TestOpenAI_DurationParsing(t *testing.T) {
	p := quota.NewOpenAI()
	cases := []struct {
		name, hdr string
		wantNil   bool
	}{
		{"hours", "2h", false},
		{"minutes", "30m", false},
		{"seconds", "45s", false},
		{"milliseconds", "500ms", false},
		{"days", "1d", false},
		{"uppercase rejected", "1H", true},
		{"garbage rejected", "soon", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := http.Header{}
			h.Set("x-ratelimit-reset-tokens", c.hdr)
			snap := p.Extract(h)
			if c.wantNil {
				require.True(t, snap == nil || snap.Quota5h.ResetAt == nil,
					"expected ResetAt to be unset for %q, got %+v", c.hdr, snap)
			} else {
				require.NotNil(t, snap)
				require.NotNil(t, snap.Quota5h.ResetAt, "expected ResetAt parsed for %q", c.hdr)
			}
		})
	}
}
