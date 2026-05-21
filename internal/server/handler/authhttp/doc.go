// Package authhttp 实现 /api/auth/* + /api/user/* + /api/admin/* 路由 handler。
//
// 为节省篇幅与减少 import cycle 风险,M1-02 将所有 handler 集中在一个包,
// 并提供 RegisterRoutes(eng, deps) 单一入口。未来按需拆细。
package authhttp
