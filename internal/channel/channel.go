package channel

import (
	"encoding/json"
	"time"
)

// Channel 是 channels 表的 ORM 模型与领域对象。
type Channel struct {
	ID          int64           `gorm:"primaryKey;column:id"         json:"id"`
	Name        string          `gorm:"column:name;size:64"           json:"name"`
	Provider    string          `gorm:"column:provider;size:32"       json:"provider"`
	BaseURL     string          `gorm:"column:base_url;size:256"      json:"base_url"`
	Credentials string          `gorm:"column:credentials;type:text"  json:"-"` // 密文
	Priority    int16           `gorm:"column:priority"               json:"priority"`
	Weight      int             `gorm:"column:weight"                 json:"weight"`
	Status      int8            `gorm:"column:status"                 json:"status"`
	Tags            json.RawMessage `gorm:"column:tags;type:json"             json:"tags"`
	Extra           json.RawMessage `gorm:"column:extra;type:json"            json:"extra"`
	PoolEnabled     int8            `gorm:"column:pool_enabled"               json:"pool_enabled"`
	AccountStrategy string          `gorm:"column:account_strategy;size:16"   json:"account_strategy"`
	AccountTopK     int8            `gorm:"column:account_top_k"              json:"account_top_k"`
	CreatedAt       time.Time       `gorm:"column:created_at"                 json:"created_at"`
	UpdatedAt   time.Time       `gorm:"column:updated_at"             json:"updated_at"`
	DeletedAt   *time.Time      `gorm:"column:deleted_at;index"       json:"deleted_at,omitempty"`

	// 运行时字段(不持久化)
	Cred           Credential               `gorm:"-" json:"-"`
	ModelMap       map[string]string        `gorm:"-" json:"-"`
	ModelOverrides map[string]ModelOverride `gorm:"-" json:"-"`
}

// TableName implements gorm.Tabler.
func (Channel) TableName() string { return "channels" }

// ModelMapping 是 channel_model_mappings 表的 ORM 模型。
type ModelMapping struct {
	ID             int64     `gorm:"primaryKey;column:id"                       json:"id"`
	ChannelID      int64     `gorm:"column:channel_id"                          json:"channel_id"`
	ClientModel    string    `gorm:"column:client_model;size:128"               json:"client_model"`
	UpstreamModel  string    `gorm:"column:upstream_model;size:128"             json:"upstream_model"`
	InputRatio     *float64  `gorm:"column:input_ratio;type:decimal(10,4)"      json:"input_ratio"`
	OutputRatio    *float64  `gorm:"column:output_ratio;type:decimal(10,4)"     json:"output_ratio"`
	CachedRatio    *float64  `gorm:"column:cached_ratio;type:decimal(10,4)"     json:"cached_ratio"`
	ReasoningRatio *float64  `gorm:"column:reasoning_ratio;type:decimal(10,4)"  json:"reasoning_ratio"`
	CreatedAt      time.Time `gorm:"column:created_at"                          json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at"                          json:"updated_at"`
}

// TableName implements gorm.Tabler.
func (ModelMapping) TableName() string { return "channel_model_mappings" }

// Credential 是渠道凭证的应用层结构。
type Credential struct {
	APIKey  string            `json:"api_key"`
	Secret  string            `json:"secret,omitempty"`
	Region  string            `json:"region,omitempty"`
	BaseURL string            `json:"base_url,omitempty"`
	Extra   map[string]string `json:"extra,omitempty"`
}

// IsZero 判断凭证是否为空。
func (c Credential) IsZero() bool { return c.APIKey == "" }

// ModelOverride 是单渠道模型倍率覆盖。nil 字段用模型字典默认值。
type ModelOverride struct {
	InputRatio     *float64
	OutputRatio    *float64
	CachedRatio    *float64
	ReasoningRatio *float64
}

// SelectHint 是 Selector 的辅助参数。
type SelectHint struct {
	Excluded []int64  // 本次请求已失败的 channel id
	Tags     []string // 要求 channel 必须包含的 tag
	UserID   int64
	TokenID  int64
	GroupID  int64
}

// MaskAPIKey 脱敏 API key:首 3 + **** + 末 4;不足 8 字符全 ****。
func MaskAPIKey(key string) string {
	if len(key) < 8 {
		return "****"
	}
	return key[:3] + "****" + key[len(key)-4:]
}
