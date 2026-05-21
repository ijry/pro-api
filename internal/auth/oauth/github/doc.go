// Package github 提供 GitHub OAuth2 provider 实现。
//
// 流程:BuildAuthURL → (用户在 GitHub 同意)→ Exchange 用 code 换 token + 拉
// /user 与 /user/emails(primary verified email 兜底)。
package github
