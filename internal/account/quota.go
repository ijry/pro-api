package account

import (
	"context"
	"net/http"
)

// ProviderExtractor 是各 provider 实现的接口(由 quota/ 子包提供)。
type ProviderExtractor interface {
	Extract(h http.Header) *QuotaSnapshot
}

// quotaTracker 默认 QuotaTracker 实现,持有 provider 注册表。
type quotaTracker struct {
	repo      Repo
	providers map[string]ProviderExtractor
}

// NewQuotaTracker 构造 QuotaTracker。providers 由 wire 注入。
func NewQuotaTracker(r Repo, providers map[string]ProviderExtractor) QuotaTracker {
	return &quotaTracker{repo: r, providers: providers}
}

func (t *quotaTracker) ExtractFromResponse(provider string, h http.Header) *QuotaSnapshot {
	p, ok := t.providers[provider]
	if !ok {
		return nil
	}
	return p.Extract(h)
}

func (t *quotaTracker) UpdateAccount(ctx context.Context, accountID int64, snap *QuotaSnapshot) error {
	if snap == nil || t.repo == nil {
		return nil
	}
	a, err := t.repo.Get(ctx, accountID)
	if err != nil {
		return err
	}
	if snap.Quota5h.Total != nil {
		a.Quota5hTotal = snap.Quota5h.Total
	}
	if snap.Quota5h.Remaining != nil {
		a.Quota5hRemaining = snap.Quota5h.Remaining
	}
	if snap.Quota5h.ResetAt != nil {
		a.Quota5hResetAt = snap.Quota5h.ResetAt
	}
	if snap.QuotaWeek.Total != nil {
		a.QuotaWeekTotal = snap.QuotaWeek.Total
	}
	if snap.QuotaWeek.Remaining != nil {
		a.QuotaWeekRemaining = snap.QuotaWeek.Remaining
	}
	if snap.QuotaWeek.ResetAt != nil {
		a.QuotaWeekResetAt = snap.QuotaWeek.ResetAt
	}
	if snap.Quota5h.SyncedAt != nil {
		a.QuotaSyncedAt = snap.Quota5h.SyncedAt
	}
	return t.repo.Update(ctx, a)
}
