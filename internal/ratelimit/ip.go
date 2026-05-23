// Package ratelimit 实现 4 维滑动窗口限流(user / token / ip / model)。
//
// 顶层 API:
//   - Limiter 接口(AllowMulti / ConsumeTPM / Stats / Reset)
//   - Planner(从请求上下文派生 Check 列表)
//   - WireRateLimit(把 limiter + planner 装到 app.Application)
//
// 算法:Redis sorted-set + Lua,详见 docs/superpowers/specs/2026-05-21-里程碑1-07-限流-设计稿.md。
package ratelimit

import (
	"net"
	"net/netip"
)

// CanonicalIP 把入参 IP / IP:port 转成限流键用的 cidr 字符串。
//
//   - IPv4 → /24,例 "1.2.3.4" → "1.2.3.0/24"
//   - IPv6 → /64,例 "2001:db8::1" → "2001:db8::/64"
//   - IPv4-mapped IPv6(::ffff:x.x.x.x)按 IPv4 处理
//   - 解析失败 → 返回原串(不带 /24,后续脚本仍能跑;调用方可打 warn 日志)
//   - 空串 → 空串
func CanonicalIP(raw string) string {
	if raw == "" {
		return ""
	}
	// 去 port:支持 "1.2.3.4:5678" / "[::1]:5678"。SplitHostPort 失败时保留原值。
	if h, _, err := net.SplitHostPort(raw); err == nil {
		raw = h
	}
	addr, err := netip.ParseAddr(raw)
	if err != nil {
		return raw // fallback,不阻断业务
	}
	// IPv4-mapped IPv6 → unmap 为 IPv4
	if addr.Is4In6() {
		addr = addr.Unmap()
	}
	bits := 24
	if addr.Is6() {
		bits = 64
	}
	p := netip.PrefixFrom(addr, bits).Masked()
	return p.String()
}
