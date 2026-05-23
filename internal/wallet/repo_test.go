package wallet

import (
	"errors"
	"sync"
	"testing"

	"github.com/ijry/pro-api/pkg/apierr"
)

func TestRepo_GetOrCreate_CreatesNewWallet(t *testing.T) {
	s, _ := newTestStore(t, false)
	w, err := s.GetOrCreate(ctx(), OwnerTypeUser, 100)
	if err != nil {
		t.Fatal(err)
	}
	if w.ID == 0 || w.OwnerID != 100 || w.OwnerType != OwnerTypeUser {
		t.Fatalf("bad wallet: %+v", w)
	}
	if w.QuotaBalance != 0 || w.Currency != "USD" {
		t.Fatalf("bad defaults: %+v", w)
	}
}

func TestRepo_GetOrCreate_ReturnsExistingWallet(t *testing.T) {
	s, _ := newTestStore(t, false)
	w1, err := s.GetOrCreate(ctx(), OwnerTypeUser, 200)
	if err != nil {
		t.Fatal(err)
	}
	w2, err := s.GetOrCreate(ctx(), OwnerTypeUser, 200)
	if err != nil {
		t.Fatal(err)
	}
	if w1.ID != w2.ID {
		t.Fatalf("expected same id, got %d vs %d", w1.ID, w2.ID)
	}
}

func TestRepo_GetOrCreate_ConcurrentSameOwner_NoDuplicate(t *testing.T) {
	s, _ := newTestStore(t, false)
	var wg sync.WaitGroup
	ids := make([]int64, 10)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			w, err := s.GetOrCreate(ctx(), OwnerTypeUser, 300)
			if err == nil {
				ids[idx] = w.ID
			}
		}(i)
	}
	wg.Wait()
	first := ids[0]
	if first == 0 {
		t.Fatal("no wallet created")
	}
	for _, id := range ids {
		if id != first {
			t.Fatalf("expected all ids %d, got %d", first, id)
		}
	}
}

func TestRepo_Snapshot_ReturnsFromDB(t *testing.T) {
	s, _ := newTestStore(t, false)
	w, _ := s.GetOrCreate(ctx(), OwnerTypeUser, 400)
	got, err := s.Snapshot(ctx(), w.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != w.ID {
		t.Fatalf("mismatch")
	}
}

func TestRepo_Snapshot_NotFound(t *testing.T) {
	s, _ := newTestStore(t, false)
	_, err := s.Snapshot(ctx(), 999)
	var ae *apierr.Error
	if !errors.As(err, &ae) || ae.Code != apierr.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %v", err)
	}
}

func TestRepo_ByOwner_NotFound(t *testing.T) {
	s, _ := newTestStore(t, false)
	_, err := s.ByOwner(ctx(), OwnerTypeUser, 999)
	var ae *apierr.Error
	if !errors.As(err, &ae) || ae.Code != apierr.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %v", err)
	}
}

func TestRepo_GetOrCreate_InvalidParam(t *testing.T) {
	s, _ := newTestStore(t, false)
	_, err := s.GetOrCreate(ctx(), "", 1)
	if err == nil {
		t.Fatal("expected error for empty owner_type")
	}
	_, err = s.GetOrCreate(ctx(), OwnerTypeUser, 0)
	if err == nil {
		t.Fatal("expected error for owner_id=0")
	}
}
