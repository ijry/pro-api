package adapter

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"syscall"

	"github.com/ijry/pro-api/pkg/apierr"
)

// ClassifyHTTPStatus 根据上游响应 status code + body 类型生成 *apierr.Error。
//
//   - body 含 content_filter / content_policy → CodeUpstreamContentFilter
//   - 429              → CodeUpstreamRateLimit
//   - 401 / 403        → CodeChannelMisconfig
//   - 402              → CodeUpstreamQuota
//   - 503              → CodeUpstreamUnavail
//   - 5xx(其他)       → CodeUpstreamError
//   - 400              → CodeInvalidParam(透传上游错误 body 给客户端,help debug)
//   - 其他非 2xx      → CodeUpstreamError
//
// 上游 raw body 截至 512 byte 放入 Error.Details["upstream_body"]。
func ClassifyHTTPStatus(status int, body []byte) *apierr.Error {
	bodyStr := string(body)
	truncated := truncate(bodyStr, 512)
	details := map[string]any{"upstream_body": truncated, "upstream_status": status}

	if containsContentFilterMarker(bodyStr) {
		return apierr.New(apierr.CodeUpstreamContentFilter, "upstream content filter triggered").
			WithDetails(details)
	}
	switch {
	case status == 429:
		return apierr.New(apierr.CodeUpstreamRateLimit, "upstream rate limited").WithDetails(details)
	case status == 401 || status == 403:
		return apierr.New(apierr.CodeChannelMisconfig, "upstream auth failed").WithDetails(details)
	case status == 402:
		return apierr.New(apierr.CodeUpstreamQuota, "upstream quota exhausted").WithDetails(details)
	case status == 503:
		return apierr.New(apierr.CodeUpstreamUnavail, "upstream unavailable").WithDetails(details)
	case status >= 500:
		return apierr.New(apierr.CodeUpstreamError, fmt.Sprintf("upstream returned %d", status)).WithDetails(details)
	case status == 400:
		return apierr.New(apierr.CodeInvalidParam, "upstream rejected request").WithDetails(details)
	default:
		return apierr.New(apierr.CodeUpstreamError, fmt.Sprintf("unexpected status %d", status)).WithDetails(details)
	}
}

// ClassifyNetErr 把网络/IO 错误归类为 *apierr.Error。
//
//   - ctx canceled / deadline → CodeUpstreamTimeout
//   - net.Error 且 Timeout()  → CodeUpstreamTimeout
//   - connection refused      → CodeUpstreamUnavail
//   - 其他网络错              → CodeUpstreamError
func ClassifyNetErr(err error) *apierr.Error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) {
		return apierr.New(apierr.CodeUpstreamTimeout, "request canceled")
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return apierr.New(apierr.CodeUpstreamTimeout, "request timeout")
	}
	var nerr net.Error
	if errors.As(err, &nerr) && nerr.Timeout() {
		return apierr.New(apierr.CodeUpstreamTimeout, "network timeout")
	}
	if errors.Is(err, syscall.ECONNREFUSED) {
		return apierr.New(apierr.CodeUpstreamUnavail, "connection refused")
	}
	return apierr.New(apierr.CodeUpstreamError, err.Error())
}

func containsContentFilterMarker(s string) bool {
	s = strings.ToLower(s)
	return strings.Contains(s, "content_filter") ||
		strings.Contains(s, "content_policy") ||
		strings.Contains(s, "responsibleaipolicy")
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
