package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ijry/pro-api/internal/notice"
	"github.com/ijry/pro-api/internal/server/middleware"
	"github.com/ijry/pro-api/pkg/apierr"
)

// --- mock notice.Service ---

type mockSvc struct {
	mu      sync.Mutex
	data    map[int64]*notice.Notice
	nextID  int64
	now     time.Time
}

func newMock() *mockSvc {
	return &mockSvc{
		data:   map[int64]*notice.Notice{},
		nextID: 1,
		now:    time.Date(2026, 5, 21, 8, 0, 0, 0, time.UTC),
	}
}

func (m *mockSvc) Create(ctx context.Context, in notice.CreateInput, createdBy int64) (*notice.Notice, error) {
	if in.Title == "" {
		return nil, apierr.New(apierr.CodeMissingParam, "title 必填")
	}
	if in.Content == "" {
		return nil, apierr.New(apierr.CodeMissingParam, "content 必填")
	}
	if in.Level != "" && !notice.IsValidLevel(in.Level) {
		return nil, apierr.New(apierr.CodeInvalidParam, "level")
	}
	if in.Target != "" && !notice.IsValidTarget(in.Target) {
		return nil, apierr.New(apierr.CodeInvalidParam, "target")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	id := m.nextID
	m.nextID++
	level := in.Level
	if level == "" {
		level = notice.LevelInfo
	}
	target := in.Target
	if target == "" {
		target = notice.TargetAll
	}
	n := &notice.Notice{
		ID:        id,
		Title:     in.Title,
		Content:   in.Content,
		Level:     level,
		Target:    target,
		Status:    notice.StatusDraft,
		PublishAt: in.PublishAt,
		ExpiresAt: in.ExpiresAt,
		Pinned:    in.Pinned,
		CreatedBy: createdBy,
		CreatedAt: m.now,
		UpdatedAt: m.now,
	}
	m.data[id] = n
	return n, nil
}

func (m *mockSvc) List(ctx context.Context, f notice.ListFilter) ([]*notice.Notice, int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []*notice.Notice
	for _, n := range m.data {
		if f.Status >= 0 && n.Status != f.Status {
			continue
		}
		out = append(out, n)
	}
	return out, int64(len(out)), nil
}

func (m *mockSvc) Get(ctx context.Context, id int64) (*notice.Notice, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	n, ok := m.data[id]
	if !ok {
		return nil, apierr.New(apierr.CodeNotFound, "公告不存在")
	}
	return n, nil
}

func (m *mockSvc) Update(ctx context.Context, id int64, p notice.UpdatePatch, actor int64) (*notice.Notice, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	n, ok := m.data[id]
	if !ok {
		return nil, apierr.New(apierr.CodeNotFound, "公告不存在")
	}
	if p.Title != nil {
		n.Title = *p.Title
	}
	if p.ExpiresAtNull {
		n.ExpiresAt = nil
	} else if p.ExpiresAt != nil {
		t := *p.ExpiresAt
		n.ExpiresAt = &t
	}
	return n, nil
}

func (m *mockSvc) Delete(ctx context.Context, id int64, actor int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.data[id]; !ok {
		return apierr.New(apierr.CodeNotFound, "公告不存在")
	}
	delete(m.data, id)
	return nil
}

func (m *mockSvc) Publish(ctx context.Context, id int64, actor int64) (*notice.Notice, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	n, ok := m.data[id]
	if !ok {
		return nil, apierr.New(apierr.CodeNotFound, "公告不存在")
	}
	if n.Status == notice.StatusPublished {
		return n, nil
	}
	n.Status = notice.StatusPublished
	if n.PublishAt == nil {
		t := m.now
		n.PublishAt = &t
	}
	return n, nil
}

func (m *mockSvc) Unpublish(ctx context.Context, id int64, actor int64) (*notice.Notice, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	n, ok := m.data[id]
	if !ok {
		return nil, apierr.New(apierr.CodeNotFound, "公告不存在")
	}
	n.Status = notice.StatusArchived
	return n, nil
}

func (m *mockSvc) ListForUser(ctx context.Context, userID int64, page, size int) ([]*notice.UserNotice, int64, error) {
	return nil, 0, nil
}

func (m *mockSvc) GetForUser(ctx context.Context, userID int64, id int64) (*notice.UserNotice, error) {
	return nil, nil
}

func (m *mockSvc) MarkRead(ctx context.Context, userID, noticeID int64) error {
	return nil
}

func (m *mockSvc) UnreadCountForUser(ctx context.Context, userID int64) int {
	return 0
}

func (m *mockSvc) ListPublic(ctx context.Context, page, size int) ([]*notice.Notice, int64, error) {
	return nil, 0, nil
}

// --- setup ---

func newAdminRouter() (*gin.Engine, *mockSvc) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorResponse("json"))
	svc := newMock()
	h := NewNoticeHandler(svc, func(c *gin.Context) int64 { return 999 })
	h.Register(r)
	return r, svc
}

