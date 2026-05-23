package channel

import (
	"context"
	"time"

	"errors"
	"github.com/ijry/pro-api/pkg/apierr"
)

// RetryFn 是 Selector 选中 channel 后业务执行的回调。
type RetryFn func(ctx context.Context, c *Channel) error

// WithRetry 包装"Select → 调用 → 失败 → 重试"流程。
func WithRetry(
	ctx context.Context,
	sel Selector,
	attempts int,
	model string,
	hint SelectHint,
	fn RetryFn,
) error {
	if attempts < 1 {
		attempts = 1
	}

	var lastErr error
	excluded := append([]int64{}, hint.Excluded...)

	for i := 0; i < attempts; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		h := hint
		h.Excluded = excluded
		ch, err := sel.Select(ctx, model, h)
		if err != nil {
			if lastErr != nil {
				return lastErr
			}
			return err
		}

		// HALF_OPEN 抢 probe 令牌
		var release func()
		if cs, ok := sel.(*channelSelector); ok {
			breaker := cs.getBreaker()
			if breaker.State(ch.ID) == StateHalfOpen {
				acquired, rel, _ := breaker.AcquireProbe(ctx, ch.ID, defaultProbeTimeout)
				if !acquired {
					excluded = append(excluded, ch.ID)
					continue
				}
				release = rel
			}
		}

		start := time.Now()
		err = fn(ctx, ch)
		elapsed := time.Since(start)
		if release != nil {
			release()
		}

		if err == nil {
			sel.ReportSuccess(ch.ID, elapsed)
			return nil
		}

		if !isRetryable(err) {
			sel.ReportFailure(ch.ID, err)
			return err
		}
		sel.ReportFailure(ch.ID, err)
		lastErr = err
		excluded = append(excluded, ch.ID)
	}
	return lastErr
}

// isRetryable 判断错误是否可重试。
func isRetryable(err error) bool {
	var e *apierr.Error; ok := errors.As(err, &e)
	if !ok {
		return false
	}
	switch e.Code {
	case apierr.CodeUpstreamError,
		apierr.CodeUpstreamTimeout,
		apierr.CodeUpstreamUnavail,
		apierr.CodeUpstreamRateLimit:
		return true
	}
	return false
}
