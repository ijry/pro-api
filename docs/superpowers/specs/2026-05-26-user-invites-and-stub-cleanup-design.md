# 用户邀请页 + Stub 清理 — 设计稿

- **日期:** 2026-05-26
- **范围:** 用户端 `/invites` 页(B 方案:邀请码 + 统计 + 列表 + 规则)+ 4 处 stub 页改写 + 2 处死代码删除
- **目标:** demo 演示站(GitHub Pages 部署)上不允许任何"页面建设中…"
- **路径选择:** 路径 1 — Demo-only 前端 + mock fixtures,后端 HTTP route 留给 M2c

## 背景

后端 `internal/invite/` 模块在 M2b 已经实现(model / repo / service),`users.invite_code` + `users.invited_by` 字段已经迁移,`invite_records` 表已存在,但用户端 HTTP 路由尚未暴露。本设计在前端定义 API 契约 + mock,后端 M2c 直接对齐契约实现即可。

现有 5 处"页面建设中…":
1. `web/user/src/pages/placeholder.vue`(被 `/invites` 路由引用)
2. `web/user/src/pages/404.vue`
3. `web/admin/src/views/NotFound.vue`(含 `TODO 待实现`)
4. `web/admin/src/views/Forbidden.vue`(含 `TODO 待实现`)
5. `web/admin/src/layouts/AppLayout.vue:47-49` 的 `m3_coming` disabled 菜单项

另:`web/user/src/pages/Home.vue` 是死代码(home 实际用 `index.vue`)。

## 范围

### 包含

1. 新建用户端 `/invites` 页(`web/user/src/pages/invites.vue`)
2. 新建 `web/user/src/api/invite.ts` 客户端 + TypeScript 类型契约
3. 新建 3 份 mock fixture(summary / invitees / records)
4. `web/shared/src/mock/routes.ts` 增 3 条 `/api/user/invites/*` 路由
5. 改写 `web/user/src/pages/404.vue` 为真正的 404 页
6. 改写 `web/admin/src/views/NotFound.vue` 为真正的 404 页
7. 改写 `web/admin/src/views/Forbidden.vue` 为真正的 403 页
8. 删 `m3_coming` 菜单项 + `nav.m3_coming` i18n key
9. 删 `web/user/src/pages/Home.vue`(死代码)
10. 删 `web/user/src/pages/placeholder.vue`(`/invites` 改指向 `invites.vue` 后不再被引用)
11. i18n key 增量(zh + en):`invites.*`、`notfound.*`、`forbidden.*`、`common.back_home`、`auth.relogin`

### 不包含

- 后端 Go HTTP route(`internal/server/handler/userhttp/invites.go`)— M2c 负责
- 邀请等级 / 排行榜 / 进阶奖励 — C 方案,不在本轮
- 邀请页之外的 i18n 全量补完

## API 契约

3 个端点,全部 GET,均要求登录。Mock 阶段直接读 fixture,M2c 后端实现需对齐字段名与类型。

### `GET /api/user/invites/me`

返回当前用户的邀请概览。

```ts
interface InviteSummary {
  invite_code: string                  // users.invite_code,例 "PROAQ7K3"
  share_url: string                    // 后端拼接,基于 setting site.base_url
  rebate_ratio: number                 // 0.10 = 10%,来自 setting invite.rebate_ratio
  stats: {
    invitee_count: number              // 累计邀请人数
    rebate_credits_total: number       // 累计返佣 credits
    rebate_credits_month: number       // 本月返佣 credits
  }
}
```

### `GET /api/user/invites/invitees?page=&size=`

返回我邀请的人(分页)。

```ts
interface Invitee {
  user_id: number
  display_name: string
  email_masked: string                 // 隐私脱敏:"j***@gmail.com"
  registered_at: string                // ISO 8601
  total_rebate_credits: number         // 该 invitee 贡献的累计返佣
}

interface InviteesResp {
  items: Invitee[]
  total: number
  page: number
  size: number
}
```

