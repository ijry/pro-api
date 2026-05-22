package openai

import "encoding/json"

// ChatRequestDTO 是 OpenAI /v1/chat/completions 入口请求的 wire DTO。
type ChatRequestDTO struct {
	Model               string             `json:"model"`
	Messages            []MessageDTO       `json:"messages"`
	MaxTokens           *int               `json:"max_tokens,omitempty"`
	MaxCompletionTokens *int               `json:"max_completion_tokens,omitempty"` // o1 系列
	Temperature         *float64           `json:"temperature,omitempty"`
	TopP                *float64           `json:"top_p,omitempty"`
	Stream              bool               `json:"stream,omitempty"`
	StreamOptions       *StreamOptionsDTO  `json:"stream_options,omitempty"`
	Stop                json.RawMessage    `json:"stop,omitempty"` // string | []string
	Tools               []ToolDTO          `json:"tools,omitempty"`
	ToolChoice          json.RawMessage    `json:"tool_choice,omitempty"` // string | object
	ResponseFormat      *ResponseFormatDTO `json:"response_format,omitempty"`
	Seed                *int               `json:"seed,omitempty"`
	User                string             `json:"user,omitempty"`
}

// MessageDTO 是一条消息的 wire 表示。
type MessageDTO struct {
	Role       string          `json:"role"`
	Content    json.RawMessage `json:"content,omitempty"` // string | []ContentPart
	Name       string          `json:"name,omitempty"`
	ToolCallID string          `json:"tool_call_id,omitempty"`
	ToolCalls  []ToolCallDTO   `json:"tool_calls,omitempty"`
}

// ContentPartDTO 是多模态消息的一部分。
type ContentPartDTO struct {
	Type     string       `json:"type"`
	Text     string       `json:"text,omitempty"`
	ImageURL *ImageURLDTO `json:"image_url,omitempty"`
}

// ImageURLDTO 是图片 URL 的 wire 表示。
type ImageURLDTO struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"`
}

// ToolDTO 是工具声明的 wire 表示。
type ToolDTO struct {
	Type     string         `json:"type"`
	Function FunctionDefDTO `json:"function"`
}

// FunctionDefDTO 是 function 工具的定义。
type FunctionDefDTO struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters,omitempty"`
}

// ToolCallDTO 是 assistant 触发的 tool 调用 wire 表示。
type ToolCallDTO struct {
	ID       string          `json:"id,omitempty"`
	Type     string          `json:"type,omitempty"`
	Function FunctionCallDTO `json:"function"`
	Index    int             `json:"index,omitempty"` // streaming delta
}

// FunctionCallDTO 是 function call 的 wire 表示。
type FunctionCallDTO struct {
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
}

// ResponseFormatDTO 控制结构化输出。
type ResponseFormatDTO struct {
	Type       string         `json:"type"`
	JSONSchema map[string]any `json:"json_schema,omitempty"`
}

// StreamOptionsDTO 控制流式响应行为。
type StreamOptionsDTO struct {
	IncludeUsage bool `json:"include_usage,omitempty"`
}

// ChatResponseDTO 是 chat completion 响应的 wire 格式。
type ChatResponseDTO struct {
	ID                string      `json:"id"`
	Object            string      `json:"object"`
	Created           int64       `json:"created"`
	Model             string      `json:"model"`
	SystemFingerprint string      `json:"system_fingerprint,omitempty"`
	Choices           []ChoiceDTO `json:"choices"`
	Usage             UsageDTO    `json:"usage"`
}

// ChoiceDTO 是单条 candidate。
type ChoiceDTO struct {
	Index        int        `json:"index"`
	Message      MessageDTO `json:"message"`
	FinishReason string     `json:"finish_reason,omitempty"`
}

