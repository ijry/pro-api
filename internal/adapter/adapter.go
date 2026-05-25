package adapter

import (
	"context"

	"github.com/ijry/pro-api/internal/protocol/ir"
)

// Capability 是 adapter 支持的能力位。
type Capability uint32

const (
	CapChat Capability = 1 << iota
	CapStream
	CapCompletion
	CapEmbedding
	CapVision
	CapToolUse
	CapReasoning // o1 / deepseek-reasoner / 思维链
	CapImage     // DALL-E 图片生成
	CapTTS       // 文字转语音
	CapSTT       // 语音转文字
	CapRerank    // 文档重排
)

// Has 判断是否包含某项能力。
func (c Capability) Has(x Capability) bool { return c&x != 0 }

// Credential 是 adapter 向上游发请求所需的认证信息。
//
// Extra 用于承载 provider 私有字段(如 azure deployments / api_version)。
type Credential struct {
	APIKey  string
	Secret  string
	Region  string
	BaseURL string
	Extra   map[string]string
}

// Adapter 是单家上游 LLM 提供商的统一接入接口。
//
// 实现要求:
//   - 所有方法必须用 ctx 控制超时(http.NewRequestWithContext)
//   - 错误必须归类为 *apierr.Error(通过 ClassifyHTTPStatus / ClassifyNetErr)
//   - Chat 与 ChatStream 共用一份 request 序列化路径(避免分支不一致)
//   - 不打日志(由 relay 层负责),仅 zap.Debug 级 trace
//   - 不 panic — 用 recover 兜底转 CodeInternal
type Adapter interface {
	Name() string
	Capabilities() Capability
	SupportedModels() []string

	Chat(ctx context.Context, req *ir.ChatRequest, cred Credential) (*ir.ChatResponse, error)
	ChatStream(ctx context.Context, req *ir.ChatRequest, cred Credential) (StreamReader, error)
	Embed(ctx context.Context, req *ir.EmbedRequest, cred Credential) (*ir.EmbedResponse, error)
}

// StreamReader 是流式响应的迭代接口。
//
// 调用方约定:
//   - Next 返回 io.EOF 表示流末
//   - Close 必须由调用方 defer 调用(避免 HTTP 连接泄漏)
//   - Next 在 ctx 取消时返回 CodeUpstreamTimeout 错误
type StreamReader interface {
	Next(ctx context.Context) (*ir.ChatChunk, error)
	Close() error
}

// ImageAdapter 支持图片生成的 adapter 可选接口。
type ImageAdapter interface {
	GenerateImage(ctx context.Context, req *ir.ImageRequest, cred Credential) (*ir.ImageResponse, error)
}

// SpeechAdapter 支持文字转语音的 adapter 可选接口。
type SpeechAdapter interface {
	TextToSpeech(ctx context.Context, req *ir.SpeechRequest, cred Credential) (*ir.SpeechResponse, error)
}

// TranscribeAdapter 支持语音转文字的 adapter 可选接口。
type TranscribeAdapter interface {
	Transcribe(ctx context.Context, req *ir.TranscribeRequest, cred Credential) (*ir.TranscribeResponse, error)
}

// RerankAdapter 支持文档重排的 adapter 可选接口。
type RerankAdapter interface {
	Rerank(ctx context.Context, req *ir.RerankRequest, cred Credential) (*ir.RerankResponse, error)
}