### `GET /api/user/invites/records?page=&size=`

返回返佣明细记录(分页)。

```ts
interface InviteRecord {
  id: number
  invitee_id: number
  invitee_display_name: string
  order_id: number
  rebate_cents: number                 // 原币 CNY 分
  rebate_credits: number               // 转换后的 credits
  created_at: string                   // ISO 8601
}

interface RecordsResp {
  items: InviteRecord[]
  total: number
  page: number
  size: number
}
```

### 关键决策说明

- **`share_url` 由后端返回**:URL 拼接逻辑应该在后端读 `setting.site.base_url`,前端只负责复制。Mock 写死 `https://demo.proapi.example/register?invite_code=PROAQ7K3`。
- **`email_masked` 字段**:列表展示其他用户邮箱必须脱敏。后端实现时在 service 层处理。
- **`rebate_credits` 显示**:简化为 1 credit = 1 CNY 分,显示统一为 `¥X.XX`(`credits / 100`),不随 locale 切换货币符号。后续接 `setting.display_currency` 是 M2c 之后的事。
- **不暴露 `invited_by`**:本轮不展示"邀请我的人是谁"。

## 前端组件结构

### `web/user/src/pages/invites.vue`

```
┌─────────────────────────────────────────────────────────────────┐
│ <h1> 邀请好友                                                    │
│ <p>  邀请新用户注册并完成充值,你可得 {ratio}% 返佣              │
└─────────────────────────────────────────────────────────────────┘

┌──────────────────────────┐  ┌──────────────────────────────────┐
│ 我的邀请码               │  │ 邀请人数 │ 累计返佣 │ 本月返佣   │
│ ┌────────────────────┐   │  │  47     │ ¥358.20 │ ¥42.10     │
│ │  PROAQ7K3   [复制] │   │  └──────────────────────────────────┘
│ └────────────────────┘   │
│ 分享链接                 │
│ ┌────────────────────┐   │
│ │ https://...  [复制]│   │
│ └────────────────────┘   │
└──────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│ [被邀请人(47)] [返佣记录(N)]    ← Tabs,N = API total           │
├─────────────────────────────────────────────────────────────────┤
│ 被邀请人 tab: 用户 │ 注册时间 │ 累计返佣                          │
│ 返佣记录 tab: 用户 │ 订单号  │ 返佣金额 │ 时间                    │
│                                          [上一页 1/N 下一页]    │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│ 📜 返佣规则                                                      │
│ • 被邀请人完成首次充值时,你获得充值金额的 10% 返佣              │
│ • 返佣直接进入你的钱包                                          │
│ • 邀请关系永久绑定,后续充值持续返佣                            │
└─────────────────────────────────────────────────────────────────┘
```

**实现细节:**

- 主容器 `max-w-7xl mx-auto`,跟 AppLayout 已有 main 容器一致
- 不抽子组件 — 全部内联在 `invites.vue` 中(YAGNI)
- 复用现有组件:`Card`、`Button`、`ClipboardButton`、`Pagination`、`EmptyState`
- Tab 切换:`ref<'invitees' | 'records'>` + 两个 button 加 class,不引入 Tabs 组件
- 状态:`ref` + `onMounted`,不进 Pinia(数据只在本页使用)
- 加载:进入页面先取 `/me`(包含统计),tab 切换时按需取对应列表
- 错误处理:加载失败显示 `EmptyState` + 重试按钮(跟 `logs/index.vue` 一致)
- 空数据:显示 `EmptyState` 提示"还没有邀请记录"

**响应式:**

- 上半部分两块卡:`md:` 断点以下竖排,以上左右排
- 表格列在 `sm:` 以下隐藏次要列(email_masked、累计返佣列)

## Stub 页改写

### `web/user/src/pages/404.vue`(用户端 404)

AuthLayout 已经提供居中渐变背景,内容只放在 layout 的 slot 里。

