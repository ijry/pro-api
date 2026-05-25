package user

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newUserDB(t *testing.T) *gorm.DB {
	t.Helper()
	dbName := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", dbName)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY,
			username TEXT NOT NULL UNIQUE,
			email TEXT UNIQUE,
			password_hash TEXT,
			display_name TEXT,
			avatar TEXT,
			role INTEGER NOT NULL DEFAULT 0,
			status INTEGER NOT NULL DEFAULT 0,
			group_id INTEGER,
			invite_code TEXT UNIQUE,
			invited_by INTEGER NOT NULL DEFAULT 0,
			email_verified_at DATETIME,
			last_login_at DATETIME,
			last_login_ip TEXT,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			deleted_at DATETIME
		)
	`).Error; err != nil {
		t.Fatal(err)
	}
	return db
}

func ptr[T any](v T) *T { return &v }

func TestRepository_CreateAndGetByID(t *testing.T) {
	db := newUserDB(t)
	repo := NewRepository(db)
	now := time.Now().UTC()
	u := &User{
		ID:        123,
		Username:  "alice",
		Email:     ptr("alice@example.com"),
		GroupID:   ptr(int64(1)),
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := repo.Create(context.Background(), u); err != nil {
		t.Fatal(err)
	}
	got, err := repo.GetByID(context.Background(), 123)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || got.Username != "alice" {
		t.Fatalf("got %+v", got)
	}
}

func TestRepository_GetByID_NotFound(t *testing.T) {
	db := newUserDB(t)
	repo := NewRepository(db)
	got, err := repo.GetByID(context.Background(), 999)
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Fatalf("want nil, got %+v", got)
	}
}

func TestRepository_GetByUsernameAndEmail(t *testing.T) {
	db := newUserDB(t)
	repo := NewRepository(db)
	now := time.Now().UTC()
	_ = repo.Create(context.Background(), &User{
		ID: 1, Username: "alice", Email: ptr("alice@example.com"), CreatedAt: now, UpdatedAt: now,
	})

	got, _ := repo.GetByUsername(context.Background(), "alice")
	if got == nil || got.ID != 1 {
		t.Fatalf("by username miss")
	}
	got2, _ := repo.GetByEmail(context.Background(), "alice@example.com")
	if got2 == nil || got2.ID != 1 {
		t.Fatalf("by email miss")
	}
	miss, _ := repo.GetByEmail(context.Background(), "nobody@example.com")
	if miss != nil {
		t.Fatalf("want nil for miss")
	}
}

func TestRepository_Update(t *testing.T) {
	db := newUserDB(t)
	repo := NewRepository(db)
	now := time.Now().UTC()
	_ = repo.Create(context.Background(), &User{
		ID: 1, Username: "alice", CreatedAt: now, UpdatedAt: now,
	})
	if err := repo.UpdateFields(context.Background(), 1, map[string]any{
		"display_name": "Alice L",
		"role":         int8(2),
		"updated_at":   time.Now().UTC(),
	}); err != nil {
		t.Fatal(err)
	}
	got, _ := repo.GetByID(context.Background(), 1)
	if got.DisplayName == nil || *got.DisplayName != "Alice L" {
		t.Fatalf("display_name not updated: %+v", got.DisplayName)
	}
	if got.Role != 2 {
		t.Fatalf("role not updated: %d", got.Role)
	}
}

func TestRepository_List_FilterAndPagination(t *testing.T) {
	db := newUserDB(t)
	repo := NewRepository(db)
	now := time.Now().UTC()
	for i := 1; i <= 5; i++ {
		_ = repo.Create(context.Background(), &User{
			ID:        int64(i),
			Username:  fmt.Sprintf("user%d", i),
			Email:     ptr(fmt.Sprintf("u%d@example.com", i)),
			Role:      int8(i % 3),
			Status:    int8(i % 2),
			CreatedAt: now.Add(time.Duration(i) * time.Second),
			UpdatedAt: now,
		})
	}
	items, total, err := repo.List(context.Background(), ListFilter{
		Keyword: "user", Page: 1, Size: 3,
	})
	if err != nil {
		t.Fatal(err)
	}
	if total != 5 {
		t.Fatalf("want total 5, got %d", total)
	}
	if len(items) != 3 {
		t.Fatalf("want 3 items, got %d", len(items))
	}
	// 默认按 created_at desc
	if items[0].ID != 5 {
		t.Fatalf("want first item id=5, got %d", items[0].ID)
	}

	role0 := int8(0)
	itemsR, _, _ := repo.List(context.Background(), ListFilter{Role: &role0, Page: 1, Size: 100})
	for _, it := range itemsR {
		if it.Role != 0 {
			t.Fatalf("filter by role failed: got role=%d", it.Role)
		}
	}
}

func TestRepository_SoftDelete(t *testing.T) {
	db := newUserDB(t)
	repo := NewRepository(db)
	now := time.Now().UTC()
	_ = repo.Create(context.Background(), &User{ID: 1, Username: "x", CreatedAt: now, UpdatedAt: now})
	if err := repo.SoftDelete(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
	got, _ := repo.GetByID(context.Background(), 1)
	if got != nil {
		t.Fatalf("want soft-deleted (excluded), got %+v", got)
	}
}
