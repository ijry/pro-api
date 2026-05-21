package token

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ijry/pro-api/internal/server/middleware"
	tokensvc "github.com/ijry/pro-api/internal/token"
)

func newAdminEngine(_ *testing.T, store tokensvc.Store) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorResponse("json"))
	ah := NewAdminHandler(store)
	g := r.Group("/api/admin/tokens")
	ah.Register(g)
	return r
}

func TestAdminHandler_List_All(t *testing.T) {
	store, cleanup := newHandlerFixture(t)
	defer cleanup()
	_, _ = createForUser(t, store, 1, "alice")
	_, _ = createForUser(t, store, 2, "bob")
	_, _ = createForUser(t, store, 3, "carol")

	r := newAdminEngine(t, store)
	rec := doJSON(t, r, http.MethodGet, "/api/admin/tokens", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var body ListResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body.Total != 3 {
		t.Fatalf("want total=3, got %d", body.Total)
	}
}

func TestAdminHandler_List_FilterByUser(t *testing.T) {
	store, cleanup := newHandlerFixture(t)
	defer cleanup()
	_, _ = createForUser(t, store, 1, "a1")
	_, _ = createForUser(t, store, 1, "a2")
	_, _ = createForUser(t, store, 2, "b1")

	r := newAdminEngine(t, store)
	rec := doJSON(t, r, http.MethodGet, "/api/admin/tokens?user_id=1", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	var body ListResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body.Total != 2 {
		t.Fatalf("want total=2 for user_id=1, got %d", body.Total)
	}
}

func TestAdminHandler_Patch_AnyUserToken(t *testing.T) {
	store, cleanup := newHandlerFixture(t)
	defer cleanup()
	id, _ := createForUser(t, store, 2, "victim")
	r := newAdminEngine(t, store)
	rpm := 7
	rec := doJSON(t, r, http.MethodPatch, "/api/admin/tokens/"+intToStr(id), PatchRequest{RPMLimit: &rpm})
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAdminHandler_Delete_AnyUserToken(t *testing.T) {
	store, cleanup := newHandlerFixture(t)
	defer cleanup()
	id, _ := createForUser(t, store, 2, "victim")
	r := newAdminEngine(t, store)
	rec := doJSON(t, r, http.MethodDelete, "/api/admin/tokens/"+intToStr(id), nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAdminHandler_NoRegenerateEndpoint(t *testing.T) {
	store, cleanup := newHandlerFixture(t)
	defer cleanup()
	id, _ := createForUser(t, store, 1, "t")
	r := newAdminEngine(t, store)
	rec := doJSON(t, r, http.MethodPost, "/api/admin/tokens/"+intToStr(id)+"/regenerate", nil)
	// 路由不存在 → 404
	if rec.Code != http.StatusNotFound {
		t.Fatalf("admin should not have regenerate endpoint, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAdminHandler_Detail(t *testing.T) {
	store, cleanup := newHandlerFixture(t)
	defer cleanup()
	id, _ := createForUser(t, store, 2, "x")
	r := newAdminEngine(t, store)
	rec := doJSON(t, r, http.MethodGet, "/api/admin/tokens/"+intToStr(id), nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	var body ViewDTO
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body.UserID != 2 {
		t.Fatalf("want user_id=2, got %d", body.UserID)
	}
}
