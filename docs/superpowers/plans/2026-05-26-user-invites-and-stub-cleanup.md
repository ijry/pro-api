# 用户邀请页 + Stub 清理 — 实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在用户端新建真实可用的 `/invites` 页(B 方案 = 邀请码 + 3 统计 + 2-tab 列表 + 规则),并清理 4 处"页面建设中…"stub 与 2 处死代码,使 demo 演示站不再出现任何未完成页面。

**Architecture:** Demo-only 前端 + mock fixtures 路径。新页面 `invites.vue` 走 `web/user/src/api/invite.ts` 客户端,fixture 落在 `web/shared/src/mock/data/`,mock 路由通过 `web/shared/src/mock/routes.ts` 注入。Stub 页(用户 404、admin 404/403)用现有 UI 库(用户端 Tailwind / admin 端 naive-ui `NResult`)重写。后端 HTTP route 按本计划暴露的契约在 M2c 实施。

**Tech Stack:** Vue 3 setup + vue-i18n 9 + vue-router 4 + Tailwind/UnoCSS(用户)+ naive-ui(admin);mock 走 `VITE_DEMO_MOCK=true` 分支 + `web/shared/src/mock/routes.ts`;构建 `pnpm typecheck` / `pnpm build:demo`。

**Spec:** `docs/superpowers/specs/2026-05-26-user-invites-and-stub-cleanup-design.md`

---

## File Structure

**Create:**
- `web/shared/src/mock/data/invite-summary.json` — `InviteSummary` 单对象
- `web/shared/src/mock/data/invite-invitees.json` — 15 条 `Invitee[]`
- `web/shared/src/mock/data/invite-records.json` — 30 条 `InviteRecord[]`
- `web/user/src/api/invite.ts` — TypeScript 契约 + 3 GET 方法
- `web/user/src/pages/invites.vue` — 整页(不拆子组件,YAGNI)

**Modify:**
- `web/shared/src/mock/routes.ts` — 新增 3 条 `/api/user/invites/*` 路由
- `web/user/src/router/index.ts:108-113` — `/invites` 改指向 `invites.vue`
- `web/user/src/pages/404.vue` — 整体改写为真正的 404 页
- `web/user/src/i18n/zh.json` + `en.json` — 新增 `invites.*` / `notfound.*` / `common.back_home`
- `web/admin/src/views/NotFound.vue` — 整体改写为 `NResult status="404"`
- `web/admin/src/views/Forbidden.vue` — 整体改写为 `NResult status="403"`
- `web/admin/src/layouts/AppLayout.vue:47-49` — 删 m3_coming 菜单项与 divider
- `web/admin/src/i18n/zh.json` + `en.json` — 新增 `notfound.*` / `forbidden.*` / `common.back_home` / `auth.relogin`,删 `nav.m3_coming`

**Delete:**
- `web/user/src/pages/Home.vue` — 死代码(实际首页是 `index.vue`)
- `web/user/src/pages/placeholder.vue` — `/invites` 改指向 `invites.vue` 后无引用

---

## Task 1: Mock fixtures(3 个 JSON)

**Files:**
- Create: `web/shared/src/mock/data/invite-summary.json`
- Create: `web/shared/src/mock/data/invite-invitees.json`
- Create: `web/shared/src/mock/data/invite-records.json`

- [ ] **Step 1.1: 写 invite-summary.json**

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

- [ ] **Step 1.2: 写 invite-invitees.json(15 条)**

```json
[
  { "user_id": 9001, "display_name": "John D.",    "email_masked": "j***@gmail.com",     "registered_at": "2026-05-25T10:30:00Z", "total_rebate_credits": 1240 },
  { "user_id": 9002, "display_name": "Alice L.",   "email_masked": "a***@outlook.com",   "registered_at": "2026-05-24T18:12:00Z", "total_rebate_credits": 980 },
  { "user_id": 9003, "display_name": "Bob W.",     "email_masked": "b***@163.com",       "registered_at": "2026-05-24T09:05:00Z", "total_rebate_credits": 320 },
  { "user_id": 9004, "display_name": "Cathy M.",   "email_masked": "c***@qq.com",        "registered_at": "2026-05-23T16:42:00Z", "total_rebate_credits": 2150 },
  { "user_id": 9005, "display_name": "David K.",   "email_masked": "d***@gmail.com",     "registered_at": "2026-05-22T11:21:00Z", "total_rebate_credits": 410 },
  { "user_id": 9006, "display_name": "Emma Z.",    "email_masked": "e***@hotmail.com",   "registered_at": "2026-05-21T20:08:00Z", "total_rebate_credits": 1670 },
  { "user_id": 9007, "display_name": "Frank L.",   "email_masked": "f***@126.com",       "registered_at": "2026-05-20T14:55:00Z", "total_rebate_credits": 88 },
  { "user_id": 9008, "display_name": "Grace P.",   "email_masked": "g***@yahoo.com",     "registered_at": "2026-05-19T09:30:00Z", "total_rebate_credits": 2400 },
  { "user_id": 9009, "display_name": "Henry T.",   "email_masked": "h***@gmail.com",     "registered_at": "2026-05-18T17:11:00Z", "total_rebate_credits": 0 },
  { "user_id": 9010, "display_name": "Ivy R.",     "email_masked": "i***@outlook.com",   "registered_at": "2026-05-17T08:45:00Z", "total_rebate_credits": 540 },
  { "user_id": 9011, "display_name": "Jack S.",    "email_masked": "j***@qq.com",        "registered_at": "2026-05-16T22:19:00Z", "total_rebate_credits": 1320 },
  { "user_id": 9012, "display_name": "Karen B.",   "email_masked": "k***@gmail.com",     "registered_at": "2026-05-15T13:02:00Z", "total_rebate_credits": 760 },
  { "user_id": 9013, "display_name": "Leo H.",     "email_masked": "l***@163.com",       "registered_at": "2026-05-14T19:33:00Z", "total_rebate_credits": 200 },
  { "user_id": 9014, "display_name": "Mia C.",     "email_masked": "m***@hotmail.com",   "registered_at": "2026-05-13T07:50:00Z", "total_rebate_credits": 1890 },
  { "user_id": 9015, "display_name": "Nathan F.",  "email_masked": "n***@gmail.com",     "registered_at": "2026-05-12T15:27:00Z", "total_rebate_credits": 0 }
]
```

