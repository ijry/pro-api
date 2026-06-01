package account

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/ijry/pro-api/pkg/apierr"
)

// ClaudeAuthJSON 解析 claude.ai auth.json 格式:
// {"claudeAiOauth":{"access_token":"...","refresh_token":"...","expires_at":"...","scopes":[...],"email":"..."}}
type ClaudeAuthJSON struct{}

func (ClaudeAuthJSON) Format() string { return "claude_authjson" }

func (ClaudeAuthJSON) Match(b []byte) bool {
	if !strings.HasPrefix(strings.TrimSpace(string(b)), "{") {
		return false
	}
	var probe struct {
		OAuth *json.RawMessage `json:"claudeAiOauth"`
	}
	if err := json.Unmarshal(b, &probe); err != nil {
		return false
	}
	return probe.OAuth != nil
}

func (ClaudeAuthJSON) Parse(_ context.Context, b []byte) ([]*Account, error) {
	var raw struct {
		ClaudeAiOauth struct {
			AT        string    `json:"access_token"`
			RT        string    `json:"refresh_token"`
			ExpiresAt time.Time `json:"expires_at"`
			Scopes    []string  `json:"scopes"`
			Email     string    `json:"email"`
		} `json:"claudeAiOauth"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, apierr.New(apierr.CodeAccountImportFields, "claude_authjson: "+err.Error())
	}
	o := raw.ClaudeAiOauth
	a := &Account{
		Provider:     "anthropic",
		CredType:     "oauth",
		Email:        o.Email,
		ImportSource: "import_authjson",
		Status:       StatusActive,
		Weight:       100,
		Cred: AccountCred{
			AccessToken:  o.AT,
			RefreshToken: o.RT,
			ExpiresAt:    o.ExpiresAt,
			Scope:        strings.Join(o.Scopes, " "),
		},
	}
	if !o.ExpiresAt.IsZero() {
		e := o.ExpiresAt
		a.AccessTokenExpiresAt = &e
	}
	return []*Account{a}, nil
}
