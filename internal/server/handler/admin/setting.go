package admin

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ijry/pro-api/internal/audit"
	"github.com/ijry/pro-api/internal/server/middleware"
	"github.com/ijry/pro-api/internal/setting"
	"github.com/ijry/pro-api/pkg/apierr"
)

// Encryptor 是 setting handler 加密的最小依赖。
type Encryptor interface {
	Encrypt(plain string) (string, error)
}

// Mailer 是 SMTP 测试的可选依赖;若 nil 走 stub。
type Mailer interface {
	SendTestMail(cfg SMTPConfig, to string) error
}

// SMTPConfig 是 SMTP 测试 body 解析后的配置。
type SMTPConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	UseTLS   bool   `json:"use_tls"`
}

// SettingHandler 是系统设置 admin handler。
type SettingHandler struct {
	Store   setting.Store
	Crypto  Encryptor
	Mailer  Mailer
	Audit   audit.Logger
	ActorOf func(c *gin.Context) int64
}

// NewSettingHandler 构造。
func NewSettingHandler(store setting.Store, crypto Encryptor, mailer Mailer, audLog audit.Logger, actorOf func(*gin.Context) int64) *SettingHandler {
	if actorOf == nil {
		actorOf = func(*gin.Context) int64 { return 0 }
	}
	if audLog == nil {
		audLog = audit.NewNoop()
	}
	return &SettingHandler{Store: store, Crypto: crypto, Mailer: mailer, Audit: audLog, ActorOf: actorOf}
}

// Register 把 4 个 admin 路由挂到 r(预期已加 SessionAuth + RoleGate(admin))。
func (h *SettingHandler) Register(r gin.IRouter) {
	r.GET("/settings", h.List)
	r.GET("/settings/:key", h.Get)
	r.PATCH("/settings/:key", h.Patch)
	r.POST("/settings/test_smtp", h.TestSMTP)
}

// settingItem 是 list/get 返回的 item 结构。
type settingItem struct {
	Key         string          `json:"key"`
	Value       json.RawMessage `json:"value"`
	Description string          `json:"description"`
	IsSensitive bool            `json:"is_sensitive"`
	UpdatedAt   string          `json:"updated_at"`
}

// groupOut 是 list 返回的分组结构。
type groupOut struct {
	Name  string        `json:"name"`
	Label string        `json:"label"`
	Items []settingItem `json:"items"`
}

// List GET /settings — 全部 KV 按分组返回,敏感字段脱敏。
func (h *SettingHandler) List(c *gin.Context) {
	rows, err := h.Store.ListAll(c.Request.Context())
	if err != nil {
		middleware.SetErr(c, apierr.Wrap(apierr.CodeDatabase, "list settings failed", err))
		return
	}
	buckets := map[setting.Group][]settingItem{}
	for _, row := range rows {
		item := toSettingItem(row)
		buckets[setting.GroupOf(row.Key)] = append(buckets[setting.GroupOf(row.Key)], item)
	}
	out := make([]groupOut, 0, len(setting.AllGroups()))
	for _, g := range setting.AllGroups() {
		items := buckets[g]
		if items == nil {
			items = []settingItem{}
		}
		out = append(out, groupOut{
			Name:  string(g),
			Label: setting.GroupLabel(g),
			Items: items,
		})
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"groups": out}})
}

// Get GET /settings/:key — 单查。
func (h *SettingHandler) Get(c *gin.Context) {
	key := c.Param("key")
	raw, ok := h.Store.Get(c.Request.Context(), key)
	if !ok {
		middleware.SetErr(c, apierr.New(apierr.CodeNotFound, "未知配置项"))
		return
	}
	row := setting.Setting{Key: key, Value: raw}
	item := toSettingItem(row)
	c.JSON(http.StatusOK, gin.H{"data": item})
}

// patchReq 是 PATCH body;value 是 raw JSON 以保留 null / bool / string / number / object 区分。
type patchReq struct {
	Value *json.RawMessage `json:"value"`
}

