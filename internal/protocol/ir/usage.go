package ir

// Usage 是 chat completion 的 token 用量。
type Usage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
	CachedTokens     int // prompt 命中缓存的 token
	ReasoningTokens  int // o1 / deepseek-reasoner 思维链 token
}
