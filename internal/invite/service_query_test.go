package invite

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ijry/pro-api/internal/setting"
	"github.com/ijry/pro-api/internal/util/clock"
	"github.com/ijry/pro-api/internal/util/idgen"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// fakeSetting mirrors auth/service_test.go helper.
type fakeSetting struct{ kv map[string]any }

func (f *fakeSetting) Get(_ context.Context, key string) (json.RawMessage, bool) {
	v, ok := f.kv[key]
	if !ok {
		return nil, false
	}
	b, _ := json.Marshal(v)
	return b, true
}
func (f *fakeSetting) GetString(_ context.Context, key, def string) string {
	v, ok := f.kv[key]
	if !ok {
		return def
	}
	return v.(string)
}
func (f *fakeSetting) GetBool(_ context.Context, _ string, def bool) bool { return def }
func (f *fakeSetting) GetInt(_ context.Context, _ string, def int) int    { return def }
func (f *fakeSetting) GetFloat(_ context.Context, key string, def float64) float64 {
	v, ok := f.kv[key]
	if !ok {
		return def
	}
	return v.(float64)
}
func (f *fakeSetting) GetJSON(_ context.Context, _ string, _ any) error      { return nil }
func (f *fakeSetting) Put(_ context.Context, _ string, _ any, _ int64) error { return nil }
func (f *fakeSetting) Close() error                                          { return nil }
func (f *fakeSetting) GetSecret(_ context.Context, _ string, _ setting.Decryptor) (string, error) {
	return "", nil
}
func (f *fakeSetting) ListAll(_ context.Context) ([]setting.Setting, error) { return nil, nil }