- [ ] **Step 1.3: 写 invite-records.json(30 条,按 created_at 倒序)**

```json
[
  { "id": 70030, "invitee_id": 9001, "invitee_display_name": "John D.",   "order_id": 50030, "rebate_cents": 1240, "rebate_credits": 1240, "created_at": "2026-05-25T14:22:00Z" },
  { "id": 70029, "invitee_id": 9004, "invitee_display_name": "Cathy M.",  "order_id": 50029, "rebate_cents": 980,  "rebate_credits": 980,  "created_at": "2026-05-25T11:08:00Z" },
  { "id": 70028, "invitee_id": 9008, "invitee_display_name": "Grace P.",  "order_id": 50028, "rebate_cents": 2400, "rebate_credits": 2400, "created_at": "2026-05-24T19:45:00Z" },
  { "id": 70027, "invitee_id": 9002, "invitee_display_name": "Alice L.",  "order_id": 50027, "rebate_cents": 500,  "rebate_credits": 500,  "created_at": "2026-05-24T16:12:00Z" },
  { "id": 70026, "invitee_id": 9006, "invitee_display_name": "Emma Z.",   "order_id": 50026, "rebate_cents": 1670, "rebate_credits": 1670, "created_at": "2026-05-24T09:38:00Z" },
  { "id": 70025, "invitee_id": 9011, "invitee_display_name": "Jack S.",   "order_id": 50025, "rebate_cents": 820,  "rebate_credits": 820,  "created_at": "2026-05-23T22:14:00Z" },
  { "id": 70024, "invitee_id": 9001, "invitee_display_name": "John D.",   "order_id": 50024, "rebate_cents": 320,  "rebate_credits": 320,  "created_at": "2026-05-23T18:55:00Z" },
  { "id": 70023, "invitee_id": 9014, "invitee_display_name": "Mia C.",    "order_id": 50023, "rebate_cents": 1100, "rebate_credits": 1100, "created_at": "2026-05-23T12:30:00Z" },
  { "id": 70022, "invitee_id": 9004, "invitee_display_name": "Cathy M.",  "order_id": 50022, "rebate_cents": 670,  "rebate_credits": 670,  "created_at": "2026-05-22T20:11:00Z" },
  { "id": 70021, "invitee_id": 9010, "invitee_display_name": "Ivy R.",    "order_id": 50021, "rebate_cents": 540,  "rebate_credits": 540,  "created_at": "2026-05-22T15:47:00Z" },
  { "id": 70020, "invitee_id": 9008, "invitee_display_name": "Grace P.",  "order_id": 50020, "rebate_cents": 1200, "rebate_credits": 1200, "created_at": "2026-05-21T19:22:00Z" },
  { "id": 70019, "invitee_id": 9012, "invitee_display_name": "Karen B.",  "order_id": 50019, "rebate_cents": 460,  "rebate_credits": 460,  "created_at": "2026-05-21T14:08:00Z" },
  { "id": 70018, "invitee_id": 9002, "invitee_display_name": "Alice L.",  "order_id": 50018, "rebate_cents": 480,  "rebate_credits": 480,  "created_at": "2026-05-20T21:30:00Z" },
  { "id": 70017, "invitee_id": 9014, "invitee_display_name": "Mia C.",    "order_id": 50017, "rebate_cents": 790,  "rebate_credits": 790,  "created_at": "2026-05-20T17:55:00Z" },
  { "id": 70016, "invitee_id": 9005, "invitee_display_name": "David K.",  "order_id": 50016, "rebate_cents": 410,  "rebate_credits": 410,  "created_at": "2026-05-19T13:20:00Z" },
  { "id": 70015, "invitee_id": 9001, "invitee_display_name": "John D.",   "order_id": 50015, "rebate_cents": 240,  "rebate_credits": 240,  "created_at": "2026-05-18T16:45:00Z" },
  { "id": 70014, "invitee_id": 9011, "invitee_display_name": "Jack S.",   "order_id": 50014, "rebate_cents": 500,  "rebate_credits": 500,  "created_at": "2026-05-17T11:08:00Z" },
  { "id": 70013, "invitee_id": 9008, "invitee_display_name": "Grace P.",  "order_id": 50013, "rebate_cents": 380,  "rebate_credits": 380,  "created_at": "2026-05-16T22:50:00Z" },
  { "id": 70012, "invitee_id": 9006, "invitee_display_name": "Emma Z.",   "order_id": 50012, "rebate_cents": 920,  "rebate_credits": 920,  "created_at": "2026-05-15T15:30:00Z" },
  { "id": 70011, "invitee_id": 9013, "invitee_display_name": "Leo H.",    "order_id": 50011, "rebate_cents": 200,  "rebate_credits": 200,  "created_at": "2026-05-14T19:18:00Z" },
  { "id": 70010, "invitee_id": 9004, "invitee_display_name": "Cathy M.",  "order_id": 50010, "rebate_cents": 500,  "rebate_credits": 500,  "created_at": "2026-05-13T10:25:00Z" },
  { "id": 70009, "invitee_id": 9012, "invitee_display_name": "Karen B.",  "order_id": 50009, "rebate_cents": 300,  "rebate_credits": 300,  "created_at": "2026-05-12T18:00:00Z" },
  { "id": 70008, "invitee_id": 9014, "invitee_display_name": "Mia C.",    "order_id": 50008, "rebate_cents": 0,    "rebate_credits": 0,    "created_at": "2026-05-11T12:34:00Z" },
  { "id": 70007, "invitee_id": 9003, "invitee_display_name": "Bob W.",    "order_id": 50007, "rebate_cents": 320,  "rebate_credits": 320,  "created_at": "2026-05-10T20:40:00Z" },
  { "id": 70006, "invitee_id": 9007, "invitee_display_name": "Frank L.",  "order_id": 50006, "rebate_cents": 88,   "rebate_credits": 88,   "created_at": "2026-05-09T14:15:00Z" },
  { "id": 70005, "invitee_id": 9001, "invitee_display_name": "John D.",   "order_id": 50005, "rebate_cents": 680,  "rebate_credits": 680,  "created_at": "2026-05-08T08:00:00Z" },
  { "id": 70004, "invitee_id": 9010, "invitee_display_name": "Ivy R.",    "order_id": 50004, "rebate_cents": 0,    "rebate_credits": 0,    "created_at": "2026-05-07T22:22:00Z" },
  { "id": 70003, "invitee_id": 9006, "invitee_display_name": "Emma Z.",   "order_id": 50003, "rebate_cents": 750,  "rebate_credits": 750,  "created_at": "2026-05-06T16:10:00Z" },
  { "id": 70002, "invitee_id": 9002, "invitee_display_name": "Alice L.",  "order_id": 50002, "rebate_cents": 0,    "rebate_credits": 0,    "created_at": "2026-05-05T11:33:00Z" },
  { "id": 70001, "invitee_id": 9011, "invitee_display_name": "Jack S.",   "order_id": 50001, "rebate_cents": 0,    "rebate_credits": 0,    "created_at": "2026-05-04T09:09:00Z" }
]
```

