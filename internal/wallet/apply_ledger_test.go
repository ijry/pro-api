package wallet

import (
	"sync"
	"testing"
)

func TestApplyLedgerBatch_SingleDebit_Consistent(t *testing.T) {
	s, _ := newTestStore(t, false)
	w, _ := s.GetOrCreate(ctx(), OwnerTypeUser, 1)
	_ = s.Credit(ctx(), w.ID, 1000, RefTypeManual, nil, "", 0)
	err := s.ApplyLedgerBatch(ctx(), []LedgerEvent{{
		WalletID:    w.ID,
		Direction:   DirectionDebit,
		AmountQuota: 300,
		RefType:     RefTypeUsage,
		Description: "x",
	}})
	if err != nil {
		t.Fatal(err)
	}
	snap, _ := s.Snapshot(ctx(), w.ID)
	if snap.QuotaBalance != 700 {
		t.Fatalf("balance = %d, want 700", snap.QuotaBalance)
	}
	if snap.QuotaTotalConsumed != 300 {
		t.Fatalf("consumed = %d, want 300", snap.QuotaTotalConsumed)
	}
}

func TestApplyLedgerBatch_MultipleEvents_BalanceAfterChained(t *testing.T) {
	s, _ := newTestStore(t, false)
	w, _ := s.GetOrCreate(ctx(), OwnerTypeUser, 2)
	_ = s.Credit(ctx(), w.ID, 1000, RefTypeManual, nil, "", 0)
	events := []LedgerEvent{
		{WalletID: w.ID, Direction: DirectionDebit, AmountQuota: 200, RefType: RefTypeUsage},
		{WalletID: w.ID, Direction: DirectionDebit, AmountQuota: 100, RefType: RefTypeUsage},
		{WalletID: w.ID, Direction: DirectionCredit, AmountQuota: 50, RefType: RefTypeRefund},
	}
	if err := s.ApplyLedgerBatch(ctx(), events); err != nil {
		t.Fatal(err)
	}
	snap, _ := s.Snapshot(ctx(), w.ID)
	if snap.QuotaBalance != 1000-200-100+50 {
		t.Fatalf("balance = %d, want %d", snap.QuotaBalance, 750)
	}
	if snap.QuotaTotalConsumed != 300 {
		t.Fatalf("consumed = %d, want 300", snap.QuotaTotalConsumed)
	}
	// ledger 中 balance_after 应该按顺序串
	type row struct {
		AmountQuota  int64
		BalanceAfter int64
		Direction    string
	}
	var rows []row
	s.db.Raw("SELECT amount_quota AS amount_quota, balance_after AS balance_after, direction AS direction FROM ledger_entries WHERE wallet_id = ? AND ref_type IN ('usage','refund') ORDER BY id", w.ID).Scan(&rows)
	if len(rows) != 3 {
		t.Fatalf("ledger count = %d, want 3", len(rows))
	}
	want := []int64{800, 700, 750}
	for i, r := range rows {
		if r.BalanceAfter != want[i] {
			t.Fatalf("row %d balance_after = %d, want %d", i, r.BalanceAfter, want[i])
		}
	}
}

func TestApplyLedgerBatch_EmptyEvents_NoOp(t *testing.T) {
	s, _ := newTestStore(t, false)
	if err := s.ApplyLedgerBatch(ctx(), nil); err != nil {
		t.Fatal(err)
	}
}

func TestApplyLedgerBatch_MixedWalletIDs_Rejected(t *testing.T) {
	s, _ := newTestStore(t, false)
	err := s.ApplyLedgerBatch(ctx(), []LedgerEvent{
		{WalletID: 1, Direction: DirectionDebit, AmountQuota: 10, RefType: RefTypeUsage},
		{WalletID: 2, Direction: DirectionDebit, AmountQuota: 10, RefType: RefTypeUsage},
	})
	if err == nil {
		t.Fatal("expected error for mixed wallet_id")
	}
}

func TestApplyLedgerBatch_InvalidDirection_Rejected(t *testing.T) {
	s, _ := newTestStore(t, false)
	err := s.ApplyLedgerBatch(ctx(), []LedgerEvent{
		{WalletID: 1, Direction: "weird", AmountQuota: 10, RefType: RefTypeUsage},
	})
	if err == nil {
		t.Fatal("expected error for bad direction")
	}
}

func TestApplyLedgerBatch_WalletLocked_NoLostUpdate(t *testing.T) {
	// 简化:并发 100 笔 -10 debit,期望最终 balance = initial - 1000
	s, _ := newTestStore(t, false)
	w, _ := s.GetOrCreate(ctx(), OwnerTypeUser, 100)
	_ = s.Credit(ctx(), w.ID, 100000, RefTypeManual, nil, "", 0)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = s.ApplyLedgerBatch(ctx(), []LedgerEvent{{
				WalletID: w.ID, Direction: DirectionDebit, AmountQuota: 10, RefType: RefTypeUsage,
			}})
		}()
	}
	wg.Wait()
	snap, _ := s.Snapshot(ctx(), w.ID)
	if snap.QuotaBalance != 100000-1000 {
		t.Fatalf("balance = %d, want %d", snap.QuotaBalance, 99000)
	}
	if snap.QuotaTotalConsumed != 1000 {
		t.Fatalf("consumed = %d, want 1000", snap.QuotaTotalConsumed)
	}
}
