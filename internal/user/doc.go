// Package user 提供 users 表的 GORM 模型、仓储与业务服务。
//
// 角色 / 状态枚举见本包常量;CRUD 走 Repository,业务规则(如默认分组、邮箱验证)
// 在 Service 层处理。
package user
