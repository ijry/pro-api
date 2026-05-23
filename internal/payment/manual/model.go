// Package manual 提供手动充值申请的领域模型与服务。
//
// 状态机:
//
//	pending -> approved   (admin Approve, 调 wallet.Credit)
//	pending -> rejected   (admin Reject, 不调钱包)
//	pending -> canceled   (user Cancel)
//
// 非 pending 状态不可再变更。
package manual

import "time"

// Status 枚举。值与 DB tinyint/smallint 一致。
const (
	StatusPending  int8 = 0
	StatusApproved int8 = 1
	StatusRejected int8 = 2
	StatusCanceled int8 = 3
)

// Currency 常量。
const (
	CurrencyCNY = "CNY"
)

// Recharge 是 manual_recharges 表的 GORM 模型。
//
// AmountMoney 单位是"厘"(1 元 = 10000 厘);AmountQuota 在 Approve 时填充。
type Recharge struct {
	ID            int64      `gorm:"primaryKey;column:id"`
	UserID        int64      `gorm:"column:user_id;index"`
	AmountMoney   int64      `gorm:"column:amount_money"`
	Currency      string     `gorm:"column:currency;size:8"`
	AmountQuota   int64      `gorm:"column:amount_quota"`
	Status        int8       `gorm:"column:status"`
	ApplicantNote string     `gorm:"column:applicant_note;size:512"`
	ReviewerID    *int64     `gorm:"column:reviewer_id"`
	ReviewNote    string     `gorm:"column:review_note;size:512"`
	ReviewedAt    *time.Time `gorm:"column:reviewed_at"`
	CreatedAt     time.Time  `gorm:"column:created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at"`
}

// TableName 固定表名。
func (Recharge) TableName() string { return "manual_recharges" }

// StatusName 返回可读名(用于审计 / 日志)。
func StatusName(s int8) string {
	switch s {
	case StatusPending:
		return "pending"
	case StatusApproved:
		return "approved"
	case StatusRejected:
		return "rejected"
	case StatusCanceled:
		return "canceled"
	default:
		return "unknown"
	}
}
