package manual

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/ijry/pro-api/internal/audit"
	"github.com/ijry/pro-api/pkg/apierr"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ----- fakes -----

type fakeIDGen struct{ next int64 }

func (f *fakeIDGen) Generate() int64 {
	f.next++
	return f.next
}

type fakeClock struct{ now time.Time }

func (f *fakeClock) Now() time.Time { return f.now }

type fakeSetting struct {
	bools   map[string]bool
	ints    map[string]int
	floats  map[string]float64
	strings map[string]string
}

func newFakeSetting() *fakeSetting {
	return &fakeSetting{
		bools: map[string]bool{}, ints: map[string]int{},
		floats: map[string]float64{}, strings: map[string]string{},
	}
}

func (f *fakeSetting) GetBool(ctx context.Context, key string, def bool) bool {
	if v, ok := f.bools[key]; ok {
		return v
	}
	return def
}
func (f *fakeSetting) GetInt(ctx context.Context, key string, def int) int {
	if v, ok := f.ints[key]; ok {
		return v
	}
	return def
}
func (f *fakeSetting) GetFloat(ctx context.Context, key string, def float64) float64 {
	if v, ok := f.floats[key]; ok {
		return v
	}
	return def
}
func (f *fakeSetting) GetString(ctx context.Context, key string, def string) string {
	if v, ok := f.strings[key]; ok {
		return v
	}
	return def
}

type fakeWallet struct {
	credits     []creditCall
	creditErr   error
	balanceByID map[int64]int64
}

type creditCall struct {
	UserID  int64
	Amount  int64
	RefType string
	RefID   int64
	Desc    string
}

func (f *fakeWallet) Credit(ctx context.Context, userID int64, amount int64, refType string, refID int64, desc string) error {
	if f.creditErr != nil {
		return f.creditErr
	}
	f.credits = append(f.credits, creditCall{UserID: userID, Amount: amount, RefType: refType, RefID: refID, Desc: desc})
	if f.balanceByID == nil {
		f.balanceByID = map[int64]int64{}
	}
	f.balanceByID[userID] += amount
	return nil
}

type fakeAudit struct{ entries []audit.Entry }

func (f *fakeAudit) Log(ctx context.Context, e audit.Entry) error {
	f.entries = append(f.entries, e)
	return nil
}

// ----- builder -----

func newSvc(t *testing.T) (*service, *fakeWallet, *fakeAudit, *fakeSetting, *fakeClock) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("sqlite: %v", err)
	}
	if err := db.Exec(`
		CREATE TABLE manual_recharges (
			id              INTEGER PRIMARY KEY,
			user_id         INTEGER NOT NULL,
			amount_money    INTEGER NOT NULL,
			currency        TEXT NOT NULL DEFAULT 'CNY',
			amount_quota    INTEGER NOT NULL DEFAULT 0,
			status          INTEGER NOT NULL DEFAULT 0,
			applicant_note  TEXT NOT NULL DEFAULT '',
			reviewer_id     INTEGER,
			review_note     TEXT NOT NULL DEFAULT '',
			reviewed_at     DATETIME,
			created_at      DATETIME NOT NULL,
			updated_at      DATETIME NOT NULL
		)
	`).Error; err != nil {
		t.Fatalf("schema: %v", err)
	}
	wallet := &fakeWallet{}
	aud := &fakeAudit{}
	st := newFakeSetting()
	clk := &fakeClock{now: time.Date(2026, 5, 21, 8, 0, 0, 0, time.UTC)}
	svc := New(Config{
		DB:      db,
		Setting: st,
		Wallet:  wallet,
		Audit:   aud,
		IDGen:   &fakeIDGen{},
		Clock:   clk,
		Log:     zap.NewNop(),
	}).(*service)
	return svc, wallet, aud, st, clk
}

// ----- Apply -----

