package pricing

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

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := "file:" + t.Name() + "?mode=memory&cache=shared&_busy_timeout=5000"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.Exec(`
		CREATE TABLE pricing_rules (
			id INTEGER PRIMARY KEY,
			scope TEXT NOT NULL,
			group_id INTEGER,
			model TEXT,
			input_ratio REAL,
			output_ratio REAL,
			cached_ratio REAL,
			reasoning_ratio REAL,
			priority INTEGER NOT NULL DEFAULT 100,
			status INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		);
	`).Error; err != nil {
		t.Fatal(err)
	}
	return db
}

// newTestService 构造一个测试 service。
func newTestService(t *testing.T, useRedis bool, opts ...func(*Config)) (*service, *redis.Client) {
	t.Helper()
	db := newTestDB(t)
	var rdb *redis.Client
	if useRedis {
		s := miniredis.RunT(t)
		rdb = redis.NewClient(&redis.Options{Addr: s.Addr()})
	}
	cfg := Config{
		DB:    db,
		Cache: rdb,
		Log:   zap.NewNop(),
		Clock: clock.Real,
		IDGen: &stubIDGen{},
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	svc, err := New(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = svc.Close() })
	return svc.(*service), rdb
}

// ratioPtr 是构造 *float64 的便利函数。
func ratioPtr(f float64) *float64 { return &f }

// idPtr 是构造 *int64 的便利函数。
func idPtr(n int64) *int64 { return &n }

// strPtr 是构造 *string 的便利函数。
func strPtr(s string) *string { return &s }
