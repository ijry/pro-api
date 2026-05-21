// Package tokenize 给不同模型族提供 token 计数。
//
// 主要用途:
//   - billing.EstimateMax 预估请求最大消耗
//   - 上游未返回 usage 时本地兜底计数
package tokenize

// Tokenizer 是一个具体的 token 计数算法。
//
// 实现要求:
//   - 同一 model + text 输入,Count 必须确定性
//   - 错误时返回 0 + error;调用方应当退化到 approximate
type Tokenizer interface {
	Count(model, text string) (int, error)
	CountMessages(model string, messages []Message) (int, error)
	Name() string
}

// Message 是 chat 消息的 token 计算载体。
type Message struct {
	Role    string
	Content string
	Name    string
}
