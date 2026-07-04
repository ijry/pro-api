package admin

// AccountHandler 测试。
//
// 注意:RoleGate(2)/RoleGate(3) 的 RBAC 强制由路由组装时挂载的 middleware 保证;
// 本 handler 测试不重复覆盖中间件层(由 cmd/proapi/main.go 与 role_gate_test 联合保证)。
// 这里只覆盖 handler 行为:成功路径、邮箱脱敏、dry_run、审计写入等。

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
	"github.com/ijry/pro-api/internal/account"
	"github.com/ijry/pro-api/internal/audit"
	"github.com/ijry/pro-api/internal/server/middleware"
)

// --- fakes ---

type fakeRepoEvent struct {
	accountID int64
	eventType string
	payload   any
}

// fakeRepo 是 admin handler 测试用的极简内存 Repo。
type fakeRepo struct {
	mu     sync.Mutex
	nextID int64
	items  map[int64]*account.Account
	events []fakeRepoEvent
}

func newFakeRepo() *fakeRepo { return &fakeRepo{items: map[int64]*account.Account{}} }

func (f *fakeRepo) Create(_ context.Context, a *account.Account) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nextID++
	if a.ID == 0 {
		a.ID = f.nextID
	}
	if a.CreatedAt.IsZero() {
		a.CreatedAt = time.Now().UTC()
	}
	a.UpdatedAt = time.Now().UTC()
	f.items[a.ID] = a
	return nil
}

func (f *fakeRepo) Update(_ context.Context, a *account.Account) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	cur, ok := f.items[a.ID]
	if !ok {
		return account.ErrNotFound
	}
	// Mirror "non-zero field" GORM semantics loosely: apply struct overlay.
	if a.Name != "" {
		cur.Name = a.Name
	}
	if a.Priority != 0 {
		cur.Priority = a.Priority
	}
	if a.Weight != 0 {
		cur.Weight = a.Weight
	}
	cur.Status = a.Status
	cur.CooldownUntil = a.CooldownUntil
	cur.ConsecFailures = a.ConsecFailures
	cur.UpdatedAt = time.Now().UTC()
	return nil
}

func (f *fakeRepo) Reactivate(_ context.Context, id int64) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	a, ok := f.items[id]
	if !ok {
		return account.ErrNotFound
	}
	a.Status = account.StatusActive
	a.CooldownUntil = nil
	a.ConsecFailures = 0
	return nil
}

func (f *fakeRepo) ResetFailures(_ context.Context, id int64) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	a, ok := f.items[id]
	if !ok {
		return account.ErrNotFound
	}
	a.ConsecFailures = 0
	now := time.Now().UTC()
	a.LastSuccessAt = &now
	a.LastUsedAt = &now
	return nil
}

func (f *fakeRepo) Get(_ context.Context, id int64) (*account.Account, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	a, ok := f.items[id]
	if !ok {
		return nil, account.ErrNotFound
	}
	return a, nil
}

func (f *fakeRepo) ListByChannel(_ context.Context, channelID int64) ([]*account.Account, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := []*account.Account{}
	for _, a := range f.items {
		if a.ChannelID == channelID {
			out = append(out, a)
		}
	}
	return out, nil
}

func (f *fakeRepo) ListForRefresher(context.Context, time.Time, int) ([]*account.Account, error) {
	return nil, nil
}

func (f *fakeRepo) ListForReaper(context.Context, time.Time, int) ([]*account.Account, error) {
	return nil, nil
}

func (f *fakeRepo) Delete(_ context.Context, id int64) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.items[id]; !ok {
		return account.ErrNotFound
	}
	delete(f.items, id)
	return nil
}

func (f *fakeRepo) AppendEvent(_ context.Context, id int64, t string, p any) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.events = append(f.events, fakeRepoEvent{id, t, p})
	return nil
}

func (f *fakeRepo) ListEvents(_ context.Context, accountID int64, page, size int) ([]*account.AccountEvent, int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []*account.AccountEvent
	var n int64
	for i, ev := range f.events {
		if ev.accountID == accountID {
			n++
			payload, _ := json.Marshal(ev.payload)
			out = append(out, &account.AccountEvent{
				ID:        int64(i + 1),
				AccountID: ev.accountID,
				EventType: ev.eventType,
				Payload:   payload,
				CreatedAt: time.Now().UTC(),
			})
		}
	}
	// We don't actually paginate in the fake; tests check total only.
	_ = page
	_ = size
	return out, n, nil
}

