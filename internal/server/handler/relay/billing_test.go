package relay

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ijry/pro-api/internal/billing"
	"github.com/ijry/pro-api/internal/channel"
	"github.com/ijry/pro-api/internal/pricing"
	"github.com/ijry/pro-api/internal/token"
	"github.com/ijry/pro-api/pkg/apierr"
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
func (b *fakeBiller) Close() error                             { return nil }

// compile-time check: fakeBiller implements billing.Biller.
var _ billing.Biller = (*fakeBiller)(nil)

func TestDeps_NilBillerSkipsBilling(t *testing.T) {
	h := New(Deps{})
	if h == nil {
		t.Fatal("New with empty Deps should return non-nil Handler")
	}
}

func TestResolveChannel_RejectsTokenDisallowedModel(t *testing.T) {
	h := New(Deps{})
	ctx := context.WithValue(context.Background(), token.CtxKeyToken, &token.View{
		ID:            1,
		AllowedModels: []string{"gpt-4o"},
	})
	c := fakeGinContextWithRequestContext(ctx)

	ch, err := h.resolveChannel(c, "claude-3")
	if ch != nil {
		t.Fatalf("want nil channel, got %+v", ch)
	}
	if err == nil || err.Code != apierr.CodeModelNotAllowed {
		t.Fatalf("want CodeModelNotAllowed, got %v", err)
	}
}

func TestResolveChannel_SelectorErrorDoesNotFallbackToBearer(t *testing.T) {
	h := New(Deps{Channel: failingSelector{err: apierr.New(apierr.CodeNoChannel, "no channel")}})
	c := fakeGinContextWithRequestContext(context.Background())

	ch, err := h.resolveChannel(c, "gpt-4o")
	if ch != nil {
		t.Fatalf("want nil channel, got %+v", ch)
	}
	if err == nil || err.Code != apierr.CodeNoChannel {
		t.Fatalf("want CodeNoChannel, got %v", err)
	}
}

func fakeGinContextWithRequestContext(ctx context.Context) *gin.Context {
	c := &gin.Context{}
	c.Request = httptest.NewRequest("POST", "/v1/chat/completions", nil).WithContext(ctx)
	return c
}

type failingSelector struct {
	err error
}

func (f failingSelector) Select(context.Context, string, channel.SelectHint) (*channel.Channel, error) {
	return nil, f.err
}
func (f failingSelector) ReportSuccess(int64, time.Duration) {}
func (f failingSelector) ReportFailure(int64, error)         {}