// Patch PATCH /settings/:key — 修改值。
func (h *SettingHandler) Patch(c *gin.Context) {
	key := c.Param("key")
	var req patchReq
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.SetErr(c, apierr.New(apierr.CodeInvalidParam, "请求体不合法"))
		return
	}
	if req.Value == nil {
		middleware.SetErr(c, apierr.New(apierr.CodeMissingParam, "value 必填"))
		return
	}
	// key 必须已存在
	old, ok := h.Store.Get(c.Request.Context(), key)
	if !ok {
		middleware.SetErr(c, apierr.New(apierr.CodeNotFound, "未知配置项,如需新增请联系开发"))
		return
	}
	raw := []byte(*req.Value)
	// 占位 -> 不更新
	if setting.IsPlaceholderValue(raw) {
		row := setting.Setting{Key: key, Value: old}
		c.JSON(http.StatusOK, gin.H{"data": toSettingItem(row)})
		return
	}
	// 敏感字段:value 必须是 JSON string
	isSensitive := setting.IsSensitive(key)
	finalRaw := raw
	if isSensitive {
		var plain string
		if err := json.Unmarshal(raw, &plain); err != nil {
			middleware.SetErr(c, apierr.New(apierr.CodeInvalidParam, "敏感字段值必须为字符串"))
			return
		}
		if h.Crypto == nil {
			middleware.SetErr(c, apierr.New(apierr.CodeInternal, "Crypto 未注入,无法加密敏感字段"))
			return
		}
		ct, err := h.Crypto.Encrypt(plain)
		if err != nil {
			middleware.SetErr(c, apierr.Wrap(apierr.CodeInternal, "encrypt failed", err))
			return
		}
		ctRaw, err := json.Marshal(ct)
		if err != nil {
			middleware.SetErr(c, apierr.Wrap(apierr.CodeInternal, "marshal ciphertext failed", err))
			return
		}
		finalRaw = ctRaw
	}
	// 写库:Store.Put 接受 any 值;此处用 json.RawMessage(*) 保证最终原样写入(避免双重编码)
	if err := h.Store.Put(c.Request.Context(), key, json.RawMessage(finalRaw), h.ActorOf(c)); err != nil {
		middleware.SetErr(c, apierr.Wrap(apierr.CodeDatabase, "put setting failed", err))
		return
	}
	// 审计(脱敏 before/after)
	h.auditUpdate(c, key, old, finalRaw)
	// 响应也走脱敏(若 sensitive,value 替换为占位)
	row := setting.Setting{Key: key, Value: finalRaw}
	c.JSON(http.StatusOK, gin.H{"data": toSettingItem(row)})
}

// smtpReq 是 test_smtp body。所有字段可选 — 缺失时从 setting 拿默认。
type smtpReq struct {
	Host     *string `json:"host"`
	Port     *int    `json:"port"`
	Username *string `json:"username"`
	Password *string `json:"password"`
	UseTLS   *bool   `json:"use_tls"`
	To       string  `json:"to"`
}

