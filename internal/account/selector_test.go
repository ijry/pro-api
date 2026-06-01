package account_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/ijry/pro-api/internal/account"
	"github.com/ijry/pro-api/internal/channel"
	"github.com/stretchr/testify/require"
)

func mkAcc(id, ch int64, rem int64) *account.Account {
	total := int64(10000)
	return &account.Account{
		ID:               id,
		ChannelID:        ch,
		Status:           account.StatusActive,
		Weight:           100,
		Priority:         0,
		Provider:         "anthropic",
		Quota5hTotal:     &total,
		Quota5hRemaining: &rem,
	}
}

func TestSelector_TopK_PrefersHigherRemaining(t *testing.T) {
	repo := newFakeRepo()
	ctx := context.Background()
	_ = repo.Create(ctx, mkAcc(0, 5, 9000))
	_ = repo.Create(ctx, mkAcc(0, 5, 5000))
	_ = repo.Create(ctx, mkAcc(0, 5, 1000))
	_ = repo.Create(ctx, mkAcc(0, 5, 500))

	ch := &channel.Channel{ID: 5, AccountStrategy: "top_k", AccountTopK: 2, Tags: json.RawMessage("[]")}

	sel := account.NewSelector(repo, nil, 42)
	hits := map[int64]int{}
	for i := 0; i < 200; i++ {
		a, err := sel.Select(ctx, ch, account.SelectHint{})
		require.NoError(t, err)
		hits[a.ID]++
	}
	// 余量排序后,前 2 名(9000/5000)应被选中;后两名不应被选
	require.Greater(t, hits[1]+hits[2], 0)
	require.Equal(t, 0, hits[3])
	require.Equal(t, 0, hits[4])
}

func TestSelector_NoActiveReturnsErr(t *testing.T) {
	repo := newFakeRepo()
	sel := account.NewSelector(repo, nil, 1)
	ch := &channel.Channel{ID: 6, AccountStrategy: "top_k", AccountTopK: 3, Tags: json.RawMessage("[]")}
	_, err := sel.Select(context.Background(), ch, account.SelectHint{})
	require.ErrorIs(t, err, account.ErrPoolEmpty)
}

func TestSelector_ExcludedFiltered(t *testing.T) {
	repo := newFakeRepo()
	ctx := context.Background()
	a := mkAcc(0, 7, 9000)
	_ = repo.Create(ctx, a)
	sel := account.NewSelector(repo, nil, 1)
	ch := &channel.Channel{ID: 7, AccountStrategy: "top_k", AccountTopK: 1, Tags: json.RawMessage("[]")}
	_, err := sel.Select(ctx, ch, account.SelectHint{Excluded: []int64{a.ID}})
	require.ErrorIs(t, err, account.ErrPoolEmpty)
}
