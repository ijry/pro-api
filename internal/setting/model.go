// Package setting 加载/写入运行时可改的系统设置。
//
// 三层一致性:本地一级缓存(TTL 60s) → Redis(TTL 10min) → DB(权威)
// Put 时:DB UPSERT → Redis DEL → Pub/Sub 广播失效本地缓存
package setting

import (
	"encoding/json"
	"time"
)

// Setting 是 system_settings 表的 GORM 模型。
type Setting struct {
	Key         string          `gorm:"primaryKey;column:key;size:128"`
	Value       json.RawMessage `gorm:"column:value;type:json"`
	Description string          `gorm:"column:description;size:256"`
	UpdatedBy   *int64          `gorm:"column:updated_by"`
	UpdatedAt   time.Time       `gorm:"column:updated_at"`
}

// TableName 固定表名。
func (Setting) TableName() string { return "system_settings" }
