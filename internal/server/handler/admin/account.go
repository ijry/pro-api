package admin

// AccountHandler 渠道账号 admin REST API。
//
// 路由组(由 cmd/proapi/main.go 装配):
//   - 所有 endpoint 都在 SessionAuth + RoleGate(2)(admin)之后;
//   - 仅 /accounts/:id/credentials/peek 额外加 RoleGate(3)(superadmin)。
//
// 错误响应:走项目统一的 middleware.SetErr + ErrorResponse("json") 中间件。
// 成功响应:gin.H{"data": ...}。
//
// 审计:所有 mutation 都通过 audit.Logger.Log 写一条 Entry,
//      Action 形如 "account.create" / "account.patch" / "account.delete" 等。
//      peek_credentials 是唯一的查询型审计,因为它返回明文凭证。

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ijry/pro-api/internal/account"
	accountoauth "github.com/ijry/pro-api/internal/account/oauth"
	"github.com/ijry/pro-api/internal/audit"
	"github.com/ijry/pro-api/internal/server/middleware"
	"github.com/ijry/pro-api/pkg/apierr"
)

// AccountHandler 是号池 admin handler。
//
// Facade 已聚合 Repo / Selector / Breaker / Refresher / Probe / Importer。
type AccountHandler struct {
	Facade  *account.Facade
	Audit   audit.Logger
	ActorOf func(c *gin.Context) int64
}

// NewAccountHandler 构造。actorOf == nil 时用 0 占位;audit == nil 时用 noop。
func NewAccountHandler(f *account.Facade, a audit.Logger, actorOf func(*gin.Context) int64) *AccountHandler {
	if actorOf == nil {
		actorOf = func(*gin.Context) int64 { return 0 }
	}
	if a == nil {
		a = audit.NewNoop()
	}
	return &AccountHandler{Facade: f, Audit: a, ActorOf: actorOf}
}

// Register 在 r 下挂 14 个 endpoint。/credentials/peek 的 RoleGate(3) 由调用方在路由组层面加。
func (h *AccountHandler) Register(r gin.IRouter) {
	r.GET("/accounts", h.List)
	r.GET("/accounts/stats/overview", h.Stats)
	r.POST("/accounts/oauth/start", h.OAuthStart)
	r.GET("/accounts/oauth/callback", h.OAuthCallback)
	r.GET("/accounts/:id", h.Get)
	r.POST("/accounts", h.Create)
	r.POST("/accounts/import", h.Import)
	r.PATCH("/accounts/:id", h.Patch)
	r.DELETE("/accounts/:id", h.Delete)
	r.POST("/accounts/:id/enable", h.Enable)
	r.POST("/accounts/:id/disable", h.Disable)
	r.POST("/accounts/:id/test", h.Test)
	r.POST("/accounts/:id/refresh_token", h.RefreshToken)
	r.POST("/accounts/:id/clear_cooldown", h.ClearCooldown)
	r.GET("/accounts/:id/events", h.Events)
	r.GET("/accounts/:id/credentials/peek", h.PeekCredentials)
}

// --- helpers ---

// parseIDParam 解 :id;失败时已 SetErr。
func parseIDParam(c *gin.Context) (int64, bool) {
	s := c.Param("id")
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil || id <= 0 {
		middleware.SetErr(c, apierr.New(apierr.CodeInvalidParam, "id 必须为正整数"))
		return 0, false
	}
	return id, true
}

// writeAcctErr 把 service/repo 返回的 error 转 apierr 写到 ctx。
// account.ErrNotFound / CodeAccountNotFound 走 CodeAccountNotFound。
func writeAcctErr(c *gin.Context, err error) {
	var ae *apierr.Error
	if asAPIErr(err, &ae) {
		middleware.SetErr(c, ae)
		return
	}
	middleware.SetErr(c, apierr.Wrap(apierr.CodeInternal, "internal error", err))
}

// auditOne 写一条 audit;ActorID 取自 ctx,IP 取自 c.ClientIP()。
func (h *AccountHandler) auditOne(c *gin.Context, action string, targetID int64, before, after any) {
	actor := h.ActorOf(c)
	tid := targetID
	entry := audit.Entry{
		Action:     action,
		TargetType: "account",
		TargetID:   &tid,
		IP:         clientIP(c),
	}
	if actor != 0 {
		entry.ActorID = &actor
	}
	if before != nil {
		if b, err := json.Marshal(before); err == nil {
			entry.Before = b
		}
	}
	if after != nil {
		if b, err := json.Marshal(after); err == nil {
			entry.After = b
		}
	}
	_ = h.Audit.Log(c.Request.Context(), entry)
}