- [ ] **Step 1.4: 校验 JSON**

Run:
```bash
node -e "JSON.parse(require('fs').readFileSync('web/shared/src/mock/data/invite-summary.json'))" && \
node -e "console.log(JSON.parse(require('fs').readFileSync('web/shared/src/mock/data/invite-invitees.json')).length)" && \
node -e "console.log(JSON.parse(require('fs').readFileSync('web/shared/src/mock/data/invite-records.json')).length)"
```
Expected: 无报错;两条数字输出分别是 `15` 和 `30`。

- [ ] **Step 1.5: Commit**

```bash
git add web/shared/src/mock/data/invite-summary.json web/shared/src/mock/data/invite-invitees.json web/shared/src/mock/data/invite-records.json
git commit -m "feat(mock): 邀请页 3 份 fixture(summary/invitees/records)"
```

---

## Task 2: 用户端 invite API 客户端

**Files:**
- Create: `web/user/src/api/invite.ts`

- [ ] **Step 2.1: 写 API 客户端**

照 `web/user/src/api/ledger.ts` 的风格(typed `get<T>` + `xxxApi` 导出对象)。

```ts
import { get } from './http'

export interface InviteSummary {
  invite_code: string
  share_url: string
  rebate_ratio: number
  stats: {
    invitee_count: number
    rebate_credits_total: number
    rebate_credits_month: number
  }
}

export interface Invitee {
  user_id: number
  display_name: string
  email_masked: string
  registered_at: string
  total_rebate_credits: number
}

export interface InviteRecord {
  id: number
  invitee_id: number
  invitee_display_name: string
  order_id: number
  rebate_cents: number
  rebate_credits: number
  created_at: string
}

export interface PageResp<T> {
  items: T[]
  total: number
  page: number
  size: number
}

export const inviteApi = {
  getSummary: () => get<InviteSummary>('/api/user/invites/me'),
  listInvitees: (page = 1, size = 10) =>
    get<PageResp<Invitee>>(`/api/user/invites/invitees?page=${page}&size=${size}`),
  listRecords: (page = 1, size = 10) =>
    get<PageResp<InviteRecord>>(`/api/user/invites/records?page=${page}&size=${size}`),
}
```

- [ ] **Step 2.2: 类型检查**

Run: `pnpm --filter @proapi/user typecheck`
Expected: 0 error(此文件尚无引用,仅类型自洽即可)。

- [ ] **Step 2.3: Commit**

```bash
git add web/user/src/api/invite.ts
git commit -m "feat(user): 添加 invite API 客户端与类型契约"
```

---

## Task 3: 注册 3 条 mock 路由

**Files:**
- Modify: `web/shared/src/mock/routes.ts`

- [ ] **Step 3.1: 在 import 区追加 3 行**

定位文件顶部 `import` 区块(`web/shared/src/mock/routes.ts:1-23`),在 `import oauthBindings from './data/oauth-bindings.json'` 之后追加:

```ts
import inviteSummary from './data/invite-summary.json'
import inviteInvitees from './data/invite-invitees.json'
import inviteRecords from './data/invite-records.json'
```

- [ ] **Step 3.2: 在 user 区追加 3 条路由**

定位 `// ============ user =============` 之后、`{ pattern: /^\/api\/public\/models/...` 之前(约 `web/shared/src/mock/routes.ts:127`),在 `notices` 路由组之后、`public` 路由组之前追加:

```ts
  { pattern: /^\/api\/user\/invites\/me$/,                       handler: () => clone(inviteSummary) },
  { pattern: /^\/api\/user\/invites\/invitees(\?.*)?$/,          handler: (_m, _u, p) => paginate(inviteInvitees as any[], p as PageParams) },
  { pattern: /^\/api\/user\/invites\/records(\?.*)?$/,           handler: (_m, _u, p) => paginate(inviteRecords as any[], p as PageParams) },
```

- [ ] **Step 3.3: 类型检查(确认 import 与字段不冲突)**

Run: `pnpm --filter @proapi/user typecheck && pnpm --filter @proapi/admin typecheck`
Expected: 0 error。

- [ ] **Step 3.4: 烟测 mock 路由**

