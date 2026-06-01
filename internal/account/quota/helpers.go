package quota

import (
	"net/http"
	"strconv"
	"time"

	"github.com/ijry/pro-api/internal/account"
)

// parseWindow fills w from three rate-limit headers: total, remaining, and an
// RFC3339 reset timestamp. Absent or unparseable headers leave the field nil.
func parseWindow(h http.Header, w *account.QuotaWindow, totalKey, remKey, resetKey string) {
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

// isZero reports whether w has no populated fields.
func isZero(w account.QuotaWindow) bool {
	return w.Total == nil && w.Remaining == nil && w.ResetAt == nil
}
