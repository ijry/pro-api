// Package auth 是 user/group/session/oauth/verifycode 的顶层编排服务。
//
// 业务能力(对应 spec §3.8):
//   - Register / Login / Logout / EmailCodeLogin
//   - SendEmailCode / ForgotPassword / ResetPassword / ChangePassword
//   - GithubOAuthStart / Callback / Bind / Unbind
package auth
