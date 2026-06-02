package account_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/ijry/pro-api/internal/account"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

type stubOAuth struct {
	call func(provider, rt string) (*account.AccountCred, error)
}

func (s stubOAuth) Start(context.Context, string, int64) (string, string, error) {
	return "", "", nil
}
func (s stubOAuth) Callback(context.Context, string, string) (*account.Account, error) {
	return nil, nil
}
func (s stubOAuth) ExchangeRefreshToken(_ context.Context, p, rt string) (*account.AccountCred, error) {
	return s.call(p, rt)
}

func TestRefresher_RefreshOne_UpdatesCred(t *testing.T) {
	repo := newFakeRepo()
	ctx := context.Background()
	exp := time.Now().Add(time.Minute)
	a := &account.Account{
		ChannelID: 1, Provider: "anthropic", CredType: "oauth",
		Status: account.StatusActive, Weight: 100,
		Cred:                 account.AccountCred{AccessToken: "old", RefreshToken: "rt-1", ExpiresAt: exp},
		AccessTokenExpiresAt: &exp,
	}
	_ = repo.Create(ctx, a)

	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	o := stubOAuth{call: func(p, rt string) (*account.AccountCred, error) {
		require.Equal(t, "anthropic", p)
		require.Equal(t, "rt-1", rt)
		return &account.AccountCred{
			AccessToken: "new", RefreshToken: "rt-2",
			ExpiresAt: time.Now().Add(time.Hour),
		}, nil
	}}
	r := account.NewRefresher(repo, rdb, o, nil)
	require.NoError(t, r.RefreshOne(ctx, a.ID))
	got, _ := repo.Get(ctx, a.ID)
	require.Equal(t, "new", got.Cred.AccessToken)
	require.Equal(t, int8(1), got.RefreshTokenValid)
	require.NotNil(t, got.LastRefreshedAt)
	require.NotNil(t, got.AccessTokenExpiresAt)
}

func TestRefresher_DistributedLock(t *testing.T) {
	repo := newFakeRepo()
	ctx := context.Background()
	a := &account.Account{
		ChannelID: 1, Provider: "anthropic", CredType: "oauth",
		Status: account.StatusActive,
		Cred:   account.AccountCred{RefreshToken: "rt-x"},
	}
	_ = repo.Create(ctx, a)

	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	// 模拟 lock 已被持有:手动 SET 同一个 key,SetNX 应失败
	mr.Set("account:lock:refresh:1", "held")

	var count int
	o := stubOAuth{call: func(string, string) (*account.AccountCred, error) {
		count++
		return &account.AccountCred{AccessToken: "x"}, nil
	}}
	r := account.NewRefresher(repo, rdb, o, nil)
	err := r.RefreshOne(ctx, a.ID)
	require.Error(t, err)
	require.Equal(t, 0, count, "OAuth should not be called when lock held")
}

func TestRefresher_FailureMarksRefreshTokenInvalid(t *testing.T) {
	repo := newFakeRepo()
	ctx := context.Background()
	a := &account.Account{
		ChannelID: 1, Provider: "anthropic", CredType: "oauth",
		Status: account.StatusActive,
		Cred:   account.AccountCred{RefreshToken: "rt-bad"},
	}
	_ = repo.Create(ctx, a)

	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	o := stubOAuth{call: func(string, string) (*account.AccountCred, error) {
		return nil, errors.New("invalid_grant")
	}}
	r := account.NewRefresher(repo, rdb, o, nil)
	require.Error(t, r.RefreshOne(ctx, a.ID))
	got, _ := repo.Get(ctx, a.ID)
	require.Equal(t, int8(2), got.RefreshTokenValid)
}

func TestRefresher_NonOAuthSkips(t *testing.T) {
	repo := newFakeRepo()
	ctx := context.Background()
	a := &account.Account{
		ChannelID: 1, Provider: "openai", CredType: "apikey",
		Status: account.StatusActive,
		Cred:   account.AccountCred{APIKey: "sk-..."},
	}
	_ = repo.Create(ctx, a)

	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	var called bool
	o := stubOAuth{call: func(string, string) (*account.AccountCred, error) {
		called = true
		return nil, nil
	}}
	r := account.NewRefresher(repo, rdb, o, nil)
	require.NoError(t, r.RefreshOne(ctx, a.ID))
	require.False(t, called, "non-oauth account must not invoke ExchangeRefreshToken")
}
