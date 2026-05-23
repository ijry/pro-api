package redeem

import (
	"context"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newRepoDB(t *testing.T) (*repo, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
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
		t.Fatalf("create table: %v", err)
	}
	return &repo{db: db}, db
}

func ptrTime(t time.Time) *time.Time { return &t }
func ptrInt64(v int64) *int64        { return &v }

func TestRepo_BatchInsert_PersistsAll(t *testing.T) {
	r, _ := newRepoDB(t)
	now := time.Now().UTC().Truncate(time.Millisecond)
	codes := []*Code{
		{ID: 1, CodeHash: "h1", CodePrefix: "AAAA", AmountQuota: 100, BatchNo: "B1", Status: StatusUnused, CreatedBy: 1, CreatedAt: now},
		{ID: 2, CodeHash: "h2", CodePrefix: "BBBB", AmountQuota: 100, BatchNo: "B1", Status: StatusUnused, CreatedBy: 1, CreatedAt: now},
	}
	if err := r.BatchInsert(context.Background(), codes); err != nil {
		t.Fatalf("BatchInsert: %v", err)
	}
	c, err := r.GetByHash(context.Background(), "h1")
	if err != nil {
		t.Fatalf("GetByHash: %v", err)
	}
	if c == nil || c.CodePrefix != "AAAA" {
		t.Fatalf("got %+v", c)
	}
}

func TestRepo_BatchInsert_DuplicateHash_Errors(t *testing.T) {
	r, _ := newRepoDB(t)
	now := time.Now().UTC().Truncate(time.Millisecond)
	codes := []*Code{
		{ID: 1, CodeHash: "h1", CodePrefix: "A", AmountQuota: 100, Status: StatusUnused, CreatedBy: 1, CreatedAt: now},
	}
	if err := r.BatchInsert(context.Background(), codes); err != nil {
		t.Fatalf("BatchInsert: %v", err)
	}
	// 重新插同样 hash → 应当报错(由调用方处理唯一约束冲突)
	dup := []*Code{
		{ID: 2, CodeHash: "h1", CodePrefix: "B", AmountQuota: 100, Status: StatusUnused, CreatedBy: 1, CreatedAt: now},
	}
	if err := r.BatchInsert(context.Background(), dup); err == nil {
		t.Fatal("want unique constraint error")
	}
}

