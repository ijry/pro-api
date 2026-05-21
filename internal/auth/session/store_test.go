package session

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/ijry/pro-api/internal/util/clock"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func boolPtr(b bool) *bool { return &b }

func newStoreForTest(t *testing.T) (Store, *miniredis.Miniredis, Repository, *gorm.DB) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	dbName := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	db, err := gorm.Open(sqlite.Open("file:"+dbName+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY, user_id INTEGER, ip TEXT, user_agent TEXT,
			created_at DATETIME, last_seen_at DATETIME, expires_at DATETIME, revoked_at DATETIME
		)`).Error; err != nil {
		t.Fatal(err)
	}
	repo := NewRepository(db)
	off := false
	s, err := New(Deps{DB: repo, Cache: rdb, Clock: clock.Real}, Config{
		TTL: time.Hour, Sliding: true,
		MirrorBatchSize: 1, MirrorBatchEvery: time.Minute, // batch=1 立即写
		RestoreOnStart: &off,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s, mr, repo, db
}

func TestCreate_StoresInRedisAndDB(t *testing.T) {
	s, mr, repo, _ := newStoreForTest(t)
	sess, err := s.Create(context.Background(), 1, 2, "1.1.1.1", "ua")
	if err != nil {
		t.Fatal(err)
	}
	if sess.ID == "" || sess.UserID != 1 {
		t.Fatalf("sess wrong: %+v", sess)
	}
	if !mr.Exists("session:" + sess.ID) {
		t.Fatal("redis key missing")
	}
	members, _ := mr.SMembers("session:user:1")
	found := false
	for _, m := range members {
		if m == sess.ID {
			found = true
		}
	}
	if !found {
		t.Fatal("user set missing")
	}
	got, _ := repo.Get(context.Background(), sess.ID)
	if got == nil || got.UserID != 1 {
		t.Fatalf("db row missing: %+v", got)
	}
}

func TestCreate_IDFormat(t *testing.T) {
	s, _, _, _ := newStoreForTest(t)
	sess, _ := s.Create(context.Background(), 1, 0, "", "")
	if !strings.HasPrefix(sess.ID, SessionIDPrefix) {
		t.Fatalf("want sess_ prefix, got %s", sess.ID)
	}
	if len(sess.ID) < 30 {
		t.Fatalf("id too short: %s", sess.ID)
	}
}

func TestGet_HitFromRedis(t *testing.T) {
	s, _, _, _ := newStoreForTest(t)
	sess, _ := s.Create(context.Background(), 7, 2, "1.2.3.4", "ua")
	got, err := s.Get(context.Background(), sess.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || got.UserID != 7 || got.Role != 2 {
		t.Fatalf("got %+v", got)
	}
}

func TestGet_MissReturnsNilNil(t *testing.T) {
	s, _, _, _ := newStoreForTest(t)
	got, err := s.Get(context.Background(), "sess_nope")
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Fatalf("want nil, got %+v", got)
	}
}

func TestTouch_SlidingExtendsTTL(t *testing.T) {
	s, mr, _, _ := newStoreForTest(t)
	sess, _ := s.Create(context.Background(), 1, 0, "", "")
	mr.FastForward(30 * time.Minute)
	if err := s.Touch(context.Background(), sess.ID); err != nil {
		t.Fatal(err)
	}
	ttl := mr.TTL("session:" + sess.ID)
	if ttl < 50*time.Minute {
		t.Fatalf("want TTL extended ~1h, got %v", ttl)
	}
}

func TestRevoke_RemovesFromRedisAndMarksDB(t *testing.T) {
	s, mr, repo, _ := newStoreForTest(t)
	sess, _ := s.Create(context.Background(), 1, 0, "", "")
	_ = s.Revoke(context.Background(), sess.ID)
	if mr.Exists("session:" + sess.ID) {
		t.Fatal("redis key still exists")
	}
	got, _ := repo.Get(context.Background(), sess.ID)
	if got == nil || got.RevokedAt == nil {
		t.Fatalf("db not marked revoked: %+v", got)
	}
}

func TestRevokeAllForUser_RemovesAll(t *testing.T) {
	s, _, repo, _ := newStoreForTest(t)
	_, _ = s.Create(context.Background(), 1, 0, "", "")
	_, _ = s.Create(context.Background(), 1, 0, "", "")
	other, _ := s.Create(context.Background(), 2, 0, "", "")

	if err := s.RevokeAllForUser(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
	// user 1 sessions 全 revoke
	o, _ := repo.Get(context.Background(), other.ID)
	if o.RevokedAt != nil {
		t.Fatal("user 2 session should still be active")
	}
}

func TestRestoreOnStart_RebuildsFromDB(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	dbName := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	db, _ := gorm.Open(sqlite.Open("file:"+dbName+"?mode=memory&cache=shared"), &gorm.Config{})
	_ = db.Exec(`CREATE TABLE sessions (id TEXT PRIMARY KEY, user_id INTEGER, ip TEXT, user_agent TEXT, created_at DATETIME, last_seen_at DATETIME, expires_at DATETIME, revoked_at DATETIME)`).Error
	repo := NewRepository(db)
	now := time.Now().UTC().Truncate(time.Second)
	_ = repo.Insert(context.Background(), &DBSession{ID: "sess_pre", UserID: 9, CreatedAt: now, LastSeenAt: now, ExpiresAt: now.Add(time.Hour)})

	s, err := New(Deps{DB: repo, Cache: rdb, Clock: clock.Real}, Config{
		TTL: time.Hour, RestoreOnStart: boolPtr(true),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })

	got, _ := s.Get(context.Background(), "sess_pre")
	if got == nil || got.UserID != 9 {
		t.Fatalf("not restored: %+v", got)
	}
}

func TestTouch_NoSession_NoOp(t *testing.T) {
	s, _, _, _ := newStoreForTest(t)
	if err := s.Touch(context.Background(), "sess_nope"); err != nil {
		t.Fatal(err)
	}
}
