package authhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/ijry/pro-api/internal/audit"
	"github.com/ijry/pro-api/internal/auth"
	"github.com/ijry/pro-api/internal/auth/oauth"
	"github.com/ijry/pro-api/internal/auth/session"
	"github.com/ijry/pro-api/internal/auth/verifycode"
	"github.com/ijry/pro-api/internal/group"
	"github.com/ijry/pro-api/internal/notify/email"
	"github.com/ijry/pro-api/internal/server/middleware"
	"github.com/ijry/pro-api/internal/setting"
	"github.com/ijry/pro-api/internal/user"
	"github.com/ijry/pro-api/internal/util/clock"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type fakeIDGen struct{ n int64 }

func (f *fakeIDGen) Generate() int64 { f.n++; return f.n }

// fakeSetting 内联实现 setting.Store。
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
func (f *fakeSetting) GetBool(_ context.Context, key string, def bool) bool {
	v, ok := f.kv[key]
	if !ok {
		return def
	}
	return v.(bool)
}
func (f *fakeSetting) GetInt(_ context.Context, key string, def int) int {
	v, ok := f.kv[key]
	if !ok {
		return def
	}
	return v.(int)
}
func (f *fakeSetting) GetFloat(_ context.Context, key string, def float64) float64 {
	v, ok := f.kv[key]
	if !ok {
		return def
	}
	return v.(float64)
}
func (f *fakeSetting) GetJSON(_ context.Context, key string, dest any) error {
	v, ok := f.kv[key]
	if !ok {
		return setting.ErrNotFound
	}
	b, _ := json.Marshal(v)
	return json.Unmarshal(b, dest)
}
func (f *fakeSetting) Put(_ context.Context, key string, val any, _ int64) error {
	f.kv[key] = val
	return nil
}
func (f *fakeSetting) Close() error { return nil }

type captureMailer struct{ msgs []email.Message }

func (c *captureMailer) Send(_ context.Context, m email.Message) error {
	c.msgs = append(c.msgs, m)
	return nil
}

type testHarness struct {
	t        *testing.T
	r        *gin.Engine
	auth     auth.Service
	user     user.Service
	group    group.Service
	session  session.Store
	oauthRepo oauth.Repository
	setting  *fakeSetting
	mailer   *captureMailer
	db       *gorm.DB
}

