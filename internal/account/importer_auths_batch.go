package account

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/ijry/pro-api/pkg/apierr"
)

// AuthsBatch 解析 auths-batch JSON 格式:
// {"email@example.com":{"tokens":{"access_token":"...","refresh_token":"..."}},...}
// 键是邮件地址,值是 session 对象。
type AuthsBatch struct{}

func (AuthsBatch) Format() string { return "auths_json_batch" }

func (AuthsBatch) Match(b []byte) bool {
	if !strings.HasPrefix(strings.TrimSpace(string(b)), "{") {
		return false
	}
	// Must be a map where each value has tokens.access_token
	var m map[string]struct {
		Tokens *struct {
			AT string `json:"access_token"`
		} `json:"tokens"`
	}
	if err := json.Unmarshal(b, &m); err != nil {
		return false
	}
	if len(m) == 0 {
		return false
	}
	// Check that keys look like email addresses and values have tokens
	for k, v := range m {
		if !strings.Contains(k, "@") {
			return false
		}
		if v.Tokens == nil || v.Tokens.AT == "" {
			return false
		}
	}
	return true
}

func (AuthsBatch) Parse(_ context.Context, b []byte) ([]*Account, error) {
	var m map[string]struct {
		Tokens struct {
			AT string `json:"access_token"`
			RT string `json:"refresh_token"`
		} `json:"tokens"`
	}
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, apierr.New(apierr.CodeAccountImportFields, "auths_json_batch: "+err.Error())
	}
	accounts := make([]*Account, 0, len(m))
	for email, entry := range m {
		provider := "anthropic"
		if strings.HasPrefix(entry.Tokens.RT, "rt_") {
			provider = "openai"
		}
		a := &Account{
			Provider:     provider,
			CredType:     "oauth",
			Email:        email,
			ImportSource: "import_authjson",
			Status:       StatusActive,
			Weight:       100,
			Cred: AccountCred{
				AccessToken:  entry.Tokens.AT,
				RefreshToken: entry.Tokens.RT,
			},
		}
		accounts = append(accounts, a)
	}
	return accounts, nil
}
