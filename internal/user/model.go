package user

import "time"

// Role 枚举。
const (
	RoleUser        int8 = 0 // 普通
	RoleDeptAdmin   int8 = 1 // 部门管理员(M3)
	RoleTenantAdmin int8 = 2 // 租户管理员
	RoleSuperAdmin  int8 = 3 // 超级管理员
)

// Status 枚举。
const (
	StatusActive             int8 = 0
	StatusDisabled           int8 = 1
	StatusPendingEmailVerify int8 = 2
)

// User 对应 users 表(M0 已建,M1 加 email_verified_at)。
type User struct {
	ID              int64      `gorm:"primaryKey;column:id"`
	Username        string     `gorm:"column:username;size:64;uniqueIndex"`
	Email           *string    `gorm:"column:email;size:128;uniqueIndex"`
	PasswordHash    *string    `gorm:"column:password_hash;size:128"`
	DisplayName     *string    `gorm:"column:display_name;size:64"`
	Avatar          *string    `gorm:"column:avatar;size:256"`
	Role            int8       `gorm:"column:role;default:0"`
	Status          int8       `gorm:"column:status;default:0"`
	GroupID         *int64     `gorm:"column:group_id"`
	InvitedBy       int64      `gorm:"column:invited_by;not null;default:0" json:"invited_by,omitempty"`
	EmailVerifiedAt *time.Time `gorm:"column:email_verified_at"`
	LastLoginAt     *time.Time `gorm:"column:last_login_at"`
	LastLoginIP     *string    `gorm:"column:last_login_ip;size:45"`
	CreatedAt       time.Time  `gorm:"column:created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at"`
	DeletedAt       *time.Time `gorm:"column:deleted_at;index"`
}

// TableName 表名固定 users。
func (User) TableName() string { return "users" }