```
大字号 "404"(text-6xl text-primary font-bold)
"页面不存在"(h2,text-fg)
"你访问的页面不存在或已被删除"(text-fg-muted)
[回到首页] 按钮 → router.push('/')
```

复用 `Card` + `Button`。

### `web/admin/src/views/NotFound.vue`(admin 404)

用 naive-ui 的 `NResult status="404"`,admin 全家桶已用 naive。

```vue
<template>
  <NResult status="404" :title="t('notfound.title')" :description="t('notfound.desc')">
    <template #footer>
      <NButton type="primary" @click="router.push('/')">{{ t('common.back_home') }}</NButton>
    </template>
  </NResult>
</template>
```

### `web/admin/src/views/Forbidden.vue`(admin 403)

```vue
<template>
  <NResult status="403" :title="t('forbidden.title')" :description="t('forbidden.desc')">
    <template #footer>
      <NSpace>
        <NButton @click="router.push('/')">{{ t('common.back_home') }}</NButton>
        <NButton type="primary" @click="onRelogin">{{ t('auth.relogin') }}</NButton>
      </NSpace>
    </template>
  </NResult>
</template>
```

`onRelogin` 调 `useUserStore().logout()` 后跳 `/login`。

### 删 `m3_coming` 菜单项

`web/admin/src/layouts/AppLayout.vue` 第 47–49 行整段删:

```ts
{ type: 'divider', key: 'div1' },
{ label: t('nav.m3_coming'), key: 'm3', disabled: true },
```

同步删 `web/admin/src/i18n/{zh,en}.json` 中 `nav.m3_coming` key。

### 删死代码

- `git rm web/user/src/pages/Home.vue`
- `git rm web/user/src/pages/placeholder.vue`

删除前确认无 import 引用(`/invites` 路由会改指向 `invites.vue`,届时 placeholder 不再被引用)。

## i18n 增量

### `web/user/src/i18n/zh.json` 新增

```json
{
  "invites": {
    "title": "邀请好友",
    "subtitle": "邀请新用户注册并完成充值,你可得 {ratio}% 返佣",
    "my_code": "我的邀请码",
    "share_url": "分享链接",
    "stats": {
      "invitee_count": "邀请人数",
      "rebate_total": "累计返佣",
      "rebate_month": "本月返佣"
    },
    "tab": {
      "invitees": "被邀请人",
      "records": "返佣记录"
    },
    "table": {
      "user": "用户",
      "registered_at": "注册时间",
      "total_rebate": "累计返佣",
      "order_id": "订单号",
      "rebate": "返佣金额",
      "created_at": "时间"
    },
    "empty": {
      "invitees": "还没有人通过你的邀请码注册",
      "records": "还没有返佣记录"
    },
    "rules": {
      "title": "返佣规则",
      "item1": "被邀请人完成充值时,你获得充值金额的 {ratio}% 返佣",
      "item2": "返佣直接进入你的钱包",
      "item3": "邀请关系永久绑定,后续充值持续返佣"
    }
  },
  "notfound": {
    "title": "页面不存在",
    "desc": "你访问的页面不存在或已被删除"
  },
  "common": {
    "back_home": "返回首页"
  }
}
```

`web/user/src/i18n/en.json` 同结构英文翻译,具体英文字串在实施计划中逐条枚举。

### `web/admin/src/i18n/zh.json` 新增

```json
{
  "notfound": { "title": "页面不存在", "desc": "你访问的页面不存在,请检查 URL" },
  "forbidden": { "title": "无权访问", "desc": "你的账号没有访问此页面的权限" },
  "common": { "back_home": "返回首页" },
  "auth": { "relogin": "切换账号" }
}
```

删:`nav.m3_coming`(zh + en)。

## Mock Fixtures

### `web/shared/src/mock/data/invite-summary.json`