// fakeImporter 给一行 fake 文本回 1 个 *account.Account。
type fakeImporter struct {
	format string
}

func (f *fakeImporter) Detect(payload []byte) (string, bool) {
	if len(bytes.TrimSpace(payload)) == 0 {
		return "", false
	}
	return f.format, true
}

func (f *fakeImporter) Parse(payload []byte, format string) ([]*account.Account, error) {
	if format != f.format {
		return nil, nil
	}
	return []*account.Account{
		{
			Provider: "anthropic",
			Tier:     "pro_max",
			CredType: "apikey",
			Email:    "u@example.com",
			Status:   account.StatusActive,
			Weight:   100,
			Extra:    json.RawMessage("{}"),
			Cred:     account.AccountCred{APIKey: "sk-ant-api03-XXXXXXXXXX"},
		},
	}, nil
}

// fakeProbe never errors; records each Run call.
type fakeProbe struct {
	mu    sync.Mutex
	calls []int64
}

func (f *fakeProbe) Run(_ context.Context, a *account.Account) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, a.ID)
	return nil
}

// fakeRefresher records RefreshOne calls.
type fakeRefresher struct {
	mu    sync.Mutex
	calls []int64
}

func (f *fakeRefresher) Run(context.Context) error { return nil }
func (f *fakeRefresher) RefreshOne(_ context.Context, id int64) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, id)
	return nil
}
func (f *fakeRefresher) Close() error { return nil }

// fakeOAuth records Start calls and returns a prepared account from Callback.
type fakeOAuth struct {
	mu              sync.Mutex
	startProvider   string
	startChannelID  int64
	callbackState   string
	callbackCode    string
	callbackAccount *account.Account
}

func (f *fakeOAuth) Start(_ context.Context, provider string, channelID int64) (string, string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.startProvider = provider
	f.startChannelID = channelID
	return "https://oauth.example/authorize?state=st-1", "st-1", nil
}

func (f *fakeOAuth) Callback(_ context.Context, state, code string) (*account.Account, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.callbackState = state
	f.callbackCode = code
	if f.callbackAccount != nil {
		return f.callbackAccount, nil
	}
	return &account.Account{
		ChannelID:         42,
		Provider:          "openai",
		CredType:          "oauth",
		Email:             "oauth@example.com",
		Status:            account.StatusActive,
		Weight:            100,
		RefreshTokenValid: 1,
		Cred: account.AccountCred{
			AccessToken:  "at",
			RefreshToken: "rt",
		},
	}, nil
}

func (f *fakeOAuth) ExchangeRefreshToken(context.Context, string, string) (*account.AccountCred, error) {
	return nil, nil
}

// memAudit captures audit entries for assertions.
type memAudit struct {
	mu      sync.Mutex
	entries []audit.Entry
}

func (m *memAudit) Log(_ context.Context, e audit.Entry) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries = append(m.entries, e)
	return nil
}

func (m *memAudit) actions() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]string, 0, len(m.entries))
	for _, e := range m.entries {
		out = append(out, e.Action)
	}
	return out
}

// --- harness ---

func newAccountTestHarness() (*gin.Engine, *fakeRepo, *memAudit, *fakeProbe, *fakeRefresher) {
	r, repo, aud, probe, ref, _ := newAccountTestHarnessWithOAuth()
	return r, repo, aud, probe, ref
}

func newAccountTestHarnessWithOAuth() (*gin.Engine, *fakeRepo, *memAudit, *fakeProbe, *fakeRefresher, *fakeOAuth) {
	gin.SetMode(gin.TestMode)
	repo := newFakeRepo()
	imp := &fakeImporter{format: "raw_api_key"}
	probe := &fakeProbe{}
	ref := &fakeRefresher{}
	oauth := &fakeOAuth{}
	facade := &account.Facade{
		Repo:      repo,
		Importer:  imp,
		Probe:     probe,
		Refresher: ref,
		OAuth:     oauth,
	}
	aud := &memAudit{}
	h := NewAccountHandler(facade, aud, func(c *gin.Context) int64 { return 7 })
	r := gin.New()
	r.Use(middleware.ErrorResponse("json"))
	g := r.Group("/api/admin")
	h.Register(g)
	return r, repo, aud, probe, ref, oauth
}

