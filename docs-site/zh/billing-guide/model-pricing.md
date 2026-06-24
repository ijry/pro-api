---
title: 模型价格表
outline: deep
---

# 模型价格表

> 本页展示 pro-api **默认配置** 的模型倍率。
>
> :::warning 仅供参考
> 具体部署方可能调过倍率,**以你登录的实例的 "系统设置 → 模型字典" 为准**。
>
> 倍率数字可能随上游公开价调整而过时,pro-api 在 minor 版本会更新默认值,详见 [CHANGELOG](/zh/changelog)。
> :::

## 倍率到价格的转换

```
价格(美元 / 1M tokens) = ratio × (1 / base_quota_per_dollar) × 1,000,000
                       = ratio × 2          (默认 base = 500000)
```

例:`ratio = 1.0` → `$2 / 1M tokens`;`ratio = 30` → `$60 / 1M tokens`。

如果部署方改了 `base_quota_per_dollar`,这个换算系数变。

## OpenAI 系列

| 模型 | input ratio | output ratio | cached ratio | reasoning ratio |
|---|---|---|---|---|
| gpt-4o | 1.25 | 5.0 | 0.625 | — |
| gpt-4o-mini | 0.075 | 0.3 | 0.0375 | — |
| gpt-4-turbo | 5.0 | 15.0 | — | — |
| gpt-3.5-turbo | 0.25 | 0.75 | — | — |
| o1 | 7.5 | 30.0 | 3.75 | 30.0 |
| o1-mini | 1.5 | 6.0 | 0.75 | 6.0 |
| o3-mini | 0.55 | 2.2 | 0.275 | 2.2 |
| text-embedding-3-small | 0.01 | — | — | — |
| text-embedding-3-large | 0.065 | — | — | — |

## Anthropic 系列

> M1 仅支持 **出向**(后端调用),Anthropic 入口在 M2。

| 模型 | input ratio | output ratio | cached ratio |
|---|---|---|---|
| claude-3.5-sonnet | 1.5 | 7.5 | 0.15 |
| claude-3.5-haiku | 0.4 | 2.0 | 0.04 |
| claude-3-haiku | 0.125 | 0.625 | 0.0125 |
| claude-3-opus | 7.5 | 37.5 | 0.75 |

## Gemini 系列

| 模型 | input ratio | output ratio |
|---|---|---|
| gemini-1.5-pro | 0.625 | 2.5 |
| gemini-1.5-flash | 0.0375 | 0.15 |
| gemini-2.0-flash | 0.05 | 0.2 |
| text-embedding-004 | 0.0 | — |

> Gemini 2.0 Flash 默认免费档,pro-api 仍按上游公开价的"按使用"倍率配置(防上游限免到期)。

## DeepSeek 系列

| 模型 | input ratio | output ratio | cached ratio | reasoning ratio |
|---|---|---|---|---|
| deepseek-chat | 0.07 | 0.14 | 0.014 | — |
| deepseek-reasoner | 0.275 | 1.1 | 0.0275 | 1.1 |

国产模型里性价比最高的之一,适合做 fallback 或主力。

## Moonshot / Kimi

| 模型 | input ratio | output ratio |
|---|---|---|
| moonshot-v1-8k | 0.6 | 0.6 |
| moonshot-v1-32k | 1.2 | 1.2 |
| moonshot-v1-128k | 3.0 | 3.0 |

> Moonshot 的 input/output 同价是它家的定价策略。

## 智谱 GLM

| 模型 | input ratio | output ratio |
|---|---|---|
| glm-4 | 0.5 | 0.5 |
| glm-4-flash | 0.0 | 0.0 |
| glm-4-air | 0.025 | 0.025 |
| glm-4-long | 0.05 | 0.05 |

`glm-4-flash` 默认免费档(2024 起),pro-api 默认 ratio 设 0,但建议部署方按市场策略加价。

## 通义千问

| 模型 | input ratio | output ratio |
|---|---|---|
| qwen-turbo | 0.15 | 0.3 |
| qwen-plus | 0.4 | 0.6 |
| qwen-max | 1.2 | 1.2 |

## 豆包

| 模型 | input ratio | output ratio |
|---|---|---|
| doubao-pro | 0.4 | 0.8 |
| doubao-lite | 0.15 | 0.3 |

## 倍率口径

- 上述倍率参考各家 **公开价的 USD 等价**,对照 OpenAI `gpt-4o` 的简单比例制定
- pro-api 部署方可基于市场策略调整(加成 / 让利 / 套餐)
- ratio 后台可改:**系统设置 → 模型字典 → 编辑**

## 模型名映射

不同上游对"同一模型"的命名可能不同,例:

- Anthropic 上游:`claude-3-5-sonnet-20241022`
- 用户友好名:`claude-3.5-sonnet`

pro-api 通过渠道的 **模型映射**(`client_model → upstream_model`)统一对外名,详见 [渠道管理](../modules/channel-management.md)。

## 更新频率

- **上游公开价改动** → pro-api 主仓库可能在 minor 版本里更新默认 ratio
- **大版本前** 看 [CHANGELOG](/zh/changelog) 的 `Changed: model pricing` 段
- **部署方自定义** → 改完不会被后续 pro-api 版本覆盖(只影响 seed 阶段)

## 关键要点

- **价格表数字可能过时**,务必以"系统设置 → 模型字典"为准
- 公式 `ratio × 2 = $/M tokens` 只在默认 `base_quota_per_dollar = 500000` 时成立
- **实际 `base_quota_per_dollar` 改动会影响所有用户**,部署方调整要慎重
- 不要在 README / 公开宣传里直接贴本页数字,标 "默认值,以实际为准"
