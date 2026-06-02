package token

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/ijry/pro-api/internal/audit"
	"github.com/ijry/pro-api/pkg/apierr"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// captureAudit 记录所有 audit.Log 调用,便于断言。
type captureAudit struct {
	mu      sync.Mutex
	entries []audit.Entry
}

func (c *captureAudit) Log(_ context.Context, e audit.Entry) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = append(c.entries, e)
	return nil
}

func (c *captureAudit) actions() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]string, len(c.entries))
	for i, e := range c.entries {
		out[i] = e.Action
	}
	return out
}

func newServiceFixture(t *testing.T) (*service, *captureAudit, *redis.Client) {
	t.Helper()
	dsn := "file:" + t.Name() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`
		CREATE TABLE api_tokens (
			id INTEGER PRIMARY KEY,
			user_id INTEGER NOT NULL,
			group_id INTEGER NOT NULL DEFAULT 0,
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
	cap := &captureAudit{}
	store, err := New(Config{
		DB:               db,
		Cache:            rdb,
		Log:              zap.NewNop(),
		IDGen:            &stubIDGen{},
		Audit:            cap,
		CacheTTL:         time.Second,
		NegativeCacheTTL: 500 * time.Millisecond,
		FlushInterval:    24 * time.Hour, // 不让 ticker 自动触发,测试手动 close 来 flush
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = store.Close()
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	})
	return store.(*service), cap, rdb
}

func TestService_Create_AuditLogged(t *testing.T) {
	s, cap, _ := newServiceFixture(t)
	plaintext, view, err := s.Create(context.Background(), CreateInput{UserID: 1, Name: "t"})
	if err != nil {
		t.Fatal(err)
	}
	if plaintext == "" || view == nil {
		t.Fatal("want plaintext + view")
	}
	if got := cap.actions(); len(got) != 1 || got[0] != "token.create" {
		t.Fatalf("audit not logged: %v", got)
	}
}

func TestService_Authenticate_CachesPositive(t *testing.T) {
	s, _, rdb := newServiceFixture(t)
	plaintext, _, _ := s.Create(context.Background(), CreateInput{UserID: 1, Name: "t"})
	if _, err := s.Authenticate(context.Background(), plaintext); err != nil {
		t.Fatal(err)
	}
	// Redis 中应有缓存
	h := hashPlaintext(plaintext)
	if n, _ := rdb.Exists(context.Background(), cacheKey(h)).Result(); n != 1 {
		t.Fatal("expected positive cache entry")
	}
	// 第二次应当命中 cache(此处无法直接断言,但不应报错)
	if _, err := s.Authenticate(context.Background(), plaintext); err != nil {
		t.Fatal(err)
	}
}

func TestService_Authenticate_NegativeCacheOnNotFound(t *testing.T) {
	s, _, rdb := newServiceFixture(t)
	// 形式合法但 DB 里没有
	bad := "pa-deadbeef0000000000000000000000000000000000000000"
	_, err := s.Authenticate(context.Background(), bad)
	var ae *apierr.Error
	if !errors.As(err, &ae) || ae.Code != apierr.CodeInvalidToken {
		t.Fatalf("want invalid_token, got %v", err)
	}
	h := hashPlaintext(bad)
	val, _ := rdb.Get(context.Background(), cacheKey(h)).Result()
	if val != negativeMarker {
		t.Fatalf("want negative marker, got %q", val)
	}
	// 第二次仍是 invalid
	_, err = s.Authenticate(context.Background(), bad)
	if !errors.As(err, &ae) || ae.Code != apierr.CodeInvalidToken {
		t.Fatalf("want invalid_token from negative cache, got %v", err)
	}
}

func TestService_Authenticate_BadFormat_ReturnsInvalid(t *testing.T) {
	s, _, _ := newServiceFixture(t)
	_, err := s.Authenticate(context.Background(), "bearer-typo")
	var ae *apierr.Error
	if !errors.As(err, &ae) || ae.Code != apierr.CodeInvalidToken {
		t.Fatalf("want invalid_token, got %v", err)
	}
}

func TestService_Revoke_Audited_CacheInvalidated(t *testing.T) {
	s, cap, rdb := newServiceFixture(t)
	plaintext, view, _ := s.Create(context.Background(), CreateInput{UserID: 1, Name: "t"})
	// 触发缓存
	_, _ = s.Authenticate(context.Background(), plaintext)

	if err := s.Revoke(context.Background(), view.ID); err != nil {
		t.Fatal(err)
	}
	h := hashPlaintext(plaintext)
	if n, _ := rdb.Exists(context.Background(), cacheKey(h)).Result(); n != 0 {
		t.Fatal("cache should be invalidated after revoke")
	}
	// 再 Authenticate 应当 disabled
	_, err := s.Authenticate(context.Background(), plaintext)
	var ae *apierr.Error
	if !errors.As(err, &ae) || ae.Code != apierr.CodeInvalidToken {
		t.Fatalf("want invalid_token after revoke, got %v", err)
	}
	if a := cap.actions(); len(a) < 2 || a[len(a)-1] != "token.revoke" {
		t.Fatalf("revoke audit missing: %v", a)
	}
}

func TestService_Regenerate_OldKeyFailsNewKeyWorks(t *testing.T) {
	s, _, _ := newServiceFixture(t)
	old, view, _ := s.Create(context.Background(), CreateInput{UserID: 1, Name: "t"})
	new, _, err := s.Regenerate(context.Background(), view.ID)
	if err != nil {
		t.Fatal(err)
	}
	if old == new {
		t.Fatal("plaintext should change")
	}
	if _, err := s.Authenticate(context.Background(), old); err == nil {
		t.Fatal("old key should fail")
	}
	if _, err := s.Authenticate(context.Background(), new); err != nil {
		t.Fatalf("new key should succeed: %v", err)
	}
}

func TestService_Update_InvalidatesCache(t *testing.T) {
	s, _, rdb := newServiceFixture(t)
	plaintext, view, _ := s.Create(context.Background(), CreateInput{UserID: 1, Name: "t"})
	_, _ = s.Authenticate(context.Background(), plaintext)

	rpm := 99
	if _, err := s.Update(context.Background(), view.ID, UpdatePatch{RPMLimit: &rpm}); err != nil {
		t.Fatal(err)
	}
	h := hashPlaintext(plaintext)
	if n, _ := rdb.Exists(context.Background(), cacheKey(h)).Result(); n != 0 {
		t.Fatal("cache should be invalidated after update")
	}
}

func TestService_IncrementUsage_FlushedOnClose(t *testing.T) {
	s, _, _ := newServiceFixture(t)
	_, view, _ := s.Create(context.Background(), CreateInput{UserID: 1, Name: "t"})
	s.IncrementUsage(view.ID, 123)
	// Close 触发 final flush
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	var got Token
	if err := s.cfg.DB.Where("id = ?", view.ID).First(&got).Error; err != nil {
		t.Fatal(err)
	}
	if got.QuotaUsed != 123 {
		t.Fatalf("want 123, got %d", got.QuotaUsed)
	}
}

func TestService_Close_Idempotent(t *testing.T) {
	s, _, _ := newServiceFixture(t)
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestService_Create_BadIP_ReturnsInvalidParam(t *testing.T) {
	s, _, _ := newServiceFixture(t)
	_, _, err := s.Create(context.Background(), CreateInput{
		UserID:     1,
		Name:       "t",
		AllowedIPs: []string{"garbage"},
	})
	var ae *apierr.Error
	if !errors.As(err, &ae) || ae.Code != apierr.CodeInvalidParam {
		t.Fatalf("want CodeInvalidParam, got %v", err)
	}
}
