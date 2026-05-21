package token

import "context"

// ModelLister 是 /v1/models handler 对渠道模块的最小依赖。
// 由 M1-05 渠道调度实现并通过 application 注入。
//
// 设计稿:M1-03 spec §4.8。
type ModelLister interface {
	// ActiveModels 返回当前所有可调用的 client_model(按字典序去重)。
	// 实现需聚合所有 status=enabled 且未熔断渠道的 channel_model_mappings.client_model。
	ActiveModels(ctx context.Context) []string

	// ModelInfo 返回模型元信息(created / owned_by);未知模型返回 ok=false。
	ModelInfo(model string) (ModelMeta, bool)
}

// ModelMeta 是 /v1/models 响应中 model object 的元信息。
type ModelMeta struct {
	// Created 是模型上线时间(unix seconds)。未知时 0。
	Created int64
	// OwnedBy 是模型所有者标签,如 "openai" / "anthropic"。
	OwnedBy string
}
