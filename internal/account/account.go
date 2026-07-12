package account

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"
)

// Status 是账号生命周期状态。
type Status int8

const (
	StatusActive   Status = 0
	StatusCooldown Status = 1
	StatusExpired  Status = 2
	StatusDisabled Status = 3
	StatusInvalid  Status = 4
)

// 额度模式:决定账号的 5h/周额度从哪来、是否参与后台探测。
const (
	// QuotaModeAuto 由后台探测/转发响应头自动回填额度(官方号)。
	QuotaModeAuto = "auto"
	// QuotaModeManual 手动设定额度,按 token 用量扣减 remaining;不参与后台探测(中转站号)。
	QuotaModeManual = "manual"
	// QuotaModeNone 不探测、不追踪额度,selector 视为恒满额(纯直连中转站号)。
	QuotaModeNone = "none"
)

// NormalizeQuotaMode 把空串或非法值归一到 QuotaModeAuto。
func NormalizeQuotaMode(m string) string {
	switch m {
	case QuotaModeManual, QuotaModeNone:
		return m
	default:
		return QuotaModeAuto
	}
}

// Account 是 accounts 表的 ORM 模型与领域对象。
type Account struct {
	ID                   int64           `gorm:"primaryKey;column:id"                json:"id"`
	ChannelID            int64           `gorm:"column:channel_id;index"             json:"channel_id"`
	Tag                  string          `gorm:"column:tag;size:64"                  json:"tag"`
	Name                 string          `gorm:"column:name;size:128"                json:"name"`
	Provider             string          `gorm:"column:provider;size:32"             json:"provider"`
	Tier                 string          `gorm:"column:tier;size:32"                 json:"tier"`
	QuotaMode            string          `gorm:"column:quota_mode;size:8"            json:"quota_mode"`
	CredType             string          `gorm:"column:cred_type;size:16"            json:"cred_type"`
	Email                string          `gorm:"column:email;size:128"               json:"email"`
	ExternalAccountID    string          `gorm:"column:external_account_id;size:64"  json:"external_account_id"`
	Credentials          string          `gorm:"column:credentials;type:text"        json:"-"`
	Priority             int16           `gorm:"column:priority"                     json:"priority"`
	Weight               int             `gorm:"column:weight"                       json:"weight"`
	Status               Status          `gorm:"column:status"                       json:"status"`
	CooldownUntil        *time.Time      `gorm:"column:cooldown_until"               json:"cooldown_until,omitempty"`
	LastFailureAt        *time.Time      `gorm:"column:last_failure_at"              json:"last_failure_at,omitempty"`
	LastFailureReason    string          `gorm:"column:last_failure_reason;size:256" json:"last_failure_reason"`
	ConsecFailures       int             `gorm:"column:consec_failures"              json:"consec_failures"`
	LastSuccessAt        *time.Time      `gorm:"column:last_success_at"              json:"last_success_at,omitempty"`
	LastUsedAt           *time.Time      `gorm:"column:last_used_at"                 json:"last_used_at,omitempty"`
	Quota5hTotal         *int64          `gorm:"column:quota_5h_total"               json:"quota_5h_total,omitempty"`
	Quota5hRemaining     *int64          `gorm:"column:quota_5h_remaining"           json:"quota_5h_remaining,omitempty"`
	Quota5hResetAt       *time.Time      `gorm:"column:quota_5h_reset_at"            json:"quota_5h_reset_at,omitempty"`
	QuotaWeekTotal       *int64          `gorm:"column:quota_week_total"             json:"quota_week_total,omitempty"`
	QuotaWeekRemaining   *int64          `gorm:"column:quota_week_remaining"         json:"quota_week_remaining,omitempty"`
	QuotaWeekResetAt     *time.Time      `gorm:"column:quota_week_reset_at"          json:"quota_week_reset_at,omitempty"`
	QuotaSyncedAt        *time.Time      `gorm:"column:quota_synced_at"              json:"quota_synced_at,omitempty"`
	AccessTokenExpiresAt *time.Time      `gorm:"column:access_token_expires_at"      json:"access_token_expires_at,omitempty"`
	RefreshTokenValid    int8            `gorm:"column:refresh_token_valid"          json:"refresh_token_valid"`
	LastRefreshedAt      *time.Time      `gorm:"column:last_refreshed_at"            json:"last_refreshed_at,omitempty"`
	ImportSource         string          `gorm:"column:import_source;size:32"        json:"import_source"`
	Extra                json.RawMessage `gorm:"column:extra;type:json"              json:"extra"`
	CreatedAt            time.Time       `gorm:"column:created_at"                   json:"created_at"`
	UpdatedAt            time.Time       `gorm:"column:updated_at"                   json:"updated_at"`
	DeletedAt            *time.Time      `gorm:"column:deleted_at;index"             json:"deleted_at,omitempty"`

	Cred AccountCred `gorm:"-" json:"-"`
}

// TableName implements gorm.Tabler.
func (Account) TableName() string { return "accounts" }

// AccountCred 是凭证明文结构(加密后写 Credentials 字段)。
type AccountCred struct {
	AccessToken   string          `json:"access_token,omitempty"`
	RefreshToken  string          `json:"refresh_token,omitempty"`
	IDToken       string          `json:"id_token,omitempty"`
	TokenType     string          `json:"token_type,omitempty"`
	Scope         string          `json:"scope,omitempty"`
	ExpiresAt     time.Time       `json:"expires_at,omitempty"`
	APIKey        string          `json:"api_key,omitempty"`
	RawSession    json.RawMessage `json:"raw_session,omitempty"`
	OAuthClientID string          `json:"oauth_client_id,omitempty"`
	OAuthIssuer   string          `json:"oauth_issuer,omitempty"`
}

// AccountEvent 是 account_events 表。
type AccountEvent struct {
	ID        int64           `gorm:"primaryKey;column:id"      json:"id"`
	AccountID int64           `gorm:"column:account_id;index"   json:"account_id"`
	EventType string          `gorm:"column:event_type;size:32" json:"event_type"`
	Payload   json.RawMessage `gorm:"column:payload;type:json"  json:"payload"`
	CreatedAt time.Time       `gorm:"column:created_at"         json:"created_at"`
}

// TableName implements gorm.Tabler.
func (AccountEvent) TableName() string { return "account_events" }

// MaskEmail 把 "user@example.com" 脱敏成 "u***@example.com"。
func MaskEmail(email string) string {
	at := strings.IndexByte(email, '@')
	if at < 1 {
		return "***"
	}
	return email[:1] + "***" + email[at:]
}

// DedupKey 计算账号去重键。优先级:external_account_id → email → sha256(access_token)。
// 返回空串表示该账号无可识别的去重维度。
func (a *Account) DedupKey() string {
	if a.ExternalAccountID != "" {
		return "ext:" + a.ExternalAccountID
	}
	if a.Email != "" {
		return "email:" + strings.ToLower(a.Email)
	}
	if a.Cred.AccessToken != "" {
		h := sha256.Sum256([]byte(a.Cred.AccessToken))
		return "atsha:" + hex.EncodeToString(h[:8])
	}
	return ""
}
