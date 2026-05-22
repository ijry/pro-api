package ir

// EmbedRequest 是 embedding 请求 IR。
type EmbedRequest struct {
	Model          string
	Input          []string
	EncodingFormat string // "float" | "base64"
	User           string
	Dimensions     *int
	ExtraParams    map[string]any
}

// EmbedResponse 是 embedding 响应 IR。
type EmbedResponse struct {
	Model string
	Data  []EmbedData
	Usage EmbedUsage
}

// EmbedData 是单条 input 对应的向量。
type EmbedData struct {
	Index        int
	Embedding    []float32
	EmbeddingB64 string
}

// EmbedUsage 是 embedding 的 token 使用量。
type EmbedUsage struct {
	PromptTokens int
	TotalTokens  int
}
