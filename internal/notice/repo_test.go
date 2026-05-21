package notice

import (
	"context"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// newRepoDB 启动一个 in-memory sqlite,建好 notices 表,返回 repo 与原始 DB。
func newRepoDB(t *testing.T) (*repo, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.Exec(`
		CREATE TABLE notices (
			id          INTEGER PRIMARY KEY,
			title       TEXT NOT NULL,
			content     TEXT NOT NULL,
			level       TEXT NOT NULL DEFAULT 'info',
			target      TEXT NOT NULL DEFAULT 'all',
			status      INTEGER NOT NULL DEFAULT 0,
			publish_at  DATETIME,
			expires_at  DATETIME,
			pinned      INTEGER NOT NULL DEFAULT 0,
			created_by  INTEGER NOT NULL,
			created_at  DATETIME NOT NULL,
			updated_at  DATETIME NOT NULL
		)
	`).Error; err != nil {
		t.Fatalf("create table: %v", err)
	}
	return &repo{db: db}, db
}

func mustNotice(t *testing.T, r Repo, n *Notice) *Notice {
	t.Helper()
	if err := r.Create(context.Background(), n); err != nil {
		t.Fatalf("Create: %v", err)
	}
	return n
}

func ptrTime(t time.Time) *time.Time { return &t }

func TestRepo_Create_AssignsRow(t *testing.T) {
	r, _ := newRepoDB(t)
	now := time.Now().UTC().Truncate(time.Second)
	n := &Notice{
		ID:        100,
		Title:     "hello",
		Content:   "world",
		Level:     "info",
		Target:    "all",
		Status:    0,
		CreatedBy: 1,
		CreatedAt: now,
		UpdatedAt: now,
	}
	mustNotice(t, r, n)
	got, err := r.GetByID(context.Background(), 100)
	if err != nil {
		t.Fatalf("GetByID err: %v", err)
	}
	if got == nil || got.Title != "hello" {
		t.Fatalf("got %+v", got)
	}
}

func TestRepo_GetByID_NotFoundReturnsNil(t *testing.T) {
	r, _ := newRepoDB(t)
	got, err := r.GetByID(context.Background(), 999)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got != nil {
		t.Fatalf("want nil, got %+v", got)
	}
}

func TestRepo_Update_OnlyChangesGivenFields(t *testing.T) {
	r, _ := newRepoDB(t)
	now := time.Now().UTC().Truncate(time.Second)
	mustNotice(t, r, &Notice{ID: 1, Title: "t1", Content: "c1", Level: "info", Target: "all", CreatedBy: 1, CreatedAt: now, UpdatedAt: now})

	if err := r.Update(context.Background(), 1, map[string]any{"title": "t2"}); err != nil {
		t.Fatalf("Update: %v", err)
	}
	got, _ := r.GetByID(context.Background(), 1)
	if got.Title != "t2" {
		t.Fatalf("title want t2, got %q", got.Title)
	}
	if got.Content != "c1" {
		t.Fatalf("content should be unchanged, got %q", got.Content)
	}
}

func TestRepo_Update_NotFoundReturnsError(t *testing.T) {
	r, _ := newRepoDB(t)
	err := r.Update(context.Background(), 999, map[string]any{"title": "x"})
	if err == nil {
		t.Fatal("want error for not found")
	}
}

func TestRepo_Delete_RemovesRow(t *testing.T) {
	r, _ := newRepoDB(t)
	now := time.Now().UTC().Truncate(time.Second)
	mustNotice(t, r, &Notice{ID: 1, Title: "t", Content: "c", Level: "info", Target: "all", CreatedBy: 1, CreatedAt: now, UpdatedAt: now})
	if err := r.Delete(context.Background(), 1); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	got, _ := r.GetByID(context.Background(), 1)
	if got != nil {
		t.Fatalf("want nil after delete, got %+v", got)
	}
}

func TestRepo_ListAdmin_FiltersByStatus(t *testing.T) {
	r, _ := newRepoDB(t)
	now := time.Now().UTC().Truncate(time.Second)
	mustNotice(t, r, &Notice{ID: 1, Title: "draft", Content: "c", Level: "info", Target: "all", Status: 0, CreatedBy: 1, CreatedAt: now, UpdatedAt: now})
	mustNotice(t, r, &Notice{ID: 2, Title: "pub", Content: "c", Level: "info", Target: "all", Status: 1, CreatedBy: 1, CreatedAt: now, UpdatedAt: now})
	mustNotice(t, r, &Notice{ID: 3, Title: "arch", Content: "c", Level: "info", Target: "all", Status: 2, CreatedBy: 1, CreatedAt: now, UpdatedAt: now})

	items, total, err := r.ListAdmin(context.Background(), 1, 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(items) != 1 || items[0].ID != 2 {
		t.Fatalf("want one (id=2), got total=%d items=%+v", total, items)
	}
}

func TestRepo_ListAdmin_StatusMinusOne_NoFilter(t *testing.T) {
	r, _ := newRepoDB(t)
	now := time.Now().UTC().Truncate(time.Second)
	mustNotice(t, r, &Notice{ID: 1, Title: "a", Content: "c", Level: "info", Target: "all", Status: 0, CreatedBy: 1, CreatedAt: now, UpdatedAt: now})
	mustNotice(t, r, &Notice{ID: 2, Title: "b", Content: "c", Level: "info", Target: "all", Status: 1, CreatedBy: 1, CreatedAt: now, UpdatedAt: now})

	items, total, err := r.ListAdmin(context.Background(), -1, 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 || len(items) != 2 {
		t.Fatalf("want 2, got total=%d len=%d", total, len(items))
	}
}

func TestRepo_ListAdmin_OrderByPinnedAndPublishAt(t *testing.T) {
	r, _ := newRepoDB(t)
	now := time.Now().UTC().Truncate(time.Second)
	t1 := now.Add(-2 * time.Hour)
	t2 := now.Add(-1 * time.Hour)
	mustNotice(t, r, &Notice{ID: 1, Title: "old", Content: "c", Level: "info", Target: "all", Status: 1, PublishAt: &t1, CreatedBy: 1, CreatedAt: now, UpdatedAt: now})
	mustNotice(t, r, &Notice{ID: 2, Title: "new", Content: "c", Level: "info", Target: "all", Status: 1, PublishAt: &t2, CreatedBy: 1, CreatedAt: now, UpdatedAt: now})
	mustNotice(t, r, &Notice{ID: 3, Title: "pinned-old", Content: "c", Level: "info", Target: "all", Status: 1, PublishAt: &t1, Pinned: true, CreatedBy: 1, CreatedAt: now, UpdatedAt: now})

	items, _, err := r.ListAdmin(context.Background(), -1, 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if items[0].ID != 3 {
		t.Fatalf("first should be pinned (id=3), got id=%d", items[0].ID)
	}
	if items[1].ID != 2 {
		t.Fatalf("second should be newer publish (id=2), got id=%d", items[1].ID)
	}
}

func TestRepo_ListVisibleForUser_ExcludesDraft(t *testing.T) {
	r, _ := newRepoDB(t)
	now := time.Now().UTC().Truncate(time.Second)
	t1 := now.Add(-1 * time.Hour)
	mustNotice(t, r, &Notice{ID: 1, Title: "draft", Content: "c", Level: "info", Target: "all", Status: 0, PublishAt: &t1, CreatedBy: 1, CreatedAt: now, UpdatedAt: now})
	mustNotice(t, r, &Notice{ID: 2, Title: "pub", Content: "c", Level: "info", Target: "all", Status: 1, PublishAt: &t1, CreatedBy: 1, CreatedAt: now, UpdatedAt: now})

	items, _, err := r.ListVisibleForUser(context.Background(), []string{"all", "user"}, now, 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].ID != 2 {
		t.Fatalf("want [2], got %+v", items)
	}
}

func TestRepo_ListVisibleForUser_ExcludesExpired(t *testing.T) {
	r, _ := newRepoDB(t)
	now := time.Now().UTC().Truncate(time.Second)
	past := now.Add(-1 * time.Hour)
	future := now.Add(1 * time.Hour)
	mustNotice(t, r, &Notice{ID: 1, Title: "expired", Content: "c", Level: "info", Target: "all", Status: 1, PublishAt: ptrTime(now.Add(-2 * time.Hour)), ExpiresAt: &past, CreatedBy: 1, CreatedAt: now, UpdatedAt: now})
	mustNotice(t, r, &Notice{ID: 2, Title: "valid", Content: "c", Level: "info", Target: "all", Status: 1, PublishAt: ptrTime(now.Add(-2 * time.Hour)), ExpiresAt: &future, CreatedBy: 1, CreatedAt: now, UpdatedAt: now})

	items, _, err := r.ListVisibleForUser(context.Background(), []string{"all", "user"}, now, 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].ID != 2 {
		t.Fatalf("want [2], got %+v", items)
	}
}

func TestRepo_ListVisibleForUser_ExcludesAdminTargetForUser(t *testing.T) {
	r, _ := newRepoDB(t)
	now := time.Now().UTC().Truncate(time.Second)
	t1 := now.Add(-1 * time.Hour)
	mustNotice(t, r, &Notice{ID: 1, Title: "admin-only", Content: "c", Level: "info", Target: "admin", Status: 1, PublishAt: &t1, CreatedBy: 1, CreatedAt: now, UpdatedAt: now})
	mustNotice(t, r, &Notice{ID: 2, Title: "all", Content: "c", Level: "info", Target: "all", Status: 1, PublishAt: &t1, CreatedBy: 1, CreatedAt: now, UpdatedAt: now})

	items, _, err := r.ListVisibleForUser(context.Background(), []string{"all", "user"}, now, 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].ID != 2 {
		t.Fatalf("want [2], got %+v", items)
	}
}

func TestRepo_ListVisibleForUser_IncludesNullPublishAt(t *testing.T) {
	r, _ := newRepoDB(t)
	now := time.Now().UTC().Truncate(time.Second)
	mustNotice(t, r, &Notice{ID: 1, Title: "no-publishat", Content: "c", Level: "info", Target: "all", Status: 1, CreatedBy: 1, CreatedAt: now, UpdatedAt: now})

	items, _, err := r.ListVisibleForUser(context.Background(), []string{"all", "user"}, now, 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].ID != 1 {
		t.Fatalf("want [1], got %+v", items)
	}
}

func TestRepo_VisibleIDsForUser_ReturnsOnlyIDs(t *testing.T) {
	r, _ := newRepoDB(t)
	now := time.Now().UTC().Truncate(time.Second)
	t1 := now.Add(-1 * time.Hour)
	mustNotice(t, r, &Notice{ID: 10, Title: "a", Content: "c", Level: "info", Target: "all", Status: 1, PublishAt: &t1, CreatedBy: 1, CreatedAt: now, UpdatedAt: now})
	mustNotice(t, r, &Notice{ID: 20, Title: "b", Content: "c", Level: "info", Target: "user", Status: 1, PublishAt: &t1, CreatedBy: 1, CreatedAt: now, UpdatedAt: now})
	mustNotice(t, r, &Notice{ID: 30, Title: "draft", Content: "c", Level: "info", Target: "all", Status: 0, PublishAt: &t1, CreatedBy: 1, CreatedAt: now, UpdatedAt: now})

	ids, err := r.VisibleIDsForUser(context.Background(), []string{"all", "user"}, now)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 2 {
		t.Fatalf("want 2 ids, got %v", ids)
	}
}

func TestRepo_CountVisibleForUser_Matches(t *testing.T) {
	r, _ := newRepoDB(t)
	now := time.Now().UTC().Truncate(time.Second)
	t1 := now.Add(-1 * time.Hour)
	mustNotice(t, r, &Notice{ID: 1, Title: "a", Content: "c", Level: "info", Target: "all", Status: 1, PublishAt: &t1, CreatedBy: 1, CreatedAt: now, UpdatedAt: now})
	mustNotice(t, r, &Notice{ID: 2, Title: "b", Content: "c", Level: "info", Target: "user", Status: 1, PublishAt: &t1, CreatedBy: 1, CreatedAt: now, UpdatedAt: now})

	count, err := r.CountVisibleForUser(context.Background(), []string{"all", "user"}, now)
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("want 2, got %d", count)
	}
}
