package token

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/ijry/pro-api/internal/app"
	"github.com/ijry/pro-api/internal/audit"
	"github.com/ijry/pro-api/internal/util/clock"
	"github.com/ijry/pro-api/internal/util/idgen"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestWireToken_AssignsStoreAndCloser(t *testing.T) {
	dsn := "file:" + t.Name() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`
		CREATE TABLE api_tokens (
			id INTEGER PRIMARY KEY,
			user_id INTEGER NOT NULL,
			name TEXT NOT NULL DEFAULT '',
			key_hash TEXT NOT NULL UNIQUE,
			key_prefix TEXT NOT NULL,
			quota_limit INTEGER,
			quota_used INTEGER NOT NULL DEFAULT 0,
			allowed_models TEXT NOT NULL DEFAULT '[]',
			allowed_ips TEXT NOT NULL DEFAULT '[]',
			rpm_limit INTEGER NOT NULL DEFAULT 0,
			tpm_limit INTEGER NOT NULL DEFAULT 0,
			expires_at DATETIME,
			last_used_at DATETIME,
			status INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			deleted_at DATETIME
		);
	`).Error; err != nil {
		t.Fatal(err)
	}

	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	idg, _ := idgen.New(1)
	application := &app.Application{
		DB:    db,
		Cache: rdb,
		Log:   zap.NewNop(),
		Clock: clock.Real,
		IDGen: idg,
		Audit: audit.NewNoop(),
	}

	if err := WireToken(application); err != nil {
		t.Fatal(err)
	}
	if application.TokenStore == nil {
		t.Fatal("TokenStore must be assigned")
	}
	store, ok := application.TokenStore.(Store)
	if !ok {
		t.Fatalf("TokenStore must implement Store, got %T", application.TokenStore)
	}
	_ = store

	// Shutdown 应触发 Close 不报错
	if err := application.Shutdown(nil); err != nil {
		t.Fatalf("shutdown error: %v", err)
	}
}
