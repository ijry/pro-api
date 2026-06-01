package quota

import (
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/ijry/pro-api/internal/account"
)

// OpenAI parses x-ratelimit-* and x-codex-week-* headers from OpenAI / Codex API responses.
type OpenAI struct{}

func NewOpenAI() *OpenAI { return &OpenAI{} }

func (o *OpenAI) Extract(h http.Header) *account.QuotaSnapshot {
	now := time.Now().UTC()
	snap := &account.QuotaSnapshot{}
	if v := h.Get("x-ratelimit-limit-tokens"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			snap.Quota5h.Total = &n
		}
	}
	if v := h.Get("x-ratelimit-remaining-tokens"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			snap.Quota5h.Remaining = &n
		}
	}
	if v := h.Get("x-ratelimit-reset-tokens"); v != "" {
		if d, ok := parseDuration(v); ok {
			t := now.Add(d)
			snap.Quota5h.ResetAt = &t
		}
	}
	if v := h.Get("x-codex-week-limit"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			snap.QuotaWeek.Total = &n
		}
	}
	if v := h.Get("x-codex-week-remaining"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			snap.QuotaWeek.Remaining = &n
		}
	}
	if isZero(snap.Quota5h) && isZero(snap.QuotaWeek) {
		return nil
	}
	snap.Quota5h.SyncedAt = &now
	snap.QuotaWeek.SyncedAt = &now
	return snap
}

var reDur = regexp.MustCompile(`^(\d+)(ms|s|m|h|d)$`)

func parseDuration(s string) (time.Duration, bool) {
	m := reDur.FindStringSubmatch(s)
	if m == nil {
		if d, err := time.ParseDuration(s); err == nil {
			return d, true
		}
		return 0, false
	}
	n, _ := strconv.ParseInt(m[1], 10, 64)
	switch m[2] {
	case "ms":
		return time.Duration(n) * time.Millisecond, true
	case "s":
		return time.Duration(n) * time.Second, true
	case "m":
		return time.Duration(n) * time.Minute, true
	case "h":
		return time.Duration(n) * time.Hour, true
	case "d":
		return time.Duration(n) * 24 * time.Hour, true
	}
	return 0, false
}