func doAcctReq(t *testing.T, r http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

// --- tests ---

// List: pre-seed two accounts and confirm 200 + items + masked email.
func TestAccountHandler_List_OK(t *testing.T) {
	r, repo, _, _, _ := newAccountTestHarness()

	a1 := &account.Account{ChannelID: 1, Name: "n1", Provider: "anthropic", Email: "user1@example.com", Status: account.StatusActive, Weight: 100, Extra: json.RawMessage("{}")}
	a2 := &account.Account{ChannelID: 1, Name: "n2", Provider: "anthropic", Email: "user2@example.com", Status: account.StatusActive, Weight: 50, Extra: json.RawMessage("{}")}
	_ = repo.Create(context.Background(), a1)
	_ = repo.Create(context.Background(), a2)

	rec := doAcctReq(t, r, http.MethodGet, "/api/admin/accounts?channel_id=1", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var body struct {
		Data struct {
			Items []map[string]any `json:"items"`
			Total int              `json:"total"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v body=%s", err, rec.Body.String())
	}
	if body.Data.Total != 2 {
		t.Fatalf("total: want 2, got %d", body.Data.Total)
	}
	if len(body.Data.Items) != 2 {
		t.Fatalf("items: want 2, got %d", len(body.Data.Items))
	}
	// Verify email masked — no raw '@' followed by raw "example.com" with full local-part.
	for _, it := range body.Data.Items {
		email, _ := it["email"].(string)
		if email == "" || !strings.Contains(email, "***@") {
			t.Fatalf("email not masked: %q", email)
		}
		if strings.Contains(email, "user1@") || strings.Contains(email, "user2@") {
			t.Fatalf("raw local-part leaked: %q", email)
		}
	}
	// Verify no leak of credentials / Credentials field.
	raw := rec.Body.String()
	if strings.Contains(raw, "credentials") || strings.Contains(raw, "access_token") {
		t.Fatalf("List should not leak credentials, body=%s", raw)
	}
}

// Create with dry_run=true: returns preview, DB has no record.
func TestAccountHandler_Create_DryRun(t *testing.T) {
	r, repo, aud, _, _ := newAccountTestHarness()

	rec := doAcctReq(t, r, http.MethodPost, "/api/admin/accounts",
		`{"channel_id":1,"text":"sk-ant-api03-XXXXXXXXXX","dry_run":true}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var body struct {
		Data struct {
			Preview map[string]any `json:"preview"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if body.Data.Preview == nil {
		t.Fatalf("expected preview, got %s", rec.Body.String())
	}
	if len(repo.items) != 0 {
		t.Fatalf("DB should be empty after dry_run, got %d", len(repo.items))
	}
	// dry_run should NOT write audit
	for _, a := range aud.actions() {
		if a == "account.create" {
			t.Fatalf("dry_run should not audit, actions=%v", aud.actions())
		}
	}
}

// Create with dry_run=false: writes 1 row, 1 audit entry with action account.create.
func TestAccountHandler_Create_Persists(t *testing.T) {
	r, repo, aud, _, _ := newAccountTestHarness()

	rec := doAcctReq(t, r, http.MethodPost, "/api/admin/accounts",
		`{"channel_id":1,"text":"sk-ant-api03-XXXXXXXXXX","dry_run":false}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var body struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if body.Data.ID == 0 {
		t.Fatalf("expected non-zero id, got %d", body.Data.ID)
	}
	if len(repo.items) != 1 {
		t.Fatalf("want 1 row, got %d", len(repo.items))
	}
	// Verify audit
	found := false
	for _, e := range aud.entries {
		if e.Action == "account.create" {
			found = true
			if e.TargetType != "account" {
				t.Fatalf("audit TargetType=%q want account", e.TargetType)
			}
		}
	}
	if !found {
		t.Fatalf("audit account.create missing, got %v", aud.actions())
	}
}

func TestAccountHandler_OAuthStart_OK(t *testing.T) {
	r, _, _, _, _, oauth := newAccountTestHarnessWithOAuth()

	rec := doAcctReq(t, r, http.MethodPost, "/api/admin/accounts/oauth/start",
		`{"provider":"openai","channel_id":42}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var body struct {
		Data struct {
			AuthURL string `json:"auth_url"`
			State   string `json:"state"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v body=%s", err, rec.Body.String())
	}
	if body.Data.AuthURL == "" || body.Data.State != "st-1" {
		t.Fatalf("unexpected oauth start response: %+v", body.Data)
	}
	if oauth.startProvider != "openai" || oauth.startChannelID != 42 {
		t.Fatalf("oauth start inputs: provider=%q channel=%d", oauth.startProvider, oauth.startChannelID)
	}
}

func TestAccountHandler_OAuthStart_MissingChannelID(t *testing.T) {
	r, _, _, _, _ := newAccountTestHarness()

	rec := doAcctReq(t, r, http.MethodPost, "/api/admin/accounts/oauth/start",
		`{"provider":"openai"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAccountHandler_OAuthCallback_PersistsAndReturnsDoneHTML(t *testing.T) {
	r, repo, aud, _, _, oauth := newAccountTestHarnessWithOAuth()

	rec := doAcctReq(t, r, http.MethodGet, "/api/admin/accounts/oauth/callback?state=st-1&code=code-1", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "account_oauth_done") {
		t.Fatalf("callback should return popup completion HTML, got %s", rec.Body.String())
	}
	if len(repo.items) != 1 {
		t.Fatalf("want 1 persisted account, got %d", len(repo.items))
	}
	var got *account.Account
	for _, a := range repo.items {
		got = a
	}
	if got.Provider != "openai" || got.CredType != "oauth" || got.ChannelID != 42 {
		t.Fatalf("unexpected persisted account: %+v", got)
	}
	if oauth.callbackState != "st-1" || oauth.callbackCode != "code-1" {
		t.Fatalf("oauth callback inputs: state=%q code=%q", oauth.callbackState, oauth.callbackCode)
	}
	if len(repo.events) != 1 || repo.events[0].eventType != "oauth_callback" || repo.events[0].accountID != got.ID {
		t.Fatalf("oauth callback event missing: %+v", repo.events)
	}
	found := false
	for _, e := range aud.entries {
		if e.Action == "account.oauth_callback" {
			found = true
			if e.TargetID == nil || *e.TargetID != got.ID {
				t.Fatalf("TargetID mismatch: %+v", e.TargetID)
			}
		}
	}
	if !found {
		t.Fatalf("audit account.oauth_callback missing, got %v", aud.actions())
	}
}

// Patch: update name + priority; row updated; audit account.patch written.
func TestAccountHandler_Patch_OK(t *testing.T) {
	r, repo, aud, _, _ := newAccountTestHarness()
	a := &account.Account{ChannelID: 1, Name: "old", Provider: "anthropic", Status: account.StatusActive, Weight: 10, Priority: 0, Extra: json.RawMessage("{}")}
	_ = repo.Create(context.Background(), a)

	path := "/api/admin/accounts/" + intToStr(a.ID)
	rec := doAcctReq(t, r, http.MethodPatch, path, `{"name":"renamed","priority":50}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if repo.items[a.ID].Name != "renamed" {
		t.Fatalf("name not updated: %q", repo.items[a.ID].Name)
	}
	if repo.items[a.ID].Priority != 50 {
		t.Fatalf("priority not updated: %d", repo.items[a.ID].Priority)
	}
	// audit
	found := false
	for _, e := range aud.entries {
		if e.Action == "account.patch" {
			found = true
		}
	}
	if !found {
		t.Fatalf("audit account.patch missing, got %v", aud.actions())
	}
}

// PeekCredentials: returns credentials in response AND writes audit with action account.peek_credentials.
// Note: this handler test only covers handler logic. The route group must apply RoleGate(3)
// to ensure only superadmins can hit this endpoint — that is wired in cmd/proapi/main.go.
func TestAccountHandler_PeekCredentials_AuditWritten(t *testing.T) {
	r, repo, aud, _, _ := newAccountTestHarness()
	a := &account.Account{
		ChannelID: 1, Provider: "anthropic", CredType: "apikey",
		Status: account.StatusActive, Weight: 100, Extra: json.RawMessage("{}"),
		Cred: account.AccountCred{APIKey: "sk-secret"},
	}
	_ = repo.Create(context.Background(), a)

	path := "/api/admin/accounts/" + intToStr(a.ID) + "/credentials/peek"
	rec := doAcctReq(t, r, http.MethodGet, path, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "credentials") {
		t.Fatalf("expected credentials in body, got %s", rec.Body.String())
	}
	// Verify exactly one peek_credentials audit
	count := 0
	for _, e := range aud.entries {
		if e.Action == "account.peek_credentials" {
			count++
			if e.TargetType != "account" {
				t.Fatalf("TargetType=%q want account", e.TargetType)
			}
			if e.TargetID == nil || *e.TargetID != a.ID {
				t.Fatalf("TargetID mismatch: %+v", e.TargetID)
			}
		}
	}
	if count != 1 {
		t.Fatalf("want 1 peek audit, got %d (actions=%v)", count, aud.actions())
	}
}

// Stats: route /accounts/stats/overview must NOT be eaten by /accounts/:id route.
func TestAccountHandler_Stats_Overview(t *testing.T) {
	r, _, _, _, _ := newAccountTestHarness()
	rec := doAcctReq(t, r, http.MethodGet, "/api/admin/accounts/stats/overview", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "by_provider") {
		t.Fatalf("expected by_provider in body, got %s", rec.Body.String())
	}
}

// multiFakeImporter 返回 n 个 apikey 账号,provider 交替 openai/anthropic,
// 用于验证 raw_apikey 批量导入时 provider 被渠道覆盖 + tag 落到每条。
type multiFakeImporter struct {
	format string
	n      int
}

func (m *multiFakeImporter) Detect(payload []byte) (string, bool) {
	if len(bytes.TrimSpace(payload)) == 0 {
		return "", false
	}
	return m.format, true
}

func (m *multiFakeImporter) Parse(_ []byte, format string) ([]*account.Account, error) {
	if format != m.format {
		return nil, nil
	}
	out := make([]*account.Account, 0, m.n)
	for i := 0; i < m.n; i++ {
		p := "anthropic"
		if i%2 == 0 {
			p = "openai"
		}
		out = append(out, &account.Account{
			Provider: p, CredType: "apikey", Status: account.StatusActive, Weight: 100,
			Cred: account.AccountCred{APIKey: "sk-" + jsonInt(int64(i))},
		})
	}
	return out, nil
}

// Import: tag 落到每条账号;raw_apikey 格式下 provider 被渠道 provider 覆盖。
func TestAccountHandler_Import_TagAndProviderFollowsChannel(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := newFakeRepo()
	facade := &account.Facade{Repo: repo, Importer: &multiFakeImporter{format: "raw_apikey", n: 2}}
	aud := &memAudit{}
	h := NewAccountHandler(facade, aud, func(c *gin.Context) int64 { return 7 })
	h.ChannelProvider = func(id int64) string { return "deepseek" }
	r := gin.New()
	r.Use(middleware.ErrorResponse("json"))
	h.Register(r.Group("/api/admin"))

	rec := doAcctReq(t, r, http.MethodPost, "/api/admin/accounts/import",
		`{"channel_id":1,"format":"apikey","tag":"k12-batch1","text":"sk-1\nsk-2"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if len(repo.items) != 2 {
		t.Fatalf("want 2 rows, got %d", len(repo.items))
	}
	for _, a := range repo.items {
		if a.Tag != "k12-batch1" {
			t.Fatalf("tag not applied: %q", a.Tag)
		}
		if a.Provider != "deepseek" {
			t.Fatalf("provider should follow channel, got %q", a.Provider)
		}
	}
}

// List: tag 查询参数按子串过滤,凭证不泄露。
func TestAccountHandler_List_TagFilter(t *testing.T) {
	r, repo, _, _, _ := newAccountTestHarness()
	_ = repo.Create(context.Background(), &account.Account{ChannelID: 1, Name: "a", Tag: "k12第一批", Provider: "anthropic", Status: account.StatusActive, Weight: 100, Extra: json.RawMessage("{}")})
	_ = repo.Create(context.Background(), &account.Account{ChannelID: 1, Name: "b", Tag: "vip", Provider: "anthropic", Status: account.StatusActive, Weight: 100, Extra: json.RawMessage("{}")})

	rec := doAcctReq(t, r, http.MethodGet, "/api/admin/accounts?channel_id=1&tag=k12", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var body struct {
		Data struct {
			Items []map[string]any `json:"items"`
			Total int              `json:"total"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if body.Data.Total != 1 {
		t.Fatalf("tag filter: want 1, got %d", body.Data.Total)
	}
	if tag, _ := body.Data.Items[0]["tag"].(string); tag != "k12第一批" {
		t.Fatalf("unexpected tag in result: %q", tag)
	}
}

// --- small helpers ---

func intToStr(i int64) string {
	// strconv.FormatInt without importing strconv here
	return jsonInt(i)
}

func jsonInt(i int64) string {
	b, _ := json.Marshal(i)
	return string(b)
}
