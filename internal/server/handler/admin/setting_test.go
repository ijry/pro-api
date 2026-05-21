package admin

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ijry/pro-api/internal/audit"
	"github.com/ijry/pro-api/internal/server/middleware"
	"github.com/ijry/pro-api/internal/setting"
)

// --- mocks ---

type mockStore struct {
	mu      sync.Mutex
	data    map[string]setting.Setting
}

func newMockStore() *mockStore {
	return &mockStore{data: map[string]setting.Setting{}}
}

func (m *mockStore) Get(ctx context.Context, key string) (json.RawMessage, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	row, ok := m.data[key]
	if !ok {
		return nil, false
	}
	return row.Value, true
}
func (m *mockStore) GetString(ctx context.Context, key string, def string) string { return def }
func (m *mockStore) GetBool(ctx context.Context, key string, def bool) bool       { return def }
func (m *mockStore) GetInt(ctx context.Context, key string, def int) int          { return def }
func (m *mockStore) GetFloat(ctx context.Context, key string, def float64) float64 {
	return def
}
func (m *mockStore) GetJSON(ctx context.Context, key string, dest any) error { return nil }
func (m *mockStore) Put(ctx context.Context, key string, val any, actor int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	var raw json.RawMessage
	switch v := val.(type) {
	case json.RawMessage:
		raw = v
	case []byte:
		raw = v
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return err
		}
		raw = b
	}
	row := m.data[key]
	row.Key = key
	row.Value = raw
	row.UpdatedAt = time.Now().UTC()
	m.data[key] = row
	return nil
}
func (m *mockStore) GetSecret(ctx context.Context, key string, dec setting.Decryptor) (string, error) {
	return "", setting.ErrNotEncrypted
}
func (m *mockStore) ListAll(ctx context.Context) ([]setting.Setting, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]setting.Setting, 0, len(m.data))
	for _, v := range m.data {
		out = append(out, v)
	}
	return out, nil
}
func (m *mockStore) Close() error { return nil }

type mockEncryptor struct{}

func (mockEncryptor) Encrypt(plain string) (string, error) {
	return "ENC(v1,n," + plain + ")", nil
}

type errEncryptor struct{}

func (errEncryptor) Encrypt(plain string) (string, error) {
	return "", errors.New("encrypt failed")
}

type mockMailer struct {
	called bool
	err    error
}

func (m *mockMailer) SendTestMail(cfg SMTPConfig, to string) error {
	m.called = true
	return m.err
}

type captureAudit struct {
	mu      sync.Mutex
	entries []audit.Entry
}

func (a *captureAudit) Log(ctx context.Context, e audit.Entry) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.entries = append(a.entries, e)
	return nil
}

// --- setup ---

func newSettingRouter(opts ...func(*SettingHandler)) (*gin.Engine, *mockStore, *captureAudit) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorResponse("json"))
	store := newMockStore()
	aud := &captureAudit{}
	h := NewSettingHandler(store, mockEncryptor{}, nil, aud, func(*gin.Context) int64 { return 999 })
	for _, o := range opts {
		o(h)
	}
	h.Register(r)
	return r, store, aud
}

// --- tests ---

