package ratelimit

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ijry/pro-api/internal/setting"
)

// 维度键命名(总纲 §6.2):
//
//	ratelimit:user:{user_id}:rpm
//	ratelimit:user:{user_id}:tpm
//	ratelimit:token:{token_id}:rpm
//	ratelimit:token:{token_id}:tpm
//	ratelimit:ip:{cidr}:rpm
//	ratelimit:model:{model}:rpm
//	ratelimit:model:{model}:tpm

// PlanInput 是构造 Check 列表所需的上下文。
// 由 RateLimit 中间件从 gin.Context 拼装。
type PlanInput struct {
	UserID    int64
	TokenID   int64
	GroupID   int64
	ChannelID int64
	IP        string
	Model     string

	// 来自 token 字段的覆盖值;0 表示用全局默认。
	TokenRPMOverride int
	TokenTPMOverride int
	// 来自 user_groups.ratio 的折算系数;0 或 >= 1 表示不调整。
	GroupRatio float64
}

// PlannerConfig 是 NewPlanner 的参数。
type PlannerConfig struct {
	Setting  setting.Store
	LocalTTL time.Duration // 默认 30s
}

// Planner 由 setting 推导限流阈值,生成 Check 列表。
type Planner struct {
	setting  setting.Store
	cache    *thresholdCache
	windowFn func(ctx context.Context) time.Duration
}

// NewPlanner 构造一个 Planner。
func NewPlanner(cfg PlannerConfig) *Planner {
	ttl := cfg.LocalTTL
	if ttl <= 0 {
		ttl = 30 * time.Second
	}
	p := &Planner{
		setting: cfg.Setting,
		cache:   newThresholdCache(ttl),
	}
	p.windowFn = p.windowFromSetting
	return p
}

// SettingStore 暴露内部 store(供测试 / 调用方调试用)。
func (p *Planner) SettingStore() setting.Store { return p.setting }

// InvalidateCache 由外部(订阅 Redis Pub/Sub 后)调用,清空本地阈值缓存。
func (p *Planner) InvalidateCache() {
	if p == nil || p.cache == nil {
		return
	}
	p.cache.purge()
}

// PlanRPM 根据 PlanInput 生成 RPM 维度的 Check 列表。
// 顺序:global → group → user → token → channel → ip → model。
// 阈值为 0 的维度被跳过。
func (p *Planner) PlanRPM(ctx context.Context, in PlanInput) []Check {
	window := p.window(ctx)
	out := make([]Check, 0, 7)

	// 全局维度
	if lim := p.globalRPM(ctx); lim > 0 {
		out = append(out, Check{
			Dimension: DimGlobalRPM,
			Key:       "ratelimit:global:rpm",
			Limit:     lim,
			Window:    window,
			Cost:      1,
		})
	}

	// 分组维度
	if in.GroupID > 0 {
		if lim := p.groupRPM(ctx, in); lim > 0 {
			out = append(out, Check{
				Dimension: DimGroupRPM,
				Key:       fmt.Sprintf("ratelimit:group:%d:rpm", in.GroupID),
				Limit:     lim,
				Window:    window,
				Cost:      1,
			})
		}
	}

	if in.UserID > 0 {
		lim := p.userRPM(ctx, in)
		if lim > 0 {
			out = append(out, Check{
				Dimension: DimUserRPM,
				Key:       fmt.Sprintf("ratelimit:user:%d:rpm", in.UserID),
				Limit:     lim,
				Window:    window,
				Cost:      1,
			})
		}
	}
	if in.TokenID > 0 {
		lim := p.tokenRPM(ctx, in)
		if lim > 0 {
			out = append(out, Check{
				Dimension: DimTokenRPM,
				Key:       fmt.Sprintf("ratelimit:token:%d:rpm", in.TokenID),
				Limit:     lim,
				Window:    window,
				Cost:      1,
			})
		}
	}

	// 渠道维度
	if in.ChannelID > 0 {
		if lim := p.channelRPM(ctx); lim > 0 {
			out = append(out, Check{
				Dimension: DimChannelRPM,
				Key:       fmt.Sprintf("ratelimit:channel:%d:rpm", in.ChannelID),
				Limit:     lim,
				Window:    window,
				Cost:      1,
			})
		}
	}

	if in.IP != "" {
		lim := p.ipRPM(ctx)
		if lim > 0 {
			out = append(out, Check{
				Dimension: DimIPRPM,
				Key:       fmt.Sprintf("ratelimit:ip:%s:rpm", CanonicalIP(in.IP)),
				Limit:     lim,
				Window:    window,
				Cost:      1,
			})
		}
	}
	if in.Model != "" {
		lim := p.modelRPM(ctx, in.Model)
		if lim > 0 {
			out = append(out, Check{
				Dimension: DimModelRPM,
				Key:       fmt.Sprintf("ratelimit:model:%s:rpm", sanitizeModel(in.Model)),
				Limit:     lim,
				Window:    window,
				Cost:      1,
			})
		}
	}
	return out
}

