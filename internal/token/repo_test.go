package token

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/ijry/pro-api/internal/util/clock"
	"github.com/ijry/pro-api/pkg/apierr"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newRepoDB(t *testing.T) *gorm.DB {
	t.Helper()
	// 每个测试用独立的 in-memory DB(共享 cache 但用 t.Name 隔离)。
	dsn := "file:" + t.Name() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	})
	// SQLite 用 TEXT 存 JSON;字段与生产 schema 同构。
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
	return db
}

// stubIDGen 是固定步进的 ID 生成器,便于断言。
type stubIDGen struct{ n int64 }

func (s *stubIDGen) Generate() int64 {
	s.n++
	return s.n
}

func newRepo(t *testing.T, db *gorm.DB) *repo {
	t.Helper()
	return &repo{db: db, idgen: &stubIDGen{}, clk: clock.Real}
}

func mustCreate(t *testing.T, r *repo, userID int64, name string) (*View, string) {
	t.Helper()
	plaintext, view, err := r.Create(context.Background(), CreateInput{
		UserID: userID,
		Name:   name,
	})
	if err != nil {
		t.Fatal(err)
	}
	return view, plaintext
}

func TestRepo_Create_GeneratesKeyAndPersists(t *testing.T) {
	r := newRepo(t, newRepoDB(t))
	plaintext, view, err := r.Create(context.Background(), CreateInput{
		UserID:        42,
		Name:          "t1",
		AllowedModels: []string{"gpt-4*"},
		AllowedIPs:    []string{"10.0.0.0/8"},
		RPMLimit:      30,
	})
	if err != nil {
		t.Fatal(err)
	}
	if plaintext == "" || plaintext[:3] != "pa-" {
		t.Fatalf("bad plaintext %q", plaintext)
	}
	if view.ID == 0 {
		t.Fatal("want id assigned")
	}
	if view.UserID != 42 {
		t.Fatalf("want user_id=42, got %d", view.UserID)
	}
	if len(view.AllowedModels) != 1 || view.AllowedModels[0] != "gpt-4*" {
		t.Fatalf("allowed_models lost: %v", view.AllowedModels)
	}
	if view.RPMLimit != 30 {
		t.Fatalf("rpm_limit lost: %d", view.RPMLimit)
	}
	if view.KeyPrefix == "" {
		t.Fatal("key_prefix must be filled")
	}
}

func TestRepo_Create_InvalidCIDR_Rejected(t *testing.T) {
	r := newRepo(t, newRepoDB(t))
	_, _, err := r.Create(context.Background(), CreateInput{
		UserID:     1,
		Name:       "t",
		AllowedIPs: []string{"not-cidr"},
	})
	if err == nil {
		t.Fatal("want error")
	}
	var ae *apierr.Error
	if !errors.As(err, &ae) || ae.Code != apierr.CodeInvalidParam {
		t.Fatalf("want CodeInvalidParam, got %v", err)
	}
}

func TestRepo_Authenticate_HitsByKeyHash(t *testing.T) {
	r := newRepo(t, newRepoDB(t))
	_, plaintext := mustCreate(t, r, 1, "t1")

	got, err := r.Authenticate(context.Background(), plaintext)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("want view")
	}
}

func TestRepo_Authenticate_BadPrefix_ReturnsInvalid(t *testing.T) {
	r := newRepo(t, newRepoDB(t))
	_, err := r.Authenticate(context.Background(), "not-pa-key")
	var ae *apierr.Error
	if !errors.As(err, &ae) || ae.Code != apierr.CodeInvalidToken {
		t.Fatalf("want CodeInvalidToken, got %v", err)
	}
}

func TestRepo_Authenticate_NotFound_ReturnsInvalid(t *testing.T) {
	r := newRepo(t, newRepoDB(t))
	_, err := r.Authenticate(context.Background(), "pa-nonexistent")
	var ae *apierr.Error
	if !errors.As(err, &ae) || ae.Code != apierr.CodeInvalidToken {
		t.Fatalf("want CodeInvalidToken, got %v", err)
	}
}

func TestRepo_Authenticate_DeletedRowNotFound(t *testing.T) {
	db := newRepoDB(t)
	r := newRepo(t, db)
	_, plaintext := mustCreate(t, r, 1, "t1")
	// 软删
	now := time.Now()
	if err := db.Exec("UPDATE api_tokens SET deleted_at = ?", now).Error; err != nil {
		t.Fatal(err)
	}
	_, err := r.Authenticate(context.Background(), plaintext)
	var ae *apierr.Error
	if !errors.As(err, &ae) || ae.Code != apierr.CodeInvalidToken {
		t.Fatalf("want CodeInvalidToken, got %v", err)
	}
}

