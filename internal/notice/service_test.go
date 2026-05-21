package notice

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/ijry/pro-api/internal/audit"
	"github.com/ijry/pro-api/pkg/apierr"
	"go.uber.org/zap"
)

// --- 测试辅助 mock ---

type fakeRepo struct {
	mu        sync.Mutex
	data      map[int64]*Notice
	visibleIDs []int64 // VisibleIDsForUser 直接返这个,便于测 UnreadCount
	failNext  error
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{data: map[int64]*Notice{}}
}

func (f *fakeRepo) Create(ctx context.Context, n *Notice) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.failNext != nil {
		err := f.failNext
		f.failNext = nil
		return err
	}
	c := *n
	f.data[n.ID] = &c
	return nil
}

func (f *fakeRepo) GetByID(ctx context.Context, id int64) (*Notice, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	n, ok := f.data[id]
	if !ok {
		return nil, nil
	}
	c := *n
	return &c, nil
}

func (f *fakeRepo) Update(ctx context.Context, id int64, fields map[string]any) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	n, ok := f.data[id]
	if !ok {
		return ErrNotFound
	}
	for k, v := range fields {
		switch k {
		case "title":
			n.Title = v.(string)
		case "content":
			n.Content = v.(string)
		case "level":
			n.Level = v.(string)
		case "target":
			n.Target = v.(string)
		case "pinned":
			n.Pinned = v.(bool)
		case "status":
			n.Status = v.(int8)
		case "publish_at":
			if v == nil {
				n.PublishAt = nil
			} else {
				t := v.(time.Time)
				n.PublishAt = &t
			}
		case "expires_at":
			if v == nil {
				n.ExpiresAt = nil
			} else {
				t := v.(time.Time)
				n.ExpiresAt = &t
			}
		case "updated_at":
			n.UpdatedAt = v.(time.Time)
		}
	}
	return nil
}

func (f *fakeRepo) Delete(ctx context.Context, id int64) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.data[id]; !ok {
		return ErrNotFound
	}
	delete(f.data, id)
	return nil
}

func (f *fakeRepo) ListAdmin(ctx context.Context, status int8, page, size int) ([]*Notice, int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []*Notice
	for _, n := range f.data {
		if status >= 0 && n.Status != status {
			continue
		}
		c := *n
		out = append(out, &c)
	}
	return out, int64(len(out)), nil
}

func (f *fakeRepo) ListVisibleForUser(ctx context.Context, targets []string, now time.Time, page, size int) ([]*Notice, int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []*Notice
	for _, n := range f.data {
		if n.Status != StatusPublished {
			continue
		}
		if n.PublishAt != nil && n.PublishAt.After(now) {
			continue
		}
		if n.ExpiresAt != nil && !n.ExpiresAt.After(now) {
			continue
		}
		if !containsStr(targets, n.Target) {
			continue
		}
		c := *n
		out = append(out, &c)
	}
	return out, int64(len(out)), nil
}

func (f *fakeRepo) CountVisibleForUser(ctx context.Context, targets []string, now time.Time) (int64, error) {
	items, _, err := f.ListVisibleForUser(ctx, targets, now, 1, 1<<30)
	return int64(len(items)), err
}

func (f *fakeRepo) VisibleIDsForUser(ctx context.Context, targets []string, now time.Time) ([]int64, error) {
	if f.visibleIDs != nil {
		return f.visibleIDs, nil
	}
	items, _, err := f.ListVisibleForUser(ctx, targets, now, 1, 1<<30)
	if err != nil {
		return nil, err
	}
	out := make([]int64, 0, len(items))
	for _, n := range items {
		out = append(out, n.ID)
	}
	return out, nil
}

func containsStr(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}

type fakeReader struct {
	mu      sync.Mutex
	sets    map[int64]map[int64]struct{}
	failGet bool
}

func newFakeReader() *fakeReader {
	return &fakeReader{sets: map[int64]map[int64]struct{}{}}
}

