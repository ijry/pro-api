// Package token 实现 pa-xxx API 调用令牌的存储、校验、限流与生命周期管理。
//
// 设计稿:docs/superpowers/specs/2026-05-21-里程碑1-03-API令牌-设计稿.md
//
// 主要导出:
//
//	Store          统一令牌存储接口(见 store.go)
//	View           解码后的业务对象
//	TokenAuth      Gin 中间件,把 "Authorization: Bearer pa-xxx" 解析为 ctx.user_id / token / group_id
//	AssertIPAllowed / AssertModelAllowed   handler 端点级白名单校验
//
// 设计要点:
//
//   - 明文 key 经 sha256 后入库(key_hash),只在 Create / Regenerate 一次性返回明文
//   - Authenticate 路径走 Redis 5min cache + 30s 负缓存,DB 兜底
//   - last_used_at / quota_used 异步 30s flush,优雅关停 final flush
//   - last_used 单调:flush SQL 用 GREATEST 避免多实例回退
//   - 通过 Pub/Sub 失效跨实例缓存(Revoke / Regenerate / Update / 异步 last_used)
package token
