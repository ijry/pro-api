// Package openai 实现 OpenAI 入口协议的编解码:HTTP body ↔ IR。
//
// 入口协议的 DTO(internal/protocol/openai/dto.go)与 IR(internal/protocol/ir)解耦,
// 允许 M2 加 Anthropic / Gemini 入口协议时各自有自己的 DTO。
package openai