Run:
```bash
node -e "
const r = require('./web/shared/src/mock/routes.ts');
" 2>&1 | head -5 || true
```
(ts-node 不一定可用,跳过亦可。下游 dev server 启动时若 import 失败即会立即暴露。)

- [ ] **Step 3.5: Commit**

```bash
git add web/shared/src/mock/routes.ts
git commit -m "feat(mock): 注册 /api/user/invites/* 3 条路由"
```

---

## Task 4: 用户端 i18n 增量

**Files:**
- Modify: `web/user/src/i18n/zh.json`
- Modify: `web/user/src/i18n/en.json`

- [ ] **Step 4.1: 阅读两份 i18n 文件的结构**

Run:
```bash
head -20 web/user/src/i18n/zh.json
head -20 web/user/src/i18n/en.json
```
Expected:确认顶层结构有 `nav` / `auth` / `home` 等顶级 key,`nav.invites` 已存在(zh: "邀请",en 同名)。

- [ ] **Step 4.2: zh.json 新增 `invites` 顶级 key**

在合适位置(例如 `home` 与 `auth.login` 之间,或文件末尾闭合 `}` 之前)追加:

```jsonc
  "invites": {
    "title": "邀请好友",
    "subtitle": "邀请新用户注册并完成充值,你可得 {ratio}% 返佣",
    "my_code": "我的邀请码",
    "share_url": "分享链接",
    "copy": "复制",
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
      "email": "邮箱",
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
    "load_failed": "加载失败",
    "retry": "重试",
    "rules": {
      "title": "返佣规则",
      "item1": "被邀请人完成充值时,你获得充值金额的 {ratio}% 返佣",
      "item2": "返佣直接进入你的钱包,可用于抵扣后续消费",
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
```

注意 JSON 没有注释、最后一个 key 后不要逗号。

- [ ] **Step 4.3: en.json 新增对应英文**

```jsonc
  "invites": {
    "title": "Invite friends",
    "subtitle": "Invite new users and get {ratio}% rebate on their top-ups",
    "my_code": "My invite code",
    "share_url": "Share link",
    "copy": "Copy",
    "stats": {
      "invitee_count": "Invitees",
      "rebate_total": "Total rebate",
      "rebate_month": "This month"
    },
    "tab": {
      "invitees": "Invitees",
      "records": "Rebate records"
    },
    "table": {
      "user": "User",
      "email": "Email",
      "registered_at": "Registered at",
      "total_rebate": "Total rebate",
      "order_id": "Order ID",
      "rebate": "Rebate",
      "created_at": "Time"
    },
    "empty": {
      "invitees": "No one has signed up with your code yet",
      "records": "No rebate records yet"
    },
    "load_failed": "Failed to load",
    "retry": "Retry",
    "rules": {
      "title": "Rebate rules",
      "item1": "When an invitee tops up, you get {ratio}% as rebate",
      "item2": "Rebate goes straight to your wallet and offsets future usage",
      "item3": "Invite binding is permanent — every later top-up keeps paying"
    }
  },
  "notfound": {
    "title": "Page not found",
    "desc": "The page you requested doesn't exist or has been removed"
  },
  "common": {
    "back_home": "Back to home"
  }
```

- [ ] **Step 4.4: JSON 校验**

Run:
```bash
node -e "JSON.parse(require('fs').readFileSync('web/user/src/i18n/zh.json'))" && \
node -e "JSON.parse(require('fs').readFileSync('web/user/src/i18n/en.json'))"
```
Expected: 无报错。

- [ ] **Step 4.5: Commit**

```bash
git add web/user/src/i18n/zh.json web/user/src/i18n/en.json
git commit -m "i18n(user): 补 invites/notfound/common.back_home 翻译"
```

---

## Task 5: invites.vue 主页面

**Files:**
- Create: `web/user/src/pages/invites.vue`

整页内联实现,不抽子组件。复用 `Card` / `Button` / `ClipboardButton` / `Pagination` / `EmptyState` / `Skeleton`。

- [ ] **Step 5.1: 写整页**

