package account_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/ijry/pro-api/internal/account"
	"github.com/stretchr/testify/require"
)

// 注:后台探测循环依赖 fakeRepo.ListForProbe 返回陈旧 active 账号(见 fake_repo_test.go)。

type stubProviderProbe struct {
	h   http.Header
	err error
}

func (s stubProviderProbe) Probe(context.Context, account.AccountCred) (http.Header, error) {
	return s.h, s.err
}

type stubQuotaTracker struct {
	extractCalled bool
	updateCalled  bool
}

func (s *stubQuotaTracker) ExtractFromResponse(_ string, _ http.Header) *account.QuotaSnapshot {
	s.extractCalled = true
	return &account.QuotaSnapshot{}
}

func (s *stubQuotaTracker) UpdateAccount(context.Context, int64, *account.QuotaSnapshot) error {
	s.updateCalled = true
	return nil
}

func TestProbe_ProbeOneCallsTracker(t *testing.T) {
	repo := newFakeRepo()
	tk := &stubQuotaTracker{}
	h := http.Header{"X": []string{"1"}}
	p := account.NewProbe(repo, tk, nil, map[string]account.ProviderProbe{
		"anthropic": stubProviderProbe{h: h},
	}, account.ProbeConfig{}, nil)
	a := &account.Account{Provider: "anthropic"}
	require.NoError(t, repo.Create(context.Background(), a))
	require.NoError(t, p.ProbeOne(context.Background(), a))
	require.True(t, tk.extractCalled)
	require.True(t, tk.updateCalled)

	// event "probed" appended on success
	require.Len(t, repo.events, 1)
	require.Equal(t, "probed", repo.events[0].eventType)
}

func TestProbe_UnknownProviderReturnsErr(t *testing.T) {
	repo := newFakeRepo()
	p := account.NewProbe(repo, &stubQuotaTracker{}, nil, map[string]account.ProviderProbe{}, account.ProbeConfig{}, nil)
	a := &account.Account{Provider: "bogus"}
	require.NoError(t, repo.Create(context.Background(), a))
	err := p.ProbeOne(context.Background(), a)
	require.Error(t, err)
}

// TestProbe_RunOnceMarksExpired 验证后台循环探测到 401 时,通过 Breaker 标记账号为 Expired。
func TestProbe_RunOnceMarksExpired(t *testing.T) {
	repo := newFakeRepo()
	br := account.NewBreaker(repo, nil, nil)
	// 401 → failTokenExpired → MarkExpired
	pp := stubProviderProbe{h: http.Header{}, err: errors.New("anthropic probe: status 401")}
	p := account.NewProbe(repo, &stubQuotaTracker{}, br, map[string]account.ProviderProbe{
		"anthropic": pp,
	}, account.ProbeConfig{}, nil)

	a := &account.Account{Provider: "anthropic", Status: account.StatusActive}
	require.NoError(t, repo.Create(context.Background(), a))

	rp, ok := p.(interface {
		RunOnce(context.Context) (int, error)
	})
	require.True(t, ok, "probeImpl should expose RunOnce")
	n, err := rp.RunOnce(context.Background())
	require.NoError(t, err)
	require.Equal(t, 1, n)

	got, err := repo.Get(context.Background(), a.ID)
	require.NoError(t, err)
	require.Equal(t, account.StatusExpired, got.Status)
}

// TestProbe_RunOnceUnknownErrKeepsActive 验证网络类未知失败不改状态,避免误伤有效账号。
func TestProbe_RunOnceUnknownErrKeepsActive(t *testing.T) {
	repo := newFakeRepo()
	br := account.NewBreaker(repo, nil, nil)
	pp := stubProviderProbe{err: errors.New("dial tcp: i/o timeout")}
	p := account.NewProbe(repo, &stubQuotaTracker{}, br, map[string]account.ProviderProbe{
		"anthropic": pp,
	}, account.ProbeConfig{}, nil)

	a := &account.Account{Provider: "anthropic", Status: account.StatusActive}
	require.NoError(t, repo.Create(context.Background(), a))

	rp := p.(interface {
		RunOnce(context.Context) (int, error)
	})
	_, err := rp.RunOnce(context.Background())
	require.NoError(t, err)

	got, err := repo.Get(context.Background(), a.ID)
	require.NoError(t, err)
	require.Equal(t, account.StatusActive, got.Status, "未知失败不应改状态")
}

// --- 手动额度扣减 (manual quota) 契约 ---

// TestDeductManualQuota_OnlyManualMode 验证 DeductManualQuota 仅对 manual 模式生效,
// 且钳到 0 不为负;auto/none 模式不受影响。
func TestDeductManualQuota_OnlyManualMode(t *testing.T) {
	repo := newFakeRepo()
	ctx := context.Background()

	mk := func(mode string, rem int64) *account.Account {
		r := rem
		a := &account.Account{Provider: "openai", Status: account.StatusActive, QuotaMode: mode, Quota5hRemaining: &r}
		require.NoError(t, repo.Create(ctx, a))
		return a
	}
	manual := mk(account.QuotaModeManual, 100)
	auto := mk(account.QuotaModeAuto, 100)
	none := mk(account.QuotaModeNone, 100)

	require.NoError(t, repo.DeductManualQuota(ctx, manual.ID, 30))
	require.NoError(t, repo.DeductManualQuota(ctx, auto.ID, 30))
	require.NoError(t, repo.DeductManualQuota(ctx, none.ID, 30))

	got, _ := repo.Get(ctx, manual.ID)
	require.Equal(t, int64(70), *got.Quota5hRemaining, "manual 应被扣减")
	got, _ = repo.Get(ctx, auto.ID)
	require.Equal(t, int64(100), *got.Quota5hRemaining, "auto 不应被扣减")
	got, _ = repo.Get(ctx, none.ID)
	require.Equal(t, int64(100), *got.Quota5hRemaining, "none 不应被扣减")

	// 扣超过余量应钳到 0,不为负。
	require.NoError(t, repo.DeductManualQuota(ctx, manual.ID, 9999))
	got, _ = repo.Get(ctx, manual.ID)
	require.Equal(t, int64(0), *got.Quota5hRemaining)
}

// TestListForProbe_SkipsManualAndNone 验证后台探测只拉 auto 模式,跳过 manual/none。
func TestListForProbe_SkipsManualAndNone(t *testing.T) {
	repo := newFakeRepo()
	ctx := context.Background()
	mk := func(mode string) int64 {
		a := &account.Account{Provider: "openai", Status: account.StatusActive, QuotaMode: mode}
		require.NoError(t, repo.Create(ctx, a))
		return a.ID
	}
	autoID := mk(account.QuotaModeAuto)
	_ = mk(account.QuotaModeManual)
	_ = mk(account.QuotaModeNone)

	list, err := repo.ListForProbe(ctx, time.Now(), 100)
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.Equal(t, autoID, list[0].ID)
}