func setupQueryDB(t *testing.T) (*gorm.DB, *Service) {
	t.Helper()
	name := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", name)), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY, username TEXT UNIQUE, email TEXT, password_hash TEXT,
			display_name TEXT, avatar TEXT, role INTEGER DEFAULT 0, status INTEGER DEFAULT 0,
			group_id INTEGER, invite_code TEXT UNIQUE, invited_by INTEGER NOT NULL DEFAULT 0,
			email_verified_at DATETIME, last_login_at DATETIME, last_login_ip TEXT,
			created_at DATETIME, updated_at DATETIME, deleted_at DATETIME
		);
		CREATE TABLE invite_records (
			id INTEGER PRIMARY KEY,
			inviter_id INTEGER NOT NULL,
			invitee_id INTEGER NOT NULL,
			order_id   INTEGER NOT NULL,
			rebate_cents   INTEGER NOT NULL DEFAULT 0,
			rebate_credits INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME
		);
	`).Error; err != nil {
		t.Fatal(err)
	}
	gen, err := idgen.New(1)
	if err != nil {
		t.Fatal(err)
	}
	sett := &fakeSetting{kv: map[string]any{
		"site.base_url":          "https://example.com",
		"invite.rebate_ratio":    0.10,
		"invite.credit_per_cent": 1.0,
	}}
	svc := NewService(Deps{
		Repo:    NewRepository(db),
		DB:      db,
		Setting: sett,
		IDGen:   gen,
		Clock:   clock.Real,
	})
	return db, svc
}

func TestGetSummary(t *testing.T) {
	db, svc := setupQueryDB(t)
	ctx := context.Background()

	now := time.Now()
	db.Exec(`INSERT INTO users (id,username,invite_code,invited_by,created_at,updated_at) VALUES (10,'alice','CODE123',0,?,?)`, now, now)
	db.Exec(`INSERT INTO users (id,username,invited_by,created_at,updated_at) VALUES (20,'bob',10,?,?)`, now, now)
	db.Exec(`INSERT INTO users (id,username,invited_by,created_at,updated_at) VALUES (30,'carol',10,?,?)`, now, now)
	db.Exec(`INSERT INTO invite_records VALUES (1,10,20,999,100,100,?)`, now)
	db.Exec(`INSERT INTO invite_records VALUES (2,10,30,998,200,200,?)`, now)

	resp, err := svc.GetSummary(ctx, 10)
	if err != nil {
		t.Fatal(err)
	}
	if resp.InviteCode != "CODE123" {
		t.Errorf("invite_code: got %q", resp.InviteCode)
	}
	if resp.ShareURL != "https://example.com/register?invite_code=CODE123" {
		t.Errorf("share_url: got %q", resp.ShareURL)
	}
	if resp.RebateRatio != 0.10 {
		t.Errorf("rebate_ratio: got %v", resp.RebateRatio)
	}
	if resp.Stats.InviteeCount != 2 {
		t.Errorf("invitee_count: want 2, got %d", resp.Stats.InviteeCount)
	}
	if resp.Stats.RebateCreditsTotal != 300 {
		t.Errorf("rebate_credits_total: want 300, got %d", resp.Stats.RebateCreditsTotal)
	}
}

func TestMaskEmail(t *testing.T) {
	cases := []struct{ in, want string }{
		{"john@example.com", "j***@example.com"},
		{"a@b.c", "a***@b.c"},
		{"", "***"},
		{"noatsign", "***"},
	}
	for _, c := range cases {
		got := maskEmail(c.in)
		if got != c.want {
			t.Errorf("maskEmail(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestOnOrderPaidReturnsUserLookupError(t *testing.T) {
	db, svc := setupQueryDB(t)
	ctx := context.Background()

	if err := db.Exec(`DROP TABLE users`).Error; err != nil {
		t.Fatal(err)
	}
	if err := svc.OnOrderPaid(ctx, 1, 20); err == nil {
		t.Fatal("expected user lookup error")
	}
}

func TestListInvitees(t *testing.T) {
	db, svc := setupQueryDB(t)
	ctx := context.Background()

	now := time.Now()
	db.Exec(`INSERT INTO users (id,username,display_name,email,invited_by,created_at,updated_at) VALUES (10,'alice','Alice','a@x.com',0,?,?)`, now, now)
	db.Exec(`INSERT INTO users (id,username,display_name,email,invited_by,created_at,updated_at) VALUES (20,'bob','Bob','b@x.com',10,?,?)`, now, now)
	db.Exec(`INSERT INTO users (id,username,display_name,email,invited_by,created_at,updated_at) VALUES (30,'carol','Carol','c@x.com',10,?,?)`, now, now)
	db.Exec(`INSERT INTO invite_records VALUES (1,10,20,1,50,50,?)`, now)
	db.Exec(`INSERT INTO invite_records VALUES (2,10,20,2,30,30,?)`, now)

	items, total, err := svc.ListInvitees(ctx, 10, 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 {
		t.Errorf("total: want 2, got %d", total)
	}
	if len(items) != 2 {
		t.Errorf("items len: want 2, got %d", len(items))
	}
	var bob *InviteeView
	for _, it := range items {
		if it.DisplayName == "Bob" {
			bob = it
		}
	}
	if bob == nil {
		t.Fatal("bob not in results")
	}
	if bob.TotalRebateCredits != 80 {
		t.Errorf("bob total_rebate_credits: want 80, got %d", bob.TotalRebateCredits)
	}
	if bob.EmailMasked != "b***@x.com" {
		t.Errorf("bob email_masked: got %q", bob.EmailMasked)
	}
}

func TestListRecords(t *testing.T) {
	db, svc := setupQueryDB(t)
	ctx := context.Background()

	now := time.Now()
	db.Exec(`INSERT INTO users (id,username,display_name,invited_by,created_at,updated_at) VALUES (10,'alice','Alice',0,?,?)`, now, now)
	db.Exec(`INSERT INTO users (id,username,display_name,invited_by,created_at,updated_at) VALUES (20,'bob','Bob',10,?,?)`, now, now)
	db.Exec(`INSERT INTO invite_records VALUES (1,10,20,999,100,100,?)`, now)
	db.Exec(`INSERT INTO invite_records VALUES (2,10,20,998,200,200,?)`, now)

	items, total, err := svc.ListRecords(ctx, 10, 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 {
		t.Errorf("total: want 2, got %d", total)
	}
	if len(items) != 2 {
		t.Errorf("items len: want 2, got %d", len(items))
	}
	found := false
	for _, it := range items {
		if it.InviteeDisplayName == "Bob" {
			found = true
		}
	}
	if !found {
		t.Error("display name 'Bob' not populated in any record")
	}
}