// --- list / detail ---

// listItem 是 List 返回的 item。绝不包含明文凭证或 Credentials 密文。
type listItem struct {
	ID                   int64      `json:"id"`
	ChannelID            int64      `json:"channel_id"`
	Name                 string     `json:"name"`
	Provider             string     `json:"provider"`
	Tier                 string     `json:"tier"`
	CredType             string     `json:"cred_type"`
	Email                string     `json:"email"`
	Status               int8       `json:"status"`
	Priority             int16      `json:"priority"`
	Weight               int        `json:"weight"`
	ConsecFailures       int        `json:"consec_failures"`
	RefreshTokenValid    int8       `json:"refresh_token_valid"`
	CooldownUntil        *time.Time `json:"cooldown_until,omitempty"`
	LastUsedAt           *time.Time `json:"last_used_at,omitempty"`
	LastSuccessAt        *time.Time `json:"last_success_at,omitempty"`
	LastFailureAt        *time.Time `json:"last_failure_at,omitempty"`
	AccessTokenExpiresAt *time.Time `json:"access_token_expires_at,omitempty"`
	Quota5h              quotaItem  `json:"quota_5h"`
	QuotaWeek            quotaItem  `json:"quota_week"`
}

// detailItem 在 listItem 之外再加几个详情字段。同样不含明文凭证。
type detailItem struct {
	listItem
	ImportSource      string    `json:"import_source"`
	ExternalAccountID string    `json:"external_account_id"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type quotaItem struct {
	Remaining *int64     `json:"remaining"`
	Total     *int64     `json:"total"`
	ResetAt   *time.Time `json:"reset_at"`
	Pct       *float64   `json:"pct,omitempty"`
}

func accountToListItem(a *account.Account) listItem {
	return listItem{
		ID:                   a.ID,
		ChannelID:            a.ChannelID,
		Name:                 a.Name,
		Provider:             a.Provider,
		Tier:                 a.Tier,
		CredType:             a.CredType,
		Email:                account.MaskEmail(a.Email),
		Status:               int8(a.Status),
		Priority:             a.Priority,
		Weight:               a.Weight,
		ConsecFailures:       a.ConsecFailures,
		RefreshTokenValid:    a.RefreshTokenValid,
		CooldownUntil:        a.CooldownUntil,
		LastUsedAt:           a.LastUsedAt,
		LastSuccessAt:        a.LastSuccessAt,
		LastFailureAt:        a.LastFailureAt,
		AccessTokenExpiresAt: a.AccessTokenExpiresAt,
		Quota5h:              makeQuota(a.Quota5hRemaining, a.Quota5hTotal, a.Quota5hResetAt),
		QuotaWeek:            makeQuota(a.QuotaWeekRemaining, a.QuotaWeekTotal, a.QuotaWeekResetAt),
	}
}

func accountToDetail(a *account.Account) detailItem {
	d := detailItem{
		listItem:          accountToListItem(a),
		ImportSource:      a.ImportSource,
		ExternalAccountID: a.ExternalAccountID,
		CreatedAt:         a.CreatedAt,
		UpdatedAt:         a.UpdatedAt,
	}
	return d
}

func makeQuota(rem, tot *int64, reset *time.Time) quotaItem {
	q := quotaItem{Remaining: rem, Total: tot, ResetAt: reset}
	if rem != nil && tot != nil && *tot > 0 {
		p := float64(*rem) / float64(*tot)
		q.Pct = &p
	}
	return q
}

// List GET /accounts?channel_id=N&status=S
func (h *AccountHandler) List(c *gin.Context) {
	chStr := c.Query("channel_id")
	statusStr := c.Query("status")

	var list []*account.Account
	if chStr == "" {
		middleware.SetErr(c, apierr.New(apierr.CodeMissingParam, "channel_id 必填"))
		return
	}
	chID, err := strconv.ParseInt(chStr, 10, 64)
	if err != nil || chID <= 0 {
		middleware.SetErr(c, apierr.New(apierr.CodeInvalidParam, "channel_id 必须为正整数"))
		return
	}
	list, err = h.Facade.Repo.ListByChannel(c.Request.Context(), chID)
	if err != nil {
		writeAcctErr(c, err)
		return
	}

	if statusStr != "" {
		want, err := strconv.Atoi(statusStr)
		if err != nil {
			middleware.SetErr(c, apierr.New(apierr.CodeInvalidParam, "status 必须为整数"))
			return
		}
		filtered := list[:0]
		for _, a := range list {
			if int(a.Status) == want {
				filtered = append(filtered, a)
			}
		}
		list = filtered
	}

	items := make([]listItem, 0, len(list))
	for _, a := range list {
		items = append(items, accountToListItem(a))
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{
		"items": items,
		"total": len(items),
	}})
}

// Get GET /accounts/:id
func (h *AccountHandler) Get(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	a, err := h.Facade.Repo.Get(c.Request.Context(), id)
	if err != nil {
		writeAcctErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": accountToDetail(a)})
}

// --- create / import ---

// acctCreateReq 是 POST /accounts 与 POST /accounts/import 的 JSON body。
type acctCreateReq struct {
	ChannelID int64  `json:"channel_id"`
	Name      string `json:"name"`
	Provider  string `json:"provider"`
	Tier      string `json:"tier"`
	Format    string `json:"format"`
	Text      string `json:"text"`
	DryRun    bool   `json:"dry_run"`
	Priority  int16  `json:"priority"`
	Weight    int    `json:"weight"`
}

type acctOAuthStartReq struct {
	Provider  string `json:"provider"`
	ChannelID int64  `json:"channel_id"`
}

// Create POST /accounts — 单条创建,与 Import 类似但只取第一条。dry_run 时不落库。
func (h *AccountHandler) Create(c *gin.Context) {
	var req acctCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.SetErr(c, apierr.New(apierr.CodeInvalidParam, "请求体不合法"))
		return
	}
	if req.ChannelID <= 0 {
		middleware.SetErr(c, apierr.New(apierr.CodeMissingParam, "channel_id 必填"))
		return
	}
	if req.Text == "" {
		middleware.SetErr(c, apierr.New(apierr.CodeMissingParam, "text 必填"))
		return
	}
	format := req.Format
	if format == "" {
		f, ok := h.Facade.Importer.Detect([]byte(req.Text))
		if !ok {
			middleware.SetErr(c, apierr.New(apierr.CodeAccountImportFormat, "cannot detect format"))
			return
		}
		format = f
	}
	list, err := h.Facade.Importer.Parse([]byte(req.Text), format)
	if err != nil {
		writeAcctErr(c, err)
		return
	}
	if len(list) == 0 {
		middleware.SetErr(c, apierr.New(apierr.CodeAccountImportFields, "empty parse result"))
		return
	}
	a := list[0]
	a.ChannelID = req.ChannelID
	if req.Name != "" {
		a.Name = req.Name
	}
	if req.Provider != "" {
		a.Provider = req.Provider
	}
	if req.Tier != "" {
		a.Tier = req.Tier
	}
	if req.Priority != 0 {
		a.Priority = req.Priority
	}
	if req.Weight != 0 {
		a.Weight = req.Weight
	}
	if req.DryRun {
		c.JSON(http.StatusOK, gin.H{"data": gin.H{"preview": accountToListItem(a)}})
		return
	}
	if err := h.Facade.Repo.Create(c.Request.Context(), a); err != nil {
		writeAcctErr(c, err)
		return
	}
	_ = h.Facade.Repo.AppendEvent(c.Request.Context(), a.ID, "imported",
		map[string]any{"by": h.ActorOf(c), "format": format})
	// Fire-and-forget probe; ignore errors since create response doesn't block on probe.
	if h.Facade.Probe != nil {
		cc := c.Copy()
		go func(acc *account.Account) {
			_ = h.Facade.Probe.Run(cc.Request.Context(), acc)
		}(a)
	}
	h.auditOne(c, "account.create", a.ID, nil, map[string]any{
		"id":         a.ID,
		"channel_id": a.ChannelID,
		"provider":   a.Provider,
		"format":     format,
	})
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"id": a.ID}})
}

// OAuthStart POST /accounts/oauth/start — 启动账号池 OAuth PKCE 流程。
func (h *AccountHandler) OAuthStart(c *gin.Context) {
	if h.Facade.OAuth == nil {
		middleware.SetErr(c, apierr.New(apierr.CodeInternal, "oauth 未就绪"))
		return
	}
	var req acctOAuthStartReq
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.SetErr(c, apierr.New(apierr.CodeInvalidParam, "请求体不合法"))
		return
	}
	if req.Provider == "" {
		middleware.SetErr(c, apierr.New(apierr.CodeMissingParam, "provider 必填"))
		return
	}
	if req.ChannelID <= 0 {
		middleware.SetErr(c, apierr.New(apierr.CodeMissingParam, "channel_id 必填"))
		return
	}
	authURL, state, err := h.Facade.OAuth.Start(c.Request.Context(), req.Provider, req.ChannelID)
	if err != nil {
		middleware.SetErr(c, apierr.Wrap(apierr.CodeAccountOAuthRejected, "oauth start failed", err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"auth_url": authURL, "state": state}})
}

// OAuthCallback GET /accounts/oauth/callback — OAuth provider 回调入口(no-auth,state 校验)。
func (h *AccountHandler) OAuthCallback(c *gin.Context) {
	if h.Facade.OAuth == nil {
		middleware.SetErr(c, apierr.New(apierr.CodeInternal, "oauth 未就绪"))
		return
	}
	state := c.Query("state")
	code := c.Query("code")
	if state == "" {
		middleware.SetErr(c, apierr.New(apierr.CodeMissingParam, "state 必填"))
		return
	}
	if code == "" {
		middleware.SetErr(c, apierr.New(apierr.CodeMissingParam, "code 必填"))
		return
	}
	a, err := h.Facade.OAuth.Callback(c.Request.Context(), state, code)
	if err != nil {
		if errors.Is(err, accountoauth.ErrStateNotFound) {
			middleware.SetErr(c, apierr.Wrap(apierr.CodeAccountOAuthState, "oauth state invalid", err))
			return
		}
		middleware.SetErr(c, apierr.Wrap(apierr.CodeAccountOAuthRejected, "oauth callback failed", err))
		return
	}
	if err := h.Facade.Repo.Create(c.Request.Context(), a); err != nil {
		writeAcctErr(c, err)
		return
	}
	_ = h.Facade.Repo.AppendEvent(c.Request.Context(), a.ID, "oauth_callback",
		map[string]any{"provider": a.Provider, "channel_id": a.ChannelID})
	if h.Facade.Probe != nil {
		cc := c.Copy()
		go func(acc *account.Account) {
			_ = h.Facade.Probe.Run(cc.Request.Context(), acc)
		}(a)
	}
	h.auditOne(c, "account.oauth_callback", a.ID, nil, map[string]any{
		"id":         a.ID,
		"channel_id": a.ChannelID,
		"provider":   a.Provider,
	})
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(accountOAuthDoneHTML))
}

const accountOAuthDoneHTML = `<!doctype html>
<html><head><meta charset="utf-8"><title>OAuth complete</title></head>
<body>
<script>
if (window.opener) {
  window.opener.postMessage({ type: "account_oauth_done" }, window.location.origin);
}
window.close();
</script>
OAuth complete. You can close this window.
</body></html>`

// Import POST /accounts/import — 批量导入。支持 JSON body 或 multipart file。
func (h *AccountHandler) Import(c *gin.Context) {
	var req acctCreateReq
	text := ""

	// Multipart 优先,fall back to JSON body。
	ct := c.ContentType()
	if ct == "multipart/form-data" {
		req.ChannelID, _ = strconv.ParseInt(c.PostForm("channel_id"), 10, 64)
		req.Format = c.PostForm("format")
		req.DryRun = c.PostForm("dry_run") == "true"
		fh, err := c.FormFile("file")
		if err == nil && fh != nil {
			f, err := fh.Open()
			if err == nil {
				defer func() { _ = f.Close() }()
				buf, _ := io.ReadAll(f)
				text = string(buf)
			}
		}
		if text == "" {
			text = c.PostForm("text")
		}
	} else {
		if err := c.ShouldBindJSON(&req); err != nil {
			middleware.SetErr(c, apierr.New(apierr.CodeInvalidParam, "请求体不合法"))
			return
		}
		text = req.Text
	}

	if req.ChannelID <= 0 {
		middleware.SetErr(c, apierr.New(apierr.CodeMissingParam, "channel_id 必填"))
		return
	}
	if text == "" {
		middleware.SetErr(c, apierr.New(apierr.CodeMissingParam, "text 或 file 必填"))
		return
	}
	format := req.Format
	if format == "" {
		f, ok := h.Facade.Importer.Detect([]byte(text))
		if !ok {
			middleware.SetErr(c, apierr.New(apierr.CodeAccountImportFormat, "cannot detect format"))
			return
		}
		format = f
	}
	list, err := h.Facade.Importer.Parse([]byte(text), format)
	if err != nil {
		writeAcctErr(c, err)
		return
	}

	previews := make([]listItem, 0, len(list))
	imported := make([]int64, 0, len(list))
	errs := []gin.H{}

	for i, a := range list {
		a.ChannelID = req.ChannelID
		if req.DryRun {
			previews = append(previews, accountToListItem(a))
			continue
		}
		if err := h.Facade.Repo.Create(c.Request.Context(), a); err != nil {
			errs = append(errs, gin.H{"index": i, "reason": err.Error()})
			continue
		}
		_ = h.Facade.Repo.AppendEvent(c.Request.Context(), a.ID, "imported",
			map[string]any{"by": h.ActorOf(c), "format": format})
		imported = append(imported, a.ID)
	}

	if req.DryRun {
		c.JSON(http.StatusOK, gin.H{"data": gin.H{
			"total":   len(list),
			"preview": previews,
		}})
		return
	}

	if len(imported) > 0 {
		h.auditOne(c, "account.import", 0, nil, map[string]any{
			"channel_id": req.ChannelID,
			"count":      len(imported),
			"format":     format,
		})
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{
		"total":       len(list),
		"imported":    len(imported),
		"errors":      errs,
		"account_ids": imported,
	}})
}

// --- patch / delete / state ops ---

// acctPatchReq 是 PATCH body。pointer 字段用于区分 "未设" 与 "显式置零"。
type acctPatchReq struct {
	Name     *string `json:"name"`
	Priority *int16  `json:"priority"`
	Weight   *int    `json:"weight"`
	Status   *int8   `json:"status"`
}

// Patch PATCH /accounts/:id
func (h *AccountHandler) Patch(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var p acctPatchReq
	if err := c.ShouldBindJSON(&p); err != nil {
		middleware.SetErr(c, apierr.New(apierr.CodeInvalidParam, "请求体不合法"))
		return
	}
	cur, err := h.Facade.Repo.Get(c.Request.Context(), id)
	if err != nil {
		writeAcctErr(c, err)
		return
	}
	before := accountToListItem(cur)

	if p.Name != nil {
		cur.Name = *p.Name
	}
	if p.Priority != nil {
		cur.Priority = *p.Priority
	}
	if p.Weight != nil {
		cur.Weight = *p.Weight
	}
	if p.Status != nil {
		cur.Status = account.Status(*p.Status)
	}
	// Patch 不动 Cred:Get 已把明文 hydrate 到 cur.Cred,
	// Repo.Update 会按 hydrate 出的值原样重新加密,等价于"不改凭证"。
	if err := h.Facade.Repo.Update(c.Request.Context(), cur); err != nil {
		writeAcctErr(c, err)
		return
	}
	after := accountToListItem(cur)
	h.auditOne(c, "account.patch", id, before, after)
	c.JSON(http.StatusOK, gin.H{"data": accountToDetail(cur)})
}

// Delete DELETE /accounts/:id (soft delete)
func (h *AccountHandler) Delete(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	if err := h.Facade.Repo.Delete(c.Request.Context(), id); err != nil {
		writeAcctErr(c, err)
		return
	}
	h.auditOne(c, "account.delete", id, nil, nil)
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"id": id}})
}

// Enable POST /accounts/:id/enable — 把账号置为 active 并清失败计数。
func (h *AccountHandler) Enable(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	if _, err := h.Facade.Repo.Get(c.Request.Context(), id); err != nil {
		writeAcctErr(c, err)
		return
	}
	// Reactivate 用 map-based 更新,绕过 GORM 零值跳过。
	if err := h.Facade.Repo.Reactivate(c.Request.Context(), id); err != nil {
		writeAcctErr(c, err)
		return
	}
	h.auditOne(c, "account.enable", id, nil, nil)
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"id": id, "status": int(account.StatusActive)}})
}

// Disable POST /accounts/:id/disable — 把账号置为 disabled。
func (h *AccountHandler) Disable(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	cur, err := h.Facade.Repo.Get(c.Request.Context(), id)
	if err != nil {
		writeAcctErr(c, err)
		return
	}
	cur.Status = account.StatusDisabled
	if err := h.Facade.Repo.Update(c.Request.Context(), cur); err != nil {
		writeAcctErr(c, err)
		return
	}
	h.auditOne(c, "account.disable", id, nil, nil)
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"id": id, "status": int(account.StatusDisabled)}})
}

// ClearCooldown POST /accounts/:id/clear_cooldown — 把 cooldown 账号恢复 active。
func (h *AccountHandler) ClearCooldown(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	if _, err := h.Facade.Repo.Get(c.Request.Context(), id); err != nil {
		writeAcctErr(c, err)
		return
	}
	if err := h.Facade.Repo.Reactivate(c.Request.Context(), id); err != nil {
		writeAcctErr(c, err)
		return
	}
	h.auditOne(c, "account.clear_cooldown", id, nil, nil)
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"id": id, "status": int(account.StatusActive)}})
}

// --- probe / refresh / events / peek / stats ---

// Test POST /accounts/:id/test — 调用 Probe.Run。
func (h *AccountHandler) Test(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	a, err := h.Facade.Repo.Get(c.Request.Context(), id)
	if err != nil {
		writeAcctErr(c, err)
		return
	}
	if h.Facade.Probe == nil {
		middleware.SetErr(c, apierr.New(apierr.CodeInternal, "probe 未就绪"))
		return
	}
	if err := h.Facade.Probe.Run(c.Request.Context(), a); err != nil {
		writeAcctErr(c, err)
		return
	}
	// Re-read so the response reflects probe-side updates (quota / last_success_at)
	a2, err := h.Facade.Repo.Get(c.Request.Context(), id)
	if err != nil {
		writeAcctErr(c, err)
		return
	}
	h.auditOne(c, "account.test", id, nil, nil)
	c.JSON(http.StatusOK, gin.H{"data": accountToDetail(a2)})
}

// RefreshToken POST /accounts/:id/refresh_token — 调用 Refresher.RefreshOne。
func (h *AccountHandler) RefreshToken(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	if h.Facade.Refresher == nil {
		middleware.SetErr(c, apierr.New(apierr.CodeInternal, "refresher 未就绪"))
		return
	}
	if err := h.Facade.Refresher.RefreshOne(c.Request.Context(), id); err != nil {
		writeAcctErr(c, err)
		return
	}
	h.auditOne(c, "account.refresh_token", id, nil, nil)
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"id": id}})
}

// Events GET /accounts/:id/events?page=&size=
func (h *AccountHandler) Events(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	items, total, err := h.Facade.Repo.ListEvents(c.Request.Context(), id, page, size)
	if err != nil {
		writeAcctErr(c, err)
		return
	}
	if items == nil {
		items = []*account.AccountEvent{}
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{
		"items": items,
		"total": total,
		"page":  page,
		"size":  size,
	}})
}

// PeekCredentials GET /accounts/:id/credentials/peek — 返回明文凭证。
// 路由组必须挂 RoleGate(3)(superadmin)。每次调用强制写一条 account.peek_credentials 审计。
func (h *AccountHandler) PeekCredentials(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	a, err := h.Facade.Repo.Get(c.Request.Context(), id)
	if err != nil {
		writeAcctErr(c, err)
		return
	}
	// MUST audit BEFORE returning so any panic/handler-error leaves a trail.
	h.auditOne(c, "account.peek_credentials", id, nil, map[string]any{
		"id":        id,
		"provider":  a.Provider,
		"cred_type": a.CredType,
	})
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"credentials": a.Cred}})
}

// Stats GET /accounts/stats/overview — M1 范围内返回 zero-shape;后续 M3 可加聚合。
func (h *AccountHandler) Stats(c *gin.Context) {
	// TODO(M3): aggregate by provider/status from DB (SELECT provider, status, COUNT(*) GROUP BY ...).
	c.JSON(http.StatusOK, gin.H{"data": gin.H{
		"totals":      gin.H{},
		"by_provider": gin.H{},
	}})
}