```vue
<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { inviteApi, type InviteSummary, type Invitee, type InviteRecord } from '@/api/invite'
import { useToast } from '@/composables/useToast'
import Card from '@/components/ui/Card.vue'
import ClipboardButton from '@/components/ui/ClipboardButton.vue'
import Pagination from '@/components/ui/Pagination.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import Skeleton from '@/components/ui/Skeleton.vue'

const { t, locale } = useI18n()
const toast = useToast()

const summary = ref<InviteSummary | null>(null)
const summaryLoading = ref(true)
const summaryError = ref(false)

const activeTab = ref<'invitees' | 'records'>('invitees')

const invitees = ref<Invitee[]>([])
const inviteesTotal = ref(0)
const inviteesPage = ref(1)
const inviteesLoading = ref(false)
const inviteesError = ref(false)

const records = ref<InviteRecord[]>([])
const recordsTotal = ref(0)
const recordsPage = ref(1)
const recordsLoading = ref(false)
const recordsError = ref(false)

const PAGE_SIZE = 10

const ratioPercent = computed(() =>
  summary.value ? Math.round(summary.value.rebate_ratio * 100) : 0
)

function formatRebate(credits: number) {
  return `¥${(credits / 100).toFixed(2)}`
}

function formatDate(iso: string) {
  const d = new Date(iso)
  return locale.value === 'zh'
    ? d.toLocaleString('zh-CN', { hour12: false })
    : d.toLocaleString('en-US')
}

async function loadSummary() {
  summaryLoading.value = true
  summaryError.value = false
  try {
    summary.value = await inviteApi.getSummary()
  } catch {
    summaryError.value = true
    toast.error(t('invites.load_failed'))
  } finally {
    summaryLoading.value = false
  }
}

async function loadInvitees() {
  inviteesLoading.value = true
  inviteesError.value = false
  try {
    const r = await inviteApi.listInvitees(inviteesPage.value, PAGE_SIZE)
    invitees.value = r.items
    inviteesTotal.value = r.total
  } catch {
    inviteesError.value = true
    toast.error(t('invites.load_failed'))
  } finally {
    inviteesLoading.value = false
  }
}

async function loadRecords() {
  recordsLoading.value = true
  recordsError.value = false
  try {
    const r = await inviteApi.listRecords(recordsPage.value, PAGE_SIZE)
    records.value = r.items
    recordsTotal.value = r.total
  } catch {
    recordsError.value = true
    toast.error(t('invites.load_failed'))
  } finally {
    recordsLoading.value = false
  }
}

onMounted(() => {
  loadSummary()
  loadInvitees()
})

watch(activeTab, (tab) => {
  if (tab === 'records' && records.value.length === 0 && !recordsLoading.value) {
    loadRecords()
  }
})

watch(inviteesPage, loadInvitees)
watch(recordsPage, loadRecords)
</script>

<template>
  <div class="space-y-5">
    <!-- Header -->
    <div>
      <h1 class="text-2xl font-bold text-fg">{{ t('invites.title') }}</h1>
      <p class="text-sm text-fg-muted mt-1">
        {{ t('invites.subtitle', { ratio: ratioPercent }) }}
      </p>
    </div>

    <!-- Code card + stats -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <!-- Invite code / share link -->
      <Card>
        <div v-if="summaryLoading" class="space-y-3">
          <Skeleton class="h-6 w-32" />
          <Skeleton class="h-10" />
          <Skeleton class="h-10" />
        </div>
        <div v-else-if="summaryError" class="py-6 text-center">
          <p class="text-fg-muted text-sm mb-3">{{ t('invites.load_failed') }}</p>
          <button class="btn-primary" @click="loadSummary">{{ t('invites.retry') }}</button>
        </div>
        <div v-else-if="summary" class="space-y-4">
          <div>
            <label class="block text-xs text-fg-muted mb-1.5">{{ t('invites.my_code') }}</label>
            <div class="flex items-center gap-2">
              <code class="flex-1 px-3 py-2 rounded-md bg-bg-elevated border border-border font-mono text-fg text-lg tracking-wider">
                {{ summary.invite_code }}
              </code>
              <ClipboardButton :text="summary.invite_code" />
            </div>
          </div>
          <div>
            <label class="block text-xs text-fg-muted mb-1.5">{{ t('invites.share_url') }}</label>
            <div class="flex items-center gap-2">
              <code class="flex-1 px-3 py-2 rounded-md bg-bg-elevated border border-border font-mono text-fg text-xs truncate">
                {{ summary.share_url }}
              </code>
              <ClipboardButton :text="summary.share_url" />
            </div>
          </div>
        </div>
      </Card>

      <!-- Stats -->
      <Card>
        <div v-if="summaryLoading" class="grid grid-cols-3 gap-4">
          <Skeleton v-for="i in 3" :key="i" class="h-16" />
        </div>
        <div v-else-if="summary" class="grid grid-cols-3 gap-4">
          <div class="text-center">
            <div class="text-xs text-fg-muted mb-1">{{ t('invites.stats.invitee_count') }}</div>
            <div class="text-2xl font-bold text-fg">{{ summary.stats.invitee_count }}</div>
          </div>
          <div class="text-center border-l border-r border-border">
            <div class="text-xs text-fg-muted mb-1">{{ t('invites.stats.rebate_total') }}</div>
            <div class="text-2xl font-bold text-primary">{{ formatRebate(summary.stats.rebate_credits_total) }}</div>
          </div>
          <div class="text-center">
            <div class="text-xs text-fg-muted mb-1">{{ t('invites.stats.rebate_month') }}</div>
            <div class="text-2xl font-bold text-fg">{{ formatRebate(summary.stats.rebate_credits_month) }}</div>
          </div>
        </div>
      </Card>
    </div>

    <!-- Tabs -->
    <div class="flex items-center gap-1 border-b border-border">
      <button
        class="px-4 h-9 text-sm border-b-2 -mb-px transition-colors"
        :class="activeTab === 'invitees'
          ? 'border-primary text-primary font-medium'
          : 'border-transparent text-fg-muted hover:text-fg'"
        @click="activeTab = 'invitees'"
      >
        {{ t('invites.tab.invitees') }}
        <span v-if="summary" class="ml-1 text-fg-muted">({{ summary.stats.invitee_count }})</span>
      </button>
      <button
        class="px-4 h-9 text-sm border-b-2 -mb-px transition-colors"
        :class="activeTab === 'records'
          ? 'border-primary text-primary font-medium'
          : 'border-transparent text-fg-muted hover:text-fg'"
        @click="activeTab = 'records'"
      >
        {{ t('invites.tab.records') }}
        <span v-if="recordsTotal > 0" class="ml-1 text-fg-muted">({{ recordsTotal }})</span>
      </button>
    </div>

    <!-- Invitees tab -->
    <div v-if="activeTab === 'invitees'">
      <div v-if="inviteesLoading" class="space-y-2">
        <Skeleton v-for="i in 5" :key="i" class="h-10" />
      </div>
      <EmptyState
        v-else-if="inviteesError"
        icon="i-lucide-alert-circle"
        :title="t('invites.load_failed')"
        :cta="t('invites.retry')"
        @cta="loadInvitees"
      />
      <EmptyState
        v-else-if="!invitees.length"
        icon="i-lucide-user-plus"
        :title="t('invites.empty.invitees')"
      />
      <div v-else class="overflow-x-auto rounded-lg border border-border">
        <table class="w-full text-sm min-w-[600px]">
          <thead class="bg-bg-elevated border-b border-border">
            <tr>
              <th class="text-left px-3 py-2 text-fg-muted font-medium">{{ t('invites.table.user') }}</th>
              <th class="text-left px-3 py-2 text-fg-muted font-medium hidden sm:table-cell">{{ t('invites.table.email') }}</th>
              <th class="text-left px-3 py-2 text-fg-muted font-medium">{{ t('invites.table.registered_at') }}</th>
              <th class="text-right px-3 py-2 text-fg-muted font-medium">{{ t('invites.table.total_rebate') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="u in invitees" :key="u.user_id" class="border-b border-border last:border-0 hover:bg-bg-elevated/50 transition-colors">
              <td class="px-3 py-2 text-fg">{{ u.display_name }}</td>
              <td class="px-3 py-2 text-fg-muted text-xs font-mono hidden sm:table-cell">{{ u.email_masked }}</td>
              <td class="px-3 py-2 text-fg-muted text-xs">{{ formatDate(u.registered_at) }}</td>
              <td class="px-3 py-2 text-right font-mono text-primary">{{ formatRebate(u.total_rebate_credits) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
      <Pagination v-if="inviteesTotal > PAGE_SIZE" v-model="inviteesPage" :total="inviteesTotal" :size="PAGE_SIZE" />
    </div>

    <!-- Records tab -->
    <div v-else>
      <div v-if="recordsLoading" class="space-y-2">
        <Skeleton v-for="i in 5" :key="i" class="h-10" />
      </div>
      <EmptyState
        v-else-if="recordsError"
        icon="i-lucide-alert-circle"
        :title="t('invites.load_failed')"
        :cta="t('invites.retry')"
        @cta="loadRecords"
      />
      <EmptyState
        v-else-if="!records.length"
        icon="i-lucide-coins"
        :title="t('invites.empty.records')"
      />
      <div v-else class="overflow-x-auto rounded-lg border border-border">
        <table class="w-full text-sm min-w-[600px]">
          <thead class="bg-bg-elevated border-b border-border">
            <tr>
              <th class="text-left px-3 py-2 text-fg-muted font-medium">{{ t('invites.table.user') }}</th>
              <th class="text-left px-3 py-2 text-fg-muted font-medium hidden sm:table-cell">{{ t('invites.table.order_id') }}</th>
              <th class="text-right px-3 py-2 text-fg-muted font-medium">{{ t('invites.table.rebate') }}</th>
              <th class="text-left px-3 py-2 text-fg-muted font-medium">{{ t('invites.table.created_at') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="r in records" :key="r.id" class="border-b border-border last:border-0 hover:bg-bg-elevated/50 transition-colors">
              <td class="px-3 py-2 text-fg">{{ r.invitee_display_name }}</td>
              <td class="px-3 py-2 text-fg-muted text-xs font-mono hidden sm:table-cell">#{{ r.order_id }}</td>
              <td class="px-3 py-2 text-right font-mono text-primary">{{ formatRebate(r.rebate_credits) }}</td>
              <td class="px-3 py-2 text-fg-muted text-xs">{{ formatDate(r.created_at) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
      <Pagination v-if="recordsTotal > PAGE_SIZE" v-model="recordsPage" :total="recordsTotal" :size="PAGE_SIZE" />
    </div>

    <!-- Rules -->
    <Card>
      <h2 class="text-base font-semibold text-fg mb-3 flex items-center gap-2">
        <span class="i-lucide-scroll-text w-4 h-4 text-primary" />
        {{ t('invites.rules.title') }}
      </h2>
      <ul class="space-y-2 text-sm text-fg-muted">
        <li class="flex gap-2"><span class="text-primary">•</span>{{ t('invites.rules.item1', { ratio: ratioPercent }) }}</li>
        <li class="flex gap-2"><span class="text-primary">•</span>{{ t('invites.rules.item2') }}</li>
        <li class="flex gap-2"><span class="text-primary">•</span>{{ t('invites.rules.item3') }}</li>
      </ul>
    </Card>
  </div>
</template>
```

