// Package redeem 提供兑换码批量生成与兑换的领域模型与服务。
//
// 状态机:
//
//	unused -> used      (user Redeem, 调 wallet.Credit)
//	unused -> disabled  (admin Disable)
//
// 已 used / disabled 不可再变更。
//
// 安全:明文兑换码全程不入库,只存 sha256 hex(64 字符)。
package redeem

import "time"

// Status 枚举。值与 DB tinyint/smallint 一致。
const (
	StatusUnused   int8 = 0
	StatusUsed     int8 = 1
	StatusDisabled int8 = 2
)

// Code 是 redeem_codes 表的 GORM 模型。
//
// CodeHash 是明文的 sha256(64 字符 hex);CodePrefix 是明文前 4 位,展示用。
type Code struct {
	ID          int64      `gorm:"primaryKey;column:id"`
	CodeHash    string     `gorm:"column:code_hash;size:64;uniqueIndex"`
	CodePrefix  string     `gorm:"column:code_prefix;size:8"`
	AmountQuota int64      `gorm:"column:amount_quota"`
	BatchNo     string     `gorm:"column:batch_no;size:32;index"`
	Status      int8       `gorm:"column:status"`
	UsedBy      *int64     `gorm:"column:used_by"`
	UsedAt      *time.Time `gorm:"column:used_at"`
	ExpiresAt   *time.Time `gorm:"column:expires_at"`
	CreatedBy   int64      `gorm:"column:created_by"`
	CreatedAt   time.Time  `gorm:"column:created_at"`
}

// TableName 固定表名。
func (Code) TableName() string { return "redeem_codes" }

// StatusName 返回可读名。
func StatusName(s int8) string {
	switch s {
	case StatusUnused:
		return "unused"
	case StatusUsed:
		return "used"
	case StatusDisabled:
		return "disabled"
	default:
		return "unknown"
	}
}
