package account

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/ijry/pro-api/pkg/apierr"
)

// OAuthSession 解析 oauth_session 格式:{"tokens":{...},"email":"..."}
type OAuthSession struct{}

func (OAuthSession) Format() string { return "oauth_session" }

func (OAuthSession) Match(b []byte) bool {
	if !strings.HasPrefix(strings.TrimSpace(string(b)), "{") {
		return false
	}
	var probe struct {
		OpenAIAPIKey *json.RawMessage `json:"OPENAI_API_KEY"` // present in codex_authjson; exclude
		Tokens       struct {
			AT string `json:"access_token"`
			RT string `json:"refresh_token"`
		} `json:"tokens"`
	}
	if err := json.Unmarshal(b, &probe); err != nil {
		return false
	}
	if probe.OpenAIAPIKey != nil {
		return false // this is codex_authjson, not oauth_session
	}
	return probe.Tokens.AT != "" && probe.Tokens.RT != ""
}

func (OAuthSession) Parse(_ context.Context, b []byte) ([]*Account, error) {
	var raw struct {
		Email  string `json:"email"`
		Tokens struct {
			AT        string    `json:"access_token"`
			RT        string    `json:"refresh_token"`
			IDToken   string    `json:"id_token"`
			ExpiresAt time.Time `json:"expires_at"`
			TokenType string    `json:"token_type"`
			Scope     string    `json:"scope"`
		} `json:"tokens"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, apierr.New(apierr.CodeAccountImportFields, "oauth_session: "+err.Error())
	}
	provider := "anthropic"
	if strings.HasPrefix(raw.Tokens.RT, "rt_") {
		provider = "openai"
	}
	a := &Account{
		Provider:     provider,
		CredType:     "oauth",
		Email:        raw.Email,
		ImportSource: "paste_session",
		Status:       StatusActive,
		Weight:       100,
		Cred: AccountCred{
			AccessToken:  raw.Tokens.AT,
			RefreshToken: raw.Tokens.RT,
			IDToken:      raw.Tokens.IDToken,
			ExpiresAt:    raw.Tokens.ExpiresAt,
			TokenType:    raw.Tokens.TokenType,
			Scope:        raw.Tokens.Scope,
		},
	}
	if !raw.Tokens.ExpiresAt.IsZero() {
		e := raw.Tokens.ExpiresAt
		a.AccessTokenExpiresAt = &e
	}
	return []*Account{a}, nil
}