func (r *fakeReader) MarkRead(ctx context.Context, userID, noticeID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.sets[userID] == nil {
		r.sets[userID] = map[int64]struct{}{}
	}
	r.sets[userID][noticeID] = struct{}{}
	return nil
}

func (r *fakeReader) IsRead(ctx context.Context, userID, noticeID int64) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.sets[userID][noticeID]
	return ok, nil
}

func (r *fakeReader) ReadSet(ctx context.Context, userID int64) (map[int64]struct{}, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := map[int64]struct{}{}
	for k := range r.sets[userID] {
		out[k] = struct{}{}
	}
	return out, nil
}

func (r *fakeReader) UnreadCount(ctx context.Context, userID int64, visibleIDs []int64) (int, error) {
	if r.failGet {
		return 0, errors.New("redis down")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	unread := 0
	for _, id := range visibleIDs {
		if _, ok := r.sets[userID][id]; !ok {
			unread++
		}
	}
	return unread, nil
}

type fakeIDGen struct {
	cur int64
}

func (g *fakeIDGen) Generate() int64 {
	g.cur++
	return g.cur
}

type fakeClock struct {
	now time.Time
}

func (c *fakeClock) Now() time.Time         { return c.now }
func (c *fakeClock) Sleep(d time.Duration)  {}
func (c *fakeClock) NewTicker(d time.Duration) interface {
	C() <-chan time.Time
	Stop()
} {
	return nil
}

type fakeAudit struct {
	mu      sync.Mutex
	entries []audit.Entry
}

func (a *fakeAudit) Log(ctx context.Context, e audit.Entry) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.entries = append(a.entries, e)
	return nil
}

func newTestService() (*service, *fakeRepo, *fakeReader, *fakeAudit, *fakeClock) {
	fr := newFakeRepo()
	fread := newFakeReader()
	fa := &fakeAudit{}
	fc := &fakeClock{now: time.Date(2026, 5, 21, 8, 0, 0, 0, time.UTC)}
	svc := &service{
		repo:   fr,
		reader: fread,
		idgen:  &fakeIDGen{},
		clock:  fc,
		audit:  fa,
		log:    zap.NewNop(),
	}
	return svc, fr, fread, fa, fc
}

// --- 测试用例 ---

func TestService_Create_StatusIsDraft(t *testing.T) {
	svc, _, _, _, _ := newTestService()
	n, err := svc.Create(context.Background(), CreateInput{Title: "t", Content: "c"}, 99)
	if err != nil {
		t.Fatal(err)
	}
	if n.Status != StatusDraft {
		t.Fatalf("want Draft, got %d", n.Status)
	}
}

func TestService_Create_GeneratesID(t *testing.T) {
	svc, _, _, _, _ := newTestService()
	n, _ := svc.Create(context.Background(), CreateInput{Title: "t", Content: "c"}, 99)
	if n.ID == 0 {
		t.Fatal("ID not set")
	}
}

func TestService_Create_DefaultsLevelAndTarget(t *testing.T) {
	svc, _, _, _, _ := newTestService()
	n, _ := svc.Create(context.Background(), CreateInput{Title: "t", Content: "c"}, 99)
	if n.Level != LevelInfo {
		t.Fatalf("level want info, got %s", n.Level)
	}
	if n.Target != TargetAll {
		t.Fatalf("target want all, got %s", n.Target)
	}
}

func TestService_Create_AuditLogged(t *testing.T) {
	svc, _, _, fa, _ := newTestService()
	_, _ = svc.Create(context.Background(), CreateInput{Title: "t", Content: "c"}, 99)
	if len(fa.entries) != 1 || fa.entries[0].Action != "notice.create" {
		t.Fatalf("want one notice.create audit, got %+v", fa.entries)
	}
}

func TestService_Create_MissingTitle_ReturnsError(t *testing.T) {
	svc, _, _, _, _ := newTestService()
	_, err := svc.Create(context.Background(), CreateInput{Content: "c"}, 99)
	if err == nil {
		t.Fatal("want error")
	}
	var ae *apierr.Error
	if !errors.As(err, &ae) || ae.Code != apierr.CodeMissingParam {
		t.Fatalf("want CodeMissingParam, got %v", err)
	}
}

