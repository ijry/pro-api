package ir

import "strings"

// CompletionRequest 是 legacy /v1/completions 请求的 IR。
type CompletionRequest struct {
	Model       string
	Prompt      []string
	MaxTokens   int
	Temperature *float64
	TopP        *float64
	Stream      bool
	Stop        []string
	User        string
	Suffix      string
	Echo        bool // M1 解析但不传递
	Logprobs    *int // M1 解析但不传递
	ExtraParams map[string]any
}

// ToChat 把 completion 请求转成等价的 chat 请求(单条 user 消息)。
func (r *CompletionRequest) ToChat() *ChatRequest {
	joined := strings.Join(r.Prompt, "\n")
	return &ChatRequest{
		Model:       r.Model,
		MaxTokens:   r.MaxTokens,
		Temperature: r.Temperature,
		TopP:        r.TopP,
		Stream:      r.Stream,
		Stop:        r.Stop,
		User:        r.User,
		ExtraParams: r.ExtraParams,
		Messages: []Message{
			{
				Role:    RoleUser,
				Content: []ContentPart{{Type: ContentText, Text: joined}},
			},
		},
	}
}