// PlanTPM 生成 TPM 维度的 Check 列表。
// 返回的 Check.Cost = 0(由调用方在 ConsumeTPM 前通过 FillTPMCost 回填实际 token 数)。
func (p *Planner) PlanTPM(ctx context.Context, in PlanInput) []Check {
	window := p.window(ctx)
	out := make([]Check, 0, 6)

	// 全局维度
	if lim := p.globalTPM(ctx); lim > 0 {
		out = append(out, Check{
			Dimension: DimGlobalTPM,
			Key:       "ratelimit:global:tpm",
			Limit:     lim,
			Window:    window,
			Cost:      0,
		})
	}

	// 分组维度
	if in.GroupID > 0 {
		if lim := p.groupTPM(ctx, in); lim > 0 {
			out = append(out, Check{
				Dimension: DimGroupTPM,
				Key:       fmt.Sprintf("ratelimit:group:%d:tpm", in.GroupID),
				Limit:     lim,
				Window:    window,
				Cost:      0,
			})
		}
	}

	if in.UserID > 0 {
		lim := p.userTPM(ctx, in)
		if lim > 0 {
			out = append(out, Check{
				Dimension: DimUserTPM,
				Key:       fmt.Sprintf("ratelimit:user:%d:tpm", in.UserID),
				Limit:     lim,
				Window:    window,
				Cost:      0,
			})
		}
	}
	if in.TokenID > 0 {
		lim := p.tokenTPM(ctx, in)
		if lim > 0 {
			out = append(out, Check{
				Dimension: DimTokenTPM,
				Key:       fmt.Sprintf("ratelimit:token:%d:tpm", in.TokenID),
				Limit:     lim,
				Window:    window,
				Cost:      0,
			})
		}
	}

	// 渠道维度
	if in.ChannelID > 0 {
		if lim := p.channelTPM(ctx); lim > 0 {
			out = append(out, Check{
				Dimension: DimChannelTPM,
				Key:       fmt.Sprintf("ratelimit:channel:%d:tpm", in.ChannelID),
				Limit:     lim,
				Window:    window,
				Cost:      0,
			})
		}
	}

	if in.Model != "" {
		lim := p.modelTPM(ctx, in.Model)
		if lim > 0 {
			out = append(out, Check{
				Dimension: DimModelTPM,
				Key:       fmt.Sprintf("ratelimit:model:%s:tpm", sanitizeModel(in.Model)),
				Limit:     lim,
				Window:    window,
				Cost:      0,
			})
		}
	}
	return out
}

// FillTPMCost 把 Check.Cost 统一改写为 totalTokens,返回新切片(不修改原切片)。
func FillTPMCost(checks []Check, totalTokens int) []Check {
	out := make([]Check, len(checks))
	for i, c := range checks {
		c.Cost = totalTokens
		out[i] = c
	}
	return out
}

// =============== 阈值推导(私有)===============

// userRPM = default user_rpm × 1/ratio(group 折算)。
// 注意:TokenRPMOverride 只作用于 token_rpm 维度,**不影响** user_rpm。
// 这是 M1 的决策:token 维度可以单独被吊销;user 维度跨 token 累计。两份计数互补。
func (p *Planner) userRPM(ctx context.Context, in PlanInput) int {
	def := p.cachedInt(ctx, "ratelimit.user_default_rpm", 60, "user_rpm_default")
	return scaleByGroup(def, in.GroupRatio)
}

