package account

import (
	"context"
	"fmt"
	"net/http"
)

// ProviderProbe 是单 provider 探测器接口;具体实现见 internal/account/probe/。
type ProviderProbe interface {
	Probe(ctx context.Context, cred AccountCred) (http.Header, error)
}

type probeImpl struct {
	repo         Repo
	quotaTracker QuotaTracker
	providers    map[string]ProviderProbe
}

// NewProbe 构造 Probe;providers 由 wire 注入(provider 名 → 实现)。
func NewProbe(r Repo, q QuotaTracker, providers map[string]ProviderProbe) Probe {
	return &probeImpl{repo: r, quotaTracker: q, providers: providers}
}

func (p *probeImpl) Run(ctx context.Context, a *Account) error {
	pp, ok := p.providers[a.Provider]
	if !ok {
		return fmt.Errorf("probe: unknown provider %q", a.Provider)
	}
	h, err := pp.Probe(ctx, a.Cred)
	if err != nil {
		_ = p.repo.AppendEvent(ctx, a.ID, "probed", map[string]any{"err": err.Error()})
		return err
	}
	snap := p.quotaTracker.ExtractFromResponse(a.Provider, h)
	if snap != nil {
		_ = p.quotaTracker.UpdateAccount(ctx, a.ID, snap)
	}
	_ = p.repo.AppendEvent(ctx, a.ID, "probed", snap)
	return nil
}
