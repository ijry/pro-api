package oauth

import (
	"context"
	"fmt"
	"time"

	"github.com/ijry/pro-api/internal/account"
)

// stateTTL 是 PKCE state 的有效期(授权跳转 → 回调的窗口),对应路线图规约。
const stateTTL = 10 * time.Minute

type registry struct {
	providers map[string]Provider
	store     StateStore
}

// NewFlow 把多个 per-provider 客户端 + state 存储包成 account.OAuthFlow。
func NewFlow(providers map[string]Provider, store StateStore) account.OAuthFlow {
	return &registry{providers: providers, store: store}
}

// Start 生成 PKCE verifier/challenge 与 state,落库 state,返回授权跳转 URL。
func (r *registry) Start(ctx context.Context, provider string, channelID int64) (string, string, error) {
	p, ok := r.providers[provider]
	if !ok {
		return "", "", fmt.Errorf("oauth: unknown provider %q", provider)
	}
	verifier, err := newVerifier()
	if err != nil {
		return "", "", err
	}
	state, err := newState()
	if err != nil {
		return "", "", err
	}
	if err := r.store.Save(ctx, state, StateData{
		Provider:  provider,
		ChannelID: channelID,
		Verifier:  verifier,
	}, stateTTL); err != nil {
		return "", "", err
	}
	return p.AuthCodeURL(state, challengeS256(verifier)), state, nil
}

// Callback 校验/消费 state,用授权码换凭证,组装出待入库的 Account。
func (r *registry) Callback(ctx context.Context, state, code string) (*account.Account, error) {
	d, err := r.store.Take(ctx, state)
	if err != nil {
		return nil, err
	}
	p, ok := r.providers[d.Provider]
	if !ok {
		return nil, fmt.Errorf("oauth: unknown provider %q", d.Provider)
	}
	cred, err := p.ExchangeCode(ctx, code, d.Verifier)
	if err != nil {
		return nil, err
	}
	acc := &account.Account{
		ChannelID:    d.ChannelID,
		Provider:     d.Provider,
		CredType:     "oauth",
		ImportSource: "oauth",
		Status:       account.StatusActive,
		Cred:         *cred,
	}
	if cred.RefreshToken != "" {
		acc.RefreshTokenValid = 1
	}
	if !cred.ExpiresAt.IsZero() {
		exp := cred.ExpiresAt
		acc.AccessTokenExpiresAt = &exp
	}
	return acc, nil
}

func (r *registry) ExchangeRefreshToken(ctx context.Context, provider, rt string) (*account.AccountCred, error) {
	p, ok := r.providers[provider]
	if !ok {
		return nil, fmt.Errorf("oauth: unknown provider %q", provider)
	}
	return p.ExchangeRefreshToken(ctx, rt)
}
