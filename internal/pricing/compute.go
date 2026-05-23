package pricing

import (
	"context"
	"math"
)

// RatioFor 返回某模型/分组/渠道生效的倍率(给 Reserve 估算前的预算评估)。
//
// 顺序:
//  1. 模型字典默认值(NULL → cached=input、reasoning=output)
//  2. 渠道 ModelOverrides 覆盖(channel == nil 时跳过)
//  3. pricing_rules 匹配(精确 > 粗略)
//  4. 用户分组倍率(group_ratio)
func (s *service) RatioFor(ctx context.Context, model string, groupID int64, ch ChannelInfo) Ratios {
	// 1. 模型字典默认
	var cat Catalog
	if s.catalog != nil {
		cat = s.catalog.LookupCatalog(model)
	}
	in := cat.DefaultInputRatio
	out := cat.DefaultOutputRatio
	cached := cat.DefaultCachedRatio
	reasoning := cat.DefaultReasoningRatio
	if cached == 0 {
		cached = in
	}
	if reasoning == 0 {
		reasoning = out
	}

	// 2. 渠道 ModelOverride
	if ch != nil {
		if mo, ok := ch.ModelOverrideFor(model); ok {
			if mo.Input != 0 {
				in = mo.Input
			}
			if mo.Output != 0 {
				out = mo.Output
			}
			if mo.Cached != 0 {
				cached = mo.Cached
			}
			if mo.Reasoning != 0 {
				reasoning = mo.Reasoning
			}
		}
	}

	// 3. pricing_rules 匹配(单条规则覆盖,非 nil 字段覆盖前序值)
	if r := s.matchRule(model, groupID); r != nil {
		if r.InputRatio != nil {
			in = *r.InputRatio
		}
		if r.OutputRatio != nil {
			out = *r.OutputRatio
		}
		if r.CachedRatio != nil {
			cached = *r.CachedRatio
		}
		if r.ReasoningRatio != nil {
			reasoning = *r.ReasoningRatio
		}
	}

	// 4. 分组倍率
	gr := 1.0
	if s.groupRatio != nil {
		v := s.groupRatio(ctx, groupID)
		if v > 0 {
			gr = v
		}
	}

	return Ratios{
		Input:     in,
		Output:    out,
		Cached:    cached,
		Reasoning: reasoning,
		Group:     gr,
	}
}

// Compute 计算实际 quota 消耗。
//
// 公式:quota = ceil((in*r_in + out*r_out + cached*r_cached + reasoning*r_reasoning) * group_ratio)
// 任意 token > 0 时最小返回 1(避免 0.x 被向下吞没)。
func (s *service) Compute(ctx context.Context, in ComputeInput) ComputeResult {
	rt := s.RatioFor(ctx, in.Model, in.GroupID, in.Channel)
	rawIn := float64(in.InputTokens) * rt.Input
	rawOut := float64(in.OutputTokens) * rt.Output
	rawCached := float64(in.CachedTokens) * rt.Cached
	rawReasoning := float64(in.ReasoningTokens) * rt.Reasoning
	sum := (rawIn + rawOut + rawCached + rawReasoning) * rt.Group
	quota := int64(math.Ceil(sum))
	if quota == 0 && (in.InputTokens+in.OutputTokens+in.CachedTokens+in.ReasoningTokens) > 0 {
		quota = 1
	}
	return ComputeResult{Quota: quota, Ratios: rt}
}

// EstimateMax 用 max_tokens 估算最大可能消耗(用于 Reserve)。
// 估算时 channel 未知(尚未选渠道),传 nil。
func (s *service) EstimateMax(ctx context.Context, model string, in EstimateInput) int64 {
	rt := s.RatioFor(ctx, model, 0, nil)
	if in.BillingGroupRatio > 0 {
		rt.Group = in.BillingGroupRatio
	}
	maxOut := in.MaxOutTokens
	if maxOut <= 0 {
		maxOut = s.DefaultMaxOut(ctx, model)
	}
	raw := float64(in.InputTokens)*rt.Input + float64(maxOut)*rt.Output
	q := int64(math.Ceil(raw * rt.Group))
	if q == 0 && (in.InputTokens+maxOut) > 0 {
		q = 1
	}
	return q
}

// DefaultMaxOut 返回模型默认 max_out_tokens。
// 启发式:max_out 上限 = 0.5 * max_input,封顶 4096;无字典回退到 4096。
func (s *service) DefaultMaxOut(ctx context.Context, model string) int {
	if s.catalog != nil {
		cat := s.catalog.LookupCatalog(model)
		if cat.MaxInputTokens > 0 {
			v := cat.MaxInputTokens / 2
			if v > 4096 {
				v = 4096
			}
			return v
		}
	}
	return s.defaultMaxOut
}
