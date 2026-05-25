// Package relay 提供统一的 LLM 调用编排层。
//
// M1 版本：单渠道直调，不做重试/多渠道选择（M1-05 channel 增强）。
package relay

import (
	"context"
	"fmt"

	"github.com/ijry/pro-api/internal/adapter"
	"github.com/ijry/pro-api/internal/protocol/ir"
	"github.com/ijry/pro-api/pkg/apierr"
)

// Service 是 relay 编排服务。
// M1 直接持有 adapter.Registry，按 providerName 查找 adapter 调用。
type Service struct {
	registry adapter.Registry
}

// New 构造 relay.Service。
func New(reg adapter.Registry) *Service {
	return &Service{registry: reg}
}

// Chat 执行非流式 chat completion。
func (s *Service) Chat(ctx context.Context, req *ir.ChatRequest, cred adapter.Credential, providerName string) (*ir.ChatResponse, error) {
	a, ok := s.registry.Get(providerName)
	if !ok {
		return nil, fmt.Errorf("relay: unknown provider %q", providerName)
	}
	return a.Chat(ctx, req, cred)
}

// ChatStream 执行流式 chat completion。
func (s *Service) ChatStream(ctx context.Context, req *ir.ChatRequest, cred adapter.Credential, providerName string) (adapter.StreamReader, error) {
	a, ok := s.registry.Get(providerName)
	if !ok {
		return nil, fmt.Errorf("relay: unknown provider %q", providerName)
	}
	return a.ChatStream(ctx, req, cred)
}

// Embed 执行 embedding 请求。
func (s *Service) Embed(ctx context.Context, req *ir.EmbedRequest, cred adapter.Credential, providerName string) (*ir.EmbedResponse, error) {
	a, ok := s.registry.Get(providerName)
	if !ok {
		return nil, fmt.Errorf("relay: unknown provider %q", providerName)
	}
	return a.Embed(ctx, req, cred)
}

// GenerateImage 执行图片生成请求。
func (s *Service) GenerateImage(ctx context.Context, req *ir.ImageRequest, cred adapter.Credential, providerName string) (*ir.ImageResponse, error) {
	a, ok := s.registry.Get(providerName)
	if !ok {
		return nil, fmt.Errorf("relay: unknown provider %q", providerName)
	}
	ia, ok := a.(adapter.ImageAdapter)
	if !ok {
		return nil, apierr.New(apierr.CodeInvalidParam, "provider does not support image generation")
	}
	return ia.GenerateImage(ctx, req, cred)
}

// TextToSpeech 执行文字转语音请求。
func (s *Service) TextToSpeech(ctx context.Context, req *ir.SpeechRequest, cred adapter.Credential, providerName string) (*ir.SpeechResponse, error) {
	a, ok := s.registry.Get(providerName)
	if !ok {
		return nil, fmt.Errorf("relay: unknown provider %q", providerName)
	}
	sa, ok := a.(adapter.SpeechAdapter)
	if !ok {
		return nil, apierr.New(apierr.CodeInvalidParam, "provider does not support TTS")
	}
	return sa.TextToSpeech(ctx, req, cred)
}

// Transcribe 执行语音转文字请求。
func (s *Service) Transcribe(ctx context.Context, req *ir.TranscribeRequest, cred adapter.Credential, providerName string) (*ir.TranscribeResponse, error) {
	a, ok := s.registry.Get(providerName)
	if !ok {
		return nil, fmt.Errorf("relay: unknown provider %q", providerName)
	}
	ta, ok := a.(adapter.TranscribeAdapter)
	if !ok {
		return nil, apierr.New(apierr.CodeInvalidParam, "provider does not support transcription")
	}
	return ta.Transcribe(ctx, req, cred)
}

// Rerank 执行文档重排请求。
func (s *Service) Rerank(ctx context.Context, req *ir.RerankRequest, cred adapter.Credential, providerName string) (*ir.RerankResponse, error) {
	a, ok := s.registry.Get(providerName)
	if !ok {
		return nil, fmt.Errorf("relay: unknown provider %q", providerName)
	}
	ra, ok := a.(adapter.RerankAdapter)
	if !ok {
		return nil, apierr.New(apierr.CodeInvalidParam, "provider does not support rerank")
	}
	return ra.Rerank(ctx, req, cred)
}
