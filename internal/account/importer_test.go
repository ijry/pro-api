package account_test

import (
	"os"
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