func doReq(t *testing.T, r http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

// --- tests ---

func TestAdminNotice_Create_201Returned(t *testing.T) {
	r, _ := newAdminRouter()
	rec := doReq(t, r, http.MethodPost, "/notices", `{"title":"t","content":"c"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAdminNotice_Create_RequiresTitle(t *testing.T) {
	r, _ := newAdminRouter()
	rec := doReq(t, r, http.MethodPost, "/notices", `{"content":"c"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAdminNotice_Create_InvalidLevel400(t *testing.T) {
	r, _ := newAdminRouter()
	rec := doReq(t, r, http.MethodPost, "/notices", `{"title":"t","content":"c","level":"bad"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

func TestAdminNotice_List_FilterByStatus(t *testing.T) {
	r, svc := newAdminRouter()
	// 创建 2 个 published + 1 个 draft
	svc.data[1] = &notice.Notice{ID: 1, Status: notice.StatusDraft}
	svc.data[2] = &notice.Notice{ID: 2, Status: notice.StatusPublished}
	svc.data[3] = &notice.Notice{ID: 3, Status: notice.StatusPublished}

	rec := doReq(t, r, http.MethodGet, "/notices?status=1", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	var body struct {
		Data struct {
			Total int `json:"total"`
		} `json:"data"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body.Data.Total != 2 {
		t.Fatalf("want 2, got %d", body.Data.Total)
	}
}

func TestAdminNotice_Get_404OnMissing(t *testing.T) {
	r, _ := newAdminRouter()
	rec := doReq(t, r, http.MethodGet, "/notices/999", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", rec.Code)
	}
}

func TestAdminNotice_Patch_RejectsStatusField(t *testing.T) {
	r, svc := newAdminRouter()
	svc.data[1] = &notice.Notice{ID: 1, Title: "x"}
	rec := doReq(t, r, http.MethodPatch, "/notices/1", `{"status":1}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAdminNotice_Patch_NullableExpiresAt(t *testing.T) {
	r, svc := newAdminRouter()
	future := time.Now().Add(time.Hour)
	svc.data[1] = &notice.Notice{ID: 1, Title: "x", ExpiresAt: &future}
	rec := doReq(t, r, http.MethodPatch, "/notices/1", `{"expires_at":null}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if svc.data[1].ExpiresAt != nil {
		t.Fatal("expires_at should be cleared")
	}
}

func TestAdminNotice_Delete_204(t *testing.T) {
	r, svc := newAdminRouter()
	svc.data[1] = &notice.Notice{ID: 1, Title: "x"}
	rec := doReq(t, r, http.MethodDelete, "/notices/1", "")
	if rec.Code != http.StatusNoContent {
		t.Fatalf("want 204, got %d", rec.Code)
	}
	if _, ok := svc.data[1]; ok {
		t.Fatal("not deleted")
	}
}

func TestAdminNotice_Publish_FromDraft(t *testing.T) {
	r, svc := newAdminRouter()
	svc.data[1] = &notice.Notice{ID: 1, Status: notice.StatusDraft}
	rec := doReq(t, r, http.MethodPost, "/notices/1/publish", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	if svc.data[1].Status != notice.StatusPublished {
		t.Fatalf("status not updated")
	}
}

func TestAdminNotice_Unpublish_FromPublished(t *testing.T) {
	r, svc := newAdminRouter()
	svc.data[1] = &notice.Notice{ID: 1, Status: notice.StatusPublished}
	rec := doReq(t, r, http.MethodPost, "/notices/1/unpublish", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	if svc.data[1].Status != notice.StatusArchived {
		t.Fatalf("status not Archived")
	}
}

func TestAdminNotice_Patch_InvalidIDBadRequest(t *testing.T) {
	r, _ := newAdminRouter()
	rec := doReq(t, r, http.MethodGet, "/notices/notanid", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "id 必须为整数") {
		t.Fatalf("body=%s", rec.Body.String())
	}
}
