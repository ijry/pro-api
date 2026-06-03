package token

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ijry/pro-api/internal/server/middleware"
	"github.com/ijry/pro-api/pkg/apierr"
)

// fakeStore 是只实现 Authenticate 的 Store 桩(其他方法 panic 不应被中间件触达)。
type fakeStore struct {
	view *View
	err  error
}

func (f *fakeStore) Authenticate(_ context.Context, _ string) (*View, error) {
	return f.view, f.err
}
func (f *fakeStore) Create(context.Context, CreateInput) (string, *View, error) {
	panic("not used")
}
func (f *fakeStore) List(context.Context, ListFilter) ([]*View, int64, error) {
	panic("not used")
}
func (f *fakeStore) Get(context.Context, int64) (*View, error)             { panic("not used") }
func (f *fakeStore) Update(context.Context, int64, UpdatePatch) (*View, error) {
	panic("not used")
}
func (f *fakeStore) Revoke(context.Context, int64) error                  { panic("not used") }
func (f *fakeStore) Regenerate(context.Context, int64) (string, *View, error) {
	panic("not used")
}
func (f *fakeStore) IncrementUsage(int64, int64)                            {}
func (f *fakeStore) TouchLastUsed(int64, time.Time)                         {}
func (f *fakeStore) Close() error                                           { return nil }

func setupAuthGin(t *testing.T, store Store) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorResponse("openai"))
	r.Use(TokenAuth(store))
	r.GET("/x", func(c *gin.Context) {
		v, ok := FromContext(c)
		uid := UserIDFromContext(c)
		gid := GroupIDFromContext(c)
		c.JSON(http.StatusOK, gin.H{
			"ok":       ok,
			"user_id":  uid,
			"group_id": gid,
			"view_id":  v.ID,
		})
	})
	return r
}

func TestTokenAuth_MissingHeader_ReturnsInvalidToken(t *testing.T) {
	r := setupAuthGin(t, &fakeStore{})
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestTokenAuth_NonBearer_ReturnsInvalidToken(t *testing.T) {
	r := setupAuthGin(t, &fakeStore{})
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Basic xxx")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", rec.Code)
	}
}

func TestTokenAuth_BadPrefix_ReturnsInvalidToken(t *testing.T) {
	r := setupAuthGin(t, &fakeStore{})
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer sk-not-pa")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", rec.Code)
	}
}

func TestTokenAuth_ValidKey_InjectsContext(t *testing.T) {
	store := &fakeStore{view: &View{ID: 7, UserID: 42}}
	r := setupAuthGin(t, store)
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer pa-anyplaceholder")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var body map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body["user_id"].(float64) != 42 {
		t.Fatalf("want user_id=42, got %v", body["user_id"])
	}
}

func TestTokenAuth_StoreError_Forwarded(t *testing.T) {
	store := &fakeStore{err: apierr.New(apierr.CodeTokenExpired, "expired")}
	r := setupAuthGin(t, store)
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer pa-xxxxxxxxxxxxxxxxxxxx")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", rec.Code)
	}
}

func TestTokenAuth_IPNotAllowed_ReturnsForbidden(t *testing.T) {
	store := &fakeStore{view: &View{ID: 1, UserID: 1, AllowedIPs: []string{"10.0.0.0/8"}}}
	r := setupAuthGin(t, store)
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer pa-xxxxxxxxxxxxxxxxxxxx")
	req.RemoteAddr = "1.2.3.4:9999"
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestTokenAuth_XForwardedFor_UsedAsRealIP(t *testing.T) {
	store := &fakeStore{view: &View{ID: 1, UserID: 1, AllowedIPs: []string{"10.0.0.0/8"}}}
	r := setupAuthGin(t, store)
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer pa-xxxxxxxxxxxxxxxxxxxx")
	req.Header.Set("X-Forwarded-For", "10.1.2.3, 8.8.8.8")
	req.RemoteAddr = "8.8.8.8:443"
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200 (XFF 10.x in CIDR), got %d body=%s", rec.Code, rec.Body.String())
	}
}


func TestExtractBearer_CaseInsensitiveScheme(t *testing.T) {
	for _, h := range []string{"Bearer pa-x", "bearer pa-x", "BEARER pa-x"} {
		key, ok := extractBearer(h)
		if !ok || key != "pa-x" {
			t.Fatalf("scheme %q: ok=%v key=%q", h, ok, key)
		}
	}
}

func TestExtractBearer_RejectsNonPAPrefix(t *testing.T) {
	if _, ok := extractBearer("Bearer sk-xxxx"); ok {
		t.Fatal("sk- prefix should be rejected")
	}
}

func TestFromContext_PlainContext_Works(t *testing.T) {
	// FromContext 内部用 string key(与 Gin Set 一致),测试也用同名常量。
	//nolint:staticcheck // SA1029: 与 Gin 兼容,key 必须是 string
	ctx2 := context.WithValue(context.Background(), CtxKeyToken, &View{ID: 5})
	got, ok := FromContext(ctx2)
	if !ok || got.ID != 5 {
		t.Fatalf("plain ctx FromContext failed: ok=%v got=%+v", ok, got)
	}
}