func TestService_Create_InvalidLevel_ReturnsError(t *testing.T) {
	svc, _, _, _, _ := newTestService()
	_, err := svc.Create(context.Background(), CreateInput{Title: "t", Content: "c", Level: "bad"}, 99)
	var ae *apierr.Error
	if !errors.As(err, &ae) || ae.Code != apierr.CodeInvalidParam {
		t.Fatalf("want CodeInvalidParam, got %v", err)
	}
}

func TestService_Create_InvalidTarget_ReturnsError(t *testing.T) {
	svc, _, _, _, _ := newTestService()
	_, err := svc.Create(context.Background(), CreateInput{Title: "t", Content: "c", Target: "bad"}, 99)
	var ae *apierr.Error
	if !errors.As(err, &ae) || ae.Code != apierr.CodeInvalidParam {
		t.Fatalf("want CodeInvalidParam, got %v", err)
	}
}

func TestService_Create_ExpiresBeforeNow_ReturnsError(t *testing.T) {
	svc, _, _, _, fc := newTestService()
	past := fc.now.Add(-1 * time.Hour)
	_, err := svc.Create(context.Background(), CreateInput{Title: "t", Content: "c", ExpiresAt: &past}, 99)
	var ae *apierr.Error
	if !errors.As(err, &ae) || ae.Code != apierr.CodeInvalidParam {
		t.Fatalf("want CodeInvalidParam, got %v", err)
	}
}

func TestService_Publish_DraftToPublished(t *testing.T) {
	svc, _, _, _, _ := newTestService()
	n, _ := svc.Create(context.Background(), CreateInput{Title: "t", Content: "c"}, 99)
	got, err := svc.Publish(context.Background(), n.ID, 99)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != StatusPublished {
		t.Fatalf("want Published, got %d", got.Status)
	}
}

func TestService_Publish_FillsPublishAtIfNull(t *testing.T) {
	svc, _, _, _, fc := newTestService()
	n, _ := svc.Create(context.Background(), CreateInput{Title: "t", Content: "c"}, 99)
	got, _ := svc.Publish(context.Background(), n.ID, 99)
	if got.PublishAt == nil || !got.PublishAt.Equal(fc.now) {
		t.Fatalf("want publish_at=%v, got %v", fc.now, got.PublishAt)
	}
}

func TestService_Publish_DoesNotOverridePublishAtIfSet(t *testing.T) {
	svc, _, _, _, fc := newTestService()
	pre := fc.now.Add(-2 * time.Hour)
	n, _ := svc.Create(context.Background(), CreateInput{Title: "t", Content: "c", PublishAt: &pre}, 99)
	got, _ := svc.Publish(context.Background(), n.ID, 99)
	if got.PublishAt == nil || !got.PublishAt.Equal(pre) {
		t.Fatalf("want %v, got %v", pre, got.PublishAt)
	}
}

func TestService_Publish_AlreadyPublished_Idempotent(t *testing.T) {
	svc, _, _, fa, _ := newTestService()
	n, _ := svc.Create(context.Background(), CreateInput{Title: "t", Content: "c"}, 99)
	_, _ = svc.Publish(context.Background(), n.ID, 99) // first
	preAudits := len(fa.entries)
	_, err := svc.Publish(context.Background(), n.ID, 99) // second
	if err != nil {
		t.Fatal(err)
	}
	if len(fa.entries) != preAudits {
		t.Fatal("second publish should not audit")
	}
}

func TestService_Publish_NotFound_ReturnsError(t *testing.T) {
	svc, _, _, _, _ := newTestService()
	_, err := svc.Publish(context.Background(), 999, 99)
	var ae *apierr.Error
	if !errors.As(err, &ae) || ae.Code != apierr.CodeNotFound {
		t.Fatalf("want CodeNotFound, got %v", err)
	}
}

