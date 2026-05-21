// Package session 提供 SessionStore(Redis primary + DB mirror)。
//
// 设计要点见 spec §4.5:
//   - Redis 是主存储,DB 是异步镜像
//   - Sliding TTL(默认开)
//   - Pub/Sub channel "proapi:session:revoke" 提供跨实例失效通知
//   - 启动时回放 DB 中未过期 session 到 Redis(防冷启动雪崩)
package session
