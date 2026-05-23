package redeem

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ijry/pro-api/internal/audit"
	"github.com/ijry/pro-api/pkg/apierr"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ----- fakes -----

type fakeIDGen struct {
	mu   sync.Mutex
	next int64
}

func (f *fakeIDGen) Generate() int64 {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.next++
	return f.next
}

type fakeClock struct{ now time.Time }

func (f *fakeClock) Now() time.Time { return f.now }

type fakeSetting struct {
	ints map[string]int
}

func newFakeSetting() *fakeSetting {
	return &fakeSetting{ints: map[string]int{}}
}

func (f *fakeSetting) GetInt(ctx context.Context, key string, def int) int {
	if v, ok := f.ints[key]; ok {
		return v
	}
	return def
}

type fakeWallet struct {
	mu        sync.Mutex
	credits   []creditCall
	creditErr error
}

type creditCall struct {
	UserID  int64
	Amount  int64
	RefType string
	RefID   int64
	Desc    string
}

func (f *fakeWallet) Credit(ctx context.Context, userID int64, amount int64, refType string, refID int64, desc string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.creditErr != nil {
		return f.creditErr
	}
	f.credits = append(f.credits, creditCall{UserID: userID, Amount: amount, RefType: refType, RefID: refID, Desc: desc})
	return nil
}

func (f *fakeWallet) creditCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.credits)
}

type fakeAudit struct {
	mu      sync.Mutex
	entries []audit.Entry
}