func TestService_Unpublish_PublishedToArchived(t *testing.T) {
	svc, _, _, _, _ := newTestService()
	n, _ := svc.Create(context.Background(), CreateInput{Title: "t", Content: "c"}, 99)
	_, _ = svc.Publish(context.Background(), n.ID, 99)
	got, _ := svc.Unpublish(context.Background(), n.ID, 99)
	if got.Status != StatusArchived {
		t.Fatalf("want Archived, got %d", got.Status)
	}
}

func TestService_Update_RejectsStatusField(t *testing.T) {
	// Service.Update 接受 UpdatePatch,UpdatePatch 没有 Status 字段 — 但要求 handler 拒绝 body 含 status。
	// 这里我们只测试 Update 不会修改 status。
	svc, _, _, _, _ := newTestService()
	n, _ := svc.Create(context.Background(), CreateInput{Title: "t", Content: "c"}, 99)
	_, _ = svc.Publish(context.Background(), n.ID, 99)
	newTitle := "new"
	got, err := svc.Update(context.Background(), n.ID, UpdatePatch{Title: &newTitle}, 99)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != StatusPublished {
		t.Fatalf("status should remain Published, got %d", got.Status)
	}
	if got.Title != "new" {
		t.Fatalf("want title=new, got %s", got.Title)
	}
}

func TestService_Update_NullsExpiresAt(t *testing.T) {
	svc, _, _, _, fc := newTestService()
	future := fc.now.Add(1 * time.Hour)
	n, _ := svc.Create(context.Background(), CreateInput{Title: "t", Content: "c", ExpiresAt: &future}, 99)
	got, err := svc.Update(context.Background(), n.ID, UpdatePatch{ExpiresAtNull: true}, 99)
	if err != nil {
		t.Fatal(err)
	}
	if got.ExpiresAt != nil {
		t.Fatalf("want nil expires_at, got %v", got.ExpiresAt)
	}
}

func TestService_Update_NotFound_ReturnsError(t *testing.T) {
	svc, _, _, _, _ := newTestService()
	title := "x"
	_, err := svc.Update(context.Background(), 999, UpdatePatch{Title: &title}, 99)
	var ae *apierr.Error
	if !errors.As(err, &ae) || ae.Code != apierr.CodeNotFound {
		t.Fatalf("want CodeNotFound, got %v", err)
	}
}

func TestService_Delete_RemovesAndAudits(t *testing.T) {
	svc, fr, _, fa, _ := newTestService()
	n, _ := svc.Create(context.Background(), CreateInput{Title: "t", Content: "c"}, 99)
	if err := svc.Delete(context.Background(), n.ID, 99); err != nil {
		t.Fatal(err)
	}
	if _, ok := fr.data[n.ID]; ok {
		t.Fatal("not deleted")
	}
	found := false
	for _, e := range fa.entries {
		if e.Action == "notice.delete" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("no audit")
	}
}

