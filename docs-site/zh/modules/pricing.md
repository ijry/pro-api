---
title: 定价与倍率
outline: deep
---

# 定价与倍率

> pro-api 通过 `tokens × ratio × group_ratio` 公式把上游 token 用量转换成内部 quota 消耗。本页讲清楚倍率是怎么算的、在哪改。

## 公式

```
quota = ( input_tokens     × input_ratio
        + output_tokens    × output_ratio
        + cached_tokens    × cached_ratio
        + reasoning_tokens × reasoning_ratio )
      × group_ratio
```

举例:用户在 `default` 分组(`group_ratio = 1.0`),用 `gpt-4o`(`input_ratio=1.25, output_ratio=5.0`),
输入 1000 token、输出 500 token:

```
quota = (1000 × 1.25 + 500 × 5.0) × 1.0
      = (1250 + 2500) × 1.0
      = 3750 quota
      = $0.0075 (按默认 base_quota_per_dollar=500000)
```

## 4 个 token 维度

| 维度 | 含义 | 何时有值 |
|---|---|---|
| **input** | 输入 tokens | 所有请求都有 |
| **output** | 输出 tokens | chat / completions 有;embeddings 没有 |
| **cached** | 命中缓存的 input | OpenAI prompt caching 命中时 |
| **reasoning** | 推理 tokens | o1 / o3 / deepseek-reasoner 等专用 |

每个维度有独立 ratio。上游不返回的维度按 0 算(不收钱)。

## 倍率匹配的优先级

`pricing.Match(user_group, model, channel)` 从精确到通用 7 层匹配:

| 层 | 来源 | 作用域 |
|---|---|---|
| 1 | `channel_model_mappings`(渠道-模型) | 仅该渠道 |
| 2 | `pricing_rules` scope=`group_model` | 用户分组 + 模型 |
| 3 | `pricing_rules` scope=`model` | 全局对该模型 |
| 4 | `pricing_rules` scope=`group` | 全局对该分组 |
| 5 | `pricing_rules` scope=`global` | 系统全局 |
| 6 | `model_catalogs.default_*_ratio` | 模型字典默认值 |
| 7 | `1.0` | 兜底 |

> 实际计费时,4 个 token 维度**独立匹配**:可能 `input_ratio` 走第 6 层默认值,`cached_ratio` 走第 2 层规则。

## 分组倍率 group_ratio

`group_ratio` 是**独立一层**,与 4 个 token ratio **相乘**:

```
default → 1.0(原价)
vip     → 0.8(8 折)
svip    → 0.6(6 折)
```

适合"会员制 / 阶梯折扣 / 部门内部价"等场景。

## 配倍率的方式

实际部署中你有 4 种方式调倍率,按"影响面"从大到小:

### 方式一:模型字典(全局默认)

后台 → **模型字典** → 编辑某模型 → 改 `default_input_ratio` 等。

影响所有用户、所有渠道,除非有更精确的规则覆盖。

### 方式二:pricing_rules

后台 → **定价规则** → 新增,选 scope + 填倍率。

适合"我要给 gpt-4o 全场加 20% 利润":新增 scope=`model` + model=`gpt-4o` + 修改 ratio。

### 方式三:渠道级覆盖

某渠道的"模型映射"里改某模型的 ratio。**仅影响走该渠道的请求**。

适合"我有一个便宜的备用渠道,想用更低的卖价吸引用户"。

### 方式四:分组倍率

`user_groups.ratio` 字段。直接影响该分组所有用户的所有请求。

## 实例

### 例 1:gpt-4o 加 1.2 倍利润

新增 pricing_rule:

```
scope: model
model: gpt-4o
input_ratio: 1.5     -- 默认 1.25 → 1.5(加 20%)
output_ratio: 6.0    -- 默认 5.0 → 6.0
```

### 例 2:VIP 8 折

不需要新增 pricing_rule,直接改 `user_groups`:

```sql
UPDATE user_groups SET ratio = 0.8 WHERE name = 'vip';
```

或后台 → 用户分组 → vip → 改 ratio 0.8。

### 例 3:某 OpenAI 渠道走特别价

在该渠道的模型映射里:

```
client_model: gpt-4o
upstream_model: gpt-4o-2024-08-06
input_ratio: 0.5      -- 该渠道便宜一半
output_ratio: 2.0
```

只对走该渠道的请求生效;同一用户走其他渠道仍用默认 ratio。

## quota 与人民币 / 美元

```
1 美元 = base_quota_per_dollar quota
        默认 500000
        → 1 quota = $0.000002 = 0.0002 美分
```

`system_settings.pricing.base_quota_per_dollar` 可改;改它**影响全平台所有计费**,务必谨慎。

人民币换算:

```
1 元 = base_quota_per_dollar / exchange_rate_cny_per_usd quota
     = 500000 / 7.2 ≈ 69444 quota
```

`exchange_rate_cny_per_usd` 也是运行时配置,影响人民币充值入账。

## 关键要点

- 4 个 token 维度的 ratio 都是 **per-token 乘数**,不是百分比。
- `cached_ratio` 默认 `NULL`,意为"按 input 算";建议显式设为 `0.5 × input_ratio`(与 OpenAI 实际定价对齐)。
- 渠道级 ratio 覆盖只对该渠道生效 —— 一个用户调多个渠道时实际消费可能不同(后台日志会标 channel_id 与倍率快照)。
- 改 `base_quota_per_dollar` 会让历史日志的"美元等价"看起来变 —— 不影响已 commit 的 quota 数字。
