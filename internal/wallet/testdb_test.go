package wallet

import (
	"context"
	"sync"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/ijry/pro-api/internal/util/clock"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// stubIDGen 是顺序递增的 ID 生成器(线程安全)。
type stubIDGen struct {
	mu sync.Mutex
	n  int64
}

func (s *stubIDGen) Generate() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.n++
	return s.n
}

// newTestDB 返回独立 in-memory sqlite,带 wallets + ledger_entries + quota_reservations 表。
func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := "file:" + t.Name() + "?mode=memory&cache=shared&_busy_timeout=5000"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, _ := db.DB()
	// sqlite 是单写者数据库;限制连接池为 1 让并发写在 Go 层串行化,避免 "table is locked"
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	stmts := []string{
		`CREATE TABLE wallets (
			id INTEGER PRIMARY KEY,
			owner_type TEXT NOT NULL,
			owner_id INTEGER NOT NULL,
			quota_balance INTEGER NOT NULL DEFAULT 0,
			quota_total_recharged INTEGER NOT NULL DEFAULT 0,
			quota_total_consumed INTEGER NOT NULL DEFAULT 0,
			currency TEXT NOT NULL DEFAULT 'USD',
			version INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			UNIQUE (owner_type, owner_id)
		)`,
		`CREATE TABLE ledger_entries (
			id INTEGER PRIMARY KEY,
			wallet_id INTEGER NOT NULL,
			direction TEXT NOT NULL,
			amount_quota INTEGER NOT NULL,
			amount_money INTEGER NOT NULL DEFAULT 0,
			currency TEXT NOT NULL DEFAULT 'USD',
			ref_type TEXT NOT NULL,
			ref_id INTEGER,
			balance_after INTEGER NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			created_at DATETIME NOT NULL
		)`,
	}
	for _, s := range stmts {
		if err := db.Exec(s).Error; err != nil {
			t.Fatalf("exec %q: %v", s, err)
		}
	}
	return db
}

func newTestStore(t *testing.T, useRedis bool) (*store, *redis.Client) {
	t.Helper()
	db := newTestDB(t)
	var rdb *redis.Client
	if useRedis {
		s := miniredis.RunT(t)
		rdb = redis.NewClient(&redis.Options{Addr: s.Addr()})
	}
	st, err := New(Config{
		DB:    db,
		Cache: rdb,
		Log:   zap.NewNop(),
		Clock: clock.Real,
		IDGen: &stubIDGen{},
	})
	if err != nil {
		t.Fatal(err)
	}
	return st.(*store), rdb
}

func ctx() context.Context { return context.Background() }
