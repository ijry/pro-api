package ir

// ImageRequest 是图片生成请求。
type ImageRequest struct {
	Model          string
	Prompt         string
	N              int    // 生成数量，默认 1
	Size           string // "256x256" / "512x512" / "1024x1024" / "1792x1024" / "1024x1792"
	Quality        string // "standard" / "hd"
	Style          string // "vivid" / "natural"
	ResponseFormat string // "url" / "b64_json"
	User           string
}

// ImageData 是单张生成图片。
type ImageData struct {
	URL           string
	B64JSON       string
	RevisedPrompt string
}

// ImageResponse 是图片生成响应。
type ImageResponse struct {
	Created int64
	Data    []ImageData
}

// SpeechRequest 是文字转语音请求。
type SpeechRequest struct {
	Model          string
	Input          string
	Voice          string  // "alloy" / "echo" / "fable" / "onyx" / "nova" / "shimmer"
	ResponseFormat string  // "mp3" / "opus" / "aac" / "flac" / "wav" / "pcm"
	Speed          float64 // 0.25 - 4.0，默认 1.0
}

// SpeechResponse 是语音二进制响应。
type SpeechResponse struct {
	ContentType string
	Data        []byte
}

// TranscribeRequest 是语音转文字请求。
type TranscribeRequest struct {
	Model          string
	Audio          []byte // 音频原始字节
	Filename       string // 文件名(含后缀)，用于 MIME 推断
	Language       string // ISO-639-1 语言代码
	Prompt         string
	ResponseFormat string  // "json" / "text" / "srt" / "verbose_json" / "vtt"
	Temperature    float64
}

// TranscribeResponse 是转写结果。
type TranscribeResponse struct {
	Text     string
	Language string
	Duration float64
}

// RerankRequest 是文档重排请求。
type RerankRequest struct {
	Model     string
	Query     string
	Documents []string
	TopN      int
}

// RerankResult 是单个文档的重排结果。
type RerankResult struct {
	Index          int
	RelevanceScore float64
	Document       string
}

// RerankResponse 是重排响应。
type RerankResponse struct {
	Results []RerankResult
	Usage   Usage
}
