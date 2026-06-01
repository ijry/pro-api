package oauth

import (
	"context"
	"fmt"

	"github.com/ijry/pro-api/internal/account"
)

type registry struct {
	providers map[string]Provider
}

// NewFlow 把多个 per-provider 客户端包成 account.OAuthFlow。
func NewFlow(providers map[string]Provider) account.OAuthFlow {
	return &registry{providers: providers}
}

func (r *registry) Start(ctx context.Context, provider string, channelID int64) (string, string, error) {
	p, ok := r.providers[provider]
	if !ok {
		return "", "", fmt.Errorf("oauth: unknown provider %q", provider)
	}
	return p.Start(ctx, channelID)
}

// Callback 在 M1 阶段不分发(state 中没有 provider 信息;M2 PKCE 流程会补)。
func (r *registry) Callback(_ context.Context, _, _ string) (*account.Account, error) {
	return nil, ErrNotImplemented
}

func (r *registry) ExchangeRefreshToken(ctx context.Context, provider, rt string) (*account.AccountCred, error) {
	p, ok := r.providers[provider]
	if !ok {
		return nil, fmt.Errorf("oauth: unknown provider %q", provider)
	}
	return p.ExchangeRefreshToken(ctx, rt)
}
