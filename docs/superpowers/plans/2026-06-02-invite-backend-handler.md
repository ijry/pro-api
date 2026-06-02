# Invite Backend Handler Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Expose 3 authenticated user-side REST endpoints (`GET /api/user/invites/me|invitees|records`) that match the frontend contract already shipped in the invites.vue page.

**Architecture:** Extend `internal/invite/` with 3 public query methods + view types, create `InviteHandler` in `internal/server/handler/user/`, wire via `wire.go` and `cmd/proapi/main.go`. The service creates a query-only instance (wallet=nil is safe — `OnOrderPaid` guards against nil wallet and is never called from HTTP handlers).

**Tech Stack:** Go 1.25 · Gin · GORM · SQLite (tests) · `github.com/ijry/pro-api`

---

## File Map

| Op | Path | Responsibility |
|----|------|----------------|
| Modify | `internal/invite/repo.go` | Add `CountByInviter` to interface + impl |
| Modify | `internal/invite/service.go` | View types + `maskEmail` + 3 query methods |
| Create | `internal/invite/service_query_test.go` | Tests for the 3 query methods |
| Create | `internal/server/handler/user/invite.go` | `InviteHandler` + 3 HTTP handlers + `Register` |
| Create | `internal/server/handler/user/invite_test.go` | Handler unit tests |
| Modify | `internal/server/handler/user/wire.go` | Add `WireInvite` |
| Modify | `cmd/proapi/main.go` | Add `wireInviteRoutes` + call in `wireRoutes` |

---

## Task 1: Extend repo — add CountByInviter

**Files:**
- Modify: `internal/invite/repo.go`

- [ ] **Step 1: Write a failing test**

Create `internal/invite/repo_test.go`:

```go
package invite

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	name := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", name)), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`
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
	return db
}

func TestCountByInviter(t *testing.T) {
	db := openTestDB(t)
	r := NewRepository(db)
	ctx := context.Background()

	now := time.Now()
	for i := 0; i < 3; i++ {
		db.Exec(`INSERT INTO invite_records (id,inviter_id,invitee_id,order_id,rebate_cents,rebate_credits,created_at) VALUES (?,100,?,999,10,10,?)`,
			int64(i+1), int64(200+i), now)
	}
	// unrelated inviter
	db.Exec(`INSERT INTO invite_records (id,inviter_id,invitee_id,order_id,rebate_cents,rebate_credits,created_at) VALUES (99,999,1,1,0,0,?)`, now)

	n, err := r.CountByInviter(ctx, 100)
	if err != nil {
		t.Fatal(err)
	}
	if n != 3 {
		t.Fatalf("want 3, got %d", n)
	}

	n0, _ := r.CountByInviter(ctx, 777)
	if n0 != 0 {
		t.Fatalf("want 0 for unknown inviter, got %d", n0)
	}
}
```

- [ ] **Step 2: Run test — expect FAIL**

```bash
cd /Users/admin/Documents/Repos/xyito/open/proapi
go test ./internal/invite/ -run TestCountByInviter -v
```

Expected: `FAIL — CountByInviter undefined`

- [ ] **Step 3: Add CountByInviter to repo.go**

```go
// Repository interface (replace existing block):
type Repository interface {
	Create(ctx context.Context, r *Record) error
	ListByInviter(ctx context.Context, inviterID int64, limit, offset int) ([]*Record, error)
	CountByInviter(ctx context.Context, inviterID int64) (int64, error)
}

// gormRepo implementation (add after existing ListByInviter):
func (r *gormRepo) CountByInviter(ctx context.Context, inviterID int64) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&Record{}).
		Where("inviter_id = ?", inviterID).
		Count(&n).Error
	return n, err
}
```

- [ ] **Step 4: Run test — expect PASS**

```bash
go test ./internal/invite/ -run TestCountByInviter -v
```

Expected: `PASS`

- [ ] **Step 5: Commit**

```bash
git add internal/invite/repo.go internal/invite/repo_test.go
git commit -m "feat(invite): repo.CountByInviter + test"
```

---

## Task 2: View types + query methods in service

**Files:**
- Modify: `internal/invite/service.go`
- Create: `internal/invite/service_query_test.go`

- [ ] **Step 1: Write failing tests**

Create `internal/invite/service_query_test.go`:

```go
package invite

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ijry/pro-api/internal/util/clock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// fakeIDGen and fakeSetting mirror auth/service_test.go helpers.
type fakeIDGen struct{ n int64 }

