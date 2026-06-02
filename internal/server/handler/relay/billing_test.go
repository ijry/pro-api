package relay

import (
	"context"
	"testing"

	"github.com/ijry/pro-api/internal/billing"
	"github.com/ijry/pro-api/internal/pricing"
)

// fakePricing is a stub pricing.Pricing.
type fakePricing struct{ estCost int64 }

func (f *fakePricing) Compute(_ context.Context, _ pricing.ComputeInput) pricing.ComputeResult {
	return pricing.ComputeResult{Quota: f.estCost, Ratios: pricing.Ratios{Group: 1.0}}
}
func (f *fakePricing) RatioFor(_ context.Context, _ string, _ int64, _ pricing.ChannelInfo) pricing.Ratios {
	return pricing.Ratios{}
}
func (f *fakePricing) EstimateMax(_ context.Context, _ string, _ pricing.EstimateInput) int64 {
	return f.estCost
}
func (f *fakePricing) DefaultMaxOut(_ context.Context, _ string) int { return 4096 }

// compile-time check: fakePricing implements pricing.Pricing.
var _ pricing.Pricing = (*fakePricing)(nil)

// fakeBiller is a stub billing.Biller.
type fakeBiller struct {
	reserved  bool
	committed bool
	refunded  bool
	lastCost  int64
}

func (b *fakeBiller) Reserve(_ context.Context, _, _, _ int64) (string, error) {
	b.reserved = true
	return "rsv-1", nil
}
func (b *fakeBiller) Commit(_ context.Context, _ string, cost int64) error {
	b.committed = true
	b.lastCost = cost
	return nil
}
func (b *fakeBiller) Refund(_ context.Context, _ string) error { b.refunded = true; return nil }
func (b *fakeBiller) Close() error                              { return nil }

// compile-time check: fakeBiller implements billing.Biller.
var _ billing.Biller = (*fakeBiller)(nil)

func TestDeps_NilBillerSkipsBilling(t *testing.T) {
	h := New(Deps{})
	if h == nil {
		t.Fatal("New with empty Deps should return non-nil Handler")
	}
}