```json
{
  "invite_code": "PROAQ7K3",
  "share_url": "https://demo.proapi.example/register?invite_code=PROAQ7K3",
  "rebate_ratio": 0.10,
  "stats": {
    "invitee_count": 47,
    "rebate_credits_total": 35820,
    "rebate_credits_month": 4210
  }
}
```

### `web/shared/src/mock/data/invite-invitees.json`

15 条假数据(够分页测试 — `size=10` 时有第 2 页),字段齐全:

```json
[
  { "user_id": 9001, "display_name": "John D.", "email_masked": "j***@gmail.com",
    "registered_at": "2026-05-23T10:30:00Z", "total_rebate_credits": 1240 },
  ...
]
```

### `web/shared/src/mock/data/invite-records.json`

30 条返佣记录(够分页测试),按 `created_at` 倒序:

```json
[
  { "id": 70001, "invitee_id": 9001, "invitee_display_name": "John D.",
    "order_id": 50001, "rebate_cents": 1240, "rebate_credits": 1240,
    "created_at": "2026-05-25T14:22:00Z" },
  ...
]
```

### `web/shared/src/mock/routes.ts` 新增

```ts
import inviteSummary from './data/invite-summary.json'
import inviteInvitees from './data/invite-invitees.json'
import inviteRecords from './data/invite-records.json'

// in routes table:
{ pattern: /^\/api\/user\/invites\/me$/,
  handler: () => clone(inviteSummary) },
{ pattern: /^\/api\/user\/invites\/invitees(\?.*)?$/,
  handler: (_m, _u, p) => paginate(inviteInvitees as any[], p as PageParams) },
{ pattern: /^\/api\/user\/invites\/records(\?.*)?$/,
  handler: (_m, _u, p) => paginate(inviteRecords as any[], p as PageParams) },
```

## 数据流

```
invites.vue
  onMounted → inviteApi.getSummary() → 显示邀请码/分享/统计
  activeTab='invitees' → inviteApi.listInvitees({page,size}) → 表格
  activeTab='records'  → inviteApi.listRecords({page,size}) → 表格
  分页点击 → 重新 listX()
  复制按钮 → ClipboardButton 自带 toast
```

## 错误处理

- API 失败:`try/catch` 内不抛,设置 `error.value=true`,UI 显示 `EmptyState` + 重试按钮
- 空数据:`items.length === 0` → `EmptyState` + 占位文案
- 复制失败:`ClipboardButton` 自带 toast 反馈(已有逻辑)

## 验收 / 回归

1. `pnpm typecheck`(`web/user` + `web/admin`)— 必须过,0 warning
2. `pnpm build:demo`(两端)— 必须成功
3. `grep -rn "页面建设" web/` — 必须 0 命中
4. 本地 `pnpm dev:demo` 启用户端,目测:
   - `/invites` 显示完整邀请页(邀请码 + 统计 + tab 切换 + 分页)
   - `/no-such-route` 显示真 404 页
5. 本地 `pnpm dev:demo` 启 admin 端,目测:
   - `/no-such-route` 显示真 404 页
   - `/forbidden` 显示真 403 页
   - 侧边栏不再有"M3 模块 (开发中)"项
6. 部署到 GH Pages 后从 docs nav "用户中心演示" 跳转,点击右上头像菜单(若 nav 中有)和导航后,所有可达路径都不应出现"页面建设中…"

## 后端契约对齐(M2c 实施时)

后端实现 `internal/server/handler/userhttp/invites.go` 时,3 个 handler 必须严格按本设计的字段名输出。具体:

- `email_masked` 在 service 层处理:`a@b.c → a***@b.c`(保留首字符 + `***` + `@后半`)
- `rebate_credits_total` / `rebate_credits_month` 通过 `SUM(rebate_credits) WHERE inviter_id=?` + 时间过滤
- `share_url` = `setting.GetString(ctx, "site.base_url") + "/register?invite_code=" + user.InviteCode`
- 分页对齐 `pkg/apipage` 已有约定
