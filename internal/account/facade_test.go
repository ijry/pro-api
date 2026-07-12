package account_test

import (
	"context"
	"testing"
	"time"

	"github.com/ijry/pro-api/internal/account"
	"github.com/stretchr/testify/require"
)

// waitRemaining 轮询等待账号 quota_5h_remaining 达到期望值(ReportUsage 异步扣减)。
func waitRemaining(t *testing.T, repo *fakeRepo, id int64, want int64) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		a, err := repo.Get(context.Background(), id)
		require.NoError(t, err)
		if a.Quota5hRemaining != nil && *a.Quota5hRemaining == want {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	a, _ := repo.Get(context.Background(), id)
	got := int64(-1)
	if a != nil && a.Quota5hRemaining != nil {
		got = *a.Quota5hRemaining
	}
	t.Fatalf("remaining 未达期望:want=%d got=%d", want, got)
}

func newManualAcc(id int64, total, remaining int64) *account.Account {
	return &account.Account{
		ID:               id,
		Status:           account.StatusActive,
		Provider:         "anthropic",
		QuotaMode:        account.QuotaModeManual,
		Quota5hTotal:     &total,
		Quota5hRemaining: &remaining,
	}
}

func TestFacade_ReportUsage_DeductsManual(t *testing.T) {
	repo := newFakeRepo()
	require.NoError(t, repo.Create(context.Background(), newManualAcc(1, 1000, 1000)))
	f := &account.Facade{Repo: repo}

	f.ReportUsage(1, 300)
	waitRemaining(t, repo, 1, 700)
}

func TestFacade_ReportUsage_ClampsAtZero(t *testing.T) {
	repo := newFakeRepo()
	require.NoError(t, repo.Create(context.Background(), newManualAcc(1, 1000, 100)))
	f := &account.Facade{Repo: repo}

	f.ReportUsage(1, 500) // 超过 remaining
	waitRemaining(t, repo, 1, 0)
}

func TestFacade_ReportUsage_IgnoresAutoMode(t *testing.T) {
	repo := newFakeRepo()
	total, rem := int64(1000), int64(1000)
	a := &account.Account{
		ID: 1, Status: account.StatusActive, Provider: "anthropic",
		QuotaMode: account.QuotaModeAuto, Quota5hTotal: &total, Quota5hRemaining: &rem,
	}
	require.NoError(t, repo.Create(context.Background(), a))
	f := &account.Facade{Repo: repo}

	f.ReportUsage(1, 300)
	// auto 模式不应被扣减:等一小会儿确认 remaining 未变。
	time.Sleep(100 * time.Millisecond)
	got, err := repo.Get(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, int64(1000), *got.Quota5hRemaining, "auto 模式不应按用量扣减")
}

func TestFacade_ReportUsage_NoopOnZeroAccount(t *testing.T) {
	repo := newFakeRepo()
	f := &account.Facade{Repo: repo}
	// accountID==0(未启用账号池)不应 panic、不应触碰 repo。
	f.ReportUsage(0, 300)
	f.ReportUsage(1, 0) // tokens<=0 亦为 no-op
	time.Sleep(50 * time.Millisecond)
}
