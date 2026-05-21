package tokenize

import "strings"

// Registry 按模型名查 Tokenizer。匹配规则:精确 > 通配(末尾 *)> fallback。
type Registry struct {
	exact    map[string]Tokenizer
	wildcard []wildcardEntry
	fallback Tokenizer
}

type wildcardEntry struct {
	prefix    string
	tokenizer Tokenizer
}

// NewRegistry 构造空注册表。fallback 不能为 nil。
func NewRegistry(fallback Tokenizer) *Registry {
	if fallback == nil {
		panic("tokenize: fallback must not be nil")
	}
	return &Registry{
		exact:    map[string]Tokenizer{},
		fallback: fallback,
	}
}

// Register 注册模型 pattern → tokenizer。pattern 末尾 "*" 是通配。
func (r *Registry) Register(pattern string, t Tokenizer) {
	if strings.HasSuffix(pattern, "*") {
		r.wildcard = append(r.wildcard, wildcardEntry{
			prefix:    strings.TrimSuffix(pattern, "*"),
			tokenizer: t,
		})
		return
	}
	r.exact[pattern] = t
}

// For 返回 model 对应的 tokenizer。精确优先 → 最长通配 → fallback。
func (r *Registry) For(model string) Tokenizer {
	if t, ok := r.exact[model]; ok {
		return t
	}
	var best wildcardEntry
	bestLen := -1
	for _, e := range r.wildcard {
		if strings.HasPrefix(model, e.prefix) && len(e.prefix) > bestLen {
			best = e
			bestLen = len(e.prefix)
		}
	}
	if bestLen >= 0 {
		return best.tokenizer
	}
	return r.fallback
}

// NewDefaultRegistry 构造默认注册表:
//
//   - tiktoken("cl100k_base") 给 "gpt-*" / "o*"
//   - approximate(4) 作为 fallback
func NewDefaultRegistry() (*Registry, error) {
	r := NewRegistry(NewApproximate(4))
	cl, err := NewTiktoken("cl100k_base")
	if err != nil {
		return nil, err
	}
	r.Register("gpt-*", cl)
	r.Register("o*", cl)
	return r, nil
}
