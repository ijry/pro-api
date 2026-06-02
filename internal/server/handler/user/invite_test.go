package user

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
	"github.com/ijry/pro-api/internal/setting"
	"github.com/ijry/pro-api/internal/util/clock"
	"github.com/ijry/pro-api/internal/util/idgen"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// inviteTestSetting is a minimal setting.Store stub for handler tests.
type inviteTestSetting struct{}

func (s *inviteTestSetting) Get(_ context.Context, _ string) (json.RawMessage, bool) {
	return nil, false
}
func (s *inviteTestSetting) GetString(_ context.Context, _ string, def string) string { return def }
func (s *inviteTestSetting) GetBool(_ context.Context, _ string, def bool) bool       { return def }
func (s *inviteTestSetting) GetInt(_ context.Context, _ string, def int) int           { return def }
func (s *inviteTestSetting) GetFloat(_ context.Context, _ string, def float64) float64 { return def }
func (s *inviteTestSetting) GetJSON(_ context.Context, _ string, _ any) error          { return nil }
func (s *inviteTestSetting) Put(_ context.Context, _ string, _ any, _ int64) error     { return nil }
func (s *inviteTestSetting) Close() error                                               { return nil }
func (s *inviteTestSetting) GetSecret(_ context.Context, _ string, _ setting.Decryptor) (string, error) {
	return "", nil
}
func (s *inviteTestSetting) ListAll(_ context.Context) ([]setting.Setting, error) { return nil, nil }

// buildTestInviteService creates an in-memory DB + invite.Service for handler tests.
func buildTestInviteService(t *testing.T) (*gorm.DB, *invite.Service) {
	t.Helper()
	name := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", name)), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`
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
	`).Error; err != nil {
		t.Fatal(err)
	}
	gen, err := idgen.New(1)
	if err != nil {
		t.Fatal(err)
	}
	svc := invite.NewService(invite.Deps{
		Repo:    invite.NewRepository(db),
		DB:      db,
		Setting: &inviteTestSetting{},
		IDGen:   gen,
		Clock:   clock.Real,
	})
	return db, svc
}

func buildInviteRouter(uid int64, svc *invite.Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorResponse("json"))
	h := NewInviteHandler(svc, func(*gin.Context) int64 { return uid })
	h.Register(r.Group("/invites"))
	return r
}

func TestInviteHandler_Me_Unauthenticated(t *testing.T) {
	_, svc := buildTestInviteService(t)
	r := buildInviteRouter(0, svc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/invites/me", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func TestInviteHandler_Me_OK(t *testing.T) {
	db, svc := buildTestInviteService(t)
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
	db, svc := buildTestInviteService(t)
	db.Exec(`INSERT INTO users (id,username,invited_by,created_at,updated_at) VALUES (7,'host',0,datetime('now'),datetime('now'))`)
	for i := 0; i < 12; i++ {
		db.Exec(fmt.Sprintf(
			`INSERT INTO users (id,username,invited_by,created_at,updated_at) VALUES (%d,'u%d',7,datetime('now'),datetime('now'))`,
			100+i, i,
		))
	}

	r := buildInviteRouter(7, svc)
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

func TestInviteHandler_Records_Empty(t *testing.T) {
	_, svc := buildTestInviteService(t)
	r := buildInviteRouter(42, svc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/invites/records", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
	var resp struct {
		Data struct {
			Items []json.RawMessage `json:"items"`
			Total int64             `json:"total"`
		} `json:"data"`
	}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Data.Total != 0 {
		t.Errorf("want total=0, got %d", resp.Data.Total)
	}
}