func (f *fakeIDGen) Generate() int64 { f.n++; return f.n }

type fakeSetting struct{ kv map[string]any }

func (f *fakeSetting) Get(_ context.Context, _ string) ([]byte, bool)           { return nil, false }
func (f *fakeSetting) GetString(_ context.Context, key, def string) string {
	v, ok := f.kv[key]
	if !ok { return def }
	return v.(string)
}
func (f *fakeSetting) GetFloat(_ context.Context, key string, def float64) float64 {
	v, ok := f.kv[key]
	if !ok { return def }
	return v.(float64)
}
func (f *fakeSetting) GetBool(_ context.Context, _ string, def bool) bool { return def }
func (f *fakeSetting) GetInt(_ context.Context, _ string, def int) int     { return def }

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
	sett := &fakeSetting{kv: map[string]any{
		"site.base_url":        "https://example.com",
		"invite.rebate_ratio":  0.10,
		"invite.credit_per_cent": 1.0,
	}}
	svc := NewService(Deps{
		Repo:    NewRepository(db),
		DB:      db,
		Setting: sett,
		IDGen:   &fakeIDGen{},
		Clock:   clock.Real,
	})
	return db, svc
}

func TestGetSummary(t *testing.T) {
	db, svc := setupQueryDB(t)
	ctx := context.Background()

	now := time.Now()
	// User 10 with invite code, invited 2 others
	db.Exec(`INSERT INTO users (id,username,invite_code,invited_by,created_at,updated_at) VALUES (10,'alice','CODE123',0,?,?)`, now, now)
	db.Exec(`INSERT INTO users (id,username,invited_by,created_at,updated_at) VALUES (20,'bob',10,?,?)`, now, now)
	db.Exec(`INSERT INTO users (id,username,invited_by,created_at,updated_at) VALUES (30,'carol',10,?,?)`, now, now)
	// Two rebate records for user 10 as inviter
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

func TestListInvitees(t *testing.T) {
	db, svc := setupQueryDB(t)
	ctx := context.Background()

	now := time.Now()
	db.Exec(`INSERT INTO users (id,username,display_name,email,invited_by,created_at,updated_at) VALUES (10,'alice','Alice','a@x.com',0,?,?)`, now, now)
	db.Exec(`INSERT INTO users (id,username,display_name,email,invited_by,created_at,updated_at) VALUES (20,'bob','Bob','b@x.com',10,?,?)`, now, now)
	db.Exec(`INSERT INTO users (id,username,display_name,email,invited_by,created_at,updated_at) VALUES (30,'carol','Carol','c@x.com',10,?,?)`, now, now)
	db.Exec(`INSERT INTO invite_records VALUES (1,10,20,1,50,50,?)`, now)
	db.Exec(`INSERT INTO invite_records VALUES (2,10,20,2,30,30,?)`, now) // two records for bob

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
	// Find bob in results
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
	if items[0].InviteeDisplayName != "Bob" && items[1].InviteeDisplayName != "Bob" {
		t.Error("display name not populated")
	}
}
```

- [ ] **Step 2: Run tests — expect FAIL**

```bash
go test ./internal/invite/ -run "TestGetSummary|TestMaskEmail|TestListInvitees|TestListRecords" -v
```

Expected: `FAIL — GetSummary/ListInvitees/ListRecords/maskEmail undefined`

- [ ] **Step 3: Add view types + helpers + methods to service.go**

Add after the existing imports (you'll need `"strings"` and `"time"`):

```go
// --- View types (match frontend JSON contract) ---

// SummaryResp is the response for GET /api/user/invites/me.
type SummaryResp struct {
	InviteCode  string    `json:"invite_code"`
	ShareURL    string    `json:"share_url"`
	RebateRatio float64   `json:"rebate_ratio"`
	Stats       StatsResp `json:"stats"`
}

// StatsResp holds the three counters inside SummaryResp.
type StatsResp struct {
	InviteeCount        int64 `json:"invitee_count"`
	RebateCreditsTotal  int64 `json:"rebate_credits_total"`
	RebateCreditsMonth  int64 `json:"rebate_credits_month"`
}

// InviteeView is one row in GET /api/user/invites/invitees.
type InviteeView struct {
	UserID             int64  `json:"user_id"`
	DisplayName        string `json:"display_name"`
	EmailMasked        string `json:"email_masked"`
	RegisteredAt       string `json:"registered_at"`
	TotalRebateCredits int64  `json:"total_rebate_credits"`
}

// RecordView is one row in GET /api/user/invites/records.
type RecordView struct {
	ID                 int64  `json:"id"`
	InviteeID          int64  `json:"invitee_id"`
	InviteeDisplayName string `json:"invitee_display_name"`
	OrderID            int64  `json:"order_id"`
	RebateCents        int64  `json:"rebate_cents"`
	RebateCredits      int64  `json:"rebate_credits"`
	CreatedAt          string `json:"created_at"`
}

// maskEmail converts "john@example.com" → "j***@example.com".
func maskEmail(email string) string {
	at := strings.Index(email, "@")
	if at <= 0 {
		return "***"
	}
	return string(email[0]) + "***" + email[at:]
}

// GetSummary returns the current user's invite summary (code, share URL, stats).
func (s *Service) GetSummary(ctx context.Context, userID int64) (*SummaryResp, error) {
	var u struct {
		InviteCode *string
	}
	if err := s.deps.DB.WithContext(ctx).Table("users").
		Select("invite_code").
		Where("id = ? AND deleted_at IS NULL", userID).
		Scan(&u).Error; err != nil {
		return nil, fmt.Errorf("invite: get user: %w", err)
	}
	inviteCode := ""
	if u.InviteCode != nil {
		inviteCode = *u.InviteCode
	}

	var inviteeCount int64
	if err := s.deps.DB.WithContext(ctx).Table("users").
		Where("invited_by = ? AND deleted_at IS NULL", userID).
		Count(&inviteeCount).Error; err != nil {
		return nil, fmt.Errorf("invite: count invitees: %w", err)
	}

	var totalCredits int64
	s.deps.DB.WithContext(ctx).Table("invite_records").
		Select("COALESCE(SUM(rebate_credits), 0)").
		Where("inviter_id = ?", userID).
		Scan(&totalCredits)

	now := s.deps.Clock.Now().UTC()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	var monthCredits int64
	s.deps.DB.WithContext(ctx).Table("invite_records").
		Select("COALESCE(SUM(rebate_credits), 0)").
		Where("inviter_id = ? AND created_at >= ?", userID, monthStart).
		Scan(&monthCredits)

	baseURL := s.deps.Setting.GetString(ctx, "site.base_url", "")
	shareURL := ""
	if inviteCode != "" {
		shareURL = baseURL + "/register?invite_code=" + inviteCode
	}

	return &SummaryResp{
		InviteCode:  inviteCode,
		ShareURL:    shareURL,
		RebateRatio: s.getRebateRatio(ctx),
		Stats: StatsResp{
			InviteeCount:       inviteeCount,
			RebateCreditsTotal: totalCredits,
			RebateCreditsMonth: monthCredits,
		},
	}, nil
}

// ListInvitees returns paginated users who registered via this user's invite code.
func (s *Service) ListInvitees(ctx context.Context, inviterID int64, page, size int) ([]*InviteeView, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 10
	}
	offset := (page - 1) * size

	var total int64
	if err := s.deps.DB.WithContext(ctx).Table("users").
		Where("invited_by = ? AND deleted_at IS NULL", inviterID).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("invite: count invitees: %w", err)
	}
	if total == 0 {
		return nil, 0, nil
	}

	type row struct {
		ID                 int64
		DisplayName        *string
		Email              *string
		CreatedAt          time.Time
		TotalRebateCredits int64
	}
	var rows []row
	err := s.deps.DB.WithContext(ctx).Table("users").
		Select("users.id, users.display_name, users.email, users.created_at, COALESCE(SUM(ir.rebate_credits), 0) AS total_rebate_credits").
		Joins("LEFT JOIN invite_records ir ON ir.invitee_id = users.id AND ir.inviter_id = ?", inviterID).
		Where("users.invited_by = ? AND users.deleted_at IS NULL", inviterID).
		Group("users.id, users.display_name, users.email, users.created_at").
		Order("users.created_at DESC").
		Offset(offset).Limit(size).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, fmt.Errorf("invite: list invitees: %w", err)
	}

	views := make([]*InviteeView, len(rows))
	for i, r := range rows {
		dn := ""
		if r.DisplayName != nil {
			dn = *r.DisplayName
		}
		email := ""
		if r.Email != nil {
			email = *r.Email
		}
		views[i] = &InviteeView{
			UserID:             r.ID,
			DisplayName:        dn,
			EmailMasked:        maskEmail(email),
			RegisteredAt:       r.CreatedAt.UTC().Format(time.RFC3339),
			TotalRebateCredits: r.TotalRebateCredits,
		}
	}
	return views, total, nil
}

