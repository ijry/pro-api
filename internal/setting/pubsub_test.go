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

func openSettingDB(t *testing.T, name string) *gorm.DB {
	t.Helper()
	dbName := strings.NewReplacer("/", "_", " ", "_").Replace(name)
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", dbName)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS system_settings (
			key TEXT PRIMARY KEY, value TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '', updated_by INTEGER, updated_at DATETIME NOT NULL
		);
	`).Error; err != nil {
		t.Fatal(err)
	}
	return db
}

func TestNew_StartsInvalidator(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	db := openSettingDB(t, t.Name())

	s, err := New(Config{DB: db, Cache: rdb, Log: zap.NewNop()})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	// 模拟本地缓存有值
	concrete := s.(*store)
	concrete.local.Store("x", cachedValue{raw: []byte(`"cached"`), ts: time.Now()})

	// 直接 Publish 失效消息
	if err := rdb.Publish(context.Background(), redisInvalidateCh, "x").Err(); err != nil {
		t.Fatal(err)
	}

	// 等 invalidator 处理
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if _, ok := concrete.local.Load("x"); !ok {
			return // success
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("expected local cache cleared after publish")
}

func TestPut_BroadcastsToOtherInstance(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb1 := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	rdb2 := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	db := openSettingDB(t, t.Name())

	a, _ := New(Config{DB: db, Cache: rdb1, Log: zap.NewNop()})
	b, _ := New(Config{DB: db, Cache: rdb2, Log: zap.NewNop()})
	defer a.Close()
	defer b.Close()

	// 实例 B 先 seed 一个本地缓存项
	concreteB := b.(*store)
	concreteB.local.Store("k", cachedValue{raw: []byte(`"v1"`), ts: time.Now()})

	// 实例 A 写新值(触发广播)
	if err := a.Put(context.Background(), "k", "v2", 1); err != nil {
		t.Fatal(err)
	}

	// 等广播传播
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if _, ok := concreteB.local.Load("k"); !ok {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("expected B local cache cleared after A.Put")
}

func TestClose_StopsInvalidator(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	s, _ := New(Config{Cache: rdb, Log: zap.NewNop()})
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	// 第二次 Close 不应 panic
	_ = s.Close()
}
