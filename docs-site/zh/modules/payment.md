---
title: 充值方式
outline: deep
---

# 充值方式

> M1 版本支持两种充值方式:**手动充值申请**(用户提单 → 管理员审批)与 **兑换码**(管理员预先批量生成,用户兑换)。
>
> 在线支付(Stripe / 支付宝 / 微信支付)在 M2 加入。

## 手动充值申请

### 用户视角

1. 用户前台 → **充值** → **手动充值申请**
2. 填表单:
   - 申请金额(CNY 或 USD)
   - 备注(贴转账凭证 URL / 订单号 / 转账截图链接)
3. 提交 → 状态 `pending`,等待管理员审批
4. 审批通过 → 钱包余额自动入账,会有公告通知;审批驳回 → `rejected`,可查看驳回原因
5. `pending` 状态下用户可**自主取消**申请

### 管理员视角

1. admin 后台 → **充值审批**
2. 列表展示 `pending` 申请,按时间排序
3. 点击"审批"→ 填:
   - 实际入账金额(可与申请金额不同,例:汇率波动)
   - 备注
   - 选择 通过 / 驳回
4. 入账或驳回 → 状态变更 + 审计日志记录(谁、什么时间、做了什么)

### 流程图

```
[用户提交]
    │ status=pending
    ▼
[管理员审批]
    │
    ├── approve → wallet.Credit + ledger 入账 → status=approved
    ├── reject  → status=rejected
    └── (用户提前取消) → status=canceled
```

### 字段说明

| 字段 | 含义 |
|---|---|
| `id` | 申请单 ID(Snowflake) |
| `user_id` | 申请人 |
| `amount_money` | 申请金额(单位:厘,1 元 = 100 厘 = 1000 厘?见下) |
| `currency` | CNY / USD |
| `amount_quota` | 审批时按当时汇率换算的 quota |
| `status` | 0 pending / 1 approved / 2 rejected / 3 canceled |
| `applicant_note` | 用户备注 |
| `reviewer_id` / `review_note` / `reviewed_at` | 管理员审批信息 |
| `created_at` / `updated_at` | |

> `amount_money` 用 **分**(¥)/ **美分**($)作为最小单位整数,避免浮点精度问题。

### 限额

| 限额 | 配置 key | 默认 |
|---|---|---|
| 单次最小金额(元) | `manual_recharge.min_amount_cny` | 100 |
| 单次最大金额(元) | `manual_recharge.max_amount_cny` | 100000 |
| 是否开放 | `manual_recharge.enabled` | true |

## 兑换码

### 用途

- 给员工发福利
- 给老用户回馈 / 投诉补偿
- 线下营销活动批量发放
- 临时邀请码(配合"邀请激活"功能)

### 管理员生成

1. admin 后台 → **兑换码** → **批量生成**
2. 填:
   - 批次号(自动生成或手填,如 `2026Q2-promo`)
   - 数量(1 - 10000)
   - 单码额度(quota,或换算的元)
   - 有效期(默认 `redeem.default_expires_days = 365` 天)
3. 提交 → 生成后**下载 CSV**(只在生成时**完整明文展示一次**,DB 存 sha256)

```csv
batch_no,code,quota,expires_at
2026Q2-promo,ABCD-EFGH-IJKL-MNPQ,1000000,2027-05-22
2026Q2-promo,RSTU-VWXY-Z234-5678,1000000,2027-05-22
...
```

### 用户兑换

1. 用户前台 → **充值** → **兑换码**
2. 输入 `XXXX-XXXX-XXXX-XXXX` 16 位 → 提交
3. 系统校验:
   - 存在?(查 `sha256(code)`)
   - 未使用?
   - 未过期?
   - 未禁用?
4. 全部通过 → `wallet.Credit` 入账 + ledger 一行 + 兑换码 `status=used`

### 兑换码字符串

- 格式:`XXXX-XXXX-XXXX-XXXX`(4 段 × 4 字符,共 16 个字符)
- 字符集:大写字母 + 数字,**排除** `0 O I L 1`(易混淆)
- DB 存 `sha256(code)`,后台只展示前缀 `ABCD-****-****-****`

### 状态

| status | 含义 |
|---|---|
| 0 | unused(可兑换) |
| 1 | used(已兑换) |
| 2 | disabled(管理员手动禁用) |

### 批次管理

后台 → **兑换码** → 列表,可:

- 按 `batch_no` 筛选
- 批次内统计:已用 / 未用 / 已禁用
- 一键 **禁用整批**(批次被泄露时应急)

## 钱包账本

每次充值(无论手动还是兑换码)都会在 `ledger_entries` 表记一行:

```
direction: credit         -- 入账
amount:    1000000        -- quota
ref_type:  manual | redeem
ref_id:    <申请单 ID 或兑换码 ID>
balance_after: <入账后余额>
created_at: ...
```

详见 [计费机制](../architecture/billing.md)。

## 退款

M1 阶段 pro-api **不实现"主动退款"**:

- 管理员不能通过后台对历史请求做主动退款
- 仅在以下情况自动退预扣:
  - 上游 5xx / timeout / 网络错
  - 模型不存在 / 凭证失效
  - 客户端 4xx 请求错(不收钱)

M2 加入 admin 调账接口(走专门的反向 ledger,可审计)。

## 关键要点

- **兑换码生成后只展示一次明文**,管理员必须立即下载 CSV 妥善保管;丢失 = 必须重新生成。
- 手动充值不绑定订单系统,纯人工审批 —— 适合内部场景 / 邀请制 SaaS。
- M2 上线后保留手动充值作为兜底,在线支付走另一条流水。
- 任何充值都会在 `ledger_entries` 留 append-only 记录,无法 UPDATE / DELETE。
