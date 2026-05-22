package ir

// ChatChunk 是流式响应的一段增量(SSE 中的单帧)。
type ChatChunk struct {
	ID           string
	Model        string
	Delta        Delta
	FinishReason string
	Usage        *Usage // 通常仅在末尾 chunk 出现
}

// Delta 是增量内容。
type Delta struct {
	Role      string
	Content   string
	ToolCalls []ToolCallDelta
}

// ToolCallDelta 是 tool call 的增量片段。
type ToolCallDelta struct {
	Index    int
	ID       string
	Function ToolCallFunctionDelta
}

// ToolCallFunctionDelta 是 function 字段的增量。
type ToolCallFunctionDelta struct {
	Name      string
	Arguments string // partial JSON
}
