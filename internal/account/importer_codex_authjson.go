package account

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/ijry/pro-api/pkg/apierr"
)

// CodexAuthJSON 解析 openai codex auth.json 格式:
// {"OPENAI_API_KEY":"...","tokens":{"id_token":"...","access_token":"...","refresh_token":"rt_...","account_id":"..."}}
type CodexAuthJSON struct{}

func (CodexAuthJSON) Format() string { return "codex_authjson" }

func (CodexAuthJSON) Match(b []byte) bool {
	if !strings.HasPrefix(strings.TrimSpace(string(b)), "{") {
		return false
	}
	var probe struct {
		OpenAIAPIKey string `json:"OPENAI_API_KEY"`
		Tokens       struct {
			IDToken string `json:"id_token"`
			RT      string `json:"refresh_token"`
		} `json:"tokens"`
	}
	if err := json.Unmarshal(b, &probe); err != nil {
		return false
	}
	if probe.OpenAIAPIKey != "" {
		return true
	}
	return probe.Tokens.IDToken != "" && strings.HasPrefix(probe.Tokens.RT, "rt_")
}

func (CodexAuthJSON) Parse(_ context.Context, b []byte) ([]*Account, error) {
	var raw struct {
		OpenAIAPIKey string `json:"OPENAI_API_KEY"`
		Tokens       struct {
			IDToken   string `json:"id_token"`
			AT        string `json:"access_token"`
			RT        string `json:"refresh_token"`
			AccountID string `json:"account_id"`
		} `json:"tokens"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, apierr.New(apierr.CodeAccountImportFields, "codex_authjson: "+err.Error())
	}

	credType := "oauth"
	cred := AccountCred{
		IDToken:      raw.Tokens.IDToken,
		AccessToken:  raw.Tokens.AT,
		RefreshToken: raw.Tokens.RT,
	}
	if raw.OpenAIAPIKey != "" {
		cred.APIKey = raw.OpenAIAPIKey
		if raw.Tokens.RT == "" {
			credType = "apikey"
		}
	}

	a := &Account{
		Provider:          "openai",
		CredType:          credType,
		ExternalAccountID: raw.Tokens.AccountID,
		ImportSource:      "import_authjson",
		Status:            StatusActive,
		Weight:            100,
		Cred:              cred,
	}
	return []*Account{a}, nil
}