func TestApply_Success_ReturnsPending(t *testing.T) {
	svc, _, aud, _, clk := newSvc(t)
	rec, err := svc.Apply(context.Background(), 42, 1_000_000, "已转账")
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if rec.Status != StatusPending {
		t.Errorf("want pending, got %d", rec.Status)
	}
	if rec.UserID != 42 || rec.AmountMoney != 1_000_000 || rec.ApplicantNote != "已转账" {
		t.Errorf("fields wrong: %+v", rec)
	}
	if !rec.CreatedAt.Equal(clk.Now()) {
		t.Errorf("created_at wrong")
	}
	if len(aud.entries) != 1 || aud.entries[0].Action != "recharge.apply" {
		t.Errorf("audit wrong: %+v", aud.entries)
	}
}

func TestApply_FeatureDisabled_ReturnsInvalidParam(t *testing.T) {
	svc, _, _, st, _ := newSvc(t)
	st.bools["manual_recharge.enabled"] = false
	_, err := svc.Apply(context.Background(), 1, 1_000_000, "")
	mustErrCode(t, err, apierr.CodeInvalidParam)
}

func TestApply_BelowMin_ReturnsInvalidParam(t *testing.T) {
	svc, _, _, st, _ := newSvc(t)
	st.ints["manual_recharge.min_amount_cny"] = 100
	// 50 元 < 100 元
	_, err := svc.Apply(context.Background(), 1, 500_000, "")
	mustErrCode(t, err, apierr.CodeInvalidParam)
}

func TestApply_AboveMax_ReturnsInvalidParam(t *testing.T) {
	svc, _, _, st, _ := newSvc(t)
	st.ints["manual_recharge.max_amount_cny"] = 1000
	_, err := svc.Apply(context.Background(), 1, 100_000_000, "") // 10000 元
	mustErrCode(t, err, apierr.CodeInvalidParam)
}

func TestApply_NoteTooLong_ReturnsInvalidParam(t *testing.T) {
	svc, _, _, _, _ := newSvc(t)
	long := make([]byte, 513)
	for i := range long {
		long[i] = 'x'
	}
	_, err := svc.Apply(context.Background(), 1, 1_000_000, string(long))
	mustErrCode(t, err, apierr.CodeInvalidParam)
}

func TestApply_NegativeOrZero_ReturnsInvalidParam(t *testing.T) {
	svc, _, _, _, _ := newSvc(t)
	_, err := svc.Apply(context.Background(), 1, 0, "")
	mustErrCode(t, err, apierr.CodeInvalidParam)
	_, err = svc.Apply(context.Background(), 1, -100, "")
	mustErrCode(t, err, apierr.CodeInvalidParam)
}

// ----- Approve -----

func TestApprove_Success_CreditsWallet_Audits(t *testing.T) {
	svc, wallet, aud, st, _ := newSvc(t)
	st.floats["manual_recharge.exchange_rate_cny_per_usd"] = 7.0
	st.ints["pricing.base_quota_per_dollar"] = 500_000
	rec := mustApply(t, svc, 42, 1_000_000) // 100 元
	clearAudit(aud)
	approved, err := svc.Approve(context.Background(), rec.ID, 99, "OK")
	if err != nil {
		t.Fatalf("Approve: %v", err)
	}
	if approved.Status != StatusApproved {
		t.Errorf("want approved, got %d", approved.Status)
	}
	if approved.AmountQuota != 7_142_857 {
		t.Errorf("amount_quota wrong: %d", approved.AmountQuota)
	}
	if approved.ReviewerID == nil || *approved.ReviewerID != 99 {
		t.Errorf("reviewer_id wrong")
	}
	if approved.ReviewNote != "OK" {
		t.Errorf("review_note wrong")
	}
	if approved.ReviewedAt == nil {
		t.Errorf("reviewed_at not set")
	}
	if len(wallet.credits) != 1 {
		t.Fatalf("want 1 credit, got %d", len(wallet.credits))
	}
	if wallet.credits[0].Amount != 7_142_857 || wallet.credits[0].RefType != "manual" || wallet.credits[0].RefID != rec.ID {
		t.Errorf("credit wrong: %+v", wallet.credits[0])
	}
	if len(aud.entries) != 1 || aud.entries[0].Action != "recharge.approve" {
		t.Errorf("audit wrong: %+v", aud.entries)
	}
}

