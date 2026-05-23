package ratelimit

import (
	"context"
	"strings"
	"time"
)

// Dimension 是维度的语义标识(用于日志 / 响应头),非 redis key 一部分。
type Dimension string

const (
	DimUserRPM  Dimension = "user_rpm"
	DimUserTPM  Dimension = "user_tpm"
	DimTokenRPM Dimension = "token_rpm"
	DimTokenTPM Dimension = "token_tpm"
	DimIPRPM    Dimension = "ip_rpm"
	DimModelRPM Dimension = "model_rpm"
	DimModelTPM Dimension = "model_tpm"
)

// IsTPM 区分 RPM / TPM。
func (d Dimension) IsTPM() bool {
	return d == DimUserTPM || d == DimTokenTPM || d == DimModelTPM
}

// HeaderSuffix 用于 X-RateLimit-* 响应头。
// 例:DimUserRPM → "User-RPM";DimModelTPM → "Model-TPM"。
// 算法:按 "_" 拆,首字母大写,"rpm" / "tpm" / "ip" 整体大写,连字符 "-" 连接。
func (d Dimension) HeaderSuffix() string {
	parts := strings.Split(string(d), "_")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		switch p {
		case "rpm", "tpm", "ip":
			out = append(out, strings.ToUpper(p))
		default:
			if p == "" {
				continue
			}
			out = append(out, strings.ToUpper(p[:1])+p[1:])
		}
	}
	return strings.Join(out, "-")
}

// Check 是单维度一次检查的输入。
type Check struct {
	Dimension Dimension
	Key       string        // 完整 redis key(由 Planner 拼装,不带 "ratelimit:" 前缀也可)
	Limit     int           // 阈值;0 = 不限,直接通过
	Window    time.Duration // 窗口长度
	Cost      int           // 默认 1;TPM 可为 token 数
}

// PerDimDecision 是 AllowMulti 在每维度的结果。
type PerDimDecision struct {
	Dimension Dimension
	Allowed   bool
	Count     int       // 写入后(或被拒前)的实际计数
	Limit     int
	Reset     time.Time
}

// Decision 是 AllowMulti 的整体结果。
type Decision struct {
	Allowed   bool             // 任一维度被拒 → false
	Denied    *Check           // 被拒的那一维度;Allowed=true 时 nil
	Remaining int              // 被拒维度剩余配额;Allowed=true 时为最紧张维度的余量
	Limit     int              // 被拒维度阈值
	Reset     time.Time        // 被拒维度窗口结束时间;Allowed=true 时为最紧张维度的 reset
	Dimension Dimension        // 被拒维度;Allowed=true 时为最紧张维度
	Per       []PerDimDecision // 每维度的明细(供 stats / 日志使用)
}

// Limiter 是 4 维滑动窗口限流的统一入口。
type Limiter interface {
	// AllowMulti 短路语义:遇到第一个被拒立即返回(不再扣后续维度)。
	AllowMulti(ctx context.Context, checks []Check) Decision

	// ConsumeTPM 仅扣 TPM 维度,不阻塞调用方。即使超额也照写入。
	ConsumeTPM(ctx context.Context, checks []Check) error

	// Stats 查某个完整 redis key 的当前窗口内计数。供管理 API 使用。
	Stats(ctx context.Context, key string) (count int, oldestAt time.Time, err error)

	// Reset 清空某个 key(DEL)。供管理 API 使用。
	Reset(ctx context.Context, key string) error
}