func TestService_ListForUser_MarksIsRead(t *testing.T) {
	svc, _, fread, _, fc := newTestService()
	t1 := fc.now.Add(-1 * time.Hour)
	n1, _ := svc.Create(context.Background(), CreateInput{Title: "a", Content: "c", PublishAt: &t1}, 99)
	n2, _ := svc.Create(context.Background(), CreateInput{Title: "b", Content: "c", PublishAt: &t1}, 99)
	_, _ = svc.Publish(context.Background(), n1.ID, 99)
	_, _ = svc.Publish(context.Background(), n2.ID, 99)

	_ = fread.MarkRead(context.Background(), 1, n1.ID)
	items, total, err := svc.ListForUser(context.Background(), 1, 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 || len(items) != 2 {
		t.Fatalf("want 2, got %d/%d", total, len(items))
	}
	got := map[int64]bool{}
	for _, u := range items {
		got[u.ID] = u.IsRead
	}
	if !got[n1.ID] {
		t.Fatalf("n1 should be read")
	}
	if got[n2.ID] {
		t.Fatalf("n2 should be unread")
	}
}

func TestService_GetForUser_HiddenReturns404(t *testing.T) {
	svc, _, _, _, _ := newTestService()
	n, _ := svc.Create(context.Background(), CreateInput{Title: "t", Content: "c"}, 99) // status=draft
	_, err := svc.GetForUser(context.Background(), 1, n.ID)
	var ae *apierr.Error
	if !errors.As(err, &ae) || ae.Code != apierr.CodeNotFound {
		t.Fatalf("want CodeNotFound, got %v", err)
	}
}

func TestService_GetForUser_VisibleReturnsWithIsRead(t *testing.T) {
	svc, _, fread, _, fc := newTestService()
	t1 := fc.now.Add(-1 * time.Hour)
	n, _ := svc.Create(context.Background(), CreateInput{Title: "t", Content: "c", PublishAt: &t1}, 99)
	_, _ = svc.Publish(context.Background(), n.ID, 99)
	_ = fread.MarkRead(context.Background(), 1, n.ID)
	got, err := svc.GetForUser(context.Background(), 1, n.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !got.IsRead {
		t.Fatal("want is_read")
	}
}

func TestService_MarkRead_DelegatesToReader(t *testing.T) {
	svc, _, fread, _, _ := newTestService()
	if err := svc.MarkRead(context.Background(), 1, 42); err != nil {
		t.Fatal(err)
	}
	if _, ok := fread.sets[1][42]; !ok {
		t.Fatal("not added to reader set")
	}
}

func TestService_MarkRead_NotFoundNoticeStillReturnsOK(t *testing.T) {
	svc, _, _, _, _ := newTestService()
	if err := svc.MarkRead(context.Background(), 1, 99999); err != nil {
		t.Fatalf("want nil error even for non-existing notice, got %v", err)
	}
}

func TestService_UnreadCountForUser_Basic(t *testing.T) {
	svc, _, _, _, fc := newTestService()
	t1 := fc.now.Add(-1 * time.Hour)
	n, _ := svc.Create(context.Background(), CreateInput{Title: "t", Content: "c", PublishAt: &t1}, 99)
	_, _ = svc.Publish(context.Background(), n.ID, 99)
	if got := svc.UnreadCountForUser(context.Background(), 1); got != 1 {
		t.Fatalf("want 1, got %d", got)
	}
}

func TestService_UnreadCountForUser_NoVisibleReturnsZero(t *testing.T) {
	svc, _, _, _, _ := newTestService()
	if got := svc.UnreadCountForUser(context.Background(), 1); got != 0 {
		t.Fatalf("want 0, got %d", got)
	}
}

func TestService_UnreadCountForUser_RedisErrorReturnsZero(t *testing.T) {
	svc, _, fread, _, fc := newTestService()
	t1 := fc.now.Add(-1 * time.Hour)
	n, _ := svc.Create(context.Background(), CreateInput{Title: "t", Content: "c", PublishAt: &t1}, 99)
	_, _ = svc.Publish(context.Background(), n.ID, 99)
	fread.failGet = true
	if got := svc.UnreadCountForUser(context.Background(), 1); got != 0 {
		t.Fatalf("want 0 on error, got %d", got)
	}
	_ = n
}

func TestService_ListPublic_ExcludesUserTarget(t *testing.T) {
	svc, _, _, _, fc := newTestService()
	t1 := fc.now.Add(-1 * time.Hour)
	a, _ := svc.Create(context.Background(), CreateInput{Title: "a", Content: "c", Target: TargetAll, PublishAt: &t1}, 99)
	u, _ := svc.Create(context.Background(), CreateInput{Title: "u", Content: "c", Target: TargetUser, PublishAt: &t1}, 99)
	_, _ = svc.Publish(context.Background(), a.ID, 99)
	_, _ = svc.Publish(context.Background(), u.ID, 99)

	items, total, err := svc.ListPublic(context.Background(), 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(items) != 1 || items[0].ID != a.ID {
		t.Fatalf("want only [%d], got %+v", a.ID, items)
	}
}
