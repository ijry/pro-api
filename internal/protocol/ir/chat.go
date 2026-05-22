package ir

// ChatRequest 是中间表示的 chat completion 请求。
type ChatRequest struct {
	Model          string
	Messages       []Message
	MaxTokens      int
	Temperature    *float64
	TopP           *float64
	Stream         bool
	Stop           []string
	Seed           *int
	User           string
	Tools          []Tool
	ToolChoice     any // string | map[string]any
	ResponseFormat *ResponseFormat
	ExtraParams    map[string]any
}

// Message 是会话中的一条消息。
type Message struct {
	Role       string
	Content    []ContentPart
	Name       string
	ToolCallID string
	ToolCalls  []ToolCall // assistant 发起的 tool 调用
}

// ContentPart 是消息的一部分(多模态)。
type ContentPart struct {
	Type     string
	Text     string
	ImageURL ImageURL
}

// ImageURL 描述图片输入。
type ImageURL struct {
	URL    string
	Detail string // low / high / auto
}

// Tool 是模型可调用的工具声明。
type Tool struct {
	Type     string // 通常为 "function"
	Function FunctionTool
}

// FunctionTool 是 function 工具的定义。
type FunctionTool struct {
	Name        string
	Description string
	Parameters  map[string]any
}

// ToolCall 是 assistant 触发的一次工具调用。
type ToolCall struct {
	ID       string
	Type     string // 通常为 "function"
	Function ToolCallFunction
}

// ToolCallFunction 描述函数调用的目标与参数 JSON。
type ToolCallFunction struct {
	Name      string
	Arguments string // raw JSON string
}

// ResponseFormat 控制结构化输出。
type ResponseFormat struct {
	Type   string // "text" / "json_object" / "json_schema"
	Schema map[string]any
}