// ListRecords returns paginated rebate records where this user is the inviter.
func (s *Service) ListRecords(ctx context.Context, inviterID int64, page, size int) ([]*RecordView, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 10
	}
	offset := (page - 1) * size

	total, err := s.deps.Repo.CountByInviter(ctx, inviterID)
	if err != nil {
		return nil, 0, fmt.Errorf("invite: count records: %w", err)
	}
	if total == 0 {
		return nil, 0, nil
	}

	records, err := s.deps.Repo.ListByInviter(ctx, inviterID, size, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("invite: list records: %w", err)
	}
	if len(records) == 0 {
		return nil, total, nil
	}

	// Batch-fetch display names for invitees.
	inviteeIDs := make([]int64, len(records))
	for i, r := range records {
		inviteeIDs[i] = r.InviteeID
	}
	type nameRow struct {
		ID          int64
		DisplayName *string
	}
	var nameRows []nameRow
	s.deps.DB.WithContext(ctx).Table("users").
		Select("id, display_name").
		Where("id IN ?", inviteeIDs).
		Scan(&nameRows)
	nameMap := make(map[int64]string, len(nameRows))
	for _, nr := range nameRows {
		if nr.DisplayName != nil {
			nameMap[nr.ID] = *nr.DisplayName
		}
	}

	views := make([]*RecordView, len(records))
	for i, r := range records {
		views[i] = &RecordView{
			ID:                 r.ID,
			InviteeID:          r.InviteeID,
			InviteeDisplayName: nameMap[r.InviteeID],
			OrderID:            r.OrderID,
			RebateCents:        r.RebateCents,
			RebateCredits:      r.RebateCredits,
			CreatedAt:          r.CreatedAt.UTC().Format(time.RFC3339),
		}
	}
	return views, total, nil
}
```

Also add `"strings"` and `"time"` to the imports block in `service.go` (it already has `"context"` and `"fmt"`).

- [ ] **Step 4: Run tests — expect PASS**

```bash
go test ./internal/invite/ -run "TestGetSummary|TestMaskEmail|TestListInvitees|TestListRecords" -v
```

Expected: all PASS

- [ ] **Step 5: Run all invite tests**

```bash
go test ./internal/invite/... -v
```

Expected: all PASS

- [ ] **Step 6: Commit**

```bash
git add internal/invite/service.go internal/invite/service_query_test.go
git commit -m "feat(invite): GetSummary/ListInvitees/ListRecords + view types + maskEmail"
```

---

## Task 3: HTTP handler

**Files:**
- Create: `internal/server/handler/user/invite.go`
- Create: `internal/server/handler/user/invite_test.go`

- [ ] **Step 1: Write failing handler test**

Create `internal/server/handler/user/invite_test.go`:

```go
package user

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ijry/pro-api/internal/invite"
	"github.com/ijry/pro-api/internal/server/middleware"
)

