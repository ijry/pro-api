package ir

// 角色枚举。与 OpenAI chat completion 语义一致。
const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleTool      = "tool"
)

// ContentPart.Type 枚举。
const (
	ContentText       = "text"
	ContentImageURL   = "image_url"
	ContentToolUse    = "tool_use"    // assistant 内嵌 tool call(IR 中通常用 Message.ToolCalls)
	ContentToolResult = "tool_result" // tool message 的输出
)

// FinishReason 枚举。
const (
	FinishStop          = "stop"
	FinishLength        = "length"
	FinishToolCalls     = "tool_calls"
	FinishContentFilter = "content_filter"
	FinishError         = "error"
)
