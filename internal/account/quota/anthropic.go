package quota

import (
	"net/http"
	"time"

	"github.com/ijry/pro-api/internal/account"
)

// Anthropic 解析 anthropic-ratelimit-* 头。
type Anthropic struct{}

func NewAnthropic() *Anthropic { return &Anthropic{} }

func (a *Anthropic) Extract(h http.Header) *account.QuotaSnapshot {
	now := time.Now().UTC()
	snap := &account.QuotaSnapshot{}
	parseWindow(h, &snap.Quota5h, "anthropic-ratelimit-tokens-limit",
		"anthropic-ratelimit-tokens-remaining",
		"anthropic-ratelimit-tokens-reset")
	parseWindow(h, &snap.QuotaWeek, "anthropic-ratelimit-tokens-week-limit",
		"anthropic-ratelimit-tokens-week-remaining",
		"anthropic-ratelimit-tokens-week-reset")
	if isZero(snap.Quota5h) && isZero(snap.QuotaWeek) {
		return nil
	}
	snap.Quota5h.SyncedAt = &now
	snap.QuotaWeek.SyncedAt = &now
	return snap
}
