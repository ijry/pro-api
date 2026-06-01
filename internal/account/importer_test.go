package account_test

import (
	"os"
	"strings"
	"testing"

	"github.com/ijry/pro-api/internal/account"
	"github.com/stretchr/testify/require"
)

func loadFixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile("importer/testdata/" + name)
	require.NoError(t, err)
	return b
}

func TestImporter_Detect(t *testing.T) {
	imp := account.NewDefaultImporter(nil) // OAuthFlow not exercised by Detect

	cases := []struct {
		file   string
		format string
	}{
		{"oauth_session.json", "oauth_session"},
		{"claude_authjson.json", "claude_authjson"},
		{"codex_authjson.json", "codex_authjson"},
		{"sub2api.json", "sub2api_export"},
		{"cliproxy.json", "cliproxy_export"},
		{"auths_batch.json", "auths_json_batch"},
		{"raw_at.txt", "raw_access_token"},
		{"raw_rt.txt", "raw_refresh_token"},
		{"raw_ak.txt", "raw_apikey"},
	}
	for _, c := range cases {
		t.Run(c.format, func(t *testing.T) {
			data := loadFixture(t, c.file)
			f, ok := imp.Detect(data)
			require.True(t, ok, "should detect %s", c.format)
			require.Equal(t, c.format, f)
		})
	}
}

func TestImporter_ParseSub2API(t *testing.T) {
	imp := account.NewDefaultImporter(nil)
	data := loadFixture(t, "sub2api.json")
	list, err := imp.Parse(data, "sub2api_export")
	require.NoError(t, err)
	require.Len(t, list, 2)
	// sub2api is a JSON array, so ordering is deterministic
	require.Equal(t, "a@example.com", list[0].Email)
}

func TestImporter_ParseRawAPIKey(t *testing.T) {
	imp := account.NewDefaultImporter(nil)
	data := loadFixture(t, "raw_ak.txt")
	list, err := imp.Parse(data, "raw_apikey")
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.Equal(t, "apikey", list[0].CredType)
	require.Equal(t, "sk-ant-api03-XXXXXXXXXX", list[0].Cred.APIKey)
}

func TestImporter_RawAccessToken_JWTPrefix(t *testing.T) {
	imp := account.NewDefaultImporter(nil)
	payload := []byte("eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJhYmMifQ.sig")
	format, ok := imp.Detect(payload)
	require.True(t, ok)
	require.Equal(t, "raw_access_token", format)

	list, err := imp.Parse(payload, "raw_access_token")
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.Equal(t, "anthropic", list[0].Provider)
	require.Equal(t, "token_pasted", list[0].CredType)
	require.True(t, strings.HasPrefix(list[0].Cred.AccessToken, "eyJ"))
}

func TestImporter_CodexAuthJSON_CredTypeOAuthWhenBothPresent(t *testing.T) {
	imp := account.NewDefaultImporter(nil)
	// Both OPENAI_API_KEY and tokens.refresh_token present → CredType should be "oauth".
	payload := []byte(`{"OPENAI_API_KEY":"sk-proj-abc","tokens":{"id_token":"eyJ","access_token":"at","refresh_token":"rt_OPENAI_yy","account_id":"ac_1"}}`)
	list, err := imp.Parse(payload, "codex_authjson")
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.Equal(t, "oauth", list[0].CredType, "CredType must be 'oauth' when refresh_token present alongside API key")
	require.Equal(t, "sk-proj-abc", list[0].Cred.APIKey)
	require.Equal(t, "rt_OPENAI_yy", list[0].Cred.RefreshToken)
}

func TestImporter_CodexAuthJSON_CredTypeAPIKeyWhenOnlyKey(t *testing.T) {
	imp := account.NewDefaultImporter(nil)
	payload := []byte(`{"OPENAI_API_KEY":"sk-proj-only","tokens":{"id_token":"eyJ","access_token":"","refresh_token":"","account_id":"ac_2"}}`)
	list, err := imp.Parse(payload, "codex_authjson")
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.Equal(t, "apikey", list[0].CredType)
}

func TestImporter_CodexAuthJSON_MatchByIDTokenAndRT(t *testing.T) {
	imp := account.NewDefaultImporter(nil)
	// Empty OPENAI_API_KEY but id_token + rt_-prefixed refresh_token → should still detect as codex_authjson.
	payload := []byte(`{"OPENAI_API_KEY":"","tokens":{"id_token":"eyJabc","access_token":"at","refresh_token":"rt_X","account_id":""}}`)
	format, ok := imp.Detect(payload)
	require.True(t, ok)
	require.Equal(t, "codex_authjson", format)
}
