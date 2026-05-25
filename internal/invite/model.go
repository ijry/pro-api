package invite

import "time"

// Record is the DB model for invite_records.
type Record struct {
	ID            int64     `gorm:"primaryKey"`
	InviterID     int64     `gorm:"not null;index:idx_invite_records_inviter"`
	InviteeID     int64     `gorm:"not null;index:idx_invite_records_invitee"`
	OrderID       int64     `gorm:"not null"`
	RebateCents   int64     `gorm:"not null;default:0"`
	RebateCredits int64     `gorm:"not null;default:0"`
	CreatedAt     time.Time
}

func (Record) TableName() string { return "invite_records" }
