package account

import (
	"context"
	"strings"

	"github.com/ijry/pro-api/pkg/apierr"
)

// RawAccessToken 解析纯文本 access_token。
// 匹配前缀: sk-ant-oat01- (Anthropic) 或 sess- (OpenAI session).
type RawAccessToken struct{}

func (RawAccessToken) Format() string { return "raw_access_token" }

func (RawAccessToken) Match(b []byte) bool {
	s := strings.TrimSpace(string(b))
	if strings.HasPrefix(s, "{") || strings.HasPrefix(s, "[") {
		return false
	}
	return strings.HasPrefix(s, "sk-ant-oat01-") || strings.HasPrefix(s, "sess-")
}

func (RawAccessToken) Parse(_ context.Context, b []byte) ([]*Account, error) {
	s := strings.TrimSpace(string(b))
	if s == "" {
		return nil, apierr.New(apierr.CodeAccountImportFields, "raw_access_token: empty token")
	}
	provider := "anthropic"
	if strings.HasPrefix(s, "sess-") {
		provider = "openai"
	}
	a := &Account{
		Provider:     provider,
		CredType:     "oauth",
		ImportSource: "paste_raw",
		Status:       StatusActive,
		Weight:       100,
		Cred: AccountCred{
			AccessToken: s,
		},
	}
	return []*Account{a}, nil
}

// RawRefreshToken 解析纯文本 refresh_token,并调用 OAuthFlow 换 access_token。
// 匹配前缀: sk-ant-ort01- (Anthropic) 或 rt_ (OpenAI).
type RawRefreshToken struct {
	OAuth OAuthFlow
}

func (RawRefreshToken) Format() string { return "raw_refresh_token" }

func (RawRefreshToken) Match(b []byte) bool {
	s := strings.TrimSpace(string(b))
	if strings.HasPrefix(s, "{") || strings.HasPrefix(s, "[") {
		return false
	}
	return strings.HasPrefix(s, "sk-ant-ort01-") || strings.HasPrefix(s, "rt_")
}

func (r RawRefreshToken) Parse(ctx context.Context, b []byte) ([]*Account, error) {
	s := strings.TrimSpace(string(b))
	if s == "" {
		return nil, apierr.New(apierr.CodeAccountImportFields, "raw_refresh_token: empty token")
	}
	provider := "anthropic"
	if strings.HasPrefix(s, "rt_") {
		provider = "openai"
	}

	a := &Account{
		Provider:     provider,
		CredType:     "oauth",
		ImportSource: "paste_raw",
		Status:       StatusActive,
		Weight:       100,
		Cred: AccountCred{
			RefreshToken: s,
		},
	}

	// If an OAuthFlow is available, exchange the refresh token for an access token now.
	if r.OAuth != nil {
		cred, err := r.OAuth.ExchangeRefreshToken(ctx, provider, s)
		if err != nil {
			return nil, apierr.Wrap(apierr.CodeAccountRefreshFailed, "raw_refresh_token: exchange failed", err)
		}
		a.Cred = *cred
		a.Cred.RefreshToken = s // preserve original refresh token
		if !cred.ExpiresAt.IsZero() {
			e := cred.ExpiresAt
			a.AccessTokenExpiresAt = &e
		}
	}

	return []*Account{a}, nil
}

// RawAPIKey 解析纯文本 API key。
// 匹配前缀: sk-ant-api (Anthropic) 或 sk-proj- / sk- (OpenAI).
type RawAPIKey struct{}

func (RawAPIKey) Format() string { return "raw_apikey" }

func (RawAPIKey) Match(b []byte) bool {
	s := strings.TrimSpace(string(b))
	if strings.HasPrefix(s, "{") || strings.HasPrefix(s, "[") {
		return false
	}
	// Must not match access_token or refresh_token prefixes
	if strings.HasPrefix(s, "sk-ant-oat01-") || strings.HasPrefix(s, "sk-ant-ort01-") ||
		strings.HasPrefix(s, "sess-") || strings.HasPrefix(s, "rt_") {
		return false
	}
	return strings.HasPrefix(s, "sk-ant-api") || strings.HasPrefix(s, "sk-proj-") ||
		strings.HasPrefix(s, "sk-")
}

func (RawAPIKey) Parse(_ context.Context, b []byte) ([]*Account, error) {
	s := strings.TrimSpace(string(b))
	if s == "" {
		return nil, apierr.New(apierr.CodeAccountImportFields, "raw_apikey: empty key")
	}
	provider := "anthropic"
	if strings.HasPrefix(s, "sk-proj-") || (strings.HasPrefix(s, "sk-") && !strings.HasPrefix(s, "sk-ant-")) {
		provider = "openai"
	}
	a := &Account{
		Provider:     provider,
		CredType:     "apikey",
		ImportSource: "paste_raw",
		Status:       StatusActive,
		Weight:       100,
		Cred: AccountCred{
			APIKey: s,
		},
	}
	return []*Account{a}, nil
}
