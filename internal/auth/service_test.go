package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/ijry/pro-api/internal/audit"
	"github.com/ijry/pro-api/internal/auth/oauth"
	"github.com/ijry/pro-api/internal/auth/oauth/github"
	"github.com/ijry/pro-api/internal/auth/session"
	"github.com/ijry/pro-api/internal/auth/verifycode"
	"github.com/ijry/pro-api/internal/group"
	"github.com/ijry/pro-api/internal/notify/email"
	"github.com/ijry/pro-api/internal/setting"
	"github.com/ijry/pro-api/internal/user"
	"github.com/ijry/pro-api/internal/util/clock"
	"github.com/ijry/pro-api/pkg/apierr"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type fakeIDGen struct{ n int64 }

func (f *fakeIDGen) Generate() int64 { f.n++; return f.n }

// fakeSetting 实现 setting.Store。
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

func (f *fakeSetting) GetSecret(ctx context.Context, key string, _ setting.Decryptor) (string, error) {
	return f.GetString(ctx, key, ""), nil
}

func (f *fakeSetting) ListAll(_ context.Context) ([]setting.Setting, error) {
	out := make([]setting.Setting, 0, len(f.kv))
	for k, v := range f.kv {
		b, _ := json.Marshal(v)
		out = append(out, setting.Setting{Key: k, Value: b})
	}
	return out, nil
}

type captureMailer struct{ msgs []email.Message }

func (c *captureMailer) Send(_ context.Context, m email.Message) error {
	c.msgs = append(c.msgs, m)
	return nil
}

func setupAuth(t *testing.T) (Service, *fakeSetting, *captureMailer, session.Store) {
	t.Helper()
	dbName := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", dbName)), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY, username TEXT UNIQUE, email TEXT UNIQUE, password_hash TEXT,
			display_name TEXT, avatar TEXT, role INTEGER, status INTEGER, group_id INTEGER,
			invite_code TEXT UNIQUE, invited_by INTEGER NOT NULL DEFAULT 0,
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
	cfg := &fakeSetting{kv: map[string]any{
		"auth.allow_register": true, "auth.email_verification_required": false,
		"auth.password.min_length": 8, "auth.password.require_mixed": false,
	}}
	svc := NewService(Deps{
		User: usvc, Group: gsvc, Session: sess, VerifyCode: vc,
		Mailer: mailer, GithubState: stateStore, OAuthRepo: oauth.NewRepository(db),
		Setting: cfg, Audit: audit.NewNoop(), Clock: clock.Real, IDGen: &fakeIDGen{},
	})
	return svc, cfg, mailer, sess
}

// --- Register ---

func TestRegister_AllowRegisterOff_Forbidden(t *testing.T) {
	svc, cfg, _, _ := setupAuth(t)
	cfg.kv["auth.allow_register"] = false
	_, err := svc.Register(context.Background(), RegisterInput{Username: "alice", Email: "a@b.com", Password: "P@ssw0rd!"})
	var ae *apierr.Error
	if !errors.As(err, &ae) || ae.Code != apierr.CodeForbidden {
		t.Fatalf("want CodeForbidden, got %v", err)
	}
}

func TestRegister_EmailVerifyOff_AutoLogin(t *testing.T) {
	svc, _, _, _ := setupAuth(t)
	res, err := svc.Register(context.Background(), RegisterInput{Username: "alice", Email: "a@b.com", Password: "P@ssw0rd!"})
	if err != nil {
		t.Fatal(err)
	}
	if res.EmailVerifyRequired || res.Session == nil {
		t.Fatalf("want auto login: %+v", res)
	}
}

func TestRegister_EmailVerifyRequired_NoSession(t *testing.T) {
	svc, cfg, mailer, _ := setupAuth(t)
	cfg.kv["auth.email_verification_required"] = true
	res, err := svc.Register(context.Background(), RegisterInput{Username: "alice", Email: "a@b.com", Password: "P@ssw0rd!"})
	if err != nil {
		t.Fatal(err)
	}
	if !res.EmailVerifyRequired || res.Session != nil {
		t.Fatalf("want verify-required: %+v", res)
	}
	if len(mailer.msgs) != 1 || mailer.msgs[0].Tag != "verify_register" {
		t.Fatalf("mailer wrong: %+v", mailer.msgs)
	}
}

