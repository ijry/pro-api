package token

import (
	"encoding/json"
	"time"
)

// Token 是 api_tokens 表的 GORM 模型(数据库行)。
// JSON 字段保持原始字节,业务层使用解码后的 View。
type Token struct {
	ID            int64           `gorm:"primaryKey;column:id"`
	UserID        int64           `gorm:"column:user_id;index"`
	Name          string          `gorm:"column:name;size:64"`
	KeyHash       string          `gorm:"column:key_hash;size:64;uniqueIndex"`
	KeyPrefix     string          `gorm:"column:key_prefix;size:32"`
	QuotaLimit    *int64          `gorm:"column:quota_limit"`
	QuotaUsed     int64           `gorm:"column:quota_used"`
	AllowedModels json.RawMessage `gorm:"column:allowed_models;type:json"`
	AllowedIPs    json.RawMessage `gorm:"column:allowed_ips;type:json"`
	RPMLimit      int             `gorm:"column:rpm_limit"`
	TPMLimit      int             `gorm:"column:tpm_limit"`
	ExpiresAt     *time.Time      `gorm:"column:expires_at"`
	LastUsedAt    *time.Time      `gorm:"column:last_used_at"`
	Status        int8            `gorm:"column:status"`
	CreatedAt     time.Time       `gorm:"column:created_at"`
	UpdatedAt     time.Time       `gorm:"column:updated_at"`
	DeletedAt     *time.Time      `gorm:"column:deleted_at;index"`
}

// TableName 固定表名。
func (Token) TableName() string { return "api_tokens" }

// 状态常量。
const (
	// StatusEnabled 表示令牌可用。
	StatusEnabled int8 = 0
	// StatusDisabled 表示令牌被显式禁用(Revoke)。
	StatusDisabled int8 = 1
	// StatusExpired 是占位状态,M1 不主动维护;Authenticate 实时根据 expires_at 判定。
	StatusExpired int8 = 2
)

// View 是解码后的业务对象,在 Authenticate / 缓存 / handler 之间流转。
//
// 注意:View 不包含明文 key — 明文只在 Create / Regenerate 的返回值里一次性出现。
type View struct {
	ID            int64
	UserID        int64
	Name          string
	KeyPrefix     string
	QuotaLimit    *int64
	QuotaUsed     int64
	AllowedModels []string
	AllowedIPs    []string
	RPMLimit      int
	TPMLimit      int
	ExpiresAt     *time.Time
	LastUsedAt    *time.Time
	Status        int8
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// ToView 把 GORM 行解码为业务 View。JSON 字段解析失败按空数组处理(损坏数据不阻塞鉴权)。
func (t *Token) ToView() *View {
	v := &View{
		ID:         t.ID,
		UserID:     t.UserID,
		Name:       t.Name,
		KeyPrefix:  t.KeyPrefix,
		QuotaLimit: t.QuotaLimit,
		QuotaUsed:  t.QuotaUsed,
		RPMLimit:   t.RPMLimit,
		TPMLimit:   t.TPMLimit,
		ExpiresAt:  t.ExpiresAt,
		LastUsedAt: t.LastUsedAt,
		Status:     t.Status,
		CreatedAt:  t.CreatedAt,
		UpdatedAt:  t.UpdatedAt,
	}
	if len(t.AllowedModels) > 0 {
		_ = json.Unmarshal(t.AllowedModels, &v.AllowedModels)
	}
	if len(t.AllowedIPs) > 0 {
		_ = json.Unmarshal(t.AllowedIPs, &v.AllowedIPs)
	}
	return v
}

// UpdatePatch 表示 Update 的部分字段。任一为 nil 表示该字段不更新。
//
//	QuotaLimit 与 ClearQuotaLimit 配合:
//	  ClearQuotaLimit=true → 把 quota_limit 置为 NULL(无限)
//	  ClearQuotaLimit=false && QuotaLimit != nil → 设为给定值
//	  ClearQuotaLimit=false && QuotaLimit == nil → 不更新
//
//	ExpiresAt 与 ClearExpiresAt 同理。
type UpdatePatch struct {
	Name            *string
	QuotaLimit      *int64
	ClearQuotaLimit bool
	AllowedModels   *[]string
	AllowedIPs      *[]string
	RPMLimit        *int
	TPMLimit        *int
	ExpiresAt       *time.Time
	ClearExpiresAt  bool
	Status          *int8
}

// ListFilter 是 List 的查询参数。
type ListFilter struct {
	UserID  int64 // 0 = 不限(管理员)
	Status  *int8 // nil = 不限
	Keyword string
	Page    int
	Size    int
}
