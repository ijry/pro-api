package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/ijry/pro-api/internal/auth/session"
	"github.com/ijry/pro-api/internal/util/clock"
	"github.com/ijry/pro-api/pkg/apierr"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func boolPtr(b bool) *bool { return &b }

func newSessionStore(t *testing.T) session.Store {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	dbName := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	db, _ := gorm.Open(sqlite.Open("file:"+dbName+"?mode=memory&cache=shared"), &gorm.Config{})
	_ = db.Exec(`CREATE TABLE sessions (id TEXT PRIMARY KEY, user_id INTEGER, ip TEXT, user_agent TEXT, created_at DATETIME, last_seen_at DATETIME, expires_at DATETIME, revoked_at DATETIME)`).Error
	off := false
	s, err := session.New(session.Deps{
		DB: session.NewRepository(db), Cache: rdb, Clock: clock.Real,
	}, session.Config{TTL: time.Hour, Sliding: true, MirrorBatchSize: 1, MirrorBatchEvery: time.Minute, RestoreOnStart: &off})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func newAuthRouter(t *testing.T) (*gin.Engine, session.Store) {
	gin.SetMode(gin.TestMode)
	s := newSessionStore(t)
	r := gin.New()
	r.Use(ErrorResponse("json"), SessionAuth(s, clock.Real))
	r.GET("/me", func(c *gin.Context) {
		c.JSON(200, gin.H{"uid": UserID(c), "role": Role(c), "sid": SessionID(c)})
	})
	return r, s
}

func TestSessionAuth_NoCookie_NotLoggedIn(t *testing.T) {
	r, _ := newAuthRouter(t)
	req := httptest.NewRequest("GET", "/me", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func TestSessionAuth_ValidCookie_InjectsCtx(t *testing.T) {
	r, s := newAuthRouter(t)
	sess, _ := s.Create(context.Background(), 7, 2, "", "")
	req := httptest.NewRequest("GET", "/me", nil)
	req.AddCookie(&http.Cookie{Name: CookieSession, Value: sess.ID})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("got %d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"uid":7`) {
		t.Fatalf("body: %s", w.Body.String())
	}
}

func TestSessionAuth_FallbackHeader(t *testing.T) {
	r, s := newAuthRouter(t)
	sess, _ := s.Create(context.Background(), 7, 2, "", "")
	req := httptest.NewRequest("GET", "/me", nil)
	req.Header.Set(HeaderSession, sess.ID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("got %d", w.Code)
	}
}

func TestSessionAuth_ExpiredSession(t *testing.T) {
	r, _ := newAuthRouter(t)
	req := httptest.NewRequest("GET", "/me", nil)
	req.AddCookie(&http.Cookie{Name: CookieSession, Value: "sess_notexist"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "20002") {
		t.Fatalf("want CodeSessionExpired (20002): %s", w.Body.String())
	}
}

// --- CSRF ---

func newCSRFRouter(whitelist []string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(ErrorResponse("json"), CSRF(whitelist))
	r.POST("/protected", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
	r.POST("/auth/login", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
	r.GET("/get", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
	return r
}

func TestCSRF_GETNotChecked(t *testing.T) {
	r := newCSRFRouter(nil)
	req := httptest.NewRequest("GET", "/get", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("got %d", w.Code)
	}
}

func TestCSRF_POSTMissingHeader_Forbidden(t *testing.T) {
	r := newCSRFRouter(nil)
	req := httptest.NewRequest("POST", "/protected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d", w.Code)
	}
}

func TestCSRF_POSTHeaderMatch_OK(t *testing.T) {
	r := newCSRFRouter(nil)
	req := httptest.NewRequest("POST", "/protected", nil)
	req.AddCookie(&http.Cookie{Name: CookieCSRF, Value: "TOK"})
	req.Header.Set(HeaderCSRF, "TOK")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("got %d", w.Code)
	}
}

func TestCSRF_POSTHeaderMismatch_Forbidden(t *testing.T) {
	r := newCSRFRouter(nil)
	req := httptest.NewRequest("POST", "/protected", nil)
	req.AddCookie(&http.Cookie{Name: CookieCSRF, Value: "A"})
	req.Header.Set(HeaderCSRF, "B")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d", w.Code)
	}
}

func TestCSRF_WhitelistedPath_Skipped(t *testing.T) {
	r := newCSRFRouter([]string{"/auth/login"})
	req := httptest.NewRequest("POST", "/auth/login", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("got %d", w.Code)
	}
}

// --- RoleGate ---

func TestRoleGate_LowerRole_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(ErrorResponse("json"), func(c *gin.Context) { c.Set(CtxKeyRole, int8(0)); c.Next() }, RoleGate(2))
	r.GET("/admin", func(c *gin.Context) { c.JSON(200, gin.H{}) })
	req := httptest.NewRequest("GET", "/admin", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d", w.Code)
	}
}

func TestRoleGate_HigherRole_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(ErrorResponse("json"), func(c *gin.Context) { c.Set(CtxKeyRole, int8(3)); c.Next() }, RoleGate(2))
	r.GET("/admin", func(c *gin.Context) { c.JSON(200, gin.H{}) })
	req := httptest.NewRequest("GET", "/admin", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("got %d", w.Code)
	}
}

func TestDeriveCSRFToken_Stable(t *testing.T) {
	v1 := DeriveCSRFToken([]byte("key"), "sess_x")
	v2 := DeriveCSRFToken([]byte("key"), "sess_x")
	if v1 != v2 || len(v1) != 32 {
		t.Fatalf("want stable 32-char token, got %q %q", v1, v2)
	}
}

var _ = apierr.CodeForbidden
var _ = boolPtr
