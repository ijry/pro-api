package adapter

import (
	"github.com/ijry/pro-api/internal/util/tokenize"
)

// RegisterTokenizers 在 wire 时调用,把 9 家 adapter 的 tokenize 规则灌到 registry。
//
// M1 决策:仅 OpenAI 系列用真 tiktoken,其余字符 ÷ 4 近似(详见 spec §4.9 / §11.13)。
//
//   - 对于无法构造的 tiktoken encoding,跳过该 pattern(用 registry 的 fallback)。
//   - 该函数不返回错误(失败仅打 warn 由调用者决定);取 tiktoken 失败时 fallback 到 approximate(4)。
func RegisterTokenizers(reg *tokenize.Registry) {
	if reg == nil {
		return
	}
	approx := tokenize.NewApproximate(4)
	gpt4o, err := tokenize.NewTiktoken("o200k_base")
	if err != nil {
		gpt4o = nil
	}
	gpt4, err := tokenize.NewTiktoken("cl100k_base")
	if err != nil {
		gpt4 = nil
	}
	pick := func(tk *tokenize.Tiktoken) tokenize.Tokenizer {
		if tk == nil {
			return approx
		}
		return tk
	}

	// OpenAI 系列
	reg.Register("gpt-4o*", pick(gpt4o))
	reg.Register("o1*", pick(gpt4o))
	reg.Register("o3*", pick(gpt4o))
	reg.Register("gpt-4*", pick(gpt4))
	reg.Register("gpt-3.5*", pick(gpt4))
	reg.Register("text-embedding-3*", pick(gpt4o))
	reg.Register("text-embedding-ada-002", pick(gpt4))

	// Anthropic / Gemini(近似)
	reg.Register("claude-*", approx)
	reg.Register("gemini-*", approx)
	reg.Register("text-embedding-004", approx)

	// DeepSeek / Moonshot(用 cl100k 近似)
	reg.Register("deepseek-*", pick(gpt4))
	reg.Register("moonshot-*", pick(gpt4))

	// Zhipu / Qwen / Doubao
	reg.Register("glm-*", pick(gpt4))
	reg.Register("embedding-3", pick(gpt4))
	reg.Register("qwen-*", pick(gpt4))
	reg.Register("text-embedding-v*", pick(gpt4))
	reg.Register("doubao-*", pick(gpt4))
	reg.Register("ep-*", pick(gpt4))
}
