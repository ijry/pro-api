package setting

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newStoreWithDB(t *testing.T) (*store, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	// 每个测试一个独立的 in-memory DB(用 t.Name() 作 dbname,清掉非法字符)
	dbName := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", dbName)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS system_settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			updated_by INTEGER,
			updated_at DATETIME NOT NULL
		);
	`).Error; err != nil {
		t.Fatal(err)
	}

	s := &store{
		db:  db,
		rdb: rdb,
		log: zap.NewNop(),
		ttl: 60 * time.Second,
	}
	return s, mr
}

func TestPut_WritesDB_AndInvalidatesRedis(t *testing.T) {
	s, mr := newStoreWithDB(t)
	_ = mr.Set("setting:flag", `"old"`)
	if err := s.Put(context.Background(), "flag", "new", 1); err != nil {
		t.Fatal(err)
	}
	if mr.Exists("setting:flag") {
		t.Fatal("expected Redis key deleted after Put")
	}
	var got Setting
	if err := s.db.Where("`key` = ?", "flag").First(&got).Error; err != nil {
		t.Fatal(err)
	}
	if string(got.Value) != `"new"` {
		t.Fatalf("want \"new\", got %s", got.Value)
	}
	if got.UpdatedBy == nil || *got.UpdatedBy != 1 {
		t.Fatalf("want updated_by=1, got %v", got.UpdatedBy)
	}
}

func TestPut_UpsertExistingKey(t *testing.T) {
	s, _ := newStoreWithDB(t)
	_ = s.Put(context.Background(), "k", 1, 1)
	_ = s.Put(context.Background(), "k", 2, 2)
	var got Setting
	_ = s.db.Where("`key` = ?", "k").First(&got).Error
	if string(got.Value) != `2` {
		t.Fatalf("want 2, got %s", got.Value)
	}
	if *got.UpdatedBy != 2 {
		t.Fatalf("want actor=2, got %v", *got.UpdatedBy)
	}
}

func TestGet_FromDB_WhenRedisAndLocalMiss(t *testing.T) {
	s, _ := newStoreWithDB(t)
	now := time.Now()
	_ = s.db.Create(&Setting{Key: "x", Value: []byte(`"from-db"`), UpdatedAt: now}).Error
	v, ok := s.Get(context.Background(), "x")
	if !ok {
		t.Fatal("want ok from DB")
	}
	if string(v) != `"from-db"` {
		t.Fatalf("got %s", v)
	}
}

func TestGet_DBHit_BackfillsRedis(t *testing.T) {
	s, mr := newStoreWithDB(t)
	_ = s.db.Create(&Setting{Key: "y", Value: []byte(`123`), UpdatedAt: time.Now()}).Error
	_, _ = s.Get(context.Background(), "y")
	if !mr.Exists("setting:y") {
		t.Fatal("expected redis backfill")
	}
}

func TestPut_WithStringValue_EncodedAsJSONString(t *testing.T) {
	s, _ := newStoreWithDB(t)
	_ = s.Put(context.Background(), "name", "alice", 1)
	got := s.GetString(context.Background(), "name", "")
	if got != "alice" {
		t.Fatalf("want alice, got %s", got)
	}
}

func TestPut_WithStructValue_RoundTrip(t *testing.T) {
	s, _ := newStoreWithDB(t)
	type cfg struct {
		Enabled bool `json:"enabled"`
		Limit   int  `json:"limit"`
	}
	in := cfg{Enabled: true, Limit: 60}
	if err := s.Put(context.Background(), "cfg", in, 1); err != nil {
		t.Fatal(err)
	}
	var out cfg
	if err := s.GetJSON(context.Background(), "cfg", &out); err != nil {
		t.Fatal(err)
	}
	if out != in {
		t.Fatalf("round-trip: want %+v, got %+v", in, out)
	}
}
