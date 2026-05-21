package middleware

import "github.com/proapi/proapi/pkg/apierr"

// openAITypeMap 把 apierr.Code 映射到 OpenAI 错误协议的 type 字段。
var openAITypeMap = map[apierr.Code]string{
	apierr.CodeInvalidToken:           "invalid_request_error",
	apierr.CodeTokenExpired:           "invalid_request_error",
	apierr.CodeNotLoggedIn:            "invalid_request_error",
	apierr.CodeSessionExpired:         "invalid_request_error",
	apierr.CodeIPNotAllowed:           "invalid_request_error",
	apierr.CodeModelNotAllowed:        "invalid_request_error",
	apierr.CodeForbidden:              "invalid_request_error",
	apierr.CodeMissingParam:           "invalid_request_error",
	apierr.CodeInvalidParam:           "invalid_request_error",
	apierr.CodeEmailNotVerified:       "invalid_request_error",
	apierr.CodeCaptchaInvalid:         "invalid_request_error",
	apierr.CodeBalanceInsufficient:    "insufficient_quota",
	apierr.CodeDeptBudgetExceeded:     "insufficient_quota",
	apierr.CodeRateLimitUser:          "rate_limit_exceeded",
	apierr.CodeRateLimitToken:         "rate_limit_exceeded",
	apierr.CodeRateLimitIP:            "rate_limit_exceeded",
	apierr.CodeRateLimitGlobal:        "rate_limit_exceeded",
	apierr.CodeUpstreamRateLimit:      "rate_limit_exceeded",
	apierr.CodeUpstreamError:          "api_error",
	apierr.CodeUpstreamTimeout:        "api_error",
	apierr.CodeUpstreamUnavail:        "api_error",
	apierr.CodeUpstreamContentFilter:  "content_filter",
	apierr.CodeUpstreamQuota:          "insufficient_quota",
	apierr.CodeNoChannel:              "api_error",
	apierr.CodeModelNotSupported:      "invalid_request_error",
	apierr.CodeChannelDisabled:        "api_error",
	apierr.CodeChannelMisconfig:       "api_error",
	apierr.CodeReservationNotFound:    "api_error",
	apierr.CodeReservationCommitted:   "api_error",
}

// openAICodeMap 把 apierr.Code 映射到 OpenAI 错误协议的 code 字段(短串)。
var openAICodeMap = map[apierr.Code]string{
	apierr.CodeInvalidToken:          "invalid_api_key",
	apierr.CodeTokenExpired:          "invalid_api_key",
	apierr.CodeBalanceInsufficient:   "insufficient_quota",
	apierr.CodeRateLimitUser:         "rate_limit_exceeded",
	apierr.CodeRateLimitToken:        "rate_limit_exceeded",
	apierr.CodeModelNotSupported:     "model_not_found",
	apierr.CodeUpstreamRateLimit:     "rate_limit_exceeded",
	apierr.CodeUpstreamQuota:         "insufficient_quota",
	apierr.CodeUpstreamContentFilter: "content_filter",
}

// openAIType 取 type,缺省返回 "api_error"。
func openAIType(c apierr.Code) string {
	if v, ok := openAITypeMap[c]; ok {
		return v
	}
	return "api_error"
}

// openAICode 取 code 字符串,缺省返回空串。
func openAICode(c apierr.Code) string { return openAICodeMap[c] }