func TestRepo_GetByHash_NotFound_ReturnsNil(t *testing.T) {
	r, _ := newRepoDB(t)
	c, err := r.GetByHash(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if c != nil {
		t.Fatalf("want nil")
	}
}

func TestRepo_GetByID(t *testing.T) {
	r, _ := newRepoDB(t)
	now := time.Now().UTC().Truncate(time.Millisecond)
	mustInsert(t, r, &Code{ID: 1, CodeHash: "h1", CodePrefix: "AAAA", AmountQuota: 100, CreatedBy: 1, CreatedAt: now})
	c, err := r.GetByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if c == nil || c.ID != 1 {
		t.Fatalf("got %+v", c)
	}
}

func TestRepo_UpdateToUsed_OnlyFromUnused(t *testing.T) {
	r, _ := newRepoDB(t)
	now := time.Now().UTC().Truncate(time.Millisecond)
	mustInsert(t, r, &Code{ID: 1, CodeHash: "h1", CodePrefix: "AAAA", AmountQuota: 100, Status: StatusUnused, CreatedBy: 1, CreatedAt: now})
	mustInsert(t, r, &Code{ID: 2, CodeHash: "h2", CodePrefix: "BBBB", AmountQuota: 100, Status: StatusUsed, CreatedBy: 1, CreatedAt: now, UsedBy: ptrInt64(7), UsedAt: ptrTime(now)})

	uid := int64(42)
	n, err := r.UpdateToUsed(context.Background(), 1, uid, now)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n != 1 {
		t.Fatalf("want 1, got %d", n)
	}
	got, _ := r.GetByID(context.Background(), 1)
	if got.Status != StatusUsed || got.UsedBy == nil || *got.UsedBy != 42 {
		t.Fatalf("not properly updated: %+v", got)
	}

	// 已 used 的 id=2,再次尝试 used 应当 0 rows
	n, err = r.UpdateToUsed(context.Background(), 2, uid, now)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n != 0 {
		t.Fatalf("want 0 (already used), got %d", n)
	}
}

func TestRepo_UpdateToDisabled_OnlyFromUnused(t *testing.T) {
	r, _ := newRepoDB(t)
	now := time.Now().UTC().Truncate(time.Millisecond)
	mustInsert(t, r, &Code{ID: 1, CodeHash: "h1", AmountQuota: 100, Status: StatusUnused, CreatedBy: 1, CreatedAt: now})
	mustInsert(t, r, &Code{ID: 2, CodeHash: "h2", AmountQuota: 100, Status: StatusUsed, CreatedBy: 1, CreatedAt: now})
	mustInsert(t, r, &Code{ID: 3, CodeHash: "h3", AmountQuota: 100, Status: StatusDisabled, CreatedBy: 1, CreatedAt: now})

	n, err := r.UpdateToDisabledBulk(context.Background(), []int64{1, 2, 3})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n != 1 {
		t.Fatalf("want only 1 affected (unused), got %d", n)
	}
}

func TestRepo_UpdateUnusedRollback(t *testing.T) {
	r, _ := newRepoDB(t)
	now := time.Now().UTC().Truncate(time.Millisecond)
	mustInsert(t, r, &Code{ID: 1, CodeHash: "h1", AmountQuota: 100, Status: StatusUsed, CreatedBy: 1, CreatedAt: now, UsedBy: ptrInt64(7), UsedAt: ptrTime(now)})

	if err := r.RollbackUsedToUnused(context.Background(), 1); err != nil {
		t.Fatalf("err: %v", err)
	}
	c, _ := r.GetByID(context.Background(), 1)
	if c.Status != StatusUnused {
		t.Fatalf("want unused, got %d", c.Status)
	}
	if c.UsedBy != nil {
		t.Fatalf("used_by should be nil")
	}
}

func TestRepo_List_FilterByBatchNo(t *testing.T) {
	r, _ := newRepoDB(t)
	now := time.Now().UTC().Truncate(time.Millisecond)
	mustInsert(t, r, &Code{ID: 1, CodeHash: "h1", BatchNo: "A", AmountQuota: 100, Status: StatusUnused, CreatedBy: 1, CreatedAt: now})
	mustInsert(t, r, &Code{ID: 2, CodeHash: "h2", BatchNo: "B", AmountQuota: 100, Status: StatusUnused, CreatedBy: 1, CreatedAt: now})

	items, total, err := r.List(context.Background(), ListFilter{BatchNo: "A", Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Fatalf("filter wrong: total=%d items=%d", total, len(items))
	}
}

func TestRepo_List_FilterByStatus(t *testing.T) {
	r, _ := newRepoDB(t)
	now := time.Now().UTC().Truncate(time.Millisecond)
	mustInsert(t, r, &Code{ID: 1, CodeHash: "h1", AmountQuota: 100, Status: StatusUnused, CreatedBy: 1, CreatedAt: now})
	mustInsert(t, r, &Code{ID: 2, CodeHash: "h2", AmountQuota: 100, Status: StatusUsed, CreatedBy: 1, CreatedAt: now})

	s := StatusUsed
	items, total, err := r.List(context.Background(), ListFilter{Status: &s, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if total != 1 || items[0].ID != 2 {
		t.Fatalf("got total=%d items=%v", total, items)
	}
}

func TestRepo_ListAll_NoPagination(t *testing.T) {
	r, _ := newRepoDB(t)
	now := time.Now().UTC().Truncate(time.Millisecond)
	for i := int64(1); i <= 5; i++ {
		mustInsert(t, r, &Code{ID: i, CodeHash: string(rune('a' + i)), AmountQuota: 100, CreatedBy: 1, CreatedAt: now})
	}
	items, err := r.ListAll(context.Background(), ListFilter{})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(items) != 5 {
		t.Fatalf("want 5, got %d", len(items))
	}
}

func mustInsert(t *testing.T, r Repo, c *Code) {
	t.Helper()
	if err := r.BatchInsert(context.Background(), []*Code{c}); err != nil {
		t.Fatalf("BatchInsert: %v", err)
	}
}
