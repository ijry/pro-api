---
title: 渠道管理
outline: deep
---

# 渠道管理

> 渠道(Channel)= 一组上游凭证 + 模型映射 + 优先级权重。proapi 的"接什么上游"与"怎么挑上游"都在这里配。

## 渠道的字段

| 字段 | 类型 | 说明 |
|---|---|---|
| `name` | string | 渠道别名(运维识别用) |
| `provider` | enum | openai / azure / anthropic / gemini / deepseek / moonshot / zhipu / qwen / doubao |
| `base_url` | string | 上游 API 根 URL;不填用 provider 默认 |
| `credentials` | encrypted blob | api_key、token 等,DB 里看到的是 `ENC(v1, ...)` |
| `priority` | int | 越大越优先;不同 priority 互不混选 |
| `weight` | int | 同一 priority 内的加权随机权重 |
| `status` | enum | 0 active / 1 disabled |
| `tags` | string[] | 渠道标签,用于路由策略(M2 用) |
| `extra` | JSON | provider 特定字段,如 Azure 的 deployment 映射 |
| 模型映射 | 子表 | `client_model → upstream_model` + 可选 ratio 覆盖 |

## 新增渠道

`admin 后台 → 渠道管理 → 新增`,步骤如下。

### 选 provider

下拉 9 选 1。**不能改** —— 不同 provider 凭证字段不同,改了就乱了。需要换 provider 请新建。

### 填凭证

各家凭证字段:

| Provider | 字段 | 说明 |
|---|---|---|
| openai | api_key | sk-xxx |
| azure | api_key + region + deployment 映射 | 必须在 extra 配 client_model → deployment |
| anthropic | api_key | sk-ant-xxx |
| gemini | api_key | Google AI Studio key |
| deepseek | api_key | sk-xxx |
| moonshot | api_key | sk-xxx |
| zhipu | api_key | id.secret |
| qwen | api_key | sk-xxx |
| doubao | api_key | 火山引擎 ak/sk |

:::danger 凭证加密
所有 `credentials` 字段由应用层用 `master_key` 做 AES-256-GCM 加密后入库。**只有内存中持有 master_key 的 proapi 进程能解密。**
:::

### 配模型映射

```
client_model       upstream_model           ratio 覆盖(可选)
─────────────────  ───────────────────────  ────────────────────
claude-3.5-sonnet  claude-3-5-sonnet-20241022   input=1.5, output=7.5
gpt-4o             gpt-4o-2024-08-06            (用全局)
gpt-4o-mini        gpt-4o-mini                  (用全局)
```

作用:

1. **改名**:把上游的版本化模型名(`claude-3-5-sonnet-20241022`)映射成用户友好的稳定名(`claude-3.5-sonnet`)。
2. **一渠道一模型多别名**:`gpt-4o-2024-08-06` 同时挂 `gpt-4o` 与 `gpt-4o-latest`。
3. **不同渠道做不同映射**:Azure 渠道把 `gpt-4o` 映到 `gpt4o-deploy-east`。

### 配倍率覆盖(可选)

针对该渠道该模型,覆盖全局倍率:`input_ratio` / `output_ratio` / `cached_ratio` / `reasoning_ratio`。
不填则按 [倍率匹配优先级](./pricing.md#倍率匹配的优先级) 找上一层。

### 设优先级与权重

```
priority 越大越优先
weight   同 priority 内加权随机
```

完整算法见 [渠道调度与熔断](../architecture/channel-scheduling.md)。

## 编辑渠道

- 凭证可更新,系统自动重新加密
- 改 priority / weight 立即生效
- 改模型映射立即生效,但**已在路由中**的请求按旧映射完成
- provider 不可改

## 禁用 / 启用

```
status = 1 (disabled)
   ↓
该渠道立即停止接收新请求,熔断状态保留
   ↓
status = 0 (active)
   ↓
立即恢复
```

## 熔断状态查看

列表中显示 `breaker_state` 列:

- `CLOSED`(绿)— 正常
- `OPEN`(红)— 隔离中
- `HALF_OPEN`(黄)— 半开探针

可点击 **重置熔断** 强制回 CLOSED + 清零失败计数。

## 测试渠道

后台提供 **测试** 按钮:后端调一次最小请求(默认用 `provider` 的最便宜模型,如 OpenAI 用 `gpt-3.5-turbo`,Anthropic 用 `claude-3-haiku`)验证凭证正确性。

返回:

- ✅ 成功:HTTP 200 + 简短响应
- ❌ 失败:具体错误(401 凭证错 / 404 模型不存在 / 网络错 等)

## 凭证加密

DB `channels.credentials` 实际格式:

```
ENC(v1, base64(nonce) | base64(ciphertext) | base64(auth_tag))
```

加密用 `master_key`(AES-256-GCM)。换 master_key 必须**所有渠道凭证重新填**。

## 多渠道策略实例

### 主备

| name | priority | weight |
|---|---|---|
| openai-main | 10 | 1 |
| openai-backup | 5 | 1 |

`main` 出问题时才走 `backup`。

### 负载分摊(降低单 key 限流)

| name | priority | weight |
|---|---|---|
| azure-east | 10 | 3 |
| azure-west | 10 | 1 |

3:1 抽样,适合"主用东区 + 偶尔走西区"。

### 省钱优先

| name | priority | weight |
|---|---|---|
| deepseek | 20 | 1 |
| openai | 10 | 1 |

DeepSeek 先用(它便宜),失败 fallback OpenAI。

## 关键要点

- 模型映射的 `client_model` 是用户在请求里写的;`upstream_model` 是 proapi 发到上游的。
- 同一 provider 多渠道很常见(用多 key 提高 quota)。
- 凭证一旦换 master_key **全部得重填**(见 [配置](../guide/configuration.md#master-key))。
- 新增 provider 不在 9 家里?需要[新增上游适配器](../development/adapter-guide.md)。
