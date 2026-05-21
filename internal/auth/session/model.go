package session

import "time"

// DBSession 是 sessions 表 GORM 模型(DB mirror)。
type DBSession struct {
	ID         string     `gorm:"primaryKey;column:id;size:48"`
	UserID     int64      `gorm:"column:user_id"`
	IP         string     `gorm:"column:ip;size:45"`
	UserAgent  string     `gorm:"column:user_agent;size:256"`
	CreatedAt  time.Time  `gorm:"column:created_at"`
	LastSeenAt time.Time  `gorm:"column:last_seen_at"`
	ExpiresAt  time.Time  `gorm:"column:expires_at"`
	RevokedAt  *time.Time `gorm:"column:revoked_at"`
}

// TableName 表名固定 sessions。
func (DBSession) TableName() string { return "sessions" }

// Session 是返回给上层的轻量结构。
type Session struct {
	ID        string
	UserID    int64
	Role      int8
	IP        string
	UserAgent string
	CreatedAt time.Time
	LastSeen  time.Time
	ExpiresAt time.Time
}
