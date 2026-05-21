package public

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ijry/pro-api/internal/notice"
	"github.com/ijry/pro-api/internal/server/middleware"
)

// publicMockSvc 是公开 handler 测试用的 notice.Service 假实现。
type publicMockSvc struct {
	publicItems []*notice.Notice
}

func (m *publicMockSvc) Create(context.Context, notice.CreateInput, int64) (*notice.Notice, error) {
	return nil, nil
}
func (m *publicMockSvc) List(context.Context, notice.ListFilter) ([]*notice.Notice, int64, error) {
	return nil, 0, nil
}
func (m *publicMockSvc) Get(context.Context, int64) (*notice.Notice, error) { return nil, nil }
func (m *publicMockSvc) Update(context.Context, int64, notice.UpdatePatch, int64) (*notice.Notice, error) {
	return nil, nil
}
func (m *publicMockSvc) Delete(context.Context, int64, int64) error                      { return nil }
func (m *publicMockSvc) Publish(context.Context, int64, int64) (*notice.Notice, error)   { return nil, nil }
func (m *publicMockSvc) Unpublish(context.Context, int64, int64) (*notice.Notice, error) { return nil, nil }
func (m *publicMockSvc) ListForUser(context.Context, int64, int, int) ([]*notice.UserNotice, int64, error) {
	return nil, 0, nil
}
func (m *publicMockSvc) GetForUser(context.Context, int64, int64) (*notice.UserNotice, error) {
	return nil, nil
}
func (m *publicMockSvc) MarkRead(context.Context, int64, int64) error      { return nil }
func (m *publicMockSvc) UnreadCountForUser(context.Context, int64) int     { return 0 }
func (m *publicMockSvc) ListPublic(ctx context.Context, page, size int) ([]*notice.Notice, int64, error) {
	return m.publicItems, int64(len(m.publicItems)), nil
}

func TestPublicNotice_List_NoAuthRequired(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorResponse("json"))
	svc := &publicMockSvc{publicItems: []*notice.Notice{{ID: 1, Title: "t", Target: "all"}}}
	NewNoticeHandler(svc).Register(r)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/notices", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestPublicNotice_List_NoIsReadField(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorResponse("json"))
	svc := &publicMockSvc{publicItems: []*notice.Notice{{ID: 1, Title: "t"}}}
	NewNoticeHandler(svc).Register(r)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/notices", nil))
	// 检查 items[0] 不含 is_read 字段
	var body map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	data := body["data"].(map[string]any)
	items := data["items"].([]any)
	first := items[0].(map[string]any)
	if _, has := first["is_read"]; has {
		t.Fatalf("public should not return is_read; got %+v", first)
	}
}

func TestPublicNotice_List_Pagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorResponse("json"))
	svc := &publicMockSvc{publicItems: []*notice.Notice{{ID: 1}, {ID: 2}}}
	NewNoticeHandler(svc).Register(r)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/notices?page=2&size=5", nil))
	var body struct {
		Data struct {
			Page int `json:"page"`
			Size int `json:"size"`
		} `json:"data"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body.Data.Page != 2 || body.Data.Size != 5 {
		t.Fatalf("want page=2 size=5, got %+v", body.Data)
	}
}