// mockInviteSvc is a stub invite.Service for handler tests.
// We can't easily embed *invite.Service (unexported fields), so we define
// a thin interface and adapt — but since InviteHandler takes *invite.Service
// directly, we use a real Service with no DB (methods return zero values on nil DB).
// Instead, we test via the exported handler methods with a stub that wraps a minimal service.
// For simplicity: construct with a real service pointing at nil deps and test only the
// routing/JSON shape; service correctness is covered by service_query_test.go.

func buildInviteRouter(uid int64, svc *invite.Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorResponse("json"))
	h := NewInviteHandler(svc, func(*gin.Context) int64 { return uid })
	h.Register(r.Group("/invites"))
	return r
}

func TestInviteHandler_Me_Unauthenticated(t *testing.T) {
	// uid=0 → 401
	svc := &invite.Service{} // zero-value service; Me will fail auth before calling svc
	r := buildInviteRouter(0, svc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/invites/me", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func TestInviteHandler_Me_OK(t *testing.T) {
	// Use a real service backed by an in-memory DB (reuse setupQueryDB via helper below)
	db, svc := buildTestService(t)
	// Seed user with invite code
	db.Exec(`INSERT INTO users (id,username,invite_code,invited_by,created_at,updated_at) VALUES (5,'u','CODE5',0,datetime('now'),datetime('now'))`)

	r := buildInviteRouter(5, svc)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/invites/me", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp struct {
		Data struct {
			InviteCode string `json:"invite_code"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.Data.InviteCode != "CODE5" {
		t.Errorf("invite_code: got %q", resp.Data.InviteCode)
	}
}

func TestInviteHandler_Invitees_Pagination(t *testing.T) {
	db, svc := buildTestService(t)
	now := "datetime('now')"
	// User 7 invited 12 others
	db.Exec(`INSERT INTO users (id,username,invited_by,created_at,updated_at) VALUES (7,'host',0,` + now + `,` + now + `)`)
	for i := 0; i < 12; i++ {
		db.Exec(`INSERT INTO users (id,username,invited_by,created_at,updated_at) VALUES (?,?,7,`+now+`,`+now+`)`,
			100+i, fmt.Sprintf("u%d", i))
	}

	r := buildInviteRouter(7, svc)

	// page 1 size 10
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/invites/invitees?page=1&size=10", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp struct {
		Data struct {
			Items []json.RawMessage `json:"items"`
			Total int64             `json:"total"`
		} `json:"data"`
	}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Data.Total != 12 {
		t.Errorf("total: want 12, got %d", resp.Data.Total)
	}
	if len(resp.Data.Items) != 10 {
		t.Errorf("page 1 items: want 10, got %d", len(resp.Data.Items))
	}
}
```

You'll also need a helper `buildTestService(t)` at the bottom of the test file (reuses the SQLite setup from `service_query_test.go`). Add this as a separate helper file or include inline:

```go
// buildTestService creates an in-memory DB + Service for handler tests.
// Placed here to avoid package-level init ordering.
func buildTestService(t *testing.T) (*gorm.DB, *invite.Service) {
	t.Helper()
	// ... (same body as setupQueryDB in service_query_test.go, but in package user)
	// We can't import internal/invite test helpers across packages,
	// so replicate the minimal setup:
	name := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", name)), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	db.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY, username TEXT UNIQUE, email TEXT,
			display_name TEXT, avatar TEXT, role INTEGER DEFAULT 0, status INTEGER DEFAULT 0,
			group_id INTEGER, invite_code TEXT UNIQUE, invited_by INTEGER NOT NULL DEFAULT 0,
			email_verified_at DATETIME, last_login_at DATETIME, last_login_ip TEXT,
			created_at DATETIME, updated_at DATETIME, deleted_at DATETIME
		);
		CREATE TABLE invite_records (
			id INTEGER PRIMARY KEY, inviter_id INTEGER NOT NULL, invitee_id INTEGER NOT NULL,
			order_id INTEGER NOT NULL, rebate_cents INTEGER NOT NULL DEFAULT 0,
			rebate_credits INTEGER NOT NULL DEFAULT 0, created_at DATETIME
		);
	`)
	svc := invite.NewService(invite.Deps{
		Repo:    invite.NewRepository(db),
		DB:      db,
		Setting: &inviteTestSetting{},
		IDGen:   &inviteTestIDGen{},
		Clock:   clock.Real,
	})
	return db, svc
}

type inviteTestSetting struct{}
func (s *inviteTestSetting) Get(_ context.Context, _ string) ([]byte, bool)               { return nil, false }
func (s *inviteTestSetting) GetString(_ context.Context, _ string, def string) string      { return def }
func (s *inviteTestSetting) GetFloat(_ context.Context, _ string, def float64) float64    { return def }
func (s *inviteTestSetting) GetBool(_ context.Context, _ string, def bool) bool            { return def }
func (s *inviteTestSetting) GetInt(_ context.Context, _ string, def int) int               { return def }

type inviteTestIDGen struct{ n int64 }
func (g *inviteTestIDGen) Generate() int64 { g.n++; return g.n }
```

Required imports for `invite_test.go`:

```go
import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ijry/pro-api/internal/invite"
	"github.com/ijry/pro-api/internal/server/middleware"
	"github.com/ijry/pro-api/internal/util/clock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)
```

- [ ] **Step 2: Run tests — expect FAIL (InviteHandler undefined)**

```bash
go test ./internal/server/handler/user/ -run "TestInviteHandler" -v
```

Expected: `FAIL — NewInviteHandler undefined`

- [ ] **Step 3: Create invite.go handler**

Create `internal/server/handler/user/invite.go`:

```go
package user

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ijry/pro-api/internal/invite"
	"github.com/ijry/pro-api/internal/server/middleware"
	"github.com/ijry/pro-api/pkg/apierr"
)

