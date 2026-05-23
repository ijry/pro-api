// Package pricing 实现定价规则的匹配、缓存、CRUD,以及 quota 计算与最大估算。
//
// 设计稿:M1-06 spec §3.2 / §4.8-§4.10。
//
// 规则匹配顺序(由精到粗):group_model → model → group → global → catalog default;
// 每层内 priority ASC → id DESC。模型支持末尾 "*" 通配(与 token 白名单同语义)。
package pricing

import "time"

// Rule 是 pricing_rules 表的 GORM 模型。
type Rule struct {
	ID             int64     `gorm:"primaryKey;column:id"`
	Scope          string    `gorm:"column:scope;size:16"`
	GroupID        *int64    `gorm:"column:group_id"`
	Model          *string   `gorm:"column:model;size:128"`
	InputRatio     *float64  `gorm:"column:input_ratio"`
	OutputRatio    *float64  `gorm:"column:output_ratio"`
	CachedRatio    *float64  `gorm:"column:cached_ratio"`
	ReasoningRatio *float64  `gorm:"column:reasoning_ratio"`
	Priority       int16     `gorm:"column:priority"`
	Status         int8      `gorm:"column:status"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`
}

// TableName 固定表名。
func (Rule) TableName() string { return "pricing_rules" }

// scope 取值。
const (
	ScopeGlobal     = "global"
	ScopeGroup      = "group"
	ScopeModel      = "model"
	ScopeGroupModel = "group_model"
)

// status 取值。
const (
	RuleStatusEnabled  int8 = 0
	RuleStatusDisabled int8 = 1
)
