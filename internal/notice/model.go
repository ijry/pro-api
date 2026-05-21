// Package notice 提供公告(notices)CRUD/发布/已读机制。
//
// 三个层次:
//   - Repo:GORM 仓储层,对应 notices 表
//   - Reader:Redis SET notice:read:{user_id} 已读集合
//   - Service:业务编排层,管理员/用户/公开三套视角
//
// 设计:
//   - status=0 草稿 / 1 已发布 / 2 已下架
//   - 已读机制纯 Redis,无 TTL,容忍偶发数据丢失
//   - Publish/Unpublish 单独 endpoint,Update 禁改 status
package notice

import "time"

// status 取值
const (
	StatusDraft     int8 = 0 // 未发布
	StatusPublished int8 = 1 // 已发布
	StatusArchived  int8 = 2 // 已下架
)

// level 取值(纯展示,后端不校验业务语义)
const (
	LevelInfo    = "info"
	LevelWarning = "warning"
	LevelDanger  = "danger"
	LevelSuccess = "success"
)

// target 取值
const (
	TargetAll   = "all"   // 所有人
	TargetUser  = "user"  // 普通用户
	TargetAdmin = "admin" // 管理员
)

// validLevels 是 level 字段合法值集合。
var validLevels = map[string]struct{}{
	LevelInfo: {}, LevelWarning: {}, LevelDanger: {}, LevelSuccess: {},
}

// validTargets 是 target 字段合法值集合。
var validTargets = map[string]struct{}{
	TargetAll: {}, TargetUser: {}, TargetAdmin: {},
}

// IsValidLevel 判断 level 是否合法。
func IsValidLevel(s string) bool { _, ok := validLevels[s]; return ok }

// IsValidTarget 判断 target 是否合法。
func IsValidTarget(s string) bool { _, ok := validTargets[s]; return ok }

// Notice 是 notices 表的 GORM 模型。
type Notice struct {
	ID        int64      `gorm:"primaryKey;column:id"            json:"id,string"`
	Title     string     `gorm:"column:title;size:128"           json:"title"`
	Content   string     `gorm:"column:content;type:text"        json:"content"`
	Level     string     `gorm:"column:level;size:16"            json:"level"`
	Target    string     `gorm:"column:target;size:16"           json:"target"`
	Status    int8       `gorm:"column:status"                   json:"status"`
	PublishAt *time.Time `gorm:"column:publish_at"               json:"publish_at"`
	ExpiresAt *time.Time `gorm:"column:expires_at"               json:"expires_at"`
	Pinned    bool       `gorm:"column:pinned"                   json:"pinned"`
	CreatedBy int64      `gorm:"column:created_by"               json:"created_by,string"`
	CreatedAt time.Time  `gorm:"column:created_at"               json:"created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at"               json:"updated_at"`
}

// TableName 固定表名。
func (Notice) TableName() string { return "notices" }

// UserNotice 是用户视角的公告;比 Notice 多 IsRead。
// 用户视角不返 Status / CreatedBy。
type UserNotice struct {
	ID        int64      `json:"id,string"`
	Title     string     `json:"title"`
	Content   string     `json:"content"`
	Level     string     `json:"level"`
	Target    string     `json:"target"`
	PublishAt *time.Time `json:"publish_at"`
	ExpiresAt *time.Time `json:"expires_at"`
	Pinned    bool       `json:"pinned"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	IsRead    bool       `json:"is_read"`
}

// ToUserNotice 转 Notice 为 UserNotice 视图。
func ToUserNotice(n *Notice, isRead bool) *UserNotice {
	return &UserNotice{
		ID:        n.ID,
		Title:     n.Title,
		Content:   n.Content,
		Level:     n.Level,
		Target:    n.Target,
		PublishAt: n.PublishAt,
		ExpiresAt: n.ExpiresAt,
		Pinned:    n.Pinned,
		CreatedAt: n.CreatedAt,
		UpdatedAt: n.UpdatedAt,
		IsRead:    isRead,
	}
}