func setupHarness(t *testing.T) *testHarness {
	t.Helper()
	gin.SetMode(gin.TestMode)
	dbName := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", dbName)), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY, username TEXT UNIQUE, email TEXT UNIQUE, password_hash TEXT,
			display_name TEXT, avatar TEXT, role INTEGER, status INTEGER, group_id INTEGER,
			email_verified_at DATETIME, last_login_at DATETIME, last_login_ip TEXT,
			created_at DATETIME, updated_at DATETIME, deleted_at DATETIME
		);
		CREATE TABLE user_groups (
			id INTEGER PRIMARY KEY, name TEXT UNIQUE, display_name TEXT, ratio REAL,
			priority INTEGER, status INTEGER, created_at DATETIME, updated_at DATETIME
		);
		INSERT INTO user_groups VALUES (1, 'default', '普通', 1.0, 0, 0, datetime('now'), datetime('now'));
		CREATE TABLE sessions (
			id TEXT PRIMARY KEY, user_id INTEGER, ip TEXT, user_agent TEXT,
			created_at DATETIME, last_seen_at DATETIME, expires_at DATETIME, revoked_at DATETIME
		);
		CREATE TABLE oauth_bindings (
			id INTEGER PRIMARY KEY, user_id INTEGER, provider TEXT, provider_uid TEXT,
			email TEXT, profile TEXT, created_at DATETIME, updated_at DATETIME,
			UNIQUE (provider, provider_uid)
		);
	`).Error; err != nil {
		t.Fatal(err)
	}
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	gsvc := group.NewService(group.NewRepository(db), clock.Real, &fakeIDGen{})
	usvc := user.NewService(user.NewRepository(db), gsvc, &fakeIDGen{}, clock.Real)
	off := false
	sess, err := session.New(session.Deps{
		DB: session.NewRepository(db), Cache: rdb, Clock: clock.Real,
	}, session.Config{TTL: time.Hour, Sliding: true, MirrorBatchSize: 1, MirrorBatchEvery: time.Minute, RestoreOnStart: &off})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sess.Close() })
	vc := verifycode.New(rdb, clock.Real, nil, verifycode.Config{})
	mailer := &captureMailer{}
	stateStore := oauth.NewStateStore(rdb, clock.Real)
	oauthRepo := oauth.NewRepository(db)
	cfg := &fakeSetting{kv: map[string]any{
		"auth.allow_register":              true,
		"auth.email_verification_required": false,
		"auth.password.min_length":         8,
		"auth.password.require_mixed":      false,
	}}
	authSvc := auth.NewService(auth.Deps{
		User: usvc, Group: gsvc, Session: sess, VerifyCode: vc,
		Mailer: mailer, GithubState: stateStore, OAuthRepo: oauthRepo,
		Setting: cfg, Audit: audit.NewNoop(), Clock: clock.Real, IDGen: &fakeIDGen{},
	})
	r := gin.New()
	r.Use(middleware.ErrorResponse("json"))
	RegisterRoutes(r, Deps{
		Auth: authSvc, User: usvc, Group: gsvc, Session: sess,
		OAuthRepo: oauthRepo, Setting: cfg, Audit: audit.NewNoop(),
		Clock: clock.Real, CSRFKey: []byte("master-key"), CookieSecure: false,
	})
	return &testHarness{t: t, r: r, auth: authSvc, user: usvc, group: gsvc, session: sess, oauthRepo: oauthRepo, setting: cfg, mailer: mailer, db: db}
}

func (h *testHarness) do(method, path string, body any, cookies map[string]string, headers map[string]string) *httptest.ResponseRecorder {
	var rd *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewReader(b)
	} else {
		rd = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, rd)
	req.Header.Set("Content-Type", "application/json")
	for k, v := range cookies {
		req.AddCookie(&http.Cookie{Name: k, Value: v})
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	h.r.ServeHTTP(w, req)
	return w
}

func extractCookie(w *httptest.ResponseRecorder, name string) string {
	for _, c := range w.Result().Cookies() {
		if c.Name == name {
			return c.Value
		}
	}
	return ""
}

// --- Register / Login ---

func TestHandler_Register_OK(t *testing.T) {
	h := setupHarness(t)
	w := h.do("POST", "/api/auth/register", map[string]any{
		"username": "alice", "email": "a@b.com", "password": "P@ssw0rd!",
	}, nil, nil)
	if w.Code != 200 {
		t.Fatalf("got %d body=%s", w.Code, w.Body.String())
	}
	sid := extractCookie(w, "proapi_session")
	if sid == "" {
		t.Fatal("want session cookie")
	}
	csrf := extractCookie(w, "proapi_csrf")
	if csrf == "" {
		t.Fatal("want csrf cookie")
	}
}

func TestHandler_Register_Disabled(t *testing.T) {
	h := setupHarness(t)
	h.setting.kv["auth.allow_register"] = false
	w := h.do("POST", "/api/auth/register", map[string]any{
		"username": "alice", "email": "a@b.com", "password": "P@ssw0rd!",
	}, nil, nil)
	if w.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d", w.Code)
	}
}

func TestHandler_Register_NeedsVerifyNoCookie(t *testing.T) {
	h := setupHarness(t)
	h.setting.kv["auth.email_verification_required"] = true
	w := h.do("POST", "/api/auth/register", map[string]any{
		"username": "alice", "email": "a@b.com", "password": "P@ssw0rd!",
	}, nil, nil)
	if w.Code != 200 {
		t.Fatalf("want 200, got %d body=%s", w.Code, w.Body.String())
	}
	if extractCookie(w, "proapi_session") != "" {
		t.Fatal("should not set session cookie")
	}
	if len(h.mailer.msgs) != 1 {
		t.Fatal("want verify mail")
	}
}

func TestHandler_Login_OK(t *testing.T) {
	h := setupHarness(t)
	_ = h.do("POST", "/api/auth/register", map[string]any{"username": "alice", "email": "a@b.com", "password": "P@ssw0rd!"}, nil, nil)
	w := h.do("POST", "/api/auth/login", map[string]any{"identity": "alice", "password": "P@ssw0rd!"}, nil, nil)
	if w.Code != 200 {
		t.Fatalf("got %d", w.Code)
	}
	if extractCookie(w, "proapi_session") == "" {
		t.Fatal("want session cookie")
	}
}

func TestHandler_Login_WrongPassword(t *testing.T) {
	h := setupHarness(t)
	_ = h.do("POST", "/api/auth/register", map[string]any{"username": "alice", "email": "a@b.com", "password": "P@ssw0rd!"}, nil, nil)
	w := h.do("POST", "/api/auth/login", map[string]any{"identity": "alice", "password": "WRONG!!!"}, nil, nil)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func TestHandler_Logout_ClearsCookie(t *testing.T) {
	h := setupHarness(t)
	w := h.do("POST", "/api/auth/register", map[string]any{"username": "alice", "email": "a@b.com", "password": "P@ssw0rd!"}, nil, nil)
	sid := extractCookie(w, "proapi_session")
	w2 := h.do("POST", "/api/auth/logout", map[string]any{}, map[string]string{"proapi_session": sid}, nil)
	if w2.Code != 200 {
		t.Fatalf("got %d", w2.Code)
	}
	// session 应已撤销
	gotSess, _ := h.session.Get(context.Background(), sid)
	if gotSess != nil {
		t.Fatal("session not revoked")
	}
}

// --- Email send_code + email login ---

func TestHandler_EmailSendCode(t *testing.T) {
	h := setupHarness(t)
	w := h.do("POST", "/api/auth/email/send_code", map[string]any{"email": "a@b.com", "purpose": "login"}, nil, nil)
	if w.Code != 200 {
		t.Fatalf("got %d body=%s", w.Code, w.Body.String())
	}
}

func TestHandler_EmailLogin_OK(t *testing.T) {
	h := setupHarness(t)
	h.setting.kv["auth.email_verification_required"] = true
	_ = h.do("POST", "/api/auth/register", map[string]any{"username": "alice", "email": "a@b.com", "password": "P@ssw0rd!"}, nil, nil)
	// 取出验证码
	body := h.mailer.msgs[0].Body
	code := extractDigits(body, 6)
	w := h.do("POST", "/api/auth/email/login", map[string]any{"email": "a@b.com", "code": code}, nil, nil)
	if w.Code != 200 {
		t.Fatalf("got %d body=%s", w.Code, w.Body.String())
	}
}

// --- /api/user/profile (SessionAuth + CSRF) ---

func TestHandler_UserProfile_NotLoggedIn(t *testing.T) {
	h := setupHarness(t)
	w := h.do("GET", "/api/user/profile", nil, nil, nil)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func TestHandler_UserProfile_OK(t *testing.T) {
	h := setupHarness(t)
	w := h.do("POST", "/api/auth/register", map[string]any{"username": "alice", "email": "a@b.com", "password": "P@ssw0rd!"}, nil, nil)
	sid := extractCookie(w, "proapi_session")
	w2 := h.do("GET", "/api/user/profile", nil, map[string]string{"proapi_session": sid}, nil)
	if w2.Code != 200 {
		t.Fatalf("got %d body=%s", w2.Code, w2.Body.String())
	}
}

func TestHandler_UserProfile_PatchRequiresCSRF(t *testing.T) {
	h := setupHarness(t)
	w := h.do("POST", "/api/auth/register", map[string]any{"username": "alice", "email": "a@b.com", "password": "P@ssw0rd!"}, nil, nil)
	sid := extractCookie(w, "proapi_session")
	csrf := extractCookie(w, "proapi_csrf")
	// 无 CSRF header → 403
	w2 := h.do("PATCH", "/api/user/profile", map[string]any{"display_name": "X"},
		map[string]string{"proapi_session": sid, "proapi_csrf": csrf}, nil)
	if w2.Code != http.StatusForbidden {
		t.Fatalf("want 403 no csrf, got %d", w2.Code)
	}
	// 带 CSRF header → 200
	w3 := h.do("PATCH", "/api/user/profile", map[string]any{"display_name": "X"},
		map[string]string{"proapi_session": sid, "proapi_csrf": csrf},
		map[string]string{"X-CSRF-Token": csrf})
	if w3.Code != 200 {
		t.Fatalf("got %d body=%s", w3.Code, w3.Body.String())
	}
}

// --- Admin ---

func TestHandler_AdminLogin_NonAdmin_Forbidden(t *testing.T) {
	h := setupHarness(t)
	_ = h.do("POST", "/api/auth/register", map[string]any{"username": "alice", "email": "a@b.com", "password": "P@ssw0rd!"}, nil, nil)
	w := h.do("POST", "/api/admin/auth/login", map[string]any{"identity": "alice", "password": "P@ssw0rd!"}, nil, nil)
	if w.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d", w.Code)
	}
}

func TestHandler_AdminLogin_AsAdmin_OK(t *testing.T) {
	h := setupHarness(t)
	// 先注册再后台直接改 role=tenant_admin
	w := h.do("POST", "/api/auth/register", map[string]any{"username": "alice", "email": "a@b.com", "password": "P@ssw0rd!"}, nil, nil)
	if w.Code != 200 {
		t.Fatal("register failed")
	}
	_, _ = h.user.Update(context.Background(), 1, user.UpdateInput{Role: int8Ptr(user.RoleTenantAdmin)})

	w2 := h.do("POST", "/api/admin/auth/login", map[string]any{"identity": "alice", "password": "P@ssw0rd!"}, nil, nil)
	if w2.Code != 200 {
		t.Fatalf("got %d body=%s", w2.Code, w2.Body.String())
	}
}

func TestHandler_AdminUsersList(t *testing.T) {
	h := setupHarness(t)
	_ = h.do("POST", "/api/auth/register", map[string]any{"username": "alice", "email": "a@b.com", "password": "P@ssw0rd!"}, nil, nil)
	_, _ = h.user.Update(context.Background(), 1, user.UpdateInput{Role: int8Ptr(user.RoleTenantAdmin)})
	wLog := h.do("POST", "/api/admin/auth/login", map[string]any{"identity": "alice", "password": "P@ssw0rd!"}, nil, nil)
	sid := extractCookie(wLog, "proapi_session")
	csrf := extractCookie(wLog, "proapi_csrf")
	w := h.do("GET", "/api/admin/users?page=1&size=10", nil,
		map[string]string{"proapi_session": sid, "proapi_csrf": csrf}, nil)
	if w.Code != 200 {
		t.Fatalf("got %d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "alice") {
		t.Fatalf("body: %s", w.Body.String())
	}
}

func TestHandler_AdminUserPatch_DisableTriggersRevokeAll(t *testing.T) {
	h := setupHarness(t)
	// admin
	_ = h.do("POST", "/api/auth/register", map[string]any{"username": "root", "email": "r@b.com", "password": "P@ssw0rd!"}, nil, nil)
	_, _ = h.user.Update(context.Background(), 1, user.UpdateInput{Role: int8Ptr(user.RoleSuperAdmin)})
	// 普通用户
	_ = h.do("POST", "/api/auth/register", map[string]any{"username": "alice", "email": "a@b.com", "password": "P@ssw0rd!"}, nil, nil)

	wAdmin := h.do("POST", "/api/admin/auth/login", map[string]any{"identity": "root", "password": "P@ssw0rd!"}, nil, nil)
	sidA := extractCookie(wAdmin, "proapi_session")
	csrfA := extractCookie(wAdmin, "proapi_csrf")

	wU := h.do("POST", "/api/auth/login", map[string]any{"identity": "alice", "password": "P@ssw0rd!"}, nil, nil)
	sidU := extractCookie(wU, "proapi_session")

	// admin 禁用 alice
	disabled := user.StatusDisabled
	w := h.do("PATCH", "/api/admin/users/2", map[string]any{"status": disabled},
		map[string]string{"proapi_session": sidA, "proapi_csrf": csrfA},
		map[string]string{"X-CSRF-Token": csrfA})
	if w.Code != 200 {
		t.Fatalf("got %d body=%s", w.Code, w.Body.String())
	}
	// alice 的 session 应被强制下线
	sess, _ := h.session.Get(context.Background(), sidU)
	if sess != nil {
		t.Fatal("user session should be revoked")
	}
}

func TestHandler_AdminUserDelete_Self_Forbidden(t *testing.T) {
	h := setupHarness(t)
	_ = h.do("POST", "/api/auth/register", map[string]any{"username": "root", "email": "r@b.com", "password": "P@ssw0rd!"}, nil, nil)
	_, _ = h.user.Update(context.Background(), 1, user.UpdateInput{Role: int8Ptr(user.RoleSuperAdmin)})
	wAdmin := h.do("POST", "/api/admin/auth/login", map[string]any{"identity": "root", "password": "P@ssw0rd!"}, nil, nil)
	sid := extractCookie(wAdmin, "proapi_session")
	csrf := extractCookie(wAdmin, "proapi_csrf")
	w := h.do("DELETE", "/api/admin/users/1", nil,
		map[string]string{"proapi_session": sid, "proapi_csrf": csrf},
		map[string]string{"X-CSRF-Token": csrf})
	if w.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestHandler_AdminResetPassword_AutoGen(t *testing.T) {
	h := setupHarness(t)
	_ = h.do("POST", "/api/auth/register", map[string]any{"username": "root", "email": "r@b.com", "password": "P@ssw0rd!"}, nil, nil)
	_, _ = h.user.Update(context.Background(), 1, user.UpdateInput{Role: int8Ptr(user.RoleSuperAdmin)})
	_ = h.do("POST", "/api/auth/register", map[string]any{"username": "alice", "email": "a@b.com", "password": "P@ssw0rd!"}, nil, nil)

	wAdmin := h.do("POST", "/api/admin/auth/login", map[string]any{"identity": "root", "password": "P@ssw0rd!"}, nil, nil)
	sid := extractCookie(wAdmin, "proapi_session")
	csrf := extractCookie(wAdmin, "proapi_csrf")
	w := h.do("POST", "/api/admin/users/2/reset_password", map[string]any{},
		map[string]string{"proapi_session": sid, "proapi_csrf": csrf},
		map[string]string{"X-CSRF-Token": csrf})
	if w.Code != 200 {
		t.Fatalf("got %d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "temp_password") {
		t.Fatalf("expected temp_password: %s", w.Body.String())
	}
}

// --- helpers ---

func int8Ptr(v int8) *int8 { return &v }

func extractDigits(s string, n int) string {
	for i := 0; i+n <= len(s); i++ {
		seg := s[i : i+n]
		ok := true
		for _, c := range seg {
			if c < '0' || c > '9' {
				ok = false
				break
			}
		}
		if ok {
			return seg
		}
	}
	return ""
}
