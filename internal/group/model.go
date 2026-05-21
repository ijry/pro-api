package group

import "time"

// Group 对应 user_groups 表。
type Group struct {
	ID          int64     `gorm:"primaryKey;column:id"`
	Name        string    `gorm:"column:name;size:64;uniqueIndex"`
	DisplayName string    `gorm:"column:display_name;size:64"`
	Ratio       float64   `gorm:"column:ratio;type:decimal(10,4);default:1.0"`
	Priority    int16     `gorm:"column:priority"`
	Status      int8      `gorm:"column:status"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

// TableName 表名固定 user_groups。
func (Group) TableName() string { return "user_groups" }

// DefaultGroupName 是种子数据中的默认分组名。
const DefaultGroupName = "default"
