// Package billing 实现预扣-提交-退款的三阶段计费。
package billing

import "time"

// Reservation 对应 quota_reservations 表。
type Reservation struct {
	ID             string    `gorm:"primaryKey;column:id"`
	WalletID       int64     `gorm:"column:wallet_id"`
	UserID         int64     `gorm:"column:user_id"`
	TokenID        int64     `gorm:"column:token_id"`
	RequestID      *int64    `gorm:"column:request_id"`
	ReservedQuota  int64     `gorm:"column:reserved_quota"`
	CommittedQuota int64     `gorm:"column:committed_quota"`
	Status         int8      `gorm:"column:status"` // 0=pending 1=committed 2=refunded 3=expired
	CreatedAt      time.Time `gorm:"column:created_at"`
	CommittedAt    *time.Time `gorm:"column:committed_at"`
	ExpiresAt      time.Time `gorm:"column:expires_at"`
}

func (Reservation) TableName() string { return "quota_reservations" }