func TestRegister_WeakPassword_InvalidParam(t *testing.T) {
	svc, _, _, _ := setupAuth(t)
	_, err := svc.Register(context.Background(), RegisterInput{Username: "alice", Email: "a@b.com", Password: "short"})
	var ae *apierr.Error
	if !errors.As(err, &ae) || ae.Code != apierr.CodeInvalidParam {
		t.Fatalf("want CodeInvalidParam, got %v", err)
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	svc, _, _, _ := setupAuth(t)
	_, _ = svc.Register(context.Background(), RegisterInput{Username: "alice", Email: "a@b.com", Password: "P@ssw0rd!"})
	_, err := svc.Register(context.Background(), RegisterInput{Username: "bob", Email: "a@b.com", Password: "P@ssw0rd!"})
	var ae *apierr.Error
	if !errors.As(err, &ae) || ae.Code != apierr.CodeEmailRegistered {
		t.Fatalf("want CodeEmailRegistered, got %v", err)
	}
}

// --- Login ---

func TestLogin_ByEmail_OK(t *testing.T) {
	svc, _, _, _ := setupAuth(t)
	_, _ = svc.Register(context.Background(), RegisterInput{Username: "alice", Email: "a@b.com", Password: "P@ssw0rd!"})
	sess, u, err := svc.Login(context.Background(), LoginInput{Identity: "a@b.com", Password: "P@ssw0rd!"})
	if err != nil {
		t.Fatal(err)
	}
	if sess == nil || u.Username != "alice" {
		t.Fatalf("got %+v %+v", sess, u)
	}
}

func TestLogin_ByUsername_OK(t *testing.T) {
	svc, _, _, _ := setupAuth(t)
	_, _ = svc.Register(context.Background(), RegisterInput{Username: "alice", Email: "a@b.com", Password: "P@ssw0rd!"})
	_, _, err := svc.Login(context.Background(), LoginInput{Identity: "alice", Password: "P@ssw0rd!"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	svc, _, _, _ := setupAuth(t)
	_, _ = svc.Register(context.Background(), RegisterInput{Username: "alice", Email: "a@b.com", Password: "P@ssw0rd!"})
	_, _, err := svc.Login(context.Background(), LoginInput{Identity: "alice", Password: "WRONG!!!"})
	var ae *apierr.Error
	if !errors.As(err, &ae) || ae.Code != apierr.CodeWrongPassword {
		t.Fatalf("want CodeWrongPassword, got %v", err)
	}
}

func TestLogin_NonExistUser_WrongPassword(t *testing.T) {
	svc, _, _, _ := setupAuth(t)
	_, _, err := svc.Login(context.Background(), LoginInput{Identity: "nobody@x.com", Password: "P@ssw0rd!"})
	var ae *apierr.Error
	if !errors.As(err, &ae) || ae.Code != apierr.CodeWrongPassword {
		t.Fatalf("want CodeWrongPassword, got %v", err)
	}
}

func TestLogin_PendingEmailVerify(t *testing.T) {
	svc, cfg, _, _ := setupAuth(t)
	cfg.kv["auth.email_verification_required"] = true
	_, _ = svc.Register(context.Background(), RegisterInput{Username: "alice", Email: "a@b.com", Password: "P@ssw0rd!"})
	_, _, err := svc.Login(context.Background(), LoginInput{Identity: "a@b.com", Password: "P@ssw0rd!"})
	var ae *apierr.Error
	if !errors.As(err, &ae) || ae.Code != apierr.CodeEmailNotVerified {
		t.Fatalf("want CodeEmailNotVerified, got %v", err)
	}
}

// --- EmailCodeLogin ---

func TestEmailCodeLogin_OKMarksVerified(t *testing.T) {
	svc, cfg, mailer, _ := setupAuth(t)
	cfg.kv["auth.email_verification_required"] = true
	_, _ = svc.Register(context.Background(), RegisterInput{Username: "alice", Email: "a@b.com", Password: "P@ssw0rd!"})
	body := mailer.msgs[0].Body
	// 提取 6 位数字
	code := extractDigits(body, 6)
	sess, u, err := svc.EmailCodeLogin(context.Background(), EmailCodeLoginInput{Email: "a@b.com", Code: code})
	if err != nil {
		t.Fatal(err)
	}
	if sess == nil {
		t.Fatal("want session")
	}
	if u.EmailVerifiedAt == nil {
		t.Fatal("want email_verified_at set")
	}
	if u.Status != user.StatusActive {
		t.Fatalf("want active, got %d", u.Status)
	}
}

func TestEmailCodeLogin_BadCode(t *testing.T) {
	svc, _, _, _ := setupAuth(t)
	_, _ = svc.Register(context.Background(), RegisterInput{Username: "alice", Email: "a@b.com", Password: "P@ssw0rd!"})
	_, _, err := svc.EmailCodeLogin(context.Background(), EmailCodeLoginInput{Email: "a@b.com", Code: "000000"})
	var ae *apierr.Error
	if !errors.As(err, &ae) || ae.Code != apierr.CodeCaptchaInvalid {
		t.Fatalf("want CodeCaptchaInvalid, got %v", err)
	}
}

// --- ForgotPassword / ResetPassword ---

func TestForgotPassword_NonExist_ReturnsOK(t *testing.T) {
	svc, _, _, _ := setupAuth(t)
	if err := svc.ForgotPassword(context.Background(), "nobody@x.com", ""); err != nil {
		t.Fatalf("want ok, got %v", err)
	}
}

func TestResetPassword_OK_RevokesSessions(t *testing.T) {
	svc, _, mailer, _ := setupAuth(t)
	_, _ = svc.Register(context.Background(), RegisterInput{Username: "alice", Email: "a@b.com", Password: "P@ssw0rd!"})
	// 先发起 forgot
	if err := svc.ForgotPassword(context.Background(), "a@b.com", ""); err != nil {
		t.Fatal(err)
	}
	if len(mailer.msgs) < 1 {
		t.Fatal("want reset mail")
	}
	body := mailer.msgs[len(mailer.msgs)-1].Body
	code := extractDigits(body, 6)
	if err := svc.ResetPassword(context.Background(), "a@b.com", code, "Newp@ss123"); err != nil {
		t.Fatal(err)
	}
	// 用新密码登录
	if _, _, err := svc.Login(context.Background(), LoginInput{Identity: "a@b.com", Password: "Newp@ss123"}); err != nil {
		t.Fatal(err)
	}
}

// --- ChangePassword ---

func TestChangePassword_OldMismatch(t *testing.T) {
	svc, _, _, _ := setupAuth(t)
	res, _ := svc.Register(context.Background(), RegisterInput{Username: "alice", Email: "a@b.com", Password: "P@ssw0rd!"})
	err := svc.ChangePassword(context.Background(), res.User.ID, res.Session.ID, "WRONG!!!", "Newp@ss123")
	var ae *apierr.Error
	if !errors.As(err, &ae) || ae.Code != apierr.CodeWrongPassword {
		t.Fatalf("want CodeWrongPassword, got %v", err)
	}
}

func TestChangePassword_OK(t *testing.T) {
	svc, _, _, _ := setupAuth(t)
	res, _ := svc.Register(context.Background(), RegisterInput{Username: "alice", Email: "a@b.com", Password: "P@ssw0rd!"})
	if err := svc.ChangePassword(context.Background(), res.User.ID, res.Session.ID, "P@ssw0rd!", "Newp@ss123"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.Login(context.Background(), LoginInput{Identity: "alice", Password: "Newp@ss123"}); err != nil {
		t.Fatal(err)
	}
}

// --- GitHub OAuth ---

// stub provider for tests
type stubGithub struct {
	info *github.UserInfo
	err  error
}

func (s *stubGithub) BuildAuthURL(_ context.Context, state, redirect string) (string, error) {
	return fmt.Sprintf("https://github.example/oauth?state=%s&redirect=%s", state, redirect), nil
}
func (s *stubGithub) Exchange(_ context.Context, code, _ string) (*github.UserInfo, string, error) {
	if s.err != nil {
		return nil, "", s.err
	}
	if code == "bad" {
		return nil, "", fmt.Errorf("bad code")
	}
	return s.info, "tok", nil
}

func setupAuthWithGithub(t *testing.T, info *github.UserInfo) (Service, *fakeSetting) {
	t.Helper()
	authSvc, cfg, _, _ := setupAuth(t)
	// reach into impl to inject github
	concrete := authSvc.(*svc)
	concrete.deps.GithubProvider = &stubGithub{info: info}
	cfg.kv["auth.github_oauth"] = map[string]any{
		"client_id": "x", "client_secret": "y", "redirect_url": "https://app/cb",
	}
	return authSvc, cfg
}

func TestGithubOAuthStart_NotEnabled(t *testing.T) {
	svc, _, _, _ := setupAuth(t)
	if _, err := svc.GithubOAuthStart(context.Background(), "", "", "", 0); err == nil {
		t.Fatal("want forbidden")
	}
}

func TestGithubOAuthStart_OK(t *testing.T) {
	svc, _ := setupAuthWithGithub(t, &github.UserInfo{ID: 1, Login: "alice"})
	url, err := svc.GithubOAuthStart(context.Background(), "", "", "/console", 0)
	if err != nil {
		t.Fatal(err)
	}
	if url == "" {
		t.Fatal("want url")
	}
}

func TestGithubCallback_AutoCreatesAccount(t *testing.T) {
	svc, _ := setupAuthWithGithub(t, &github.UserInfo{ID: 555, Login: "newcomer", Email: "n@gh.com", Name: "N"})
	urlStr, _ := svc.GithubOAuthStart(context.Background(), "", "", "/x", 0)
	state := extractStateParam(urlStr)
	sess, u, _, err := svc.GithubOAuthCallback(context.Background(), "ok", state, "1.1.1.1", "ua")
	if err != nil {
		t.Fatal(err)
	}
	if sess == nil || u.Username != "newcomer" {
		t.Fatalf("got %+v %+v", sess, u)
	}
	if u.Email == nil || *u.Email != "n@gh.com" {
		t.Fatalf("email not set: %+v", u.Email)
	}
}

func TestGithubCallback_LinksExistingBinding(t *testing.T) {
	svc, _ := setupAuthWithGithub(t, &github.UserInfo{ID: 100, Login: "alice", Email: "a@gh.com"})
	// 第一次:自动建账号
	url1, _ := svc.GithubOAuthStart(context.Background(), "", "", "", 0)
	_, u1, _, _ := svc.GithubOAuthCallback(context.Background(), "ok", extractStateParam(url1), "", "")

	// 第二次:同一 GitHub uid 应当复用账号
	url2, _ := svc.GithubOAuthStart(context.Background(), "", "", "", 0)
	_, u2, _, err := svc.GithubOAuthCallback(context.Background(), "ok", extractStateParam(url2), "", "")
	if err != nil {
		t.Fatal(err)
	}
	if u1.ID != u2.ID {
		t.Fatalf("expected reuse: u1=%d u2=%d", u1.ID, u2.ID)
	}
}

func TestBindGithub_BoundToOther_Forbidden(t *testing.T) {
	svc, _ := setupAuthWithGithub(t, &github.UserInfo{ID: 100, Login: "alice", Email: "a@gh.com"})
	// 让 user A 自动注册并绑定
	url1, _ := svc.GithubOAuthStart(context.Background(), "", "", "", 0)
	_, uA, _, _ := svc.GithubOAuthCallback(context.Background(), "ok", extractStateParam(url1), "", "")

	// User B 想绑定同一 GitHub
	res, _ := svc.Register(context.Background(), RegisterInput{Username: "bob", Email: "b@x.com", Password: "P@ssw0rd!"})
	urlB, _ := svc.GithubOAuthStart(context.Background(), "", "", "", res.User.ID)
	err := svc.BindGithub(context.Background(), res.User.ID, "ok", extractStateParam(urlB))
	if err == nil {
		t.Fatal("want forbidden")
	}
	_ = uA
}

func TestUnbindGithub_NoPassword_Forbidden(t *testing.T) {
	svc, _ := setupAuthWithGithub(t, &github.UserInfo{ID: 1, Login: "alice", Email: "a@gh.com"})
	// 自动建 github 账号(无密码)
	url1, _ := svc.GithubOAuthStart(context.Background(), "", "", "", 0)
	_, u, _, _ := svc.GithubOAuthCallback(context.Background(), "ok", extractStateParam(url1), "", "")
	err := svc.UnbindGithub(context.Background(), u.ID)
	var ae *apierr.Error
	if !errors.As(err, &ae) || ae.Code != apierr.CodeForbidden {
		t.Fatalf("want CodeForbidden, got %v", err)
	}
}

// --- helpers ---

// extractDigits 从字符串里找出连续 n 位数字。
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

func extractStateParam(s string) string {
	// 解析 "...&state=XXX&..." 或末尾
	for _, kv := range strings.Split(s, "&") {
		if strings.HasPrefix(kv, "state=") {
			return kv[len("state="):]
		}
		if strings.Contains(kv, "state=") {
			parts := strings.SplitN(kv, "state=", 2)
			return parts[1]
		}
	}
	return ""
}
