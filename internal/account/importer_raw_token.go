package account

import (
	"context"
	"strings"

	"github.com/ijry/pro-api/pkg/apierr"
)

// RawAccessToken 解析纯文本 access_token。
// 匹配前缀: sk-ant-oat01- (Anthropic) 或 eyJ (JWT).
type RawAccessToken struct{}

func (RawAccessToken) Format() string { return "raw_access_token" }

func (RawAccessToken) Match(b []byte) bool {
	s := strings.TrimSpace(string(b))
	if strings.HasPrefix(s, "{") || strings.HasPrefix(s, "[") {
		return false
	}
	return strings.HasPrefix(s, "sk-ant-oat01-") || strings.HasPrefix(s, "eyJ")
}

func (RawAccessToken) Parse(_ context.Context, b []byte) ([]*Account, error) {
	s := strings.TrimSpace(string(b))
	if s == "" {
		return nil, apierr.New(apierr.CodeAccountImportFields, "raw_access_token: empty token")
	}
	a := &Account{
		Provider:     "anthropic",
		CredType:     "token_pasted",
		ImportSource: "paste_tokens",
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
		ImportSource: "paste_tokens",
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

// Parse 按行拆分:每行一个 API key,逐行 trim 并跳过空行,每行产出一个账号。
// 单行输入即 len==1 的特例。provider 按前缀推断(anthropic / openai);
// 中转站场景下调用方可用渠道 provider 覆盖(见 handler)。
func (RawAPIKey) Parse(_ context.Context, b []byte) ([]*Account, error) {
	lines := strings.Split(string(b), "\n")
	accounts := make([]*Account, 0, len(lines))
	for _, line := range lines {
		s := strings.TrimSpace(line)
		if s == "" {
			continue
		}
		provider := "anthropic"
		if strings.HasPrefix(s, "sk-proj-") || (strings.HasPrefix(s, "sk-") && !strings.HasPrefix(s, "sk-ant-")) {
			provider = "openai"
		}
		accounts = append(accounts, &Account{
			Provider:     provider,
			CredType:     "apikey",
			ImportSource: "paste_apikey",
			Status:       StatusActive,
			Weight:       100,
			Cred: AccountCred{
				APIKey: s,
			},
		})
	}
	if len(accounts) == 0 {
		return nil, apierr.New(apierr.CodeAccountImportFields, "raw_apikey: empty key")
	}
	return accounts, nil
}