- [ ] **Step 5.2: 类型检查**

Run: `pnpm --filter @proapi/user typecheck`
Expected: 0 error。

如有报错(常见:`useToast` 路径、组件路径),按现有页面(如 `web/user/src/pages/logs/index.vue`)对照修正。

- [ ] **Step 5.3: Commit**

```bash
git add web/user/src/pages/invites.vue
git commit -m "feat(user): 实现 /invites 页(邀请码 + 统计 + 列表 + 规则)"
```

---

## Task 6: 路由切换 placeholder → invites.vue

**Files:**
- Modify: `web/user/src/router/index.ts:108-113`

- [ ] **Step 6.1: 编辑路由**

原代码:
```ts
  {
    path: '/invites',
    name: 'invites',
    component: () => import('@/pages/placeholder.vue'),
    meta: { auth: true, title: 'nav.invites', layout: 'app' },
  },
```

改为:
```ts
  {
    path: '/invites',
    name: 'invites',
    component: () => import('@/pages/invites.vue'),
    meta: { auth: true, title: 'nav.invites', layout: 'app' },
  },
```

- [ ] **Step 6.2: 类型检查**

Run: `pnpm --filter @proapi/user typecheck`
Expected: 0 error。

- [ ] **Step 6.3: Commit**

```bash
git add web/user/src/router/index.ts
git commit -m "feat(user): /invites 路由指向真页面 invites.vue"
```

---

## Task 7: 用户端 404 页改写

**Files:**
- Modify: `web/user/src/pages/404.vue`

复用 AuthLayout(已在 router 中配置 `layout: 'auth'`)。

- [ ] **Step 7.1: 整体改写**

将文件改为:

```vue
<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import Card from '@/components/ui/Card.vue'
import Button from '@/components/ui/Button.vue'

const router = useRouter()
const { t } = useI18n()
</script>

<template>
  <Card class="max-w-md w-full text-center">
    <div class="text-6xl font-bold text-primary mb-2 leading-none">404</div>
    <h2 class="text-xl font-semibold text-fg mb-2">{{ t('notfound.title') }}</h2>
    <p class="text-fg-muted text-sm mb-6">{{ t('notfound.desc') }}</p>
    <Button @click="router.push('/')">{{ t('common.back_home') }}</Button>
  </Card>
</template>
```

