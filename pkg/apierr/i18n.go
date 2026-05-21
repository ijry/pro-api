package apierr

// Lang 是支持的语言。M1 仅 zh / en。
type Lang string

const (
	LangZH Lang = "zh"
	LangEN Lang = "en"
)

var defaultMessages = map[Lang]map[Code]string{
	LangZH: {
		CodeInternal:              "内部错误,请稍后重试",
		CodeDatabase:              "数据库错误",
		CodeCache:                 "缓存错误",
		CodeUpstreamUnavail:       "上游不可达",
		CodeNotLoggedIn:           "请先登录",
		CodeSessionExpired:        "会话已过期,请重新登录",
		CodeInvalidToken:          "令牌无效",
		CodeTokenExpired:          "令牌已过期",
		CodeIPNotAllowed:          "IP 不在白名单",
		CodeModelNotAllowed:       "该令牌无权调用此模型",
		CodeForbidden:             "权限不足",
		CodeMissingParam:          "缺少必要参数",
		CodeInvalidParam:          "参数格式错误",
		CodeEmailRegistered:       "邮箱已被注册",
		CodeUsernameTaken:         "用户名已被占用",
		CodeWrongPassword:         "邮箱或密码错误",
		CodeBalanceInsufficient:   "余额不足,请充值",
		CodeNoChannel:             "暂无可用渠道,请稍后重试",
		CodeModelNotSupported:     "不支持的模型",
		CodeOrderNotFound:         "订单不存在",
		CodeRedeemInvalid:         "兑换码无效或已使用",
		CodeInviteInvalid:         "邀请码无效",
		CodeDeptBudgetExceeded:    "部门预算不足",
		CodeRateLimitUser:         "请求过于频繁,请稍后再试",
		CodeRateLimitToken:        "令牌请求过于频繁",
		CodeRateLimitIP:           "IP 请求过于频繁",
		CodeRateLimitGlobal:       "系统繁忙,请稍后再试",
		CodeUpstreamError:         "上游返回错误",
		CodeUpstreamTimeout:       "上游响应超时",
		CodeUpstreamContentFilter: "上游内容审查拒绝",
		CodeEmailNotVerified:      "邮箱未验证,请先验证邮箱",
		CodeCaptchaInvalid:        "验证码错误或已过期",
		CodeUpstreamRateLimit:     "上游限流",
		CodeUpstreamQuota:         "上游 quota 已用完",
		CodeReservationNotFound:   "预扣记录不存在或已过期",
		CodeReservationCommitted:  "预扣记录已提交,不可重复操作",
		CodeChannelDisabled:       "渠道已禁用",
		CodeChannelMisconfig:      "渠道配置错误",
		CodeNotFound:              "资源不存在",
	},
	LangEN: {
		CodeInternal:              "Internal error, please retry",
		CodeDatabase:              "Database error",
		CodeCache:                 "Cache error",
		CodeUpstreamUnavail:       "Upstream unavailable",
		CodeNotLoggedIn:           "Please sign in",
		CodeSessionExpired:        "Session expired, please sign in again",
		CodeInvalidToken:          "Invalid token",
		CodeTokenExpired:          "Token expired",
		CodeIPNotAllowed:          "IP not allowed",
		CodeModelNotAllowed:       "Token not authorized for this model",
		CodeForbidden:             "Permission denied",
		CodeMissingParam:          "Missing required parameter",
		CodeInvalidParam:          "Invalid parameter",
		CodeEmailRegistered:       "Email already registered",
		CodeUsernameTaken:         "Username taken",
		CodeWrongPassword:         "Invalid email or password",
		CodeBalanceInsufficient:   "Insufficient balance, please top up",
		CodeNoChannel:             "No available channel, please retry later",
		CodeModelNotSupported:     "Unsupported model",
		CodeOrderNotFound:         "Order not found",
		CodeRedeemInvalid:         "Invalid or used redeem code",
		CodeInviteInvalid:         "Invalid invite code",
		CodeDeptBudgetExceeded:    "Department budget exceeded",
		CodeRateLimitUser:         "Too many requests, please slow down",
		CodeRateLimitToken:        "Too many requests from this token",
		CodeRateLimitIP:           "Too many requests from this IP",
		CodeRateLimitGlobal:       "System busy, please retry later",
		CodeUpstreamError:         "Upstream error",
		CodeUpstreamTimeout:       "Upstream timeout",
		CodeUpstreamContentFilter: "Upstream content filter",
		CodeEmailNotVerified:      "Email not verified",
		CodeCaptchaInvalid:        "Invalid or expired captcha",
		CodeUpstreamRateLimit:     "Upstream rate limited",
		CodeUpstreamQuota:         "Upstream quota exhausted",
		CodeReservationNotFound:   "Reservation not found or expired",
		CodeReservationCommitted:  "Reservation already committed",
		CodeChannelDisabled:       "Channel disabled",
		CodeChannelMisconfig:      "Channel misconfigured",
		CodeNotFound:              "Resource not found",
	},
}

// Message 返回 code 在 lang 下的默认消息。
// lang 未知时降级到 en;code 未知时返回空串。
func Message(lang Lang, code Code) string {
	if m, ok := defaultMessages[lang]; ok {
		if s, ok := m[code]; ok {
			return s
		}
	}
	if s, ok := defaultMessages[LangEN][code]; ok {
		return s
	}
	return ""
}

// Localized 是 New 的本地化版本,自动从 lang 取消息。
func Localized(lang Lang, code Code) *Error {
	return New(code, Message(lang, code))
}
