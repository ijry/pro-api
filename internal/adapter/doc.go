// Package adapter 定义上游 LLM 提供商的统一接入接口与共用工具。
//
// 设计要点(详见 docs/superpowers/specs/2026-05-21-里程碑1-04-适配器层-设计稿.md):
//   - 9 家 adapter 实现统一 Adapter 接口
//   - relay 层通过 Registry 按 provider 名查 adapter
//   - 共用 HTTP client、SSE 行读取器、错误归类与 tokenizer 注册
package adapter
