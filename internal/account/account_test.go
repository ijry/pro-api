package account_test

import (
	"strings"
	"testing"

	"github.com/ijry/pro-api/internal/account"
	"github.com/stretchr/testify/require"
)

func TestAccount_DedupKey(t *testing.T) {
	require.Equal(t, "ext:ac_1", (&account.Account{ExternalAccountID: "ac_1"}).DedupKey())
	require.Equal(t, "email:u@example.com",
		(&account.Account{Email: "U@Example.com"}).DedupKey())
	require.True(t, strings.HasPrefix(
		(&account.Account{Cred: account.AccountCred{AccessToken: "tok"}}).DedupKey(),
		"atsha:"))
	require.Equal(t, "", (&account.Account{}).DedupKey()) // empty fallback
}
