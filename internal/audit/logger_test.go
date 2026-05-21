package audit

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ijry/pro-api/internal/util/idgen"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newAuditDB(t *testing.T) *gorm.DB {
	t.Helper()
	dbName := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", dbName)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS audit_logs (
			id INTEGER PRIMARY KEY,
			created_at DATETIME NOT NULL,
			actor_id INTEGER,
			actor_role INTEGER NOT NULL DEFAULT 0,
			action TEXT NOT NULL,
			target_type TEXT NOT NULL,
			target_id INTEGER,
			"before" TEXT,
			"after" TEXT,
			ip TEXT NOT NULL DEFAULT ''
		);
	`).Error; err != nil {
		t.Fatal(err)
	}
	return db
}

func TestLog_WritesEntry(t *testing.T) {
	db := newAuditDB(t)
	gen, _ := idgen.New(1)
	logger := NewDB(db, zap.NewNop(), gen)

	actorID := int64(42)
	err := logger.Log(context.Background(), Entry{
		ActorID:    &actorID,
		Action:     "user.update",
		TargetType: "user",
		TargetID:   &actorID,
	})
	if err != nil {
		t.Fatal(err)
	}
	var count int64
	_ = db.Model(&Entry{}).Count(&count).Error
	if count != 1 {
		t.Fatalf("want 1 row, got %d", count)
	}
}

func TestLog_DBError_IsSwallowed(t *testing.T) {
	db := newAuditDB(t)
	_ = db.Exec(`INSERT INTO audit_logs (id, created_at, action, target_type, ip) VALUES (1, datetime('now'), 'x', 'y', '')`).Error
	core, recorded := observer.New(zap.ErrorLevel)
	logger := NewDB(db, zap.New(core), &fixedGen{id: 1})
	if err := logger.Log(context.Background(), Entry{Action: "x", TargetType: "y"}); err != nil {
		t.Fatalf("expected swallowed error, got %v", err)
	}
	if recorded.Len() == 0 {
		t.Fatal("expected error log")
	}
}

type fixedGen struct{ id int64 }

func (f *fixedGen) Generate() int64 { return f.id }

func TestNoop_DoesNothing(t *testing.T) {
	noop := NewNoop()
	if err := noop.Log(context.Background(), Entry{Action: "x", TargetType: "y"}); err != nil {
		t.Fatal(err)
	}
}

func TestLog_FillsIDAndCreatedAt(t *testing.T) {
	db := newAuditDB(t)
	gen, _ := idgen.New(1)
	logger := NewDB(db, zap.NewNop(), gen)
	_ = logger.Log(context.Background(), Entry{Action: "x", TargetType: "y"})
	var row Entry
	_ = db.First(&row).Error
	if row.ID == 0 {
		t.Fatal("expected id filled by idgen")
	}
	if row.CreatedAt.IsZero() || time.Since(row.CreatedAt) > time.Second {
		t.Fatalf("expected created_at filled, got %v", row.CreatedAt)
	}
}