// UsageDTO 是 token 用量。
type UsageDTO struct {
	PromptTokens            int                         `json:"prompt_tokens"`
	CompletionTokens        int                         `json:"completion_tokens"`
	TotalTokens             int                         `json:"total_tokens"`
	PromptTokensDetails     *PromptTokensDetailsDTO     `json:"prompt_tokens_details,omitempty"`
	CompletionTokensDetails *CompletionTokensDetailsDTO `json:"completion_tokens_details,omitempty"`
}

// PromptTokensDetailsDTO 描述 prompt 内细节。
type PromptTokensDetailsDTO struct {
	CachedTokens int `json:"cached_tokens"`
}

// CompletionTokensDetailsDTO 描述 completion 内细节。
type CompletionTokensDetailsDTO struct {
	ReasoningTokens int `json:"reasoning_tokens"`
}

// ChatChunkDTO 是 SSE chunk 的 wire 格式。
type ChatChunkDTO struct {
	ID      string           `json:"id"`
	Object  string           `json:"object"`
	Created int64            `json:"created"`
	Model   string           `json:"model"`
	Choices []ChoiceChunkDTO `json:"choices"`
	Usage   *UsageDTO        `json:"usage,omitempty"`
}

// ChoiceChunkDTO 是流式 choice。
type ChoiceChunkDTO struct {
	Index        int      `json:"index"`
	Delta        DeltaDTO `json:"delta"`
	FinishReason string   `json:"finish_reason,omitempty"`
}

// DeltaDTO 是流式 delta。
type DeltaDTO struct {
	Role      string        `json:"role,omitempty"`
	Content   string        `json:"content,omitempty"`
	ToolCalls []ToolCallDTO `json:"tool_calls,omitempty"`
}

// EmbedRequestDTO 是 /v1/embeddings 入口请求。
type EmbedRequestDTO struct {
	Model          string          `json:"model"`
	Input          json.RawMessage `json:"input"`           // string | []string
	EncodingFormat string          `json:"encoding_format,omitempty"`
	User           string          `json:"user,omitempty"`
	Dimensions     *int            `json:"dimensions,omitempty"`
}

// EmbedResponseDTO 是 embedding 响应。
type EmbedResponseDTO struct {
	Object string         `json:"object"`
	Data   []EmbedDataDTO `json:"data"`
	Model  string         `json:"model"`
	Usage  EmbedUsageDTO  `json:"usage"`
}

// EmbedDataDTO 是单条 input 对应的向量。
type EmbedDataDTO struct {
	Object    string `json:"object"`
	Index     int    `json:"index"`
	Embedding any    `json:"embedding"` // []float32 或 base64 string
}

// EmbedUsageDTO 是 embedding 的用量。
type EmbedUsageDTO struct {
	PromptTokens int `json:"prompt_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// CompletionRequestDTO 是 legacy /v1/completions 入口请求。
type CompletionRequestDTO struct {
	Model       string          `json:"model"`
	Prompt      json.RawMessage `json:"prompt"` // string | []string
	MaxTokens   *int            `json:"max_tokens,omitempty"`
	Temperature *float64        `json:"temperature,omitempty"`
	TopP        *float64        `json:"top_p,omitempty"`
	Stream      bool            `json:"stream,omitempty"`
	Stop        json.RawMessage `json:"stop,omitempty"`
	User        string          `json:"user,omitempty"`
	Suffix      string          `json:"suffix,omitempty"`
	Echo        bool            `json:"echo,omitempty"`
	Logprobs    *int            `json:"logprobs,omitempty"`
}

// CompletionResponseDTO 是 legacy /v1/completions 响应。
type CompletionResponseDTO struct {
	ID      string                `json:"id"`
	Object  string                `json:"object"`
	Created int64                 `json:"created"`
	Model   string                `json:"model"`
	Choices []CompletionChoiceDTO `json:"choices"`
	Usage   UsageDTO              `json:"usage"`
}

// CompletionChoiceDTO 是 legacy completion 的 choice。
type CompletionChoiceDTO struct {
	Index        int    `json:"index"`
	Text         string `json:"text"`
	FinishReason string `json:"finish_reason,omitempty"`
	Logprobs     any    `json:"logprobs"`
}
