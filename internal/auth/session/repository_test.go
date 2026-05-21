package session

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newSessionDB(t *testing.T) *gorm.DB {
	t.Helper()
	dbName := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", dbName)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			user_id INTEGER NOT NULL,
			ip TEXT NOT NULL DEFAULT '',
			user_agent TEXT NOT NULL DEFAULT '',
			created_at DATETIME NOT NULL,
			last_seen_at DATETIME NOT NULL,
			expires_at DATETIME NOT NULL,
			revoked_at DATETIME
		)
	`).Error; err != nil {
		t.Fatal(err)
	}
	return db
}

func TestRepository_InsertAndGet(t *testing.T) {
	db := newSessionDB(t)
	repo := NewRepository(db)
	now := time.Now().UTC().Truncate(time.Second)
	s := &DBSession{ID: "sess_abc", UserID: 1, CreatedAt: now, LastSeenAt: now, ExpiresAt: now.Add(time.Hour)}
	if err := repo.Insert(context.Background(), s); err != nil {
		t.Fatal(err)
	}
	got, _ := repo.Get(context.Background(), "sess_abc")
	if got == nil || got.UserID != 1 {
		t.Fatalf("got %+v", got)
	}
}

func TestRepository_UpdateLastSeen(t *testing.T) {
	db := newSessionDB(t)
	repo := NewRepository(db)
	now := time.Now().UTC().Truncate(time.Second)
	_ = repo.Insert(context.Background(), &DBSession{ID: "x", UserID: 1, CreatedAt: now, LastSeenAt: now, ExpiresAt: now.Add(time.Hour)})
	newTime := now.Add(30 * time.Minute)
	if err := repo.UpdateLastSeen(context.Background(), "x", newTime, newTime.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	got, _ := repo.Get(context.Background(), "x")
	if !got.LastSeenAt.Equal(newTime) {
		t.Fatalf("not updated: %v", got.LastSeenAt)
	}
}

func TestRepository_MarkRevoked(t *testing.T) {
	db := newSessionDB(t)
	repo := NewRepository(db)
	now := time.Now().UTC().Truncate(time.Second)
	_ = repo.Insert(context.Background(), &DBSession{ID: "x", UserID: 1, CreatedAt: now, LastSeenAt: now, ExpiresAt: now.Add(time.Hour)})
	_ = repo.MarkRevoked(context.Background(), "x", now)
	got, _ := repo.Get(context.Background(), "x")
	if got.RevokedAt == nil {
		t.Fatal("want revoked_at set")
	}
}

func TestRepository_MarkAllRevokedForUser(t *testing.T) {
	db := newSessionDB(t)
	repo := NewRepository(db)
	now := time.Now().UTC().Truncate(time.Second)
	_ = repo.Insert(context.Background(), &DBSession{ID: "a", UserID: 1, CreatedAt: now, LastSeenAt: now, ExpiresAt: now.Add(time.Hour)})
	_ = repo.Insert(context.Background(), &DBSession{ID: "b", UserID: 1, CreatedAt: now, LastSeenAt: now, ExpiresAt: now.Add(time.Hour)})
	_ = repo.Insert(context.Background(), &DBSession{ID: "c", UserID: 2, CreatedAt: now, LastSeenAt: now, ExpiresAt: now.Add(time.Hour)})
	_ = repo.MarkAllRevokedForUser(context.Background(), 1, now)
	a, _ := repo.Get(context.Background(), "a")
	c, _ := repo.Get(context.Background(), "c")
	if a.RevokedAt == nil {
		t.Fatal("a should be revoked")
	}
	if c.RevokedAt != nil {
		t.Fatal("c should not be revoked")
	}
}

func TestRepository_ListActive(t *testing.T) {
	db := newSessionDB(t)
	repo := NewRepository(db)
	now := time.Now().UTC().Truncate(time.Second)
	_ = repo.Insert(context.Background(), &DBSession{ID: "a", UserID: 1, CreatedAt: now, LastSeenAt: now, ExpiresAt: now.Add(time.Hour)})
	_ = repo.Insert(context.Background(), &DBSession{ID: "b", UserID: 1, CreatedAt: now, LastSeenAt: now, ExpiresAt: now.Add(-time.Hour)})
	items, err := repo.ListActive(context.Background(), now, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].ID != "a" {
		t.Fatalf("ListActive wrong: %+v", items)
	}
}

func TestRepository_DeleteExpired(t *testing.T) {
	db := newSessionDB(t)
	repo := NewRepository(db)
	now := time.Now().UTC().Truncate(time.Second)
	_ = repo.Insert(context.Background(), &DBSession{ID: "old", UserID: 1, CreatedAt: now, LastSeenAt: now, ExpiresAt: now.Add(-8 * 24 * time.Hour)})
	_ = repo.Insert(context.Background(), &DBSession{ID: "new", UserID: 1, CreatedAt: now, LastSeenAt: now, ExpiresAt: now.Add(time.Hour)})
	n, err := repo.DeleteExpired(context.Background(), now.Add(-7*24*time.Hour), 100)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("want 1 deleted, got %d", n)
	}
}