func (f *fakeAudit) Log(ctx context.Context, e audit.Entry) error {
	f.mu.Lock()
	defer f.mu.Unlock()
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
		CREATE TABLE redeem_codes (
			id           INTEGER PRIMARY KEY,
			code_hash    TEXT NOT NULL UNIQUE,
			code_prefix  TEXT NOT NULL,
			amount_quota INTEGER NOT NULL,
			batch_no     TEXT NOT NULL DEFAULT '',
			status       INTEGER NOT NULL DEFAULT 0,
			used_by      INTEGER,
			used_at      DATETIME,
			expires_at   DATETIME,
			created_by   INTEGER NOT NULL,
			created_at   DATETIME NOT NULL
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

// ----- BatchCreate -----

func TestBatchCreate_Success_ReturnsPlaintexts(t *testing.T) {
	svc, _, aud, _, _ := newSvc(t)
	plains, batch, ids, err := svc.BatchCreate(context.Background(), 1, 3, 100, nil, "")
	if err != nil {
		t.Fatalf("BatchCreate: %v", err)
	}
	if len(plains) != 3 || len(ids) != 3 {
		t.Fatalf("want 3 each, got plains=%d ids=%d", len(plains), len(ids))
	}
	if batch == "" {
		t.Fatal("auto batch_no should be set")
	}
	// 校验明文长度 16,且能 normalize 回去查到 hash
	for i, p := range plains {
		if len(p) != codeLen {
			t.Errorf("plain[%d] len wrong: %d", i, len(p))
		}
		got, err := svc.repo.GetByHash(context.Background(), hashCode(p))
		if err != nil || got == nil {
			t.Errorf("plain[%d] not in DB: %v", i, err)
		}
	}
	// 审计:1 条 batch_create,不含明文
	if len(aud.entries) != 1 || aud.entries[0].Action != "redeem.batch_create" {
		t.Errorf("audit wrong: %v", aud.entries)
	}
	auditBody := string(aud.entries[0].After)
	for _, p := range plains {
		if strings.Contains(auditBody, p) {
			t.Errorf("audit should not contain plaintext")
		}
	}
}

func TestBatchCreate_CountZero_ReturnsInvalidParam(t *testing.T) {
	svc, _, _, _, _ := newSvc(t)
	_, _, _, err := svc.BatchCreate(context.Background(), 1, 0, 100, nil, "")
	mustErrCode(t, err, apierr.CodeInvalidParam)
}

func TestBatchCreate_Count1001_ReturnsInvalidParam(t *testing.T) {
	svc, _, _, _, _ := newSvc(t)
	_, _, _, err := svc.BatchCreate(context.Background(), 1, 1001, 100, nil, "")
	mustErrCode(t, err, apierr.CodeInvalidParam)
}

func TestBatchCreate_AmountQuotaZero_ReturnsInvalidParam(t *testing.T) {
	svc, _, _, _, _ := newSvc(t)
	_, _, _, err := svc.BatchCreate(context.Background(), 1, 1, 0, nil, "")
	mustErrCode(t, err, apierr.CodeInvalidParam)
}

func TestBatchCreate_DefaultExpiresAt_FromSetting(t *testing.T) {
	svc, _, _, st, clk := newSvc(t)
	st.ints["redeem.default_expires_days"] = 365
	_, _, ids, err := svc.BatchCreate(context.Background(), 1, 1, 100, nil, "")
	if err != nil {
		t.Fatalf("BatchCreate: %v", err)
	}
	c, err := svc.repo.GetByID(context.Background(), ids[0])
	if err != nil || c == nil {
		t.Fatal("not found")
	}
	if c.ExpiresAt == nil {
		t.Fatal("expires_at should be set")
	}
	want := clk.Now().UTC().AddDate(0, 0, 365)
	if c.ExpiresAt.Sub(want).Abs() > time.Second {
		t.Fatalf("expires_at off: got %v want %v", c.ExpiresAt, want)
	}
}

func TestBatchCreate_ExpiresAtSettingZero_NeverExpires(t *testing.T) {
	svc, _, _, st, _ := newSvc(t)
	st.ints["redeem.default_expires_days"] = 0
	_, _, ids, err := svc.BatchCreate(context.Background(), 1, 1, 100, nil, "")
	if err != nil {
		t.Fatalf("BatchCreate: %v", err)
	}
	c, _ := svc.repo.GetByID(context.Background(), ids[0])
	if c.ExpiresAt != nil {
		t.Fatalf("want nil expires_at, got %v", c.ExpiresAt)
	}
}

func TestBatchCreate_CustomBatchNo_UsesIt(t *testing.T) {
	svc, _, _, _, _ := newSvc(t)
	_, batch, ids, err := svc.BatchCreate(context.Background(), 1, 1, 100, nil, "promo-1")
	if err != nil {
		t.Fatalf("BatchCreate: %v", err)
	}
	if batch != "promo-1" {
		t.Fatalf("batch_no wrong: %s", batch)
	}
	c, _ := svc.repo.GetByID(context.Background(), ids[0])
	if c.BatchNo != "promo-1" {
		t.Fatalf("got %s", c.BatchNo)
	}
}

func TestBatchCreate_PrefixCorrect(t *testing.T) {
	svc, _, _, _, _ := newSvc(t)
	plains, _, ids, err := svc.BatchCreate(context.Background(), 1, 5, 100, nil, "")
	if err != nil {
		t.Fatalf("BatchCreate: %v", err)
	}
	for i, p := range plains {
		c, _ := svc.repo.GetByID(context.Background(), ids[i])
		if c.CodePrefix != p[:4] {
			t.Errorf("prefix mismatch: code=%s prefix=%s", p, c.CodePrefix)
		}
	}
}

// ----- Redeem -----

func TestRedeem_Success_UpdatesStatusAndCreditsWallet(t *testing.T) {
	svc, wallet, aud, _, clk := newSvc(t)
	plains, _, _, _ := svc.BatchCreate(context.Background(), 1, 1, 500, nil, "")
	clearAudit(aud)

	res, err := svc.Redeem(context.Background(), 42, plains[0])
	if err != nil {
		t.Fatalf("Redeem: %v", err)
	}
	if res.AmountQuota != 500 {
		t.Errorf("amount wrong: %d", res.AmountQuota)
	}
	if !res.UsedAt.Equal(clk.Now().UTC()) {
		t.Errorf("used_at wrong: %v vs %v", res.UsedAt, clk.Now().UTC())
	}
	if wallet.creditCount() != 1 {
		t.Errorf("want 1 credit, got %d", wallet.creditCount())
	}
	if wallet.credits[0].RefType != "redeem" || wallet.credits[0].UserID != 42 || wallet.credits[0].Amount != 500 {
		t.Errorf("credit wrong: %+v", wallet.credits[0])
	}
	if len(aud.entries) != 1 || aud.entries[0].Action != "redeem.use" {
		t.Errorf("audit wrong: %v", aud.entries)
	}
}

func TestRedeem_FormatInvalid_ReturnsFormatReason(t *testing.T) {
	svc, _, _, _, _ := newSvc(t)
	_, err := svc.Redeem(context.Background(), 1, "too-short")
	mustErrCodeAndReason(t, err, apierr.CodeRedeemInvalid, "format")
}

func TestRedeem_NotFound_ReturnsNotFoundReason(t *testing.T) {
	svc, _, _, _, _ := newSvc(t)
	_, err := svc.Redeem(context.Background(), 1, "ABCD2345EFGH6789")
	mustErrCodeAndReason(t, err, apierr.CodeRedeemInvalid, "not_found")
}

func TestRedeem_AlreadyUsed_ReturnsUsedReason(t *testing.T) {
	svc, _, _, _, _ := newSvc(t)
	plains, _, _, _ := svc.BatchCreate(context.Background(), 1, 1, 100, nil, "")
	if _, err := svc.Redeem(context.Background(), 42, plains[0]); err != nil {
		t.Fatalf("first: %v", err)
	}
	_, err := svc.Redeem(context.Background(), 42, plains[0])
	mustErrCodeAndReason(t, err, apierr.CodeRedeemInvalid, "used")
}

func TestRedeem_Disabled_ReturnsDisabledReason(t *testing.T) {
	svc, _, _, _, _ := newSvc(t)
	plains, _, ids, _ := svc.BatchCreate(context.Background(), 1, 1, 100, nil, "")
	if _, err := svc.Disable(context.Background(), 99, []int64{ids[0]}, "test"); err != nil {
		t.Fatalf("Disable: %v", err)
	}
	_, err := svc.Redeem(context.Background(), 42, plains[0])
	mustErrCodeAndReason(t, err, apierr.CodeRedeemInvalid, "disabled")
}

func TestRedeem_Expired_ReturnsExpiredReason(t *testing.T) {
	svc, _, _, _, clk := newSvc(t)
	past := clk.Now().UTC().Add(-1 * time.Hour)
	plains, _, _, err := svc.BatchCreate(context.Background(), 1, 1, 100, &past, "")
	if err != nil {
		t.Fatalf("BatchCreate: %v", err)
	}
	_, err = svc.Redeem(context.Background(), 42, plains[0])
	mustErrCodeAndReason(t, err, apierr.CodeRedeemInvalid, "expired")
}

func TestRedeem_NormalizesInput_WithHyphenAndLowercase(t *testing.T) {
	svc, _, _, _, _ := newSvc(t)
	plains, _, _, _ := svc.BatchCreate(context.Background(), 1, 1, 100, nil, "")
	// 取明文,加入分隔符 + 转小写
	fancy := strings.ToLower(format(plains[0]))
	res, err := svc.Redeem(context.Background(), 42, fancy)
	if err != nil {
		t.Fatalf("Redeem: %v", err)
	}
	if res.AmountQuota != 100 {
		t.Fatal("wrong amount")
	}
}

func TestRedeem_WalletCreditFails_RollsBackStatus(t *testing.T) {
	svc, wallet, _, _, _ := newSvc(t)
	plains, _, ids, _ := svc.BatchCreate(context.Background(), 1, 1, 100, nil, "")
	wallet.creditErr = errors.New("wallet boom")

	_, err := svc.Redeem(context.Background(), 42, plains[0])
	if err == nil {
		t.Fatal("want error")
	}
	c, _ := svc.repo.GetByID(context.Background(), ids[0])
	if c.Status != StatusUnused {
		t.Errorf("status not rolled back: %d", c.Status)
	}
	if c.UsedBy != nil {
		t.Errorf("used_by not cleared")
	}
}

// ----- Disable -----

func TestDisable_OnlyUnusedTransitions(t *testing.T) {
	svc, _, _, _, _ := newSvc(t)
	_, _, ids, _ := svc.BatchCreate(context.Background(), 1, 3, 100, nil, "")
	// 第 0 张兑了
	plains, _, _, _ := svc.BatchCreate(context.Background(), 1, 1, 100, nil, "")
	usedID := mustGetIDByPlain(t, svc, plains[0])
	_, _ = svc.Redeem(context.Background(), 1, plains[0])

	disabled, err := svc.Disable(context.Background(), 99, append(ids, usedID), "test")
	if err != nil {
		t.Fatalf("Disable: %v", err)
	}
	if disabled != 3 {
		t.Fatalf("want 3 disabled (used skipped), got %d", disabled)
	}
}

func TestDisable_AuditOnce(t *testing.T) {
	svc, _, aud, _, _ := newSvc(t)
	_, _, ids, _ := svc.BatchCreate(context.Background(), 1, 1, 100, nil, "")
	clearAudit(aud)
	_, err := svc.Disable(context.Background(), 99, ids, "leak")
	if err != nil {
		t.Fatalf("Disable: %v", err)
	}
	if len(aud.entries) != 1 || aud.entries[0].Action != "redeem.disable" {
		t.Errorf("audit wrong: %+v", aud.entries)
	}
}

// ----- List / Get / Export -----

func TestList_FilterByBatch(t *testing.T) {
	svc, _, _, _, _ := newSvc(t)
	_, _, _, _ = svc.BatchCreate(context.Background(), 1, 2, 100, nil, "A")
	_, _, _, _ = svc.BatchCreate(context.Background(), 1, 3, 100, nil, "B")

	items, total, err := svc.List(context.Background(), ListFilter{BatchNo: "B", Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total != 3 || len(items) != 3 {
		t.Fatalf("want 3 each, got total=%d items=%d", total, len(items))
	}
}

func TestGet_NotFound_ReturnsOrderNotFound(t *testing.T) {
	svc, _, _, _, _ := newSvc(t)
	_, err := svc.Get(context.Background(), 999)
	mustErrCode(t, err, apierr.CodeOrderNotFound)
}

func TestExport_WritesCSV_NoPlaintext(t *testing.T) {
	svc, _, aud, _, _ := newSvc(t)
	_, _, _, _ = svc.BatchCreate(context.Background(), 1, 2, 100, nil, "P")
	clearAudit(aud)

	var buf bytes.Buffer
	if err := svc.Export(context.Background(), &buf, ListFilter{BatchNo: "P"}); err != nil {
		t.Fatalf("Export: %v", err)
	}
	r := csv.NewReader(&buf)
	rows, err := r.ReadAll()
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}
	// 1 header + 2 data rows
	if len(rows) != 3 {
		t.Fatalf("want 3 rows, got %d: %v", len(rows), rows)
	}
	// 头部含 id/code_prefix/code_display 不含 plaintext
	header := strings.Join(rows[0], ",")
	if !strings.Contains(header, "code_prefix") {
		t.Errorf("header missing code_prefix")
	}
	if strings.Contains(header, "plaintext") {
		t.Errorf("header should not contain plaintext")
	}
	// audit
	if len(aud.entries) != 1 || aud.entries[0].Action != "redeem.export" {
		t.Errorf("audit wrong: %+v", aud.entries)
	}
}

// ----- helpers -----

func mustGetIDByPlain(t *testing.T, svc *service, plain string) int64 {
	t.Helper()
	c, err := svc.repo.GetByHash(context.Background(), hashCode(plain))
	if err != nil || c == nil {
		t.Fatalf("can't find by plain: %v", err)
	}
	return c.ID
}

func clearAudit(a *fakeAudit) {
	a.mu.Lock()
	a.entries = nil
	a.mu.Unlock()
}

func mustErrCode(t *testing.T, err error, want apierr.Code) {
	t.Helper()
	if err == nil {
		t.Fatalf("want code %d, got nil", want)
	}
	var e *apierr.Error
	if !errors.As(err, &e) {
		t.Fatalf("not apierr.Error: %T %v", err, err)
	}
	if e.Code != want {
		t.Fatalf("want code %d, got %d: %s", want, e.Code, e.Message)
	}
}

func mustErrCodeAndReason(t *testing.T, err error, want apierr.Code, reason string) {
	t.Helper()
	mustErrCode(t, err, want)
	var e *apierr.Error
	_ = errors.As(err, &e)
	if e.Details["reason"] != reason {
		t.Fatalf("want reason %q, got %v", reason, e.Details["reason"])
	}
}
