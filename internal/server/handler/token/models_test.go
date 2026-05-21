package token

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ijry/pro-api/internal/server/middleware"
	tokensvc "github.com/ijry/pro-api/internal/token"
)

// stubModelLister 实现 ModelLister 用于测试。
type stubModelLister struct {
	active []string
	meta   map[string]tokensvc.ModelMeta
}

func (s *stubModelLister) ActiveModels(_ context.Context) []string { return s.active }
func (s *stubModelLister) ModelInfo(m string) (tokensvc.ModelMeta, bool) {
	if s.meta == nil {
		return tokensvc.ModelMeta{}, false
	}
	mm, ok := s.meta[m]
	return mm, ok
}

func newModelsEngine(t *testing.T, lister tokensvc.ModelLister, view *tokensvc.View) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorResponse("openai"))
	// 注入预设 view(模拟 TokenAuth 已通过)
	r.Use(func(c *gin.Context) {
		if view != nil {
			c.Set(tokensvc.CtxKeyToken, view)
		}
		c.Next()
	})
	h := NewModelsHandler(lister)
	h.Register(r.Group("/v1"))
	return r
}

func TestModelsHandler_NoFilter_ReturnsAll(t *testing.T) {
	lister := &stubModelLister{active: []string{"gpt-4o", "claude-3"}}
	view := &tokensvc.View{ID: 1}
	r := newModelsEngine(t, lister, view)
	req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var body ModelsResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body.Object != "list" || len(body.Data) != 2 {
		t.Fatalf("unexpected: %+v", body)
	}
}

func TestModelsHandler_WithAllowedModels_Filtered(t *testing.T) {
	lister := &stubModelLister{active: []string{"gpt-4o", "claude-3", "gpt-4-turbo"}}
	view := &tokensvc.View{ID: 1, AllowedModels: []string{"gpt-4o"}}
	r := newModelsEngine(t, lister, view)
	req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	var body ModelsResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if len(body.Data) != 1 || body.Data[0].ID != "gpt-4o" {
		t.Fatalf("filter wrong: %+v", body)
	}
}

func TestModelsHandler_WildcardFilter_Matches(t *testing.T) {
	lister := &stubModelLister{active: []string{"gpt-4o", "claude-3", "gpt-4-turbo"}}
	view := &tokensvc.View{ID: 1, AllowedModels: []string{"gpt-4*"}}
	r := newModelsEngine(t, lister, view)
	req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	var body ModelsResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if len(body.Data) != 2 {
		t.Fatalf("want 2 gpt-4*, got %+v", body)
	}
}

func TestModelsHandler_NoChannel_ReturnsEmptyList(t *testing.T) {
	lister := &stubModelLister{active: []string{}}
	view := &tokensvc.View{ID: 1}
	r := newModelsEngine(t, lister, view)
	req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	var body ModelsResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body.Object != "list" || len(body.Data) != 0 {
		t.Fatalf("want empty list, got %+v", body)
	}
}

func TestModelsHandler_ResponseFormatMatchesOpenAI(t *testing.T) {
	lister := &stubModelLister{
		active: []string{"gpt-4o"},
		meta: map[string]tokensvc.ModelMeta{
			"gpt-4o": {Created: 1714521600, OwnedBy: "openai"},
		},
	}
	view := &tokensvc.View{ID: 1}
	r := newModelsEngine(t, lister, view)
	req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	var body ModelsResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if len(body.Data) != 1 {
		t.Fatal("want 1 model")
	}
	got := body.Data[0]
	if got.Object != "model" || got.Created != 1714521600 || got.OwnedBy != "openai" {
		t.Fatalf("response shape wrong: %+v", got)
	}
}

func TestModelsHandler_MissingToken_Returns401(t *testing.T) {
	lister := &stubModelLister{active: []string{"gpt-4o"}}
	r := newModelsEngine(t, lister, nil)
	req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestModelsHandler_NilLister_EmptyList(t *testing.T) {
	view := &tokensvc.View{ID: 1}
	r := newModelsEngine(t, nil, view)
	req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	var body ModelsResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body.Object != "list" || len(body.Data) != 0 {
		t.Fatalf("nil lister should return empty list, got %+v", body)
	}
}