func TestApprove_NotPending_ReturnsInvalidParam(t *testing.T) {
	svc, _, _, st, _ := newSvc(t)
	st.floats["manual_recharge.exchange_rate_cny_per_usd"] = 7.0
	st.ints["pricing.base_quota_per_dollar"] = 500_000
	rec := mustApply(t, svc, 1, 1_000_000)
	if _, err := svc.Approve(context.Background(), rec.ID, 99, ""); err != nil {
		t.Fatalf("first approve: %v", err)
	}
	// 再次 approve 应当失败
	_, err := svc.Approve(context.Background(), rec.ID, 99, "")
	mustErrCode(t, err, apierr.CodeInvalidParam)
}

func TestApprove_NotFound_ReturnsOrderNotFound(t *testing.T) {
	svc, _, _, _, _ := newSvc(t)
	_, err := svc.Approve(context.Background(), 99999, 1, "")
	mustErrCode(t, err, apierr.CodeOrderNotFound)
}

func TestApprove_ExchangeRateMisconfigured_ReturnsInternal(t *testing.T) {
	svc, _, _, st, _ := newSvc(t)
	st.floats["manual_recharge.exchange_rate_cny_per_usd"] = 0
	st.ints["pricing.base_quota_per_dollar"] = 500_000
	rec := mustApply(t, svc, 1, 1_000_000)
	_, err := svc.Approve(context.Background(), rec.ID, 99, "")
	mustErrCode(t, err, apierr.CodeInternal)
}

func TestApprove_WalletCreditFails_RollsBackStatus(t *testing.T) {
	svc, wallet, _, st, _ := newSvc(t)
	st.floats["manual_recharge.exchange_rate_cny_per_usd"] = 7.0
	st.ints["pricing.base_quota_per_dollar"] = 500_000
	wallet.creditErr = errors.New("wallet boom")
	rec := mustApply(t, svc, 1, 1_000_000)

	_, err := svc.Approve(context.Background(), rec.ID, 99, "")
	if err == nil {
		t.Fatal("want error")
	}
	// 状态应当回滚到 pending,amount_quota = 0
	got, _ := svc.repo.GetByID(context.Background(), rec.ID)
	if got.Status != StatusPending {
		t.Errorf("status not rolled back: %d", got.Status)
	}
	if got.AmountQuota != 0 {
		t.Errorf("amount_quota not reset: %d", got.AmountQuota)
	}
	if got.ReviewerID != nil {
		t.Errorf("reviewer_id not cleared")
	}
}

// ----- Reject -----

func TestReject_Success_NoCredit(t *testing.T) {
	svc, wallet, aud, _, _ := newSvc(t)
	rec := mustApply(t, svc, 1, 1_000_000)
	clearAudit(aud)
	rejected, err := svc.Reject(context.Background(), rec.ID, 99, "工行无流水")
	if err != nil {
		t.Fatalf("Reject: %v", err)
	}
	if rejected.Status != StatusRejected {
		t.Errorf("want rejected, got %d", rejected.Status)
	}
	if rejected.AmountQuota != 0 {
		t.Errorf("amount_quota should remain 0")
	}
	if len(wallet.credits) != 0 {
		t.Errorf("wallet should not be credited on reject")
	}
	if len(aud.entries) != 1 || aud.entries[0].Action != "recharge.reject" {
		t.Errorf("audit wrong")
	}
}

func TestReject_NotPending_ReturnsInvalidParam(t *testing.T) {
	svc, _, _, _, _ := newSvc(t)
	rec := mustApply(t, svc, 1, 1_000_000)
	if _, err := svc.Reject(context.Background(), rec.ID, 99, ""); err != nil {
		t.Fatalf("first reject: %v", err)
	}
	_, err := svc.Reject(context.Background(), rec.ID, 99, "")
	mustErrCode(t, err, apierr.CodeInvalidParam)
}

// ----- Cancel -----

