// Package verifycode 提供 6 位数字邮箱/手机验证码生成与校验。
//
// 设计要点见 spec §3.4、§4.6:
//   - 6 位数字,Redis TTL 5 分钟
//   - 同一 (purpose,email) 重复 Generate 直接覆盖
//   - 60s 节流键,避免穷举发送
//   - Verify 单次有效(命中后立刻删除)
package verifycode
