package pricing

import "strings"

// matchRule 按"最精确者胜 + priority 数字小者胜 + 同 priority id desc"顺序找规则。
//
// 顺序:group_model → model → group → global → catalog default(nil 表示无匹配)。
// 每个层级内按 (priority ASC, id DESC) 排;层级间按上面顺序。
func (s *service) matchRule(model string, groupID int64) *Rule {
	rules := s.cache.All()
	// 1. group_model
	if r := pickBest(rules, func(r *Rule) bool {
		return r.Scope == ScopeGroupModel &&
			r.GroupID != nil && *r.GroupID == groupID &&
			r.Model != nil && modelMatch(*r.Model, model)
	}); r != nil {
		return r
	}
	// 2. model
	if r := pickBest(rules, func(r *Rule) bool {
		return r.Scope == ScopeModel &&
			r.Model != nil && modelMatch(*r.Model, model)
	}); r != nil {
		return r
	}
	// 3. group
	if r := pickBest(rules, func(r *Rule) bool {
		return r.Scope == ScopeGroup &&
			r.GroupID != nil && *r.GroupID == groupID
	}); r != nil {
		return r
	}
	// 4. global
	return pickBest(rules, func(r *Rule) bool {
		return r.Scope == ScopeGlobal
	})
}

// pickBest 在 rules 中找第一个满足 pred 的"最优"规则。
// 优先级:priority ASC(数字小者胜),同 priority 取 id DESC(最新者胜)。
func pickBest(rs []*Rule, pred func(*Rule) bool) *Rule {
	var best *Rule
	for _, r := range rs {
		if r.Status != RuleStatusEnabled {
			continue
		}
		if !pred(r) {
			continue
		}
		if best == nil {
			best = r
			continue
		}
		if r.Priority < best.Priority ||
			(r.Priority == best.Priority && r.ID > best.ID) {
			best = r
		}
	}
	return best
}

// modelMatch 支持末尾 "*" 通配(与 token 模型白名单同语义)。
func modelMatch(pattern, model string) bool {
	if pattern == "*" {
		return true
	}
	if strings.HasSuffix(pattern, "*") {
		return strings.HasPrefix(model, strings.TrimSuffix(pattern, "*"))
	}
	return pattern == model
}
