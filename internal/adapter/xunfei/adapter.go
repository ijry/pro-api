// Package xunfei provides an iFlytek Spark (讯飞星火) adapter (OpenAI-compatible).
package xunfei

import (
	"context"

	"github.com/ijry/pro-api/internal/adapter"
	oadapter "github.com/ijry/pro-api/internal/adapter/openai"
	"github.com/ijry/pro-api/internal/protocol/ir"
)

// supportedModels 列出讯飞星火 OpenAI 兼容端点的常用模型。
// 鉴权使用控制台的 APIPassword 作为 Bearer token(填入渠道 credential 的 API key)。
var supportedModels = []string{
	"4.0Ultra",
	"generalv3.5",
	"max-32k",
	"generalv3",
	"pro-128k",
	"lite",
}

// Adapter 复用 OpenAI 基座,base URL 指向讯飞星火兼容端点(基座追加 /v1)。
type Adapter struct{ base *oadapter.OpenAI }

// New 构造讯飞星火适配器。
func New() adapter.Adapter { return &Adapter{base: oadapter.New("https://spark-api-open.xf-yun.com")} }

func (a *Adapter) Name() string                     { return "xunfei" }
func (a *Adapter) Capabilities() adapter.Capability { return a.base.Capabilities() }
func (a *Adapter) SupportedModels() []string        { return supportedModels }
func (a *Adapter) Chat(ctx context.Context, req *ir.ChatRequest, cred adapter.Credential) (*ir.ChatResponse, error) {
	return a.base.Chat(ctx, req, cred)
}
func (a *Adapter) ChatStream(ctx context.Context, req *ir.ChatRequest, cred adapter.Credential) (adapter.StreamReader, error) {
	return a.base.ChatStream(ctx, req, cred)
}
func (a *Adapter) Embed(ctx context.Context, req *ir.EmbedRequest, cred adapter.Credential) (*ir.EmbedResponse, error) {
	return a.base.Embed(ctx, req, cred)
}