// InviteHandler serves the user-side invite endpoints.
type InviteHandler struct {
	Svc    *invite.Service
	UserOf func(c *gin.Context) int64
}

// NewInviteHandler constructs an InviteHandler.
func NewInviteHandler(svc *invite.Service, userOf func(c *gin.Context) int64) *InviteHandler {
	if userOf == nil {
		userOf = func(*gin.Context) int64 { return 0 }
	}
	return &InviteHandler{Svc: svc, UserOf: userOf}
}

// Register mounts the 3 invite routes onto r.
func (h *InviteHandler) Register(r gin.IRouter) {
	r.GET("/me", h.Me)
	r.GET("/invitees", h.Invitees)
	r.GET("/records", h.Records)
}

func (h *InviteHandler) requireUser(c *gin.Context) (int64, bool) {
	uid := h.UserOf(c)
	if uid <= 0 {
		middleware.SetErr(c, apierr.New(apierr.CodeNotLoggedIn, "请先登录"))
		return 0, false
	}
	return uid, true
}

// Me GET /invites/me — returns InviteSummary.
func (h *InviteHandler) Me(c *gin.Context) {
	uid, ok := h.requireUser(c)
	if !ok {
		return
	}
	resp, err := h.Svc.GetSummary(c.Request.Context(), uid)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// Invitees GET /invites/invitees?page=&size= — paginated invitees list.
func (h *InviteHandler) Invitees(c *gin.Context) {
	uid, ok := h.requireUser(c)
	if !ok {
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	items, total, err := h.Svc.ListInvitees(c.Request.Context(), uid, page, size)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{
		"items": items,
		"total": total,
		"page":  page,
		"size":  size,
	}})
}

