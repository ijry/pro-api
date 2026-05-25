package oauth

import "context"

// UserInfo is the normalized user information returned by all OAuth providers.
type UserInfo struct {
	ProviderUID string // stable unique ID on the provider side
	Name        string
	Email       string // may be empty for some providers
	Avatar      string
	Raw         []byte // raw JSON from provider
}

// Provider is the unified interface for all third-party OAuth providers.
type Provider interface {
	// Name returns the provider identifier, e.g. "google"/"wechat"/"feishu"/"dingtalk"/"discord".
	Name() string
	// BuildAuthURL builds the OAuth authorization redirect URL.
	BuildAuthURL(ctx context.Context, state, redirectURL string) (string, error)
	// Exchange exchanges the code for UserInfo and accessToken.
	Exchange(ctx context.Context, code, redirectURL string) (*UserInfo, string, error)
}