func TestRepo_Authenticate_StatusDisabled_ReturnsInvalid(t *testing.T) {
	db := newRepoDB(t)
	r := newRepo(t, db)
	v, plaintext := mustCreate(t, r, 1, "t1")
	if err := db.Exec("UPDATE api_tokens SET status = 1 WHERE id = ?", v.ID).Error; err != nil {
		t.Fatal(err)
	}
	_, err := r.Authenticate(context.Background(), plaintext)
	var ae *apierr.Error
	if !errors.As(err, &ae) || ae.Code != apierr.CodeInvalidToken {
		t.Fatalf("want CodeInvalidToken, got %v", err)
	}
}

func TestRepo_Authenticate_ExpiresAtPast_ReturnsExpired(t *testing.T) {
	db := newRepoDB(t)
	r := newRepo(t, db)
	v, plaintext := mustCreate(t, r, 1, "t1")
	past := time.Now().Add(-time.Hour)
	if err := db.Exec("UPDATE api_tokens SET expires_at = ? WHERE id = ?", past, v.ID).Error; err != nil {
		t.Fatal(err)
	}
	_, err := r.Authenticate(context.Background(), plaintext)
	var ae *apierr.Error
	if !errors.As(err, &ae) || ae.Code != apierr.CodeTokenExpired {
		t.Fatalf("want CodeTokenExpired, got %v", err)
	}
}

func TestRepo_Get_NotFound(t *testing.T) {
	r := newRepo(t, newRepoDB(t))
	_, err := r.Get(context.Background(), 99999)
	var ae *apierr.Error
	if !errors.As(err, &ae) || ae.Code != apierr.CodeNotFound {
		t.Fatalf("want CodeNotFound, got %v", err)
	}
}

func TestRepo_List_FilterByUser(t *testing.T) {
	r := newRepo(t, newRepoDB(t))
	for _, u := range []int64{1, 1, 2} {
		_, _ = mustCreate(t, r, u, "t")
	}
	got, total, err := r.List(context.Background(), ListFilter{UserID: 1, Page: 1, Size: 10})
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 || len(got) != 2 {
		t.Fatalf("want 2 rows, got total=%d len=%d", total, len(got))
	}
}

func TestRepo_List_FilterByStatus(t *testing.T) {
	db := newRepoDB(t)
	r := newRepo(t, db)
	v1, _ := mustCreate(t, r, 1, "a")
	_, _ = mustCreate(t, r, 1, "b")
	_ = db.Exec("UPDATE api_tokens SET status = 1 WHERE id = ?", v1.ID).Error
	st := StatusDisabled
	got, total, err := r.List(context.Background(), ListFilter{UserID: 1, Status: &st, Page: 1, Size: 10})
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(got) != 1 || got[0].ID != v1.ID {
		t.Fatalf("want 1 disabled, got total=%d", total)
	}
}

func TestRepo_List_Keyword(t *testing.T) {
	r := newRepo(t, newRepoDB(t))
	_, _ = mustCreate(t, r, 1, "alpha")
	_, _ = mustCreate(t, r, 1, "beta")
	got, total, err := r.List(context.Background(), ListFilter{UserID: 1, Keyword: "alph", Page: 1, Size: 10})
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(got) != 1 || got[0].Name != "alpha" {
		t.Fatalf("want alpha match, got %+v", got)
	}
}

func TestRepo_List_Pagination(t *testing.T) {
	r := newRepo(t, newRepoDB(t))
	for i := 0; i < 5; i++ {
		_, _ = mustCreate(t, r, 1, "x")
	}
	got, total, err := r.List(context.Background(), ListFilter{UserID: 1, Page: 2, Size: 2})
	if err != nil {
		t.Fatal(err)
	}
	if total != 5 || len(got) != 2 {
		t.Fatalf("want total=5 len=2, got %d / %d", total, len(got))
	}
}

