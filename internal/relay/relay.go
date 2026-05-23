// Package relay 提供统一的 LLM 调用编排层。
//
// M1 版本：单渠道直调，不做重试/多渠道选择（M1-05 channel 增强）。
package relay

import (
	"context"
	"fmt"

	"github.com/ijry/pro-api/internal/adapter"
	"github.com/ijry/pro-api/internal/protocol/ir"
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
