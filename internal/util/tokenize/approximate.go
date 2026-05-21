package tokenize

import "fmt"

// Approximate 以字节数 / divisor 估算 token。仅用作 fallback。
type Approximate struct {
	divisor int
}

// NewApproximate 构造 Approximate。divisor 必须 > 0。
func NewApproximate(divisor int) *Approximate {
	if divisor <= 0 {
		panic(fmt.Sprintf("tokenize: divisor must be > 0, got %d", divisor))
	}
	return &Approximate{divisor: divisor}
}

// Count 返回 ceil(len(text) / divisor)。
func (a *Approximate) Count(_, text string) (int, error) {
	n := len(text)
	if n == 0 {
		return 0, nil
	}
	return (n + a.divisor - 1) / a.divisor, nil
}

// CountMessages 对每条消息算 content token + 4 token 的 role overhead 近似。
func (a *Approximate) CountMessages(model string, messages []Message) (int, error) {
	total := 0
	for _, m := range messages {
		c, err := a.Count(model, m.Content)
		if err != nil {
			return 0, err
		}
		total += c + 4
	}
	return total, nil
}

// Name 返回 tokenizer 名称。
func (a *Approximate) Name() string { return "approximate" }