func TestAdminSetting_List_GroupedByPrefix(t *testing.T) {
	r, store, _ := newSettingRouter()
	now := time.Now().UTC()
	store.data["auth.allow_register"] = setting.Setting{Key: "auth.allow_register", Value: json.RawMessage(`true`), UpdatedAt: now}
	store.data["session.cookie_name"] = setting.Setting{Key: "session.cookie_name", Value: json.RawMessage(`"sid"`), UpdatedAt: now}

	rec := doReq(t, r, http.MethodGet, "/settings", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var body struct {
		Data struct {
			Groups []struct {
				Name  string `json:"name"`
				Items []struct {
					Key string `json:"key"`
				} `json:"items"`
			} `json:"groups"`
		} `json:"data"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	hasAuth, hasSession := false, false
	for _, g := range body.Data.Groups {
		if g.Name == "auth" && len(g.Items) == 1 && g.Items[0].Key == "auth.allow_register" {
			hasAuth = true
		}
		if g.Name == "session" && len(g.Items) == 1 && g.Items[0].Key == "session.cookie_name" {
			hasSession = true
		}
	}
	if !hasAuth || !hasSession {
		t.Fatalf("missing groups: %+v", body.Data.Groups)
	}
}

func TestAdminSetting_List_SensitiveValuesMasked(t *testing.T) {
	r, store, _ := newSettingRouter()
	store.data["auth.smtp.password"] = setting.Setting{Key: "auth.smtp.password", Value: json.RawMessage(`"ENC(v1,n,c)"`)}

	rec := doReq(t, r, http.MethodGet, "/settings", "")
	var body struct {
		Data struct {
			Groups []struct {
				Items []struct {
					Key         string `json:"key"`
					Value       any    `json:"value"`
					IsSensitive bool   `json:"is_sensitive"`
				} `json:"items"`
			} `json:"groups"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, g := range body.Data.Groups {
		for _, it := range g.Items {
			if it.Key == "auth.smtp.password" {
				if it.Value != "<encrypted>" {
					t.Fatalf("want masked, got %v", it.Value)
				}
				if !it.IsSensitive {
					t.Fatal("want is_sensitive=true")
				}
				found = true
			}
		}
	}
	if !found {
		t.Fatal("auth.smtp.password not in any group")
	}
}

func TestAdminSetting_Get_KnownKey(t *testing.T) {
	r, store, _ := newSettingRouter()
	store.data["auth.allow_register"] = setting.Setting{Key: "auth.allow_register", Value: json.RawMessage(`true`)}
	rec := doReq(t, r, http.MethodGet, "/settings/auth.allow_register", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
}

func TestAdminSetting_Get_UnknownKey_404(t *testing.T) {
	r, _, _ := newSettingRouter()
	rec := doReq(t, r, http.MethodGet, "/settings/missing.key", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", rec.Code)
	}
}

func TestAdminSetting_Get_SensitiveValueMasked(t *testing.T) {
	r, store, _ := newSettingRouter()
	store.data["auth.smtp.password"] = setting.Setting{Key: "auth.smtp.password", Value: json.RawMessage(`"ENC(v1,n,c)"`)}
	rec := doReq(t, r, http.MethodGet, "/settings/auth.smtp.password", "")
	var body struct {
		Data struct {
			Value any `json:"value"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Data.Value != "<encrypted>" {
		t.Fatalf("want masked, got %v", body.Data.Value)
	}
}

func TestAdminSetting_Patch_PlainValue(t *testing.T) {
	r, store, _ := newSettingRouter()
	store.data["auth.allow_register"] = setting.Setting{Key: "auth.allow_register", Value: json.RawMessage(`true`)}
	rec := doReq(t, r, http.MethodPatch, "/settings/auth.allow_register", `{"value":false}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if string(store.data["auth.allow_register"].Value) != `false` {
		t.Fatalf("want false, got %s", store.data["auth.allow_register"].Value)
	}
}

func TestAdminSetting_Patch_SensitiveValue_Encrypts(t *testing.T) {
	r, store, _ := newSettingRouter()
	store.data["auth.smtp.password"] = setting.Setting{Key: "auth.smtp.password", Value: json.RawMessage(`"old"`)}
	rec := doReq(t, r, http.MethodPatch, "/settings/auth.smtp.password", `{"value":"hunter2"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	stored := string(store.data["auth.smtp.password"].Value)
	if !strings.Contains(stored, "ENC(v1,n,hunter2)") {
		t.Fatalf("expected encrypted, got %s", stored)
	}
}

func TestAdminSetting_Patch_PlaceholderValue_NoOp(t *testing.T) {
	r, store, _ := newSettingRouter()
	store.data["auth.smtp.password"] = setting.Setting{Key: "auth.smtp.password", Value: json.RawMessage(`"ENC(v1,old,old)"`)}
	rec := doReq(t, r, http.MethodPatch, "/settings/auth.smtp.password", `{"value":"<encrypted>"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	// 不变
	if string(store.data["auth.smtp.password"].Value) != `"ENC(v1,old,old)"` {
		t.Fatalf("should not change, got %s", store.data["auth.smtp.password"].Value)
	}
}

func TestAdminSetting_Patch_SensitiveNonString_400(t *testing.T) {
	r, store, _ := newSettingRouter()
	store.data["auth.smtp.password"] = setting.Setting{Key: "auth.smtp.password", Value: json.RawMessage(`"old"`)}
	rec := doReq(t, r, http.MethodPatch, "/settings/auth.smtp.password", `{"value":123}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

func TestAdminSetting_Patch_UnknownKey_404(t *testing.T) {
	r, _, _ := newSettingRouter()
	rec := doReq(t, r, http.MethodPatch, "/settings/missing.key", `{"value":1}`)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", rec.Code)
	}
}

func TestAdminSetting_Patch_MissingValue_400(t *testing.T) {
	r, store, _ := newSettingRouter()
	store.data["k"] = setting.Setting{Key: "k", Value: json.RawMessage(`1`)}
	rec := doReq(t, r, http.MethodPatch, "/settings/k", `{}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

func TestAdminSetting_Patch_AuditMasksSensitiveDiff(t *testing.T) {
	r, store, aud := newSettingRouter()
	store.data["auth.smtp.password"] = setting.Setting{Key: "auth.smtp.password", Value: json.RawMessage(`"old"`)}
	_ = doReq(t, r, http.MethodPatch, "/settings/auth.smtp.password", `{"value":"hunter2"}`)
	if len(aud.entries) != 1 {
		t.Fatalf("want 1 audit, got %d", len(aud.entries))
	}
	e := aud.entries[0]
	// before / after 都应该是脱敏占位 (JSON encoding 会把 < 转 <)
	if !strings.Contains(string(e.Before), `encrypted`) || !strings.Contains(string(e.After), `encrypted`) {
		t.Fatalf("audit not masked: before=%s after=%s", e.Before, e.After)
	}
	// 明文不能泄露到 audit
	if strings.Contains(string(e.After), "hunter2") {
		t.Fatalf("plaintext leaked: %s", e.After)
	}
	if strings.Contains(string(e.Before), "old") {
		t.Fatalf("plaintext leaked in before: %s", e.Before)
	}
}

func TestAdminSetting_TestSMTP_StubbedWhenMailerMissing(t *testing.T) {
	r, _, _ := newSettingRouter()
	body := `{"host":"smtp.example.com","port":465,"username":"a@b.com","password":"p","to":"x@y.com"}`
	rec := doReq(t, r, http.MethodPost, "/settings/test_smtp", body)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"stubbed":true`) {
		t.Fatalf("want stubbed=true, body=%s", rec.Body.String())
	}
}

func TestAdminSetting_TestSMTP_ValidatesFields(t *testing.T) {
	r, _, _ := newSettingRouter()
	body := `{"host":"","port":465,"username":"a@b.com","password":"p","to":"x@y.com"}`
	rec := doReq(t, r, http.MethodPost, "/settings/test_smtp", body)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}

	body2 := `{"host":"s","port":99999,"username":"a","password":"p","to":"x@y.com"}`
	rec2 := doReq(t, r, http.MethodPost, "/settings/test_smtp", body2)
	if rec2.Code != http.StatusBadRequest {
		t.Fatalf("want 400 for bad port, got %d", rec2.Code)
	}

	body3 := `{"host":"s","port":465,"username":"a","password":"p","to":"badmail"}`
	rec3 := doReq(t, r, http.MethodPost, "/settings/test_smtp", body3)
	if rec3.Code != http.StatusBadRequest {
		t.Fatalf("want 400 for bad email, got %d", rec3.Code)
	}
}

func TestAdminSetting_TestSMTP_WithMailer_RealSend(t *testing.T) {
	mailer := &mockMailer{}
	r, _, _ := newSettingRouter(func(h *SettingHandler) { h.Mailer = mailer })
	body := `{"host":"s","port":465,"username":"a","password":"p","to":"x@y.com"}`
	rec := doReq(t, r, http.MethodPost, "/settings/test_smtp", body)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	if !mailer.called {
		t.Fatal("mailer not called")
	}
	if strings.Contains(rec.Body.String(), `"stubbed":true`) {
		t.Fatalf("should not stub when mailer is set: %s", rec.Body.String())
	}
}

func TestAdminSetting_TestSMTP_WithMailer_Error(t *testing.T) {
	mailer := &mockMailer{err: errors.New("dial failed")}
	r, _, _ := newSettingRouter(func(h *SettingHandler) { h.Mailer = mailer })
	body := `{"host":"s","port":465,"username":"a","password":"p","to":"x@y.com"}`
	rec := doReq(t, r, http.MethodPost, "/settings/test_smtp", body)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200 (error is data not http), got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"ok":false`) || !strings.Contains(rec.Body.String(), "dial failed") {
		t.Fatalf("want ok=false with error; body=%s", rec.Body.String())
	}
}

func TestAdminSetting_Patch_EncryptError_500(t *testing.T) {
	r, store, _ := newSettingRouter(func(h *SettingHandler) { h.Crypto = errEncryptor{} })
	store.data["auth.smtp.password"] = setting.Setting{Key: "auth.smtp.password", Value: json.RawMessage(`"old"`)}
	rec := doReq(t, r, http.MethodPatch, "/settings/auth.smtp.password", `{"value":"hunter2"}`)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("want 500, got %d", rec.Code)
	}
}
