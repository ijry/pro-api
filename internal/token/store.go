package token

import (
	"context"
	"time"
)

// Store 是 token 模块对外的统一接口。
// 所有 HTTP handler / 其他模块通过 Store 与令牌交互;实现细节(repo / cache / flusher)对外透明。
//
// 设计稿:M1-03 spec §3.1。
type Store interface {
	// Authenticate 用明文 key 校验。
	//
	//   - 形式必须是 "pa-xxx..."
	//   - 错误:apierr.CodeInvalidToken(格式/不存在/已禁用) / CodeTokenExpired
	//   - 命中后异步 schedule last_used_at flush
	Authenticate(ctx context.Context, plaintextKey string) (*View, error)

	// Create 创建令牌。
	//
	//   - 内部生成明文 + sha256 + prefix
	//   - 返回明文 key(只此一次,后续永远不可读)
	Create(ctx context.Context, in CreateInput) (plaintextKey string, view *View, err error)

	// List 列我的(用户侧)或全部(管理员侧,filter.UserID == 0 视为全部)。
	List(ctx context.Context, filter ListFilter) ([]*View, int64, error)

	// Get 按 id 取详情。不存在返回 apierr.CodeNotFound。
	Get(ctx context.Context, id int64) (*View, error)

	// Update 部分字段更新。
	Update(ctx context.Context, id int64, patch UpdatePatch) (*View, error)

	// Revoke 软禁用(status=1)。不删表,key_hash 仍占用以避免重用。
	Revoke(ctx context.Context, id int64) error

	// Regenerate 原 key_hash 失效,生成新明文。
	Regenerate(ctx context.Context, id int64) (plaintextKey string, view *View, err error)

	// IncrementUsage 给 M1-06 调:本次消费的 quota 累加到 quota_used。
	// 同步即返,不阻塞调用方;内部 batch flush。
	IncrementUsage(tokenID int64, delta int64)

	// TouchLastUsed 给 TokenAuth 中间件调:更新 last_used_at = max(old, t)。
	TouchLastUsed(tokenID int64, t time.Time)

	// Close 关停 batch flusher + Pub/Sub 订阅。final flush 同步完成。
	Close() error
}