func TestRepo_Update_PatchOnlyNonNil(t *testing.T) {
	r := newRepo(t, newRepoDB(t))
	v, _ := mustCreate(t, r, 1, "old")
	newName := "new"
	newRPM := 99
	updated, err := r.Update(context.Background(), v.ID, UpdatePatch{
		Name:     &newName,
		RPMLimit: &newRPM,
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "new" || updated.RPMLimit != 99 {
		t.Fatalf("patch failed: %+v", updated)
	}
}

func TestRepo_Update_ClearQuotaLimit(t *testing.T) {
	r := newRepo(t, newRepoDB(t))
	q := int64(100)
	plaintext, _, err := r.Create(context.Background(), CreateInput{UserID: 1, Name: "t", QuotaLimit: &q})
	_ = plaintext
	if err != nil {
		t.Fatal(err)
	}
	v, err := r.Authenticate(context.Background(), plaintext)
	if err != nil {
		t.Fatal(err)
	}
	if v.QuotaLimit == nil || *v.QuotaLimit != 100 {
		t.Fatalf("want quota_limit=100, got %v", v.QuotaLimit)
	}
	updated, err := r.Update(context.Background(), v.ID, UpdatePatch{ClearQuotaLimit: true})
	if err != nil {
		t.Fatal(err)
	}
	if updated.QuotaLimit != nil {
		t.Fatalf("want nil, got %v", *updated.QuotaLimit)
	}
}

func TestRepo_Update_ClearExpiresAt(t *testing.T) {
	db := newRepoDB(t)
	r := newRepo(t, db)
	v, _ := mustCreate(t, r, 1, "t")
	future := time.Now().Add(time.Hour)
	_ = db.Exec("UPDATE api_tokens SET expires_at = ? WHERE id = ?", future, v.ID).Error
	updated, err := r.Update(context.Background(), v.ID, UpdatePatch{ClearExpiresAt: true})
	if err != nil {
		t.Fatal(err)
	}
	if updated.ExpiresAt != nil {
		t.Fatalf("want nil expires_at, got %v", *updated.ExpiresAt)
	}
}

func TestRepo_Update_AllowedModels(t *testing.T) {
	r := newRepo(t, newRepoDB(t))
	v, _ := mustCreate(t, r, 1, "t")
	models := []string{"gpt-4o", "claude-3*"}
	updated, err := r.Update(context.Background(), v.ID, UpdatePatch{AllowedModels: &models})
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.AllowedModels) != 2 || updated.AllowedModels[1] != "claude-3*" {
		t.Fatalf("models not updated: %v", updated.AllowedModels)
	}
}

func TestRepo_Update_BadIPRejected(t *testing.T) {
	r := newRepo(t, newRepoDB(t))
	v, _ := mustCreate(t, r, 1, "t")
	ips := []string{"garbage"}
	_, err := r.Update(context.Background(), v.ID, UpdatePatch{AllowedIPs: &ips})
	var ae *apierr.Error
	if !errors.As(err, &ae) || ae.Code != apierr.CodeInvalidParam {
		t.Fatalf("want CodeInvalidParam, got %v", err)
	}
}

func TestRepo_Update_Status2Rejected(t *testing.T) {
	r := newRepo(t, newRepoDB(t))
	v, _ := mustCreate(t, r, 1, "t")
	st := StatusExpired
	_, err := r.Update(context.Background(), v.ID, UpdatePatch{Status: &st})
	if err == nil {
		t.Fatal("want reject status=2")
	}
}

func TestRepo_Revoke_SetsStatusDisabled(t *testing.T) {
	r := newRepo(t, newRepoDB(t))
	v, _ := mustCreate(t, r, 1, "t")
	keyHash, err := r.Revoke(context.Background(), v.ID)
	if err != nil {
		t.Fatal(err)
	}
	if keyHash == "" {
		t.Fatal("want returned old hash")
	}
	got, _ := r.Get(context.Background(), v.ID)
	if got.Status != StatusDisabled {
		t.Fatalf("want disabled, got %d", got.Status)
	}
}

func TestRepo_Regenerate_ChangesKeyHash(t *testing.T) {
	r := newRepo(t, newRepoDB(t))
	v, plaintext1 := mustCreate(t, r, 1, "t")
	plaintext2, updated, oldHash, err := r.Regenerate(context.Background(), v.ID)
	if err != nil {
		t.Fatal(err)
	}
	if plaintext1 == plaintext2 {
		t.Fatal("plaintext must change")
	}
	if oldHash == "" {
		t.Fatal("old hash must be returned")
	}
	if updated.ID != v.ID {
		t.Fatal("id must persist")
	}
	// 原 key 应该失效
	_, authErr := r.Authenticate(context.Background(), plaintext1)
	var ae *apierr.Error
	if !errors.As(authErr, &ae) || ae.Code != apierr.CodeInvalidToken {
		t.Fatalf("want old key invalid, got %v", authErr)
	}
	// 新 key 应当能 Auth
	if _, err := r.Authenticate(context.Background(), plaintext2); err != nil {
		t.Fatalf("new key should authenticate: %v", err)
	}
}

func TestToken_ToView_DecodesJSON(t *testing.T) {
	row := &Token{
		AllowedModels: json.RawMessage(`["gpt-4o","claude-3*"]`),
		AllowedIPs:    json.RawMessage(`["10.0.0.0/8"]`),
	}
	v := row.ToView()
	if len(v.AllowedModels) != 2 || v.AllowedModels[0] != "gpt-4o" {
		t.Fatalf("models: %v", v.AllowedModels)
	}
	if len(v.AllowedIPs) != 1 {
		t.Fatalf("ips: %v", v.AllowedIPs)
	}
}