- [ ] **Step 7.2: 类型检查**

Run: `pnpm --filter @proapi/user typecheck`
Expected: 0 error。

- [ ] **Step 7.3: Commit**

```bash
git add web/user/src/pages/404.vue
git commit -m "feat(user): 404 页改写为真正的 Not Found 页"
```

---

## Task 8: admin i18n 增量(notfound/forbidden/common/auth)

**Files:**
- Modify: `web/admin/src/i18n/zh.json`
- Modify: `web/admin/src/i18n/en.json`

- [ ] **Step 8.1: 删 `nav.m3_coming`**

`zh.json` 第 32 行(`"m3_coming": "M3 模块 (开发中)",`)整行删,注意删后上一行末尾的逗号:

```jsonc
// 改前
    "settings": "系统设置",
    "m3_coming": "M3 模块 (开发中)",
    "logout": "退出"

// 改后
    "settings": "系统设置",
    "logout": "退出"
```

`en.json` 同位置同样删除 `"m3_coming": "M3 Modules (Coming)",`。

- [ ] **Step 8.2: zh.json 新增 4 个顶级 key**

在文件末尾闭合 `}` 之前追加(注意上一个顶级 key 末尾要补逗号):

```jsonc
  "notfound": {
    "title": "页面不存在",
    "desc": "你访问的页面不存在,请检查 URL 是否正确"
  },
  "forbidden": {
    "title": "无权访问",
    "desc": "你的账号没有访问此页面的权限"
  },
  "common": {
    "back_home": "返回首页"
  }
```

如果 `auth` 已是顶级 key,在 `auth` 对象内追加 `"relogin": "切换账号"`;否则新增:

```jsonc
  "auth": {
    "relogin": "切换账号"
  }
```

**先 grep 确认是否已有 `auth` / `common` 顶级 key 再做合并,避免覆盖。**

Run: `grep -n '"auth"\|"common"' web/admin/src/i18n/zh.json`

- [ ] **Step 8.3: en.json 对应英文**

```jsonc
  "notfound": {
    "title": "Page not found",
    "desc": "The page you requested doesn't exist or has been moved"
  },
  "forbidden": {
    "title": "Forbidden",
    "desc": "Your account doesn't have permission to access this page"
  },
  "common": {
    "back_home": "Back to home"
  }
```

合并 `auth.relogin`: `"relogin": "Switch account"`。

- [ ] **Step 8.4: JSON 校验**

Run:
```bash
node -e "JSON.parse(require('fs').readFileSync('web/admin/src/i18n/zh.json'))" && \
node -e "JSON.parse(require('fs').readFileSync('web/admin/src/i18n/en.json'))" && \
! grep -n 'm3_coming' web/admin/src/i18n/zh.json web/admin/src/i18n/en.json
```
Expected: 前两条无报错;最后一条 `grep` 无输出且 `!` 取反后退出码 0。

- [ ] **Step 8.5: Commit**

```bash
git add web/admin/src/i18n/zh.json web/admin/src/i18n/en.json
git commit -m "i18n(admin): 补 notfound/forbidden/common.back_home/auth.relogin,删 nav.m3_coming"
```

---

## Task 9: admin NotFound.vue 改写

**Files:**
- Modify: `web/admin/src/views/NotFound.vue`

- [ ] **Step 9.1: 整体改写**

```vue
<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { NResult, NButton } from 'naive-ui'

const router = useRouter()
const { t } = useI18n()
</script>

<template>
  <div class="p-8 flex items-center justify-center">
    <NResult
      status="404"
      :title="t('notfound.title')"
      :description="t('notfound.desc')"
    >
      <template #footer>
        <NButton type="primary" @click="router.push('/')">{{ t('common.back_home') }}</NButton>
      </template>
    </NResult>
  </div>
</template>
```

- [ ] **Step 9.2: 类型检查**

Run: `pnpm --filter @proapi/admin typecheck`
Expected: 0 error。

- [ ] **Step 9.3: Commit**

```bash
git add web/admin/src/views/NotFound.vue
git commit -m "feat(admin): NotFound 页改写为 NResult 404"
```

---

## Task 10: admin Forbidden.vue 改写

**Files:**
- Modify: `web/admin/src/views/Forbidden.vue`

- [ ] **Step 10.1: 确认 user store API**

Run: `grep -rn "useUserStore\|userStore.logout" web/admin/src/views web/admin/src/layouts | head`
Expected: 看到 `useUserStore()` 在 admin 中的导入路径(预期 `@/stores/user`,与 `AppLayout.vue` 一致)。

- [ ] **Step 10.2: 整体改写**

```vue
<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { NResult, NButton, NSpace } from 'naive-ui'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const { t } = useI18n()
const userStore = useUserStore()

async function onRelogin() {
  await userStore.logout()
  router.push({ name: 'login' })
}
</script>

<template>
  <div class="p-8 flex items-center justify-center">
    <NResult
      status="403"
      :title="t('forbidden.title')"
      :description="t('forbidden.desc')"
    >
      <template #footer>
        <NSpace>
          <NButton @click="router.push('/')">{{ t('common.back_home') }}</NButton>
          <NButton type="primary" @click="onRelogin">{{ t('auth.relogin') }}</NButton>
        </NSpace>
      </template>
    </NResult>
  </div>
</template>
```

如 Step 10.1 显示 user store 路径不同,按实际路径替换 `@/stores/user`,或参照 `AppLayout.vue` 中的导入语句。

- [ ] **Step 10.3: 类型检查**

Run: `pnpm --filter @proapi/admin typecheck`
Expected: 0 error。

- [ ] **Step 10.4: Commit**

```bash
git add web/admin/src/views/Forbidden.vue
git commit -m "feat(admin): Forbidden 页改写为 NResult 403 + 切换账号"
```

---

## Task 11: 删 m3_coming 菜单项

**Files:**
- Modify: `web/admin/src/layouts/AppLayout.vue:47-49`

