package account

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/ijry/pro-api/pkg/apierr"
)

// CLIProxy 解析 cliproxy (cli proxy) 批量导出格式:
// {"accounts":[{"provider":"...","email":"...","credentials":{"access_token":"...","refresh_token":"..."}},...]}
type CLIProxy struct{}

func (CLIProxy) Format() string { return "cliproxy_export" }

func (CLIProxy) Match(b []byte) bool {
	if !strings.HasPrefix(strings.TrimSpace(string(b)), "{") {
		return false
	}
	var probe struct {
		Accounts *json.RawMessage `json:"accounts"`
	}
	if err := json.Unmarshal(b, &probe); err != nil {
		return false
	}
	if probe.Accounts == nil {
		return false
	}
	// Ensure it's a non-empty array with at least one entry having credentials
	var arr []struct {
		Credentials *json.RawMessage `json:"credentials"`
	}
	if err := json.Unmarshal(*probe.Accounts, &arr); err != nil {
		return false
	}
	return len(arr) > 0 && arr[0].Credentials != nil
}

func (CLIProxy) Parse(_ context.Context, b []byte) ([]*Account, error) {
	var raw struct {
		Accounts []struct {
			Provider    string `json:"provider"`
			Email       string `json:"email"`
			Credentials struct {
				AT string `json:"access_token"`
				RT string `json:"refresh_token"`
			} `json:"credentials"`
		} `json:"accounts"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, apierr.New(apierr.CodeAccountImportFields, "cliproxy_export: "+err.Error())
	}
	accounts := make([]*Account, 0, len(raw.Accounts))
	for _, entry := range raw.Accounts {
		provider := entry.Provider
		if provider == "" {
			provider = "anthropic"
			if strings.HasPrefix(entry.Credentials.RT, "rt_") {
				provider = "openai"
			}
		}
		a := &Account{
			Provider:     provider,
			CredType:     "oauth",
			Email:        entry.Email,
			ImportSource: "import_cpa",
			Status:       StatusActive,
			Weight:       100,
			Cred: AccountCred{
				AccessToken:  entry.Credentials.AT,
				RefreshToken: entry.Credentials.RT,
			},
		}
		accounts = append(accounts, a)
	}
	return accounts, nil
}
