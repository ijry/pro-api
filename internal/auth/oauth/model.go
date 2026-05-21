package oauth

import (
	"encoding/json"
	"time"
)

// Binding 对应 oauth_bindings 表。
type Binding struct {
	ID          int64           `gorm:"primaryKey;column:id"`
	UserID      int64           `gorm:"column:user_id"`
	Provider    string          `gorm:"column:provider;size:16"`
	ProviderUID string          `gorm:"column:provider_uid;size:128"`
	Email       string          `gorm:"column:email;size:128"`
	Profile     json.RawMessage `gorm:"column:profile;type:json"`
	CreatedAt   time.Time       `gorm:"column:created_at"`
	UpdatedAt   time.Time       `gorm:"column:updated_at"`
}

// TableName 表名固定 oauth_bindings。
func (Binding) TableName() string { return "oauth_bindings" }
