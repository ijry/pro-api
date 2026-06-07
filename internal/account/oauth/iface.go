package oauth

import (
	"context"

	"github.com/ijry/pro-api/internal/account"
)

// Provider 是单 provider 的 OAuth 客户端。
//
// PKCE 授权码流程的编排(生成 verifier/state、保存 state、分发 Callback)由 Flow
// 统一负责;Provider 只负责拼授权 URL 与用授权码/刷新令牌换取凭证。
type Provider interface {
	// AuthCodeURL 拼出 OAuth2 授权码 + PKCE 的授权跳转 URL。
	AuthCodeURL(state, challenge string) string
	// ExchangeCode 用授权码 + code_verifier 换取凭证(grant_type=authorization_code)。
	ExchangeCode(ctx context.Context, code, verifier string) (*account.AccountCred, error)
	// ExchangeRefreshToken 用 refresh_token 换取新凭证(grant_type=refresh_token)。
	ExchangeRefreshToken(ctx context.Context, refreshToken string) (*account.AccountCred, error)
}

// Config 是单 provider 的 OAuth 配置。
type Config struct {
	TokenURL    string
	AuthURL     string
	ClientID    string
	RedirectURI string
	Scopes      []string
}
