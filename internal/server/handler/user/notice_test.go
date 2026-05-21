package user

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ijry/pro-api/internal/notice"
	"github.com/ijry/pro-api/internal/server/middleware"
	"github.com/ijry/pro-api/pkg/apierr"
)

// userMockSvc 是 user handler 测试用的 notice.Service 假实现。
type userMockSvc struct {
	mu        sync.Mutex
	visible   map[int64]*notice.Notice
	read      map[int64]map[int64]struct{}
	hidden    map[int64]bool // 标记 id 仅存在但用户不可见(返 404)
}

func newUserMock() *userMockSvc {
	return &userMockSvc{
		visible: map[int64]*notice.Notice{},
		read:    map[int64]map[int64]struct{}{},
		hidden:  map[int64]bool{},
	}
}

func (m *userMockSvc) Create(context.Context, notice.CreateInput, int64) (*notice.Notice, error) {
	return nil, nil
}
func (m *userMockSvc) List(context.Context, notice.ListFilter) ([]*notice.Notice, int64, error) {
	return nil, 0, nil
}
func (m *userMockSvc) Get(context.Context, int64) (*notice.Notice, error) { return nil, nil }
func (m *userMockSvc) Update(context.Context, int64, notice.UpdatePatch, int64) (*notice.Notice, error) {
	return nil, nil
}
func (m *userMockSvc) Delete(context.Context, int64, int64) error                       { return nil }
func (m *userMockSvc) Publish(context.Context, int64, int64) (*notice.Notice, error)    { return nil, nil }
func (m *userMockSvc) Unpublish(context.Context, int64, int64) (*notice.Notice, error)  { return nil, nil }
func (m *userMockSvc) ListPublic(context.Context, int, int) ([]*notice.Notice, int64, error) {
	return nil, 0, nil
}

func (m *userMockSvc) ListForUser(ctx context.Context, uid int64, page, size int) ([]*notice.UserNotice, int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []*notice.UserNotice
	for _, n := range m.visible {
		_, isRead := m.read[uid][n.ID]
		out = append(out, notice.ToUserNotice(n, isRead))
	}
	return out, int64(len(out)), nil
}

func (m *userMockSvc) GetForUser(ctx context.Context, uid int64, id int64) (*notice.UserNotice, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.hidden[id] {
		return nil, apierr.New(apierr.CodeNotFound, "公告不存在")
	}
	n, ok := m.visible[id]
	if !ok {
		return nil, apierr.New(apierr.CodeNotFound, "公告不存在")
	}
	_, isRead := m.read[uid][id]
	return notice.ToUserNotice(n, isRead), nil
}

func (m *userMockSvc) MarkRead(ctx context.Context, uid, nid int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.read[uid] == nil {
		m.read[uid] = map[int64]struct{}{}
	}
	m.read[uid][nid] = struct{}{}
	return nil
}

func (m *userMockSvc) UnreadCountForUser(ctx context.Context, uid int64) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	unread := 0
	for id := range m.visible {
		if _, isRead := m.read[uid][id]; !isRead {
			unread++
		}
	}
	return unread
}

// --- setup ---

func newUserRouter(uid int64) (*gin.Engine, *userMockSvc) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorResponse("json"))
	svc := newUserMock()
	h := NewNoticeHandler(svc, func(c *gin.Context) int64 { return uid })
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

func TestUserNotice_Unauthenticated_401(t *testing.T) {
	r, _ := newUserRouter(0)
	rec := doReq(t, r, http.MethodGet, "/notices", "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", rec.Code)
	}
}

func TestUserNotice_List_OK(t *testing.T) {
	r, svc := newUserRouter(1)
	svc.visible[10] = &notice.Notice{ID: 10, Title: "t1"}
	svc.visible[20] = &notice.Notice{ID: 20, Title: "t2"}
	rec := doReq(t, r, http.MethodGet, "/notices", "")
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

func TestUserNotice_List_MarksIsRead(t *testing.T) {
	r, svc := newUserRouter(1)
	svc.visible[10] = &notice.Notice{ID: 10, Title: "t1"}
	svc.visible[20] = &notice.Notice{ID: 20, Title: "t2"}
	_ = svc.MarkRead(context.Background(), 1, 20)
	rec := doReq(t, r, http.MethodGet, "/notices", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	var body struct {
		Data struct {
			Items []struct {
				ID     string `json:"id"`
				IsRead bool   `json:"is_read"`
			} `json:"items"`
		} `json:"data"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	got := map[string]bool{}
	for _, it := range body.Data.Items {
		got[it.ID] = it.IsRead
	}
	if !got["20"] {
		t.Fatalf("20 should be read, got %+v", got)
	}
	if got["10"] {
		t.Fatalf("10 should not be read")
	}
}

func TestUserNotice_Get_404OnHidden(t *testing.T) {
	r, svc := newUserRouter(1)
	svc.hidden[999] = true
	rec := doReq(t, r, http.MethodGet, "/notices/999", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", rec.Code)
	}
}

func TestUserNotice_Read_204(t *testing.T) {
	r, svc := newUserRouter(1)
	svc.visible[10] = &notice.Notice{ID: 10}
	rec := doReq(t, r, http.MethodPost, "/notices/10/read", "")
	if rec.Code != http.StatusNoContent {
		t.Fatalf("want 204, got %d", rec.Code)
	}
	if _, ok := svc.read[1][10]; !ok {
		t.Fatal("not marked read")
	}
}

func TestUserNotice_Read_Idempotent(t *testing.T) {
	r, svc := newUserRouter(1)
	svc.visible[10] = &notice.Notice{ID: 10}
	_ = doReq(t, r, http.MethodPost, "/notices/10/read", "")
	rec := doReq(t, r, http.MethodPost, "/notices/10/read", "")
	if rec.Code != http.StatusNoContent {
		t.Fatalf("want 204, got %d", rec.Code)
	}
}

func TestUserNotice_UnreadCount_Returns3(t *testing.T) {
	r, svc := newUserRouter(1)
	svc.visible[1] = &notice.Notice{ID: 1}
	svc.visible[2] = &notice.Notice{ID: 2}
	svc.visible[3] = &notice.Notice{ID: 3}
	rec := doReq(t, r, http.MethodGet, "/notices/unread_count", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	var body struct {
		Data struct {
			UnreadCount int `json:"unread_count"`
		} `json:"data"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body.Data.UnreadCount != 3 {
		t.Fatalf("want 3, got %d", body.Data.UnreadCount)
	}
}

func TestUserNotice_UnreadCount_AllRead_Returns0(t *testing.T) {
	r, svc := newUserRouter(1)
	svc.visible[1] = &notice.Notice{ID: 1}
	_ = svc.MarkRead(context.Background(), 1, 1)
	rec := doReq(t, r, http.MethodGet, "/notices/unread_count", "")
	var body struct {
		Data struct {
			UnreadCount int `json:"unread_count"`
		} `json:"data"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body.Data.UnreadCount != 0 {
		t.Fatalf("want 0, got %d", body.Data.UnreadCount)
	}
}
