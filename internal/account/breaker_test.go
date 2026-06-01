package account_test

import (
	"context"
	"testing"
	"time"

	"github.com/ijry/pro-api/internal/account"
	"github.com/stretchr/testify/require"
)

func TestBreaker_MarkCooldown(t *testing.T) {
	repo := newFakeRepo()
	ctx := context.Background()
	a := mkAcc(0, 1, 5000)
	_ = repo.Create(ctx, a)
	b := account.NewBreaker(repo, nil, nil)
	until := time.Now().Add(10 * time.Second)
	require.NoError(t, b.MarkCooldown(ctx, a.ID, until, "429"))
	got, _ := repo.Get(ctx, a.ID)
	require.Equal(t, account.StatusCooldown, got.Status)
	require.NotNil(t, got.CooldownUntil)
}

func TestBreaker_ReaperRestores(t *testing.T) {
	repo := newFakeRepo()
	ctx := context.Background()
	a := mkAcc(0, 1, 5000)
	past := time.Now().Add(-time.Second)
	a.Status = account.StatusCooldown
	a.CooldownUntil = &past
	_ = repo.Create(ctx, a)
	b := account.NewBreaker(repo, nil, nil)
	n, err := b.RunReaperOnce(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, n)
	got, _ := repo.Get(ctx, a.ID)
	require.Equal(t, account.StatusActive, got.Status)
	require.Nil(t, got.CooldownUntil)
}

func TestBreaker_MarkInvalid(t *testing.T) {
	repo := newFakeRepo()
	ctx := context.Background()
	a := mkAcc(0, 1, 5000)
	_ = repo.Create(ctx, a)
	b := account.NewBreaker(repo, nil, nil)
	require.NoError(t, b.MarkInvalid(ctx, a.ID, "invalid_grant"))
	got, _ := repo.Get(ctx, a.ID)
	require.Equal(t, account.StatusInvalid, got.Status)
}

func TestBreaker_IncConsecFailure(t *testing.T) {
	repo := newFakeRepo()
	ctx := context.Background()
	a := mkAcc(0, 1, 5000)
	_ = repo.Create(ctx, a)
	b := account.NewBreaker(repo, nil, nil)
	n, err := b.IncConsecFailure(ctx, a.ID)
	require.NoError(t, err)
	require.Equal(t, 1, n)
	n, _ = b.IncConsecFailure(ctx, a.ID)
	require.Equal(t, 2, n)
}