// smtpResp 是 test_smtp 响应。
type smtpResp struct {
	OK      bool   `json:"ok"`
	Stubbed bool   `json:"stubbed,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// TestSMTP POST /settings/test_smtp — 校验配置 + 可选真发邮件。
func (h *SettingHandler) TestSMTP(c *gin.Context) {
	var req smtpReq
	if err := c.ShouldBindJSON(&req); err != nil {
		// 允许空 body;返回 400 仅当格式错误
		if err.Error() != "EOF" {
			middleware.SetErr(c, apierr.New(apierr.CodeInvalidParam, "请求体不合法"))
			return
		}
	}
	// 校验字段
	cfg := SMTPConfig{}
	if req.Host != nil {
		cfg.Host = *req.Host
	}
	if cfg.Host == "" {
		middleware.SetErr(c, apierr.New(apierr.CodeMissingParam, "host 必填"))
		return
	}
	if req.Port != nil {
		cfg.Port = *req.Port
	}
	if cfg.Port < 1 || cfg.Port > 65535 {
		middleware.SetErr(c, apierr.New(apierr.CodeInvalidParam, "port 必须在 [1, 65535]"))
		return
	}
	if req.Username != nil {
		cfg.Username = *req.Username
	}
	if cfg.Username == "" {
		middleware.SetErr(c, apierr.New(apierr.CodeMissingParam, "username 必填"))
		return
	}
	if req.Password != nil {
		cfg.Password = *req.Password
	}
	if cfg.Password == "" {
		middleware.SetErr(c, apierr.New(apierr.CodeMissingParam, "password 必填"))
		return
	}
	if req.UseTLS != nil {
		cfg.UseTLS = *req.UseTLS
	}
	if req.To == "" {
		middleware.SetErr(c, apierr.New(apierr.CodeMissingParam, "to 必填"))
		return
	}
	if !isLikelyEmail(req.To) {
		middleware.SetErr(c, apierr.New(apierr.CodeInvalidParam, "to 不是合法 email"))
		return
	}
	// Mailer 未注入 → stub
	if h.Mailer == nil {
		c.JSON(http.StatusOK, gin.H{"data": smtpResp{OK: true, Stubbed: true, Message: "Mailer 未就绪,仅校验配置格式"}})
		return
	}
	if err := h.Mailer.SendTestMail(cfg, req.To); err != nil {
		c.JSON(http.StatusOK, gin.H{"data": smtpResp{OK: false, Error: err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": smtpResp{OK: true}})
}

// --- helpers ---

// toSettingItem 把 store 行转 settingItem;敏感字段 / ENC 值统一脱敏。
func toSettingItem(row setting.Setting) settingItem {
	is := setting.IsSensitive(row.Key)
	value := row.Value
	if is || setting.IsEncryptedValue(row.Value) {
		mask, _ := json.Marshal(setting.EncryptedPlaceholder)
		value = mask
	}
	updatedAt := ""
	if !row.UpdatedAt.IsZero() {
		updatedAt = row.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00")
	}
	return settingItem{
		Key:         row.Key,
		Value:       value,
		Description: row.Description,
		IsSensitive: is,
		UpdatedAt:   updatedAt,
	}
}

// maskedRaw 返回 audit 用的脱敏 raw;敏感或 ENC 值统一 "<encrypted>"。
func maskedRaw(key string, raw []byte) json.RawMessage {
	if setting.IsSensitive(key) || setting.IsEncryptedValue(raw) {
		b, _ := json.Marshal(setting.EncryptedPlaceholder)
		return b
	}
	return raw
}

// auditUpdate 写一条 setting.update 审计(before/after 都脱敏)。
func (h *SettingHandler) auditUpdate(c *gin.Context, key string, before, after []byte) {
	actor := h.ActorOf(c)
	tid := hashKeyToInt64(key)
	beforeJSON, _ := json.Marshal(map[string]any{
		"key":   key,
		"value": json.RawMessage(maskedRaw(key, before)),
	})
	afterJSON, _ := json.Marshal(map[string]any{
		"key":   key,
		"value": json.RawMessage(maskedRaw(key, after)),
	})
	_ = h.Audit.Log(c.Request.Context(), audit.Entry{
		Action:     "setting.update",
		TargetType: "setting",
		TargetID:   &tid,
		ActorID:    &actor,
		Before:     beforeJSON,
		After:      afterJSON,
		IP:         clientIP(c),
	})
}

// clientIP 提取 client IP(失败返空串)。
func clientIP(c *gin.Context) string {
	h := c.Request.RemoteAddr
	host, _, err := net.SplitHostPort(h)
	if err != nil {
		return h
	}
	return host
}

// hashKeyToInt64 把 setting key 哈希到 int64(供 audit.TargetID;M1 临时方案,M3 改 string)。
func hashKeyToInt64(key string) int64 {
	sum := md5.Sum([]byte(key))
	// 取前 8 字节,转 int64
	return int64(binary.BigEndian.Uint64(sum[:8]))
}

// isLikelyEmail 简单 email 合法性判断(M1 范围)。
func isLikelyEmail(s string) bool {
	if s == "" {
		return false
	}
	at := strings.IndexByte(s, '@')
	if at <= 0 || at == len(s)-1 {
		return false
	}
	if strings.IndexByte(s[at+1:], '.') < 0 {
		return false
	}
	return true
}

var _ = errors.New
