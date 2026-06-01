package oauth

import (
	"context"
	"errors"

	"github.com/ijry/pro-api/internal/account"
)

// ErrNotImplemented 是 M1 阶段 Start / Callback 的占位错误;M2 将补全 PKCE 流程。
var ErrNotImplemented = errors.New("oauth: direct authorization not implemented in M1 (planned for M2)")

// Provider 是单 provider 的 OAuth 客户端。Start/Callback 是 M1 stub;
// ExchangeRefreshToken 是 M1 真实实现,用于 RawRefreshToken parser 与后台 Refresher。
type Provider interface {
	Start(ctx context.Context, channelID int64) (authURL, state string, err error)
	Callback(ctx context.Context, state, code string) (*account.Account, error)
	ExchangeRefreshToken(ctx context.Context, refreshToken string) (*account.AccountCred, error)
}

// Config 是单 provider 的 OAuth 配置。
type Config struct {
	TokenURL string
	AuthURL  string
	ClientID string
	Scopes   []string
}
