// Package apierr 定义统一错误码与 Error 类型,供后端各层与前端 i18n 共享。
package apierr

import "net/http"

// Code 是 5 位整数错误码。
type Code int

// 错误码分段:
//
//	1xxxx 系统/基础设施
//	2xxxx 鉴权
//	3xxxx 参数
//	4xxxx 业务
//	5xxxx 限流
//	6xxxx 上游
const (
	CodeOK              Code = 0
	CodeInternal        Code = 10000
	CodeDatabase        Code = 10001
	CodeCache           Code = 10002
	CodeUpstreamUnavail Code = 10003

	CodeNotLoggedIn     Code = 20001
	CodeSessionExpired  Code = 20002
	CodeInvalidToken    Code = 20003
	CodeTokenExpired    Code = 20004
	CodeIPNotAllowed    Code = 20005
	CodeModelNotAllowed Code = 20006
	CodeForbidden       Code = 20010

	CodeMissingParam Code = 30001
	CodeInvalidParam Code = 30002

	CodeEmailRegistered     Code = 40001
	CodeUsernameTaken       Code = 40002
	CodeWrongPassword       Code = 40003
	CodeBalanceInsufficient Code = 40004
	CodeNoChannel           Code = 40005
	CodeModelNotSupported   Code = 40006
	CodeOrderNotFound       Code = 40007
	CodeRedeemInvalid       Code = 40008
	CodeInviteInvalid       Code = 40009
	CodeDeptBudgetExceeded  Code = 40010

	CodeRateLimitUser   Code = 50001
	CodeRateLimitToken  Code = 50002
	CodeRateLimitIP     Code = 50003
	CodeRateLimitGlobal Code = 50004

	CodeUpstreamError         Code = 60001
	CodeUpstreamTimeout       Code = 60002
	CodeUpstreamContentFilter Code = 60003

	// === M1 新增 ===

	CodeEmailNotVerified Code = 20007
	CodeCaptchaInvalid   Code = 20008

	// CodeNotFound 表示资源(按 id 查询)不存在 → HTTP 404。
	// 通用资源 404,各模块都可复用(M1-03 spec §7 引用)。
	CodeNotFound Code = 40015

	CodeUpstreamRateLimit Code = 60004
	CodeUpstreamQuota     Code = 60005

	CodeReservationNotFound  Code = 40011
	CodeReservationCommitted Code = 40012

	CodeChannelDisabled  Code = 40013
	CodeChannelMisconfig Code = 40014
)

// httpStatusByCode 把错误码映射到 HTTP status。
func httpStatusByCode(c Code) int {
	switch {
	case c == CodeOK:
		return http.StatusOK
	case c == CodeNotLoggedIn || c == CodeSessionExpired || c == CodeInvalidToken || c == CodeTokenExpired:
		return http.StatusUnauthorized
	case c == CodeForbidden || c == CodeIPNotAllowed || c == CodeModelNotAllowed:
		return http.StatusForbidden
	case c == CodeMissingParam || c == CodeInvalidParam:
		return http.StatusBadRequest
	case c == CodeBalanceInsufficient || c == CodeDeptBudgetExceeded:
		return http.StatusPaymentRequired
	case c == CodeNoChannel || c == CodeModelNotSupported:
		return http.StatusNotFound
	case c == CodeOrderNotFound || c == CodeRedeemInvalid || c == CodeInviteInvalid:
		return http.StatusBadRequest
	case c == CodeEmailRegistered || c == CodeUsernameTaken:
		return http.StatusConflict
	case c == CodeWrongPassword:
		return http.StatusUnauthorized
	case c >= 50000 && c < 60000:
		return http.StatusTooManyRequests
	case c == CodeUpstreamError || c == CodeUpstreamUnavail || c == CodeUpstreamContentFilter:
		return http.StatusBadGateway
	case c == CodeUpstreamTimeout:
		return http.StatusGatewayTimeout
	case c == CodeEmailNotVerified || c == CodeCaptchaInvalid:
		return http.StatusUnauthorized
	case c == CodeUpstreamRateLimit || c == CodeUpstreamQuota:
		return http.StatusTooManyRequests
	case c == CodeReservationNotFound:
		return http.StatusGone
	case c == CodeReservationCommitted:
		return http.StatusConflict
	case c == CodeChannelDisabled || c == CodeChannelMisconfig:
		return http.StatusServiceUnavailable
	case c == CodeNotFound:
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}
