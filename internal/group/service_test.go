package group

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ijry/pro-api/internal/util/clock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newGroupDB(t *testing.T) *gorm.DB {
	t.Helper()
	dbName := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", dbName)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS user_groups (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			display_name TEXT NOT NULL DEFAULT '',
			ratio REAL NOT NULL DEFAULT 1.0,
			priority INTEGER NOT NULL DEFAULT 0,
			status INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		)
	`).Error; err != nil {
		t.Fatal(err)
	}
	return db
}

type fakeGen struct{ next int64 }

func (f *fakeGen) Generate() int64 { f.next++; return f.next }

func TestService_CreateAndGet(t *testing.T) {
	db := newGroupDB(t)
	svc := NewService(NewRepository(db), clock.Real, &fakeGen{})
	g, err := svc.Create(context.Background(), CreateInput{Name: "default", DisplayName: "普通", Ratio: 1.0})
	if err != nil {
		t.Fatal(err)
	}
	if g.ID == 0 {
		t.Fatal("want id set")
	}
	got, _ := svc.GetByID(context.Background(), g.ID)
	if got == nil || got.Name != "default" {
		t.Fatalf("got %+v", got)
	}
}

func TestService_Default_NotSeeded_Err(t *testing.T) {
	db := newGroupDB(t)
	svc := NewService(NewRepository(db), clock.Real, &fakeGen{})
	if _, err := svc.Default(context.Background()); err == nil {
		t.Fatal("want error when default group missing")
	}
}

func TestService_Default_Loads(t *testing.T) {
	db := newGroupDB(t)
	svc := NewService(NewRepository(db), clock.Real, &fakeGen{})
	_, _ = svc.Create(context.Background(), CreateInput{Name: "default"})
	g, err := svc.Default(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if g.Name != "default" {
		t.Fatalf("got %s", g.Name)
	}
}

func TestService_RatioFor_CachesAndReuses(t *testing.T) {
	db := newGroupDB(t)
	repo := NewRepository(db)
	svc := NewService(repo, clock.Real, &fakeGen{})
	g, _ := svc.Create(context.Background(), CreateInput{Name: "vip", Ratio: 0.8})

	r1, _ := svc.RatioFor(context.Background(), g.ID)
	if r1 != 0.8 {
		t.Fatalf("want 0.8, got %v", r1)
	}
	// 改 DB 不会立刻影响缓存
	_ = repo.UpdateFields(context.Background(), g.ID, map[string]any{"ratio": 2.0})
	r2, _ := svc.RatioFor(context.Background(), g.ID)
	if r2 != 0.8 {
		t.Fatalf("cache should hit, got %v", r2)
	}
}

func TestService_Update_InvalidatesCache(t *testing.T) {
	db := newGroupDB(t)
	svc := NewService(NewRepository(db), clock.Real, &fakeGen{})
	g, _ := svc.Create(context.Background(), CreateInput{Name: "vip", Ratio: 0.8})
	_, _ = svc.RatioFor(context.Background(), g.ID)
	_, _ = svc.Update(context.Background(), g.ID, CreateInput{Name: "vip", Ratio: 0.5})
	r, _ := svc.RatioFor(context.Background(), g.ID)
	if r != 0.5 {
		t.Fatalf("want 0.5 after update, got %v", r)
	}
}

func TestService_RatioFor_Missing_ReturnsOne(t *testing.T) {
	db := newGroupDB(t)
	svc := NewService(NewRepository(db), clock.Real, &fakeGen{})
	r, err := svc.RatioFor(context.Background(), 999)
	if err != nil {
		t.Fatal(err)
	}
	if r != 1.0 {
		t.Fatalf("want 1.0, got %v", r)
	}
}

func TestService_List(t *testing.T) {
	db := newGroupDB(t)
	svc := NewService(NewRepository(db), clock.Real, &fakeGen{})
	_, _ = svc.Create(context.Background(), CreateInput{Name: "a", Priority: 1})
	_, _ = svc.Create(context.Background(), CreateInput{Name: "b", Priority: 5})
	items, err := svc.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 || items[0].Name != "b" {
		t.Fatalf("priority ordering wrong: %+v", items)
	}
}

func TestService_Delete(t *testing.T) {
	db := newGroupDB(t)
	svc := NewService(NewRepository(db), clock.Real, &fakeGen{})
	g, _ := svc.Create(context.Background(), CreateInput{Name: "x"})
	if err := svc.Delete(context.Background(), g.ID); err != nil {
		t.Fatal(err)
	}
	got, _ := svc.GetByID(context.Background(), g.ID)
	if got != nil {
		t.Fatalf("want deleted")
	}
}

var _ = time.Second
