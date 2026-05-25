// Package adapterreg 装配所有 LLM adapter 到 registry。
// 独立包以避免 internal/adapter 子包与父包之间的循环依赖。
package adapterreg

import (
	"github.com/ijry/pro-api/internal/adapter"
	"github.com/ijry/pro-api/internal/adapter/anthropic"
	"github.com/ijry/pro-api/internal/adapter/azure"
	"github.com/ijry/pro-api/internal/adapter/deepseek"
	"github.com/ijry/pro-api/internal/adapter/doubao"
	"github.com/ijry/pro-api/internal/adapter/gemini"
	"github.com/ijry/pro-api/internal/adapter/groq"
	"github.com/ijry/pro-api/internal/adapter/huggingface"
	"github.com/ijry/pro-api/internal/adapter/minimax"
	"github.com/ijry/pro-api/internal/adapter/mistral"
	"github.com/ijry/pro-api/internal/adapter/moonshot"
	oadapter "github.com/ijry/pro-api/internal/adapter/openai"
	"github.com/ijry/pro-api/internal/adapter/openrouter"
	"github.com/ijry/pro-api/internal/adapter/qwen"
	"github.com/ijry/pro-api/internal/adapter/tencent"
	"github.com/ijry/pro-api/internal/adapter/yi"
	"github.com/ijry/pro-api/internal/adapter/zhipu"
	"github.com/ijry/pro-api/internal/util/tokenize"
)

// WireAdapters 向 Registry 注册所有 16 家 adapter，并注册 tokenizers。
//
// 用法：
//
//	reg := adapter.NewRegistry()
//	adapterreg.WireAdapters(reg, app.Tokenize)
//	app.AdapterReg = reg
func WireAdapters(reg adapter.Registry, tokReg *tokenize.Registry) {
	// 注册 tokenizers
	adapter.RegisterTokenizers(tokReg)

	// 原始 9 家 adapter
	reg.Register(oadapter.New(""))
	reg.Register(azure.New())
	reg.Register(anthropic.New())
	reg.Register(gemini.New())
	reg.Register(deepseek.New())
	reg.Register(moonshot.New())
	reg.Register(zhipu.New())
	reg.Register(qwen.New())
	reg.Register(doubao.New())

	// M2a 新增 8 家 adapter
	reg.Register(groq.New())
	reg.Register(mistral.New())
	reg.Register(yi.New())
	reg.Register(openrouter.New())
	reg.Register(huggingface.New())
	reg.Register(minimax.New())
	reg.Register(tencent.New())
}
