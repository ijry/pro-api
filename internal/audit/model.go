// Package audit 记录关键操作日志(谁、何时、改了什么)。
package audit

import (
	"encoding/json"
	"time"
)

// Entry 是 audit_logs 表的 GORM 模型。
type Entry struct {
	ID         int64           `gorm:"primaryKey;column:id"`
	CreatedAt  time.Time       `gorm:"column:created_at"`
	ActorID    *int64          `gorm:"column:actor_id"`
	ActorRole  int8            `gorm:"column:actor_role"`
	Action     string          `gorm:"column:action;size:64"`
	TargetType string          `gorm:"column:target_type;size:32"`
	TargetID   *int64          `gorm:"column:target_id"`
	Before     json.RawMessage `gorm:"column:before;type:json"`
	After      json.RawMessage `gorm:"column:after;type:json"`
	IP         string          `gorm:"column:ip;size:45"`
}

// TableName 固定表名。
func (Entry) TableName() string { return "audit_logs" }
