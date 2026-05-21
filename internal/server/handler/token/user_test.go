package token

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/ijry/pro-api/internal/audit"
	"github.com/ijry/pro-api/internal/server/middleware"
	tokensvc "github.com/ijry/pro-api/internal/token"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// stubIDGen 步进 id 生成器。
type stubIDGen struct {
	mu sync.Mutex
	n  int64
}

func (g *stubIDGen) Generate() int64 {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.n++
	return g.n
}

func newHandlerFixture(t *testing.T) (tokensvc.Store, func()) {
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
	store, err := tokensvc.New(tokensvc.Config{
		DB:            db,
		Cache:         rdb,
		Log:           zap.NewNop(),
		IDGen:         &stubIDGen{},
		Audit:         audit.NewNoop(),
		FlushInterval: 24 * time.Hour,
	})
	if err != nil {
		t.Fatal(err)
	}
	cleanup := func() {
		_ = store.Close()
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	}
	return store, cleanup
}

func newUserEngine(_ *testing.T, store tokensvc.Store, currentUser int64) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorResponse("json"))
	uh := NewUserHandler(store, func(_ *gin.Context) int64 { return currentUser })
	g := r.Group("/api/user/tokens")
	uh.Register(g)
	return r
}

func doJSON(t *testing.T, r *gin.Engine, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatal(err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

// createForUser 在 store 中给 user 直接建一个 token,返回 id 和 plaintext。
func createForUser(t *testing.T, store tokensvc.Store, uid int64, name string) (int64, string) {
	t.Helper()
	plaintext, view, err := store.Create(context.Background(), tokensvc.CreateInput{
		UserID: uid,
		Name:   name,
	})
	if err != nil {
		t.Fatal(err)
	}
	return view.ID, plaintext
}

func TestUserHandler_List_ReturnsOnlyOwnTokens(t *testing.T) {
	store, cleanup := newHandlerFixture(t)
	defer cleanup()
	_, _ = createForUser(t, store, 1, "alice")
	_, _ = createForUser(t, store, 2, "bob")

	r := newUserEngine(t, store, 1)
	rec := doJSON(t, r, http.MethodGet, "/api/user/tokens", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var body ListResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body.Total != 1 || len(body.Items) != 1 {
		t.Fatalf("want 1 own, got %+v", body)
	}
}

func TestUserHandler_Create_ReturnsPlaintextOnce(t *testing.T) {
	store, cleanup := newHandlerFixture(t)
	defer cleanup()
	r := newUserEngine(t, store, 1)
	rec := doJSON(t, r, http.MethodPost, "/api/user/tokens", CreateRequest{
		Name: "demo",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var body CreateResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body.PlaintextKey == "" || body.PlaintextKey[:3] != "pa-" {
		t.Fatalf("bad plaintext: %q", body.PlaintextKey)
	}
	if body.View.ID == 0 || body.View.UserID != 1 {
		t.Fatalf("bad view: %+v", body.View)
	}
}

func TestUserHandler_Create_NameRequired(t *testing.T) {
	store, cleanup := newHandlerFixture(t)
	defer cleanup()
	r := newUserEngine(t, store, 1)
	rec := doJSON(t, r, http.MethodPost, "/api/user/tokens", CreateRequest{})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestUserHandler_Create_BadAllowedIPs(t *testing.T) {
	store, cleanup := newHandlerFixture(t)
	defer cleanup()
	r := newUserEngine(t, store, 1)
	rec := doJSON(t, r, http.MethodPost, "/api/user/tokens", CreateRequest{
		Name:       "t",
		AllowedIPs: []string{"this-is-not-an-ip"},
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestUserHandler_Create_BadAllowedModels(t *testing.T) {
	store, cleanup := newHandlerFixture(t)
	defer cleanup()
	r := newUserEngine(t, store, 1)
	rec := doJSON(t, r, http.MethodPost, "/api/user/tokens", CreateRequest{
		Name:          "t",
		AllowedModels: []string{"bad model with space"},
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestUserHandler_Create_ExpiresAtInPast_Rejected(t *testing.T) {
	store, cleanup := newHandlerFixture(t)
	defer cleanup()
	r := newUserEngine(t, store, 1)
	past := time.Now().Add(-time.Hour)
	rec := doJSON(t, r, http.MethodPost, "/api/user/tokens", CreateRequest{
		Name:      "t",
		ExpiresAt: &past,
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestUserHandler_Patch_ForbiddenOnOtherUserToken(t *testing.T) {
	store, cleanup := newHandlerFixture(t)
	defer cleanup()
	otherID, _ := createForUser(t, store, 2, "bob")
	r := newUserEngine(t, store, 1)
	newName := "stolen"
	rec := doJSON(t, r, http.MethodPatch,
		"/api/user/tokens/"+intToStr(otherID),
		PatchRequest{Name: &newName})
	if rec.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestUserHandler_Patch_UpdatesOwnToken(t *testing.T) {
	store, cleanup := newHandlerFixture(t)
	defer cleanup()
	id, _ := createForUser(t, store, 1, "t")
	r := newUserEngine(t, store, 1)
	rpm := 99
	rec := doJSON(t, r, http.MethodPatch, "/api/user/tokens/"+intToStr(id), PatchRequest{RPMLimit: &rpm})
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var body ViewDTO
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body.RPMLimit != 99 {
		t.Fatalf("want rpm=99, got %d", body.RPMLimit)
	}
}

func TestUserHandler_Delete_RevokeOwnToken(t *testing.T) {
	store, cleanup := newHandlerFixture(t)
	defer cleanup()
	id, _ := createForUser(t, store, 1, "t")
	r := newUserEngine(t, store, 1)
	rec := doJSON(t, r, http.MethodDelete, "/api/user/tokens/"+intToStr(id), nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestUserHandler_Regenerate_ReturnsNewKey(t *testing.T) {
	store, cleanup := newHandlerFixture(t)
	defer cleanup()
	id, oldKey := createForUser(t, store, 1, "t")
	r := newUserEngine(t, store, 1)
	rec := doJSON(t, r, http.MethodPost, "/api/user/tokens/"+intToStr(id)+"/regenerate", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var body CreateResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body.PlaintextKey == "" || body.PlaintextKey == oldKey {
		t.Fatalf("regenerate should return new plaintext: %q -> %q", oldKey, body.PlaintextKey)
	}
}

func TestUserHandler_Unauthenticated_Returns401(t *testing.T) {
	store, cleanup := newHandlerFixture(t)
	defer cleanup()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorResponse("json"))
	uh := NewUserHandler(store, func(_ *gin.Context) int64 { return 0 })
	uh.Register(r.Group("/api/user/tokens"))
	rec := doJSON(t, r, http.MethodGet, "/api/user/tokens", nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", rec.Code)
	}
}

func intToStr(n int64) string {
	const digits = "0123456789"
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	buf := []byte{}
	for n > 0 {
		buf = append([]byte{digits[n%10]}, buf...)
		n /= 10
	}
	if neg {
		buf = append([]byte{'-'}, buf...)
	}
	return string(buf)
}
