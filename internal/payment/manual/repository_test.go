package manual

import (
	"context"
	"errors"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// newRepoDB 启动 in-memory sqlite,建好 manual_recharges 表,返回 repo 与原始 DB。
func newRepoDB(t *testing.T) (*repo, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
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
		t.Fatalf("create table: %v", err)
	}
	return &repo{db: db}, db
}

func ptrTime(t time.Time) *time.Time { return &t }
func ptrInt64(v int64) *int64        { return &v }

func TestRepo_Create_PersistsAllFields(t *testing.T) {
	r, _ := newRepoDB(t)
	now := time.Now().UTC().Truncate(time.Millisecond)
	rec := &Recharge{
		ID:            1001,
		UserID:        42,
		AmountMoney:   1_000_000,
		Currency:      "CNY",
		Status:        StatusPending,
		ApplicantNote: "已转账",
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := r.Create(context.Background(), rec); err != nil {
		t.Fatalf("Create: %v", err)
	}
	got, err := r.GetByID(context.Background(), 1001)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got == nil {
		t.Fatal("got nil")
	}
	if got.UserID != 42 || got.AmountMoney != 1_000_000 || got.Currency != "CNY" {
		t.Errorf("fields mismatch: %+v", got)
	}
	if got.Status != StatusPending {
		t.Errorf("status mismatch")
	}
}

func TestRepo_GetByID_NotFound_ReturnsNil(t *testing.T) {
	r, _ := newRepoDB(t)
	got, err := r.GetByID(context.Background(), 999)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got != nil {
		t.Fatalf("want nil, got %+v", got)
	}
}

func TestRepo_UpdateStatusFromPending_OnlyAffectsPending(t *testing.T) {
	r, _ := newRepoDB(t)
	now := time.Now().UTC().Truncate(time.Millisecond)
	mustInsert(t, r, &Recharge{ID: 1, UserID: 1, AmountMoney: 100, Status: StatusPending, CreatedAt: now, UpdatedAt: now})
	mustInsert(t, r, &Recharge{ID: 2, UserID: 1, AmountMoney: 100, Status: StatusApproved, CreatedAt: now, UpdatedAt: now})

	// 转 1 → approved 成功
	rowsAffected, err := r.UpdateStatusFromPending(context.Background(), 1, map[string]any{
		"status":     StatusApproved,
		"updated_at": now,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if rowsAffected != 1 {
		t.Fatalf("want 1 affected, got %d", rowsAffected)
	}
	// 转 2 → approved 失败(已是 approved)
	rowsAffected, err = r.UpdateStatusFromPending(context.Background(), 2, map[string]any{
		"status":     StatusApproved,
		"updated_at": now,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if rowsAffected != 0 {
		t.Fatalf("want 0 affected (not pending), got %d", rowsAffected)
	}
}

func TestRepo_Update_Fields(t *testing.T) {
	r, _ := newRepoDB(t)
	now := time.Now().UTC().Truncate(time.Millisecond)
	mustInsert(t, r, &Recharge{ID: 1, UserID: 1, AmountMoney: 100, Status: StatusPending, CreatedAt: now, UpdatedAt: now})

	if err := r.Update(context.Background(), 1, map[string]any{
		"status":      StatusCanceled,
		"updated_at":  now,
		"review_note": "",
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}
	got, _ := r.GetByID(context.Background(), 1)
	if got.Status != StatusCanceled {
		t.Errorf("want canceled, got %d", got.Status)
	}
}

func TestRepo_ListByUser_FilterByUserID(t *testing.T) {
	r, _ := newRepoDB(t)
	now := time.Now().UTC().Truncate(time.Millisecond)
	mustInsert(t, r, &Recharge{ID: 1, UserID: 10, AmountMoney: 100, Status: StatusPending, CreatedAt: now, UpdatedAt: now})
	mustInsert(t, r, &Recharge{ID: 2, UserID: 10, AmountMoney: 200, Status: StatusApproved, CreatedAt: now.Add(time.Second), UpdatedAt: now})
	mustInsert(t, r, &Recharge{ID: 3, UserID: 11, AmountMoney: 300, Status: StatusPending, CreatedAt: now.Add(2 * time.Second), UpdatedAt: now})

	items, total, err := r.List(context.Background(), ListFilter{UserID: 10, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total != 2 {
		t.Fatalf("want total 2, got %d", total)
	}
	if len(items) != 2 {
		t.Fatalf("want 2 items, got %d", len(items))
	}
}

func TestRepo_List_FilterByStatus(t *testing.T) {
	r, _ := newRepoDB(t)
	now := time.Now().UTC().Truncate(time.Millisecond)
	mustInsert(t, r, &Recharge{ID: 1, UserID: 10, AmountMoney: 100, Status: StatusPending, CreatedAt: now, UpdatedAt: now})
	mustInsert(t, r, &Recharge{ID: 2, UserID: 11, AmountMoney: 200, Status: StatusApproved, CreatedAt: now.Add(time.Second), UpdatedAt: now})

	status := StatusPending
	items, total, err := r.List(context.Background(), ListFilter{Status: &status, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total != 1 {
		t.Fatalf("want 1, got %d", total)
	}
	if items[0].ID != 1 {
		t.Fatalf("want id 1, got %d", items[0].ID)
	}
}

func TestRepo_List_OrderByCreatedAtDesc(t *testing.T) {
	r, _ := newRepoDB(t)
	now := time.Now().UTC().Truncate(time.Millisecond)
	mustInsert(t, r, &Recharge{ID: 1, UserID: 1, AmountMoney: 100, Status: StatusPending, CreatedAt: now, UpdatedAt: now})
	mustInsert(t, r, &Recharge{ID: 2, UserID: 1, AmountMoney: 200, Status: StatusPending, CreatedAt: now.Add(2 * time.Second), UpdatedAt: now})

	items, _, err := r.List(context.Background(), ListFilter{UserID: 1, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if items[0].ID != 2 || items[1].ID != 1 {
		t.Fatalf("want desc order [2,1], got [%d,%d]", items[0].ID, items[1].ID)
	}
}

func TestRepo_List_Pagination(t *testing.T) {
	r, _ := newRepoDB(t)
	now := time.Now().UTC().Truncate(time.Millisecond)
	for i := int64(1); i <= 5; i++ {
		mustInsert(t, r, &Recharge{ID: i, UserID: 1, AmountMoney: i * 100, Status: StatusPending, CreatedAt: now.Add(time.Duration(i) * time.Second), UpdatedAt: now})
	}

	items, total, err := r.List(context.Background(), ListFilter{UserID: 1, Page: 2, PageSize: 2})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total != 5 {
		t.Fatalf("want total 5, got %d", total)
	}
	if len(items) != 2 {
		t.Fatalf("want 2 items, got %d", len(items))
	}
	if items[0].ID != 3 || items[1].ID != 2 {
		t.Fatalf("pagination wrong: %d %d", items[0].ID, items[1].ID)
	}
}

func mustInsert(t *testing.T, r Repo, rec *Recharge) {
	t.Helper()
	if err := r.Create(context.Background(), rec); err != nil {
		t.Fatalf("Create: %v", err)
	}
}

// 兜底:errors 包还是要导入(防止编译失败但未引用)
var _ = errors.New
