package account

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/ijry/pro-api/pkg/apierr"
)

// Sub2API 解析 sub2api 批量导出格式:
// [{"email":"...","tokens":{"access_token":"...","refresh_token":"..."}},...}]
type Sub2API struct{}

func (Sub2API) Format() string { return "sub2api_export" }

func (Sub2API) Match(b []byte) bool {
	s := strings.TrimSpace(string(b))
	if !strings.HasPrefix(s, "[") {
		return false
	}
	// Probe: must be array of objects with tokens.access_token
	var probe []struct {
		Tokens *struct {
			AT string `json:"access_token"`
		} `json:"tokens"`
	}
	if err := json.Unmarshal(b, &probe); err != nil {
		return false
	}
	return len(probe) > 0 && probe[0].Tokens != nil && probe[0].Tokens.AT != ""
}

func (Sub2API) Parse(_ context.Context, b []byte) ([]*Account, error) {
	var rows []struct {
		Email  string `json:"email"`
		Tokens struct {
			AT string `json:"access_token"`
			RT string `json:"refresh_token"`
		} `json:"tokens"`
	}
	if err := json.Unmarshal(b, &rows); err != nil {
		return nil, apierr.New(apierr.CodeAccountImportFields, "sub2api_export: "+err.Error())
	}
	accounts := make([]*Account, 0, len(rows))
	for _, row := range rows {
		provider := "anthropic"
		if strings.HasPrefix(row.Tokens.RT, "rt_") {
			provider = "openai"
		}
		a := &Account{
			Provider:     provider,
			CredType:     "oauth",
			Email:        row.Email,
			ImportSource: "import_sub2api",
			Status:       StatusActive,
			Weight:       100,
			Cred: AccountCred{
				AccessToken:  row.Tokens.AT,
				RefreshToken: row.Tokens.RT,
			},
		}
		accounts = append(accounts, a)
	}
	return accounts, nil
}
