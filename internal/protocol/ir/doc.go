// Package ir 定义"中间表示"(Intermediate Representation),
// 在入口协议(OpenAI HTTP / SSE)与各上游 adapter 之间架起统一桥梁。
//
// IR 设计原则:
//   - 字段语义贴近 OpenAI;Anthropic/Gemini 等差异由各 adapter 内部翻译
//   - 不暴露 HTTP / JSON tag;DTO 在 internal/protocol/<入口协议> 中维护
//   - 字段可演化:加新能力(如多模态)时不破坏既有 adapter
package ir