func (p *Planner) userTPM(ctx context.Context, in PlanInput) int {
	def := p.cachedInt(ctx, "ratelimit.user_default_tpm", 100000, "user_tpm_default")
	return scaleByGroup(def, in.GroupRatio)
}

// tokenRPM:override > 0 时用 override;否则取 user_default 但不应用 group ratio。
func (p *Planner) tokenRPM(ctx context.Context, in PlanInput) int {
	if in.TokenRPMOverride > 0 {
		return in.TokenRPMOverride
	}
	return p.cachedInt(ctx, "ratelimit.user_default_rpm", 60, "user_rpm_default")
}

func (p *Planner) tokenTPM(ctx context.Context, in PlanInput) int {
	if in.TokenTPMOverride > 0 {
		return in.TokenTPMOverride
	}
	return p.cachedInt(ctx, "ratelimit.user_default_tpm", 100000, "user_tpm_default")
}

func (p *Planner) ipRPM(ctx context.Context) int {
	return p.cachedInt(ctx, "ratelimit.ip_rpm", 600, "ip_rpm")
}

func (p *Planner) modelRPM(ctx context.Context, _ string) int {
	return p.cachedInt(ctx, "ratelimit.model_default_rpm", 0, "model_rpm_default")
}

func (p *Planner) modelTPM(ctx context.Context, _ string) int {
	return p.cachedInt(ctx, "ratelimit.model_default_tpm", 0, "model_tpm_default")
}

func (p *Planner) groupRPM(ctx context.Context, in PlanInput) int {
	def := p.cachedInt(ctx, "ratelimit.group_default_rpm", 0, "group_rpm_default")
	return scaleByGroup(def, in.GroupRatio)
}

func (p *Planner) groupTPM(ctx context.Context, in PlanInput) int {
	def := p.cachedInt(ctx, "ratelimit.group_default_tpm", 0, "group_tpm_default")
	return scaleByGroup(def, in.GroupRatio)
}

func (p *Planner) channelRPM(ctx context.Context) int {
	return p.cachedInt(ctx, "ratelimit.channel_default_rpm", 0, "channel_rpm_default")
}

func (p *Planner) channelTPM(ctx context.Context) int {
	return p.cachedInt(ctx, "ratelimit.channel_default_tpm", 0, "channel_tpm_default")
}

func (p *Planner) globalRPM(ctx context.Context) int {
	return p.cachedInt(ctx, "ratelimit.global_rpm", 0, "global_rpm")
}

func (p *Planner) globalTPM(ctx context.Context) int {
	return p.cachedInt(ctx, "ratelimit.global_tpm", 0, "global_tpm")
}

// cachedInt 先查本地缓存,miss 则查 setting.Store 并回填。
func (p *Planner) cachedInt(ctx context.Context, settingKey string, def int, cacheKey string) int {
	if v, ok := p.cache.get(cacheKey); ok {
		return v
	}
	v := def
	if p.setting != nil {
		v = p.setting.GetInt(ctx, settingKey, def)
	}
	p.cache.set(cacheKey, v, 0)
	return v
}

func (p *Planner) window(ctx context.Context) time.Duration {
	if p.windowFn != nil {
		return p.windowFn(ctx)
	}
	return time.Minute
}

func (p *Planner) windowFromSetting(ctx context.Context) time.Duration {
	if p.setting == nil {
		return time.Minute
	}
	sec := p.setting.GetInt(ctx, "ratelimit.window_seconds", 60)
	if sec <= 0 {
		sec = 60
	}
	return time.Duration(sec) * time.Second
}

// sanitizeModel 把 model 名里可能影响 redis key 阅读的字符替换掉。
// M1 仅做基本清理(空格 / 控制字符)。
func sanitizeModel(m string) string {
	m = strings.TrimSpace(m)
	// 替换空白与冒号(冒号会破坏 key 阅读;Redis 接受但不直观)
	repl := strings.NewReplacer(" ", "_", "\t", "_", "\n", "_", "\r", "_")
	return repl.Replace(m)
}
