package wallet

import (
	"errors"
	"sync"
	"testing"

	"github.com/ijry/pro-api/pkg/apierr"
)

func TestCredit_AmountZero_ReturnsInvalidParam(t *testing.T) {
	s, _ := newTestStore(t, false)
	w, _ := s.GetOrCreate(ctx(), OwnerTypeUser, 1)
	err := s.Credit(ctx(), w.ID, 0, RefTypeManual, nil, "", 0)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCredit_AmountNegative_ReturnsInvalidParam(t *testing.T) {
	s, _ := newTestStore(t, false)
	w, _ := s.GetOrCreate(ctx(), OwnerTypeUser, 1)
	err := s.Credit(ctx(), w.ID, -10, RefTypeManual, nil, "", 0)
	var ae *apierr.Error
	if !errors.As(err, &ae) || ae.Code != apierr.CodeInvalidParam {
		t.Fatalf("expected CodeInvalidParam, got %v", err)
	}
}

func TestCredit_InvalidRefType(t *testing.T) {
	s, _ := newTestStore(t, false)
	w, _ := s.GetOrCreate(ctx(), OwnerTypeUser, 1)
	err := s.Credit(ctx(), w.ID, 10, "usage", nil, "", 0) // usage 不允许 Credit
	if err == nil {
		t.Fatal("expected error for invalid ref_type")
	}
}

func TestCredit_TransactionalDB_LedgerAndWalletConsistent(t *testing.T) {
	s, _ := newTestStore(t, false)
	w, _ := s.GetOrCreate(ctx(), OwnerTypeUser, 1)
	if err := s.Credit(ctx(), w.ID, 1000, RefTypeManual, nil, "test top-up", 0); err != nil {
		t.Fatal(err)
	}
	snap, _ := s.Snapshot(ctx(), w.ID)
	if snap.QuotaBalance != 1000 {
		t.Fatalf("balance = %d, want 1000", snap.QuotaBalance)
	}
	if snap.QuotaTotalRecharged != 1000 {
		t.Fatalf("recharged = %d, want 1000", snap.QuotaTotalRecharged)
	}
	if snap.Version != 1 {
		t.Fatalf("version = %d, want 1", snap.Version)
	}
	// ledger
	var count int64
	s.db.Raw("SELECT COUNT(*) FROM ledger_entries WHERE wallet_id = ?", w.ID).Scan(&count)
	if count != 1 {
		t.Fatalf("ledger count = %d, want 1", count)
	}
	var ba int64
	s.db.Raw("SELECT balance_after FROM ledger_entries WHERE wallet_id = ?", w.ID).Scan(&ba)
	if ba != 1000 {
		t.Fatalf("balance_after = %d, want 1000", ba)
	}
}

func TestCredit_SyncsRedisBalance(t *testing.T) {
	s, rdb := newTestStore(t, true)
	w, _ := s.GetOrCreate(ctx(), OwnerTypeUser, 2)
	if err := s.Credit(ctx(), w.ID, 500, RefTypeManual, nil, "", 7); err != nil {
		t.Fatal(err)
	}
	v, _ := rdb.HGet(ctx(), walletRedisKey(OwnerTypeUser, 2), hashFieldBalance).Result()
	if v != "500" {
		t.Fatalf("redis balance = %s, want 500", v)
	}
}

func TestCredit_RefundRefType_DoesNotIncreaseRecharged(t *testing.T) {
	s, _ := newTestStore(t, false)
	w, _ := s.GetOrCreate(ctx(), OwnerTypeUser, 3)
	if err := s.Credit(ctx(), w.ID, 100, RefTypeRefund, nil, "", 0); err != nil {
		t.Fatal(err)
	}
	snap, _ := s.Snapshot(ctx(), w.ID)
	if snap.QuotaTotalRecharged != 0 {
		t.Fatalf("recharged = %d, want 0 (refund should not count)", snap.QuotaTotalRecharged)
	}
	if snap.QuotaBalance != 100 {
		t.Fatalf("balance = %d, want 100", snap.QuotaBalance)
	}
}

func TestCredit_ConcurrentSameWallet_FinalBalanceCorrect(t *testing.T) {
	s, _ := newTestStore(t, false)
	w, _ := s.GetOrCreate(ctx(), OwnerTypeUser, 4)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = s.Credit(ctx(), w.ID, 50, RefTypeManual, nil, "", 0)
		}()
	}
	wg.Wait()
	snap, _ := s.Snapshot(ctx(), w.ID)
	if snap.QuotaBalance != 50*20 {
		t.Fatalf("balance = %d, want %d", snap.QuotaBalance, 50*20)
	}
}
