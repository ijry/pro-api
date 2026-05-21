package oauth

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newOauthDB(t *testing.T) *gorm.DB {
	t.Helper()
	dbName := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", dbName)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS oauth_bindings (
			id INTEGER PRIMARY KEY,
			user_id INTEGER NOT NULL,
			provider TEXT NOT NULL,
			provider_uid TEXT NOT NULL,
			email TEXT NOT NULL DEFAULT '',
			profile TEXT,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			UNIQUE (provider, provider_uid)
		)
	`).Error; err != nil {
		t.Fatal(err)
	}
	return db
}

func TestOAuth_Repository_CreateAndFind(t *testing.T) {
	db := newOauthDB(t)
	repo := NewRepository(db)
	now := time.Now().UTC().Truncate(time.Second)
	b := &Binding{ID: 1, UserID: 100, Provider: "github", ProviderUID: "98765", Email: "x@gh.com", CreatedAt: now, UpdatedAt: now}
	if err := repo.Create(context.Background(), b); err != nil {
		t.Fatal(err)
	}
	got, _ := repo.FindByProviderUID(context.Background(), "github", "98765")
	if got == nil || got.UserID != 100 {
		t.Fatalf("got %+v", got)
	}
	miss, _ := repo.FindByProviderUID(context.Background(), "github", "nope")
	if miss != nil {
		t.Fatal("want nil")
	}
}

func TestOAuth_Repository_ListByUserAndDelete(t *testing.T) {
	db := newOauthDB(t)
	repo := NewRepository(db)
	now := time.Now().UTC().Truncate(time.Second)
	_ = repo.Create(context.Background(), &Binding{ID: 1, UserID: 100, Provider: "github", ProviderUID: "1", CreatedAt: now, UpdatedAt: now})
	_ = repo.Create(context.Background(), &Binding{ID: 2, UserID: 100, Provider: "google", ProviderUID: "2", CreatedAt: now, UpdatedAt: now})
	items, _ := repo.ListByUser(context.Background(), 100)
	if len(items) != 2 {
		t.Fatalf("want 2 items, got %d", len(items))
	}
	if err := repo.DeleteByUserProvider(context.Background(), 100, "github"); err != nil {
		t.Fatal(err)
	}
	items2, _ := repo.ListByUser(context.Background(), 100)
	if len(items2) != 1 || items2[0].Provider != "google" {
		t.Fatalf("after delete: %+v", items2)
	}
}