// Records GET /invites/records?page=&size= — paginated rebate records.
func (h *InviteHandler) Records(c *gin.Context) {
	uid, ok := h.requireUser(c)
	if !ok {
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	items, total, err := h.Svc.ListRecords(c.Request.Context(), uid, page, size)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{
		"items": items,
		"total": total,
		"page":  page,
		"size":  size,
	}})
}
```

Note: `writeErr` is already defined in `notice.go` (same package) — do not redefine it.

- [ ] **Step 4: Run handler tests — expect PASS**

```bash
go test ./internal/server/handler/user/ -run "TestInviteHandler" -v
```

Expected: all PASS

- [ ] **Step 5: Commit**

```bash
git add internal/server/handler/user/invite.go internal/server/handler/user/invite_test.go
git commit -m "feat(user/handler): InviteHandler — Me/Invitees/Records + tests"
```

---

## Task 4: Wire function

**Files:**
- Modify: `internal/server/handler/user/wire.go`

- [ ] **Step 1: Add WireInvite to wire.go**

Open `internal/server/handler/user/wire.go`. Current imports are:
```go
import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/ijry/pro-api/internal/app"
	"github.com/ijry/pro-api/internal/channel"
	"github.com/ijry/pro-api/internal/notice"
	"github.com/ijry/pro-api/internal/relay"
)
```

Add `"github.com/ijry/pro-api/internal/invite"` to imports and append this function:

```go
// WireInvite constructs an InviteHandler for the user API group.
//
// The invite service is instantiated with wallet=nil because these HTTP handlers
// only call read methods (GetSummary, ListInvitees, ListRecords). OnOrderPaid —
// the only method that uses the wallet — is called exclusively from the payment
// service, which wires its own invite.Service instance with a real wallet.
func WireInvite(a *app.Application, userOf func(*gin.Context) int64) (*InviteHandler, error) {
	if a == nil {
		return nil, errors.New("user: app is nil")
	}
	svc := invite.Wire(a, nil)
	return NewInviteHandler(svc, userOf), nil
}
```

- [ ] **Step 2: Verify compile**

```bash
go build ./internal/server/handler/user/...
```

Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add internal/server/handler/user/wire.go
git commit -m "feat(user/handler): WireInvite"
```

