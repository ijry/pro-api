package channel

import (
	"context"
	"time"
)

// BreakerState 是熔断器状态。
type BreakerState string

const (
	StateClosed   BreakerState = "closed"
	StateOpen     BreakerState = "open"
	StateHalfOpen BreakerState = "half_open"
)

// Breaker 是渠道熔断器接口。
type Breaker interface {
	// State 返回某 channel 的当前熔断状态。
	State(channelID int64) BreakerState

	// Snapshot 返回完整快照。
	Snapshot(channelID int64) HealthSnapshot

	// RecordSuccess / RecordFailure 是 Selector 在 retry 中调用的入口。
	RecordSuccess(channelID int64, latency time.Duration)
	RecordFailure(channelID int64, err error)

	// AcquireProbe 在 HALF_OPEN 状态下抢令牌。
	AcquireProbe(ctx context.Context, channelID int64, timeout time.Duration) (acquired bool, release func(), err error)

	// ForceTransition 强制状态转换（管理员启用时 reset 回 closed）。
	ForceTransition(channelID int64, to BreakerState) error

	// Close 停止后台 goroutine。
	Close() error
}

// Selector 是渠道选择器接口。
type Selector interface {
	Select(ctx context.Context, model string, hint SelectHint) (*Channel, error)
	ReportSuccess(channelID int64, latency time.Duration)
	ReportFailure(channelID int64, err error)
}

// HealthSnapshot 是渠道健康快照。
type HealthSnapshot struct {
	ChannelID   int64        `json:"channel_id"`
	State       BreakerState `json:"state"`
	ConsecFail  int          `json:"consec_fail"`
	OpenedAt    *time.Time   `json:"opened_at,omitempty"`
	LastProbeAt *time.Time   `json:"last_probe_at,omitempty"`
	Recent      Counters     `json:"recent"`
}

// Counters 是近期计数。
type Counters struct {
	Success int `json:"success"`
	Failure int `json:"failure"`
	Window  int `json:"window_seconds"`
}
