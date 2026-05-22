package ir

// ChatResponse 是非流式 chat completion 响应 IR。
type ChatResponse struct {
	ID                string
	Model             string
	SystemFingerprint string
	Choices           []Choice
	Usage             Usage
}

// Choice 是单条 candidate。
type Choice struct {
	Index        int
	Message      Message
	FinishReason string
}