---

## Task 5: Register routes in main.go

**Files:**
- Modify: `cmd/proapi/main.go`

- [ ] **Step 1: Add wireInviteRoutes function**

At the bottom of `cmd/proapi/main.go`, add:

```go
// wireInviteRoutes mounts the 3 user invite endpoints under /api/user/invites.
//
// Routes: GET /me, GET /invitees, GET /records — all require SessionAuth.
func wireInviteRoutes(eng *gin.Engine, a *app.Application, log *zap.Logger) {
	sessStore := authhwire.SessionStoreFrom(a)
	if sessStore == nil {
		log.Warn("session store not available; invite routes skipped")
		return
	}
	userOf := func(c *gin.Context) int64 { return middleware.UserID(c) }
	h, err := userhdr.WireInvite(a, userOf)
	if err != nil {
		log.Warn("invite handler wiring failed", zap.Error(err))
		return
	}
	inviteG := eng.Group("/api/user/invites",
		middleware.ErrorResponse("json"),
		middleware.SessionAuth(sessStore, a.Clock),
	)
	h.Register(inviteG)
}
```

- [ ] **Step 2: Call wireInviteRoutes from wireRoutes**

In `wireRoutes`, after the Playground block (around line 267), add:

```go
// Invite 路由(邀请返佣)
wireInviteRoutes(eng, a, log)
```

- [ ] **Step 3: Verify build**

```bash
go build ./cmd/proapi/...
```

Expected: no errors. If there are import issues, check that `userhdr` alias resolves to `internal/server/handler/user` in the import block at the top of `main.go`.

- [ ] **Step 4: Commit**

```bash
git add cmd/proapi/main.go
git commit -m "feat(server): /api/user/invites/* 路由装配"
```

---

## Task 6: Full verification

**Files:** none (read-only verification)

- [ ] **Step 1: Run all invite package tests**

```bash
go test ./internal/invite/... -v
```

Expected: all tests PASS (TestCountByInviter, TestGetSummary, TestMaskEmail, TestListInvitees, TestListRecords)

- [ ] **Step 2: Run handler tests**

```bash
go test ./internal/server/handler/user/... -v
```

Expected: all tests PASS (including new TestInviteHandler_*)

- [ ] **Step 3: Full build**

```bash
go build ./...
```

Expected: no errors

- [ ] **Step 4: Commit (if anything was adjusted)**

Only if Step 1-3 required fixes. Otherwise no commit needed.

---

## Self-Review

### Spec coverage (against `docs/superpowers/specs/2026-05-26-user-invites-and-stub-cleanup-design.md`)

| Spec requirement | Task |
|---|---|
| `GET /api/user/invites/me` → `InviteSummary` | Task 2 (service) + Task 3 (handler) |
| `GET /api/user/invites/invitees?page=&size=` → paginated `Invitee` | Task 2 (service) + Task 3 (handler) |
| `GET /api/user/invites/records?page=&size=` → paginated `InviteRecord` | Task 2 (service) + Task 3 (handler) |
| `email_masked` at service layer | Task 2 (`maskEmail`) |
| `rebate_credits_total` via SUM | Task 2 (`GetSummary`) |
| `rebate_credits_month` via SUM + month filter | Task 2 (`GetSummary`) |
| `share_url` = `site.base_url` + path | Task 2 (`GetSummary`) |
| 分页对齐 (page/size/total in response) | Task 3 (handler JSON shape) |
| Route registration with SessionAuth | Task 5 |

### Placeholder scan
None found — all steps contain complete code.

### Type consistency
- `invite.SummaryResp`, `invite.InviteeView`, `invite.RecordView` defined in Task 2, used in Task 3 ✓
- `InviteHandler` defined in Task 3, wired in Task 4 ✓
- `WireInvite` defined in Task 4, called in Task 5 ✓
- `wireInviteRoutes` defined and called in Task 5 ✓