- [ ] **Step 11.1: 删两行**

定位 `web/admin/src/layouts/AppLayout.vue` 第 47–49 行:

```ts
  { label: () => h(RouterLink, { to: '/settings' }, { default: () => t('nav.settings') }), key: 'settings' },
  { type: 'divider', key: 'div1' },
  { label: t('nav.m3_coming'), key: 'm3', disabled: true },
]
```

删除中间两行,保留 `settings` 与 `]`:

```ts
  { label: () => h(RouterLink, { to: '/settings' }, { default: () => t('nav.settings') }), key: 'settings' },
]
```

- [ ] **Step 11.2: 类型 + 残留检查**

Run:
```bash
pnpm --filter @proapi/admin typecheck && \
! grep -rn "m3_coming\|m3 模块\|M3 模块\|M3 Modules" web/admin/src
```
Expected: typecheck 通过;grep 无输出。

- [ ] **Step 11.3: Commit**

```bash
git add web/admin/src/layouts/AppLayout.vue
git commit -m "chore(admin): 删除 m3_coming 占位菜单项"
```

---

## Task 12: 删死代码(Home.vue / placeholder.vue)

**Files:**
- Delete: `web/user/src/pages/Home.vue`
- Delete: `web/user/src/pages/placeholder.vue`

- [ ] **Step 12.1: 确认无引用**

Run:
```bash
grep -rn "pages/Home\|pages/placeholder\|@/pages/Home\|@/pages/placeholder" web/user/src
```
Expected: 无输出(`router/index.ts` 在 Task 6 已切到 `invites.vue`)。

如果仍有命中,先解决引用再删除。

- [ ] **Step 12.2: git rm**

```bash
git rm web/user/src/pages/Home.vue web/user/src/pages/placeholder.vue
```

- [ ] **Step 12.3: 类型检查**

Run: `pnpm --filter @proapi/user typecheck`
Expected: 0 error。

- [ ] **Step 12.4: Commit**

```bash
git commit -m "chore(user): 删死代码 Home.vue / placeholder.vue"
```

---

## Task 13: 最终回归验证

**Files:** 无修改,仅运行验收检查。

- [ ] **Step 13.1: typecheck × 2**

Run:
```bash
pnpm --filter @proapi/user typecheck && \
pnpm --filter @proapi/admin typecheck
```
Expected: 两个工作区均 0 error 0 warning。

- [ ] **Step 13.2: build:demo × 2**

Run:
```bash
pnpm --filter @proapi/user build:demo && \
pnpm --filter @proapi/admin build:demo
```
Expected: 两个 demo 构建成功,产物落在 `web/user/dist/` 和 `web/admin/dist/`。

- [ ] **Step 13.3: "页面建设中" 必须 0 命中**

Run:
```bash
! grep -rn "页面建设" web/
```
Expected: grep 无输出,`!` 取反后退出码 0。如果有命中,定位并补漏。

- [ ] **Step 13.4: dev 烟测(可选但推荐)**

Run(后台启用户 demo dev):
```bash
pnpm --filter @proapi/user dev:demo
```
浏览器访问:
- `http://localhost:5173/` — 首页正常
- `http://localhost:5173/invites` — 邀请页(看到邀请码 + 3 统计 + 2 tab + 分页 + 规则)
- 点 "复制" 按钮 — toast 出现
- 切换到"返佣记录" tab — 数据出现(分页器在 30 条记录、size=10 时显示 3 页)
- `http://localhost:5173/no-such-route` — 真 404 页(大字号 "404" + "返回首页" 按钮)

同上跑 admin:
```bash
pnpm --filter @proapi/admin dev:demo
```
- `http://localhost:5174/no-such-route` — NResult 404
- `http://localhost:5174/forbidden` — NResult 403 + "切换账号" 按钮
- 左侧菜单不再有 "M3 模块 (开发中)" 项

- [ ] **Step 13.5: 最终 commit(收尾,可选)**

如果前 12 个 task 已经独立提交,Task 13 通常不产生新提交。若发现并修复了小 bug:

```bash
git add -A
git commit -m "fix: 邀请页回归发现的小问题"
```

---

## Self-Review 备忘

**Spec coverage(11 项 vs 任务):**

| Spec 项 | Task |
|---|---|
| 1. 新建 `/invites` 页 | Task 5 + Task 6 |
| 2. 新建 `api/invite.ts` | Task 2 |
| 3. 新建 3 份 mock fixture | Task 1 |
| 4. mock routes 3 条 | Task 3 |
| 5. 改写 user 404.vue | Task 7 |
| 6. 改写 admin NotFound.vue | Task 9 |
| 7. 改写 admin Forbidden.vue | Task 10 |
| 8. 删 `m3_coming` + i18n | Task 11 + Task 8 |
| 9. 删 Home.vue | Task 12 |
| 10. 删 placeholder.vue | Task 12 |
| 11. i18n 增量(user + admin) | Task 4 + Task 8 |
| 验收:typecheck/build/grep | Task 13 |

**关键类型一致性检查:**
- `InviteSummary.stats.invitee_count` / `rebate_credits_total` / `rebate_credits_month` — Task 1 fixture + Task 2 类型 + Task 5 模板,三处字段名一致 ✓
- `Invitee.email_masked` — fixture + 类型 + 模板 hidden sm 列,三处一致 ✓
- `InviteRecord.invitee_display_name` / `rebate_credits` — fixture + 类型 + 模板,三处一致 ✓
- `inviteApi.getSummary` / `listInvitees` / `listRecords` — Task 2 定义 + Task 5 使用,签名一致 ✓
- mock route `paginate(items, p)` 返回 `{items,total,page,size}` 与 `PageResp<T>` 字段名一致 ✓

**No placeholders:** 所有任务步骤都有具体 code block / 命令 / 预期输出,无 "TBD" / "类似 Task X" 等占位 ✓
