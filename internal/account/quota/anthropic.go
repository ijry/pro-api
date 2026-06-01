package quota

import (
	"net/http"
	"strconv"
	"time"

	"github.com/ijry/pro-api/internal/account"
)

// Anthropic 解析 anthropic-ratelimit-* 头。
type Anthropic struct{}

func NewAnthropic() *Anthropic { return &Anthropic{} }

func (a *Anthropic) Extract(h http.Header) *account.QuotaSnapshot {
	snap := &account.QuotaSnapshot{}
	parse5h(h, &snap.Quota5h, "anthropic-ratelimit-tokens-limit",
		"anthropic-ratelimit-tokens-remaining",
		"anthropic-ratelimit-tokens-reset")
	parse5h(h, &snap.QuotaWeek, "anthropic-ratelimit-tokens-week-limit",
		"anthropic-ratelimit-tokens-week-remaining",
		"anthropic-ratelimit-tokens-week-reset")
	if isZero(snap.Quota5h) && isZero(snap.QuotaWeek) {
		return nil
	}
	now := time.Now().UTC()
	snap.Quota5h.SyncedAt = &now
	snap.QuotaWeek.SyncedAt = &now
	return snap
}

func parse5h(h http.Header, w *account.QuotaWindow, totalKey, remKey, resetKey string) {
	if v := h.Get(totalKey); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			w.Total = &n
		}
	}
	if v := h.Get(remKey); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			w.Remaining = &n
		}
	}
	if v := h.Get(resetKey); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			w.ResetAt = &t
		}
	}
}

func isZero(w account.QuotaWindow) bool {
	return w.Total == nil && w.Remaining == nil && w.ResetAt == nil
}