func TestCancel_Self_Pending_OK(t *testing.T) {
	svc, _, aud, _, _ := newSvc(t)
	rec := mustApply(t, svc, 42, 1_000_000)
	clearAudit(aud)
	canceled, err := svc.Cancel(context.Background(), rec.ID, 42)
	if err != nil {
		t.Fatalf("Cancel: %v", err)
	}
	if canceled.Status != StatusCanceled {
		t.Errorf("want canceled, got %d", canceled.Status)
	}
	if len(aud.entries) != 1 || aud.entries[0].Action != "recharge.cancel" {
		t.Errorf("audit wrong")
	}
}

func TestCancel_OtherUser_ReturnsOrderNotFound(t *testing.T) {
	svc, _, _, _, _ := newSvc(t)
	rec := mustApply(t, svc, 42, 1_000_000)
	// 用户 7 想取消用户 42 的单 → 应当返回 not found(不暴露存在性)
	_, err := svc.Cancel(context.Background(), rec.ID, 7)
	mustErrCode(t, err, apierr.CodeOrderNotFound)
}

func TestCancel_NotPending_ReturnsInvalidParam(t *testing.T) {
	svc, _, _, _, _ := newSvc(t)
	rec := mustApply(t, svc, 42, 1_000_000)
	_, _ = svc.Cancel(context.Background(), rec.ID, 42)
	_, err := svc.Cancel(context.Background(), rec.ID, 42)
	mustErrCode(t, err, apierr.CodeInvalidParam)
}

func TestCancel_NotFound_ReturnsOrderNotFound(t *testing.T) {
	svc, _, _, _, _ := newSvc(t)
	_, err := svc.Cancel(context.Background(), 999999, 42)
	mustErrCode(t, err, apierr.CodeOrderNotFound)
}

// ----- List / Get -----

func TestList_Self_Filtered(t *testing.T) {
	svc, _, _, _, _ := newSvc(t)
	_ = mustApply(t, svc, 42, 1_000_000)
	_ = mustApply(t, svc, 7, 2_000_000)
	items, total, err := svc.List(context.Background(), ListFilter{UserID: 42, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Fatalf("want 1, got total=%d items=%d", total, len(items))
	}
	if items[0].UserID != 42 {
		t.Fatalf("wrong user")
	}
}

func TestGet_Self_OK(t *testing.T) {
	svc, _, _, _, _ := newSvc(t)
	rec := mustApply(t, svc, 42, 1_000_000)
	got, err := svc.Get(context.Background(), rec.ID, 42)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.ID != rec.ID {
		t.Fatal("wrong id")
	}
}

func TestGet_OtherUser_ReturnsOrderNotFound(t *testing.T) {
	svc, _, _, _, _ := newSvc(t)
	rec := mustApply(t, svc, 42, 1_000_000)
	_, err := svc.Get(context.Background(), rec.ID, 7)
	mustErrCode(t, err, apierr.CodeOrderNotFound)
}

func TestGet_Admin_BypassOwner(t *testing.T) {
	svc, _, _, _, _ := newSvc(t)
	rec := mustApply(t, svc, 42, 1_000_000)
	// userID = 0 = admin
	got, err := svc.Get(context.Background(), rec.ID, 0)
	if err != nil {
		t.Fatalf("Get(admin): %v", err)
	}
	if got.ID != rec.ID {
		t.Fatal("wrong id")
	}
}

// ----- helpers -----

func mustApply(t *testing.T, svc Service, userID, amount int64) *Recharge {
	t.Helper()
	rec, err := svc.Apply(context.Background(), userID, amount, "test")
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	return rec
}

func clearAudit(a *fakeAudit) { a.entries = nil }

func mustErrCode(t *testing.T, err error, want apierr.Code) {
	t.Helper()
	if err == nil {
		t.Fatalf("want error code %d, got nil", want)
	}
	var e *apierr.Error
	if !errors.As(err, &e) {
		t.Fatalf("want apierr.Error, got %T: %v", err, err)
	}
	if e.Code != want {
		t.Fatalf("want code %d, got %d: %s", want, e.Code, e.Message)
	}
}

// 避免编译器抱怨 json 未使用
var _ = json.Marshal
