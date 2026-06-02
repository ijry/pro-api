package pricing

import (
	"context"
	"fmt"

	"github.com/ijry/pro-api/internal/app"
	"github.com/ijry/pro-api/internal/group"
)

// WirePricing assembles pricing.Service and stores it in app.PricingSvc.
func WirePricing(ctx context.Context, a *app.Application) error {
	if a == nil {
		return fmt.Errorf("WirePricing: app is nil")
	}

	var groupRatio GroupRatioLookup
	if gs, ok := a.GroupSvc.(group.Service); ok {
		groupRatio = func(ctx context.Context, id int64) float64 {
			r, _ := gs.RatioFor(ctx, id)
			return r
		}
	}

	svc, err := New(ctx, Config{
		DB:         a.DB,
		Cache:      a.Cache,
		Log:        a.Log,
		Clock:      a.Clock,
		IDGen:      a.IDGen,
		Audit:      a.Audit,
		GroupRatio: groupRatio,
	})
	if err != nil {
		return fmt.Errorf("WirePricing: %w", err)
	}
	a.PricingSvc = svc
	a.AddCloser("pricing", svc.Close)
	return nil
}

// PricingFrom retrieves the Pricing interface from app.PricingSvc. Returns nil if not wired.
func PricingFrom(a *app.Application) Pricing {
	if a == nil {
		return nil
	}
	p, _ := a.PricingSvc.(Pricing)
	return p
}
