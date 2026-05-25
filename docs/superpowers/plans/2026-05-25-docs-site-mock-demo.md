# docs-site Mock 演示模式 实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让 proapi 文档站访客无需部署后端,直接点 nav 即可访问 admin / user 的可交互演示(纯前端 mock 数据)。

**Architecture:** 在 `web/shared/src/mock/` 提供按 URL 匹配的 mock 路由表;`admin`/`user` 的 `http.ts` 在 `VITE_DEMO_MOCK=true` 时短路 axios 走 mock;router / store 在 demo 模式跳过 auth、注入 fake session;`docs-site/scripts/build-demos.js` 构建两个前端到 `docs-site/public/{admin,user}-demo/`,VitePress nav 添加"在线演示"下拉。

**Tech Stack:** Vue 3 / Pinia / vue-router / Vite / VitePress / pnpm workspaces / cross-env

**TDD 说明:** 项目前端未引入单元测试框架(无 vitest/jest)。本计划与 spec § 6.2 对齐,以 `typecheck` + 手工 `pnpm run dev:demo` / `docs:build-with-demo` 验证替代单元测试。不为引入 mock 层而临时添加测试框架。

**对应设计稿:** `docs/superpowers/specs/2026-05-25-docs-site-mock-demo-设计稿.md`(commit `d6d9220`)

---

## File Map

新建:
- `web/shared/src/mock/index.ts`
- `web/shared/src/mock/routes.ts`
- `web/shared/src/mock/helpers.ts`
- `web/shared/src/mock/data/admin-user.json`
- `web/shared/src/mock/data/user-profile.json`
- `web/shared/src/mock/data/admin-wallet.json`
- `web/shared/src/mock/data/user-wallet.json`
- `web/shared/src/mock/data/channels.json`
- `web/shared/src/mock/data/admin-tokens.json`
- `web/shared/src/mock/data/user-tokens.json`
- `web/shared/src/mock/data/models.json`
- `web/shared/src/mock/data/stats-overview.json`
- `web/shared/src/mock/data/stats-timeseries.json`
- `web/shared/src/mock/data/stats-by-model.json`
- `web/shared/src/mock/data/stats-by-channel.json`
- `web/shared/src/mock/data/stats-by-user.json`
- `web/shared/src/mock/data/ledger.json`
- `web/shared/src/mock/data/usage.json`
- `web/shared/src/mock/data/notices.json`
- `web/shared/src/mock/data/log-requests.json`
- `web/shared/src/mock/data/log-errors.json`
- `web/shared/src/mock/data/log-audit.json`
- `web/shared/src/mock/data/admin-recharges.json`
- `web/shared/src/mock/data/user-recharges.json`
- `web/shared/src/mock/data/oauth-bindings.json`
- `docs-site/scripts/build-demos.js`

修改:
- `web/shared/package.json`(exports 加 `./mock`)
- `web/shared/tsconfig.json`(开启 resolveJsonModule)
- `web/admin/src/api/http.ts`(底层 get/post/patch/del 加 mock 分支)
- `web/admin/env.d.ts`(声明 `VITE_DEMO_MOCK`)
- `web/admin/src/router/guard.ts`(demo 模式跳过 auth 检查)
- `web/admin/src/router/index.ts`(history base 用 BASE_URL)
- `web/admin/package.json`(加 `build:demo`、`dev:demo` + `cross-env` devDep)
- `web/user/src/api/http.ts`(同 admin)
- `web/user/env.d.ts`(同 admin)
- `web/user/src/router/index.ts`(同 admin)
- `web/user/package.json`(同 admin)
- `docs-site/package.json`(加 `demo:build`、`docs:build-with-demo`)
- `docs-site/.vitepress/config.ts`(zh + en nav 各加"在线演示"下拉)
- `docs-site/zh/guide/introduction.md`(banner)
- `docs-site/zh/guide/quickstart.md`(banner)
- `docs-site/en/guide/introduction.md`(banner)
- `docs-site/en/guide/quickstart.md`(banner)
- `.gitignore`(加 demo 产物)

---

## Task 1: 创建 mock 工具模块(helpers + 空骨架)

**Files:**
- Create: `web/shared/src/mock/helpers.ts`
- Create: `web/shared/src/mock/routes.ts`
- Create: `web/shared/src/mock/index.ts`
- Modify: `web/shared/package.json`

- [ ] **Step 1: 创建 `helpers.ts`**

写入 `web/shared/src/mock/helpers.ts`:

```ts
export interface PageParams {
  page?: number | string
  size?: number | string
  [k: string]: unknown
}

export interface Page<T> {
  items: T[]
  total: number
  page: number
  size: number
}

export function clone<T>(v: T): T {
  return JSON.parse(JSON.stringify(v))
}

export function paginate<T>(items: T[], params?: PageParams): Page<T> {
  const page = Math.max(1, Number(params?.page ?? 1))
  const size = Math.max(1, Number(params?.size ?? 20))
  const start = (page - 1) * size
  return {
    items: clone(items.slice(start, start + size)),
    total: items.length,
    page,
    size,
  }
}

export function ok(extra?: Record<string, unknown>) {
  return { ok: true, ...(extra ?? {}) }
}
```

- [ ] **Step 2: 创建 `routes.ts`(空表占位,Task 3 填实)**

写入 `web/shared/src/mock/routes.ts`:

```ts
import type { PageParams } from './helpers'

export type MockMethod = 'GET' | 'POST' | 'PATCH' | 'DELETE'
export type MockHandler = (method: MockMethod, url: string, params?: unknown) => unknown

export interface MockRoute {
  pattern: RegExp
  handler: MockHandler
  methods?: MockMethod[]
}

export const routes: MockRoute[] = []

export function routeMock(method: MockMethod, url: string, params?: unknown):
  { matched: boolean; data: unknown } {
  const path = url.split('?')[0]
  for (const r of routes) {
    if (r.methods && !r.methods.includes(method)) continue
    if (r.pattern.test(path)) {
      return { matched: true, data: r.handler(method, url, params) }
    }
  }
  return { matched: false, data: null }
}

export type { PageParams }
```

- [ ] **Step 3: 创建 `index.ts`(入口)**

写入 `web/shared/src/mock/index.ts`:

```ts
import { routeMock, type MockMethod } from './routes'

export interface MockResult<T> {
  matched: boolean
  data: T | null
}

export async function matchMock<T = unknown>(
  method: MockMethod,
  url: string,
  params?: unknown,
): Promise<MockResult<T>> {
  const delay = 100 + Math.random() * 200
  await new Promise((r) => setTimeout(r, delay))
  const { matched, data } = routeMock(method, url, params)
  return { matched, data: data as T | null }
}

export type { MockMethod } from './routes'
export type { Page } from './helpers'
```

- [ ] **Step 4: 注册 shared 子路径导出**

修改 `web/shared/package.json`,在 `exports` 对象中追加 `"./mock"` 一行:

```json
{
  "name": "@proapi/shared",
  "version": "0.0.0",
  "private": true,
  "type": "module",
  "main": "src/index.ts",
  "types": "src/index.ts",
  "exports": {
    ".": "./src/index.ts",
    "./api/types": "./src/api/types.ts",
    "./api/codes": "./src/api/codes.ts",
    "./mock": "./src/mock/index.ts",
    "./i18n/zh": "./src/i18n/zh.json",
    "./i18n/en": "./src/i18n/en.json"
  },
  "scripts": {
    "typecheck": "tsc --noEmit"
  }
}
```

- [ ] **Step 5: typecheck 通过**

Run: `pnpm -C web/shared typecheck`
Expected: 无错误退出码 0。

- [ ] **Step 6: 提交**

```bash
git add web/shared/src/mock/ web/shared/package.json
git commit -m "feat(mock): mock 工具模块骨架(matchMock + helpers + routes 空表)"
```

---

## Task 2: 写 mock 数据 JSON 集

**Files:**
- Create: `web/shared/src/mock/data/admin-user.json`
- Create: `web/shared/src/mock/data/user-profile.json`
- Create: `web/shared/src/mock/data/admin-wallet.json`
- Create: `web/shared/src/mock/data/user-wallet.json`
- Create: `web/shared/src/mock/data/channels.json`
- Create: `web/shared/src/mock/data/admin-tokens.json`
- Create: `web/shared/src/mock/data/user-tokens.json`
- Create: `web/shared/src/mock/data/models.json`
- Create: `web/shared/src/mock/data/stats-overview.json`
- Create: `web/shared/src/mock/data/stats-timeseries.json`
- Create: `web/shared/src/mock/data/stats-by-model.json`
- Create: `web/shared/src/mock/data/stats-by-channel.json`
- Create: `web/shared/src/mock/data/stats-by-user.json`
- Create: `web/shared/src/mock/data/ledger.json`
- Create: `web/shared/src/mock/data/usage.json`
- Create: `web/shared/src/mock/data/notices.json`
- Create: `web/shared/src/mock/data/log-requests.json`
- Create: `web/shared/src/mock/data/log-errors.json`
- Create: `web/shared/src/mock/data/log-audit.json`
- Create: `web/shared/src/mock/data/admin-recharges.json`
- Create: `web/shared/src/mock/data/user-recharges.json`
- Create: `web/shared/src/mock/data/oauth-bindings.json`

字段已对照仓库实际 `web/{admin,user}/src/api/*.ts` 中的 interface 校准(读 `auth.ts/channel.ts/token.ts/stats.ts/model.ts/ledger.ts/log.ts/notice.ts/wallet.ts/recharge.ts` + user 的 `profile.ts/wallet.ts/token.ts/usage.ts/recharge.ts`)。

- [ ] **Step 1: `admin-user.json`(对应 `AdminUser`)**

```json
{
  "id": 1,
  "username": "admin",
  "email": "admin@proapi.demo",
  "display_name": "演示管理员",
  "avatar": null,
  "role": 3,
  "status": 1,
  "group_id": 1,
  "group_name": "默认分组",
  "email_verified_at": "2026-01-01T00:00:00Z",
  "last_login_at": "2026-05-25T10:00:00Z",
  "created_at": "2026-01-01T00:00:00Z"
}
```

- [ ] **Step 2: `user-profile.json`(对应 user `UserProfile`,`id` 是字符串)**

```json
{
  "id": "u_demo_1001",
  "email": "demo@proapi.demo",
  "display_name": "演示用户",
  "avatar_url": "",
  "created_at": "2026-01-01T00:00:00Z"
}
```

- [ ] **Step 3: `admin-wallet.json`(admin `Wallet`)**

```json
{
  "id": 1001,
  "owner_type": "user",
  "owner_id": 1001,
  "quota_balance": 8650000,
  "quota_total_recharged": 10000000,
  "quota_total_consumed": 1350000,
  "currency": "USD"
}
```

- [ ] **Step 4: `user-wallet.json`(user `WalletInfo`)**

```json
{
  "balance_usd": 86.5,
  "balance_cny": 620.4,
  "total_recharged_usd": 100.0,
  "total_consumed_usd": 13.5,
  "currency": "USD"
}
```

- [ ] **Step 5: `channels.json`(对应 admin `Channel`,覆盖 16 家适配器)**

```json
[
  { "id": 1, "name": "openai-main", "provider": "openai", "base_url": "https://api.openai.com", "priority": 100, "weight": 10, "status": 1, "tags": ["primary"], "extra": {}, "credentials_masked": { "api_key": "sk-****abcd" }, "created_at": "2026-01-15T08:00:00Z", "updated_at": "2026-05-20T10:00:00Z", "health": { "state": "closed", "consec_fail": 0, "opened_at": null }, "last_test": { "at": "2026-05-25T09:00:00Z", "ok": true, "latency_ms": 312 } },
  { "id": 2, "name": "anthropic-main", "provider": "anthropic", "base_url": "https://api.anthropic.com", "priority": 90, "weight": 8, "status": 1, "tags": ["claude"], "extra": {}, "credentials_masked": { "api_key": "sk-ant-****x9y2" }, "created_at": "2026-01-20T08:00:00Z", "updated_at": "2026-05-20T10:00:00Z", "health": { "state": "closed", "consec_fail": 0, "opened_at": null }, "last_test": { "at": "2026-05-25T09:00:00Z", "ok": true, "latency_ms": 420 } },
  { "id": 3, "name": "gemini-main", "provider": "gemini", "base_url": "https://generativelanguage.googleapis.com", "priority": 80, "weight": 6, "status": 1, "tags": ["google"], "extra": {}, "credentials_masked": { "api_key": "AIza****x7q1" }, "created_at": "2026-02-01T08:00:00Z", "updated_at": "2026-05-20T10:00:00Z", "health": { "state": "closed", "consec_fail": 0, "opened_at": null }, "last_test": { "at": "2026-05-25T09:00:00Z", "ok": true, "latency_ms": 540 } },
  { "id": 4, "name": "azure-east-us", "provider": "azure", "base_url": "https://contoso.openai.azure.com", "priority": 70, "weight": 5, "status": 1, "tags": ["enterprise"], "extra": { "deployment": "gpt-4" }, "credentials_masked": { "api_key": "****1234" }, "created_at": "2026-02-10T08:00:00Z", "updated_at": "2026-05-20T10:00:00Z", "health": { "state": "closed", "consec_fail": 0, "opened_at": null }, "last_test": { "at": "2026-05-25T09:00:00Z", "ok": true, "latency_ms": 380 } },
  { "id": 5, "name": "deepseek-main", "provider": "deepseek", "base_url": "https://api.deepseek.com", "priority": 60, "weight": 5, "status": 1, "tags": ["cn"], "extra": {}, "credentials_masked": { "api_key": "sk-****ds01" }, "created_at": "2026-02-15T08:00:00Z", "updated_at": "2026-05-20T10:00:00Z", "health": { "state": "closed", "consec_fail": 0, "opened_at": null }, "last_test": { "at": "2026-05-25T09:00:00Z", "ok": true, "latency_ms": 220 } },
  { "id": 6, "name": "moonshot-main", "provider": "moonshot", "base_url": "https://api.moonshot.cn", "priority": 55, "weight": 4, "status": 1, "tags": ["cn"], "extra": {}, "credentials_masked": { "api_key": "sk-****ms01" }, "created_at": "2026-02-18T08:00:00Z", "updated_at": "2026-05-20T10:00:00Z", "health": { "state": "closed", "consec_fail": 0, "opened_at": null }, "last_test": { "at": "2026-05-25T09:00:00Z", "ok": true, "latency_ms": 280 } },
  { "id": 7, "name": "zhipu-glm", "provider": "zhipu", "base_url": "https://open.bigmodel.cn", "priority": 50, "weight": 4, "status": 1, "tags": ["cn"], "extra": {}, "credentials_masked": { "api_key": "****glm9" }, "created_at": "2026-02-20T08:00:00Z", "updated_at": "2026-05-20T10:00:00Z", "health": { "state": "closed", "consec_fail": 0, "opened_at": null }, "last_test": { "at": "2026-05-25T09:00:00Z", "ok": true, "latency_ms": 260 } },
  { "id": 8, "name": "qwen-main", "provider": "qwen", "base_url": "https://dashscope.aliyuncs.com", "priority": 50, "weight": 4, "status": 1, "tags": ["cn"], "extra": {}, "credentials_masked": { "api_key": "sk-****qw01" }, "created_at": "2026-02-25T08:00:00Z", "updated_at": "2026-05-20T10:00:00Z", "health": { "state": "closed", "consec_fail": 0, "opened_at": null }, "last_test": { "at": "2026-05-25T09:00:00Z", "ok": true, "latency_ms": 240 } },
  { "id": 9, "name": "doubao-main", "provider": "doubao", "base_url": "https://ark.cn-beijing.volces.com", "priority": 45, "weight": 3, "status": 1, "tags": ["cn"], "extra": {}, "credentials_masked": { "api_key": "****db01" }, "created_at": "2026-03-01T08:00:00Z", "updated_at": "2026-05-20T10:00:00Z", "health": { "state": "closed", "consec_fail": 0, "opened_at": null }, "last_test": { "at": "2026-05-25T09:00:00Z", "ok": true, "latency_ms": 310 } },
  { "id": 10, "name": "minimax-main", "provider": "minimax", "base_url": "https://api.minimax.chat", "priority": 40, "weight": 3, "status": 1, "tags": ["cn"], "extra": {}, "credentials_masked": { "api_key": "****mm01" }, "created_at": "2026-03-05T08:00:00Z", "updated_at": "2026-05-20T10:00:00Z", "health": { "state": "closed", "consec_fail": 0, "opened_at": null }, "last_test": { "at": "2026-05-25T09:00:00Z", "ok": true, "latency_ms": 350 } },
  { "id": 11, "name": "tencent-hunyuan", "provider": "tencent", "base_url": "https://hunyuan.tencentcloudapi.com", "priority": 40, "weight": 3, "status": 1, "tags": ["cn"], "extra": {}, "credentials_masked": { "api_key": "****tc01" }, "created_at": "2026-03-10T08:00:00Z", "updated_at": "2026-05-20T10:00:00Z", "health": { "state": "closed", "consec_fail": 0, "opened_at": null }, "last_test": { "at": "2026-05-25T09:00:00Z", "ok": true, "latency_ms": 330 } },
  { "id": 12, "name": "groq-main", "provider": "groq", "base_url": "https://api.groq.com", "priority": 35, "weight": 3, "status": 1, "tags": ["fast"], "extra": {}, "credentials_masked": { "api_key": "gsk_****grq" }, "created_at": "2026-03-15T08:00:00Z", "updated_at": "2026-05-20T10:00:00Z", "health": { "state": "closed", "consec_fail": 0, "opened_at": null }, "last_test": { "at": "2026-05-25T09:00:00Z", "ok": true, "latency_ms": 95 } },
  { "id": 13, "name": "mistral-main", "provider": "mistral", "base_url": "https://api.mistral.ai", "priority": 30, "weight": 2, "status": 1, "tags": ["eu"], "extra": {}, "credentials_masked": { "api_key": "****mst" }, "created_at": "2026-03-20T08:00:00Z", "updated_at": "2026-05-20T10:00:00Z", "health": { "state": "closed", "consec_fail": 0, "opened_at": null }, "last_test": { "at": "2026-05-25T09:00:00Z", "ok": true, "latency_ms": 410 } },
  { "id": 14, "name": "yi-main", "provider": "yi", "base_url": "https://api.lingyiwanwu.com", "priority": 28, "weight": 2, "status": 1, "tags": ["cn"], "extra": {}, "credentials_masked": { "api_key": "****yi01" }, "created_at": "2026-03-25T08:00:00Z", "updated_at": "2026-05-20T10:00:00Z", "health": { "state": "closed", "consec_fail": 0, "opened_at": null }, "last_test": { "at": "2026-05-25T09:00:00Z", "ok": true, "latency_ms": 290 } },
  { "id": 15, "name": "openrouter-main", "provider": "openrouter", "base_url": "https://openrouter.ai", "priority": 25, "weight": 2, "status": 1, "tags": ["aggregator"], "extra": {}, "credentials_masked": { "api_key": "sk-or-****abcd" }, "created_at": "2026-04-01T08:00:00Z", "updated_at": "2026-05-20T10:00:00Z", "health": { "state": "closed", "consec_fail": 0, "opened_at": null }, "last_test": { "at": "2026-05-25T09:00:00Z", "ok": true, "latency_ms": 580 } },
  { "id": 16, "name": "huggingface-main", "provider": "huggingface", "base_url": "https://api-inference.huggingface.co", "priority": 20, "weight": 1, "status": 2, "tags": ["lab"], "extra": {}, "credentials_masked": { "api_key": "hf_****hgf" }, "created_at": "2026-04-10T08:00:00Z", "updated_at": "2026-05-20T10:00:00Z", "health": { "state": "open", "consec_fail": 3, "opened_at": "2026-05-24T18:00:00Z" }, "last_test": { "at": "2026-05-25T09:00:00Z", "ok": false, "latency_ms": 0, "error": "connection timeout" } }
]
```

- [ ] **Step 6: `admin-tokens.json`(对应 admin `Token`)**

```json
[
  { "id": 101, "user_id": 1001, "name": "demo-prod-token", "key_prefix": "sk-prx-", "quota_limit": 5000000, "quota_used": 1820000, "allowed_models": ["gpt-4o", "claude-3-5-sonnet"], "allowed_ips": [], "rpm_limit": 600, "tpm_limit": 200000, "expires_at": null, "last_used_at": "2026-05-25T08:45:12Z", "status": 1, "created_at": "2026-02-01T08:00:00Z" },
  { "id": 102, "user_id": 1001, "name": "demo-dev-token", "key_prefix": "sk-dev-", "quota_limit": 1000000, "quota_used": 234000, "allowed_models": [], "allowed_ips": ["127.0.0.1/32"], "rpm_limit": 60, "tpm_limit": 30000, "expires_at": "2026-12-31T23:59:59Z", "last_used_at": "2026-05-24T17:20:00Z", "status": 1, "created_at": "2026-02-10T08:00:00Z" },
  { "id": 103, "user_id": 1002, "name": "test-user-token", "key_prefix": "sk-tst-", "quota_limit": null, "quota_used": 47000, "allowed_models": [], "allowed_ips": [], "rpm_limit": 30, "tpm_limit": 10000, "expires_at": null, "last_used_at": null, "status": 1, "created_at": "2026-03-01T08:00:00Z" },
  { "id": 104, "user_id": 1003, "name": "expired-token", "key_prefix": "sk-old-", "quota_limit": 100000, "quota_used": 100000, "allowed_models": [], "allowed_ips": [], "rpm_limit": 60, "tpm_limit": 20000, "expires_at": "2026-04-30T00:00:00Z", "last_used_at": "2026-04-28T10:00:00Z", "status": 2, "created_at": "2026-01-15T08:00:00Z" },
  { "id": 105, "user_id": 1001, "name": "disabled-token", "key_prefix": "sk-off-", "quota_limit": 200000, "quota_used": 5000, "allowed_models": [], "allowed_ips": [], "rpm_limit": 60, "tpm_limit": 20000, "expires_at": null, "last_used_at": "2026-05-10T12:00:00Z", "status": 0, "created_at": "2026-03-20T08:00:00Z" }
]
```

- [ ] **Step 7: `user-tokens.json`(对应 user `TokenView`,注意 `id: string`、`status` 字符串、字段 `prefix` 不是 `key_prefix`)**

```json
[
  { "id": "tk_demo_a1b2", "name": "默认 Token", "prefix": "sk-prx-", "status": "enabled", "quota_used": 1820000, "quota_limit": 5000000, "allowed_models": ["gpt-4o", "claude-3-5-sonnet"], "allowed_ips": [], "rpm_limit": 600, "tpm_limit": 200000, "expires_at": null, "last_used_at": "2026-05-25T08:45:12Z", "created_at": "2026-02-01T08:00:00Z" },
  { "id": "tk_demo_c3d4", "name": "开发 Token", "prefix": "sk-dev-", "status": "enabled", "quota_used": 234000, "quota_limit": 1000000, "allowed_models": [], "allowed_ips": ["127.0.0.1/32"], "rpm_limit": 60, "tpm_limit": 30000, "expires_at": "2026-12-31T23:59:59Z", "last_used_at": "2026-05-24T17:20:00Z", "created_at": "2026-02-10T08:00:00Z" }
]
```

- [ ] **Step 8: `models.json`(对应 `ModelCatalog`)**

```json
[
  { "id": 1, "name": "gpt-4o", "family": "chat", "capabilities": ["chat", "vision", "tools"], "default_input_ratio": 5.0, "default_output_ratio": 15.0, "default_cached_ratio": 2.5, "default_reasoning_ratio": null, "max_input_tokens": 128000, "status": 1, "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-05-01T00:00:00Z" },
  { "id": 2, "name": "gpt-4o-mini", "family": "chat", "capabilities": ["chat", "tools"], "default_input_ratio": 0.15, "default_output_ratio": 0.6, "default_cached_ratio": 0.075, "default_reasoning_ratio": null, "max_input_tokens": 128000, "status": 1, "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-05-01T00:00:00Z" },
  { "id": 3, "name": "claude-3-5-sonnet", "family": "chat", "capabilities": ["chat", "vision", "tools"], "default_input_ratio": 3.0, "default_output_ratio": 15.0, "default_cached_ratio": 0.3, "default_reasoning_ratio": null, "max_input_tokens": 200000, "status": 1, "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-05-01T00:00:00Z" },
  { "id": 4, "name": "claude-3-opus", "family": "chat", "capabilities": ["chat", "vision"], "default_input_ratio": 15.0, "default_output_ratio": 75.0, "default_cached_ratio": 1.5, "default_reasoning_ratio": null, "max_input_tokens": 200000, "status": 1, "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-05-01T00:00:00Z" },
  { "id": 5, "name": "gemini-1.5-pro", "family": "chat", "capabilities": ["chat", "vision", "tools"], "default_input_ratio": 1.25, "default_output_ratio": 5.0, "default_cached_ratio": 0.3125, "default_reasoning_ratio": null, "max_input_tokens": 1000000, "status": 1, "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-05-01T00:00:00Z" },
  { "id": 6, "name": "deepseek-chat", "family": "chat", "capabilities": ["chat", "tools"], "default_input_ratio": 0.14, "default_output_ratio": 0.28, "default_cached_ratio": null, "default_reasoning_ratio": null, "max_input_tokens": 32768, "status": 1, "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-05-01T00:00:00Z" },
  { "id": 7, "name": "deepseek-reasoner", "family": "chat", "capabilities": ["chat", "reasoning"], "default_input_ratio": 0.55, "default_output_ratio": 2.19, "default_cached_ratio": 0.14, "default_reasoning_ratio": 2.19, "max_input_tokens": 64000, "status": 1, "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-05-01T00:00:00Z" },
  { "id": 8, "name": "qwen-max", "family": "chat", "capabilities": ["chat"], "default_input_ratio": 1.4, "default_output_ratio": 5.6, "default_cached_ratio": null, "default_reasoning_ratio": null, "max_input_tokens": 32000, "status": 1, "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-05-01T00:00:00Z" },
  { "id": 9, "name": "moonshot-v1-128k", "family": "chat", "capabilities": ["chat"], "default_input_ratio": 2.0, "default_output_ratio": 2.0, "default_cached_ratio": null, "default_reasoning_ratio": null, "max_input_tokens": 128000, "status": 1, "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-05-01T00:00:00Z" },
  { "id": 10, "name": "text-embedding-3-large", "family": "embed", "capabilities": ["embed"], "default_input_ratio": 0.13, "default_output_ratio": 0, "default_cached_ratio": null, "default_reasoning_ratio": null, "max_input_tokens": 8191, "status": 1, "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-05-01T00:00:00Z" },
  { "id": 11, "name": "dall-e-3", "family": "image", "capabilities": ["image"], "default_input_ratio": 40.0, "default_output_ratio": 0, "default_cached_ratio": null, "default_reasoning_ratio": null, "max_input_tokens": 4000, "status": 1, "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-05-01T00:00:00Z" },
  { "id": 12, "name": "whisper-1", "family": "audio", "capabilities": ["stt"], "default_input_ratio": 6.0, "default_output_ratio": 0, "default_cached_ratio": null, "default_reasoning_ratio": null, "max_input_tokens": 0, "status": 1, "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-05-01T00:00:00Z" }
]
```

- [ ] **Step 9: `stats-overview.json`(对应 admin `Overview`)**

```json
{
  "requests_today": 18234,
  "revenue_today": 124.36,
  "active_users": 47,
  "error_rate": 0.012,
  "delta": { "requests": 0.082, "revenue": 0.115, "users": 0.04, "error_rate": -0.003 }
}
```

- [ ] **Step 10: `stats-timeseries.json`(对应 `{ points: TimeseriesPoint[] }`)**

```json
{
  "points": [
    { "ts": "2026-05-19T00:00:00Z", "requests": 14210, "errors": 152, "quota": 480000 },
    { "ts": "2026-05-20T00:00:00Z", "requests": 15834, "errors": 187, "quota": 522000 },
    { "ts": "2026-05-21T00:00:00Z", "requests": 16021, "errors": 174, "quota": 540000 },
    { "ts": "2026-05-22T00:00:00Z", "requests": 17400, "errors": 211, "quota": 590000 },
    { "ts": "2026-05-23T00:00:00Z", "requests": 17995, "errors": 198, "quota": 614000 },
    { "ts": "2026-05-24T00:00:00Z", "requests": 18012, "errors": 205, "quota": 622000 },
    { "ts": "2026-05-25T00:00:00Z", "requests": 18234, "errors": 218, "quota": 630000 }
  ]
}
```

- [ ] **Step 11: `stats-by-model.json`(`{ rows: ByModelRow[] }`)**

```json
{
  "rows": [
    { "model": "gpt-4o", "requests": 6420, "tokens_in": 2150000, "tokens_out": 312000, "quota": 13780000, "errors": 18 },
    { "model": "claude-3-5-sonnet", "requests": 4180, "tokens_in": 1840000, "tokens_out": 256000, "quota": 9430000, "errors": 12 },
    { "model": "gpt-4o-mini", "requests": 3920, "tokens_in": 980000, "tokens_out": 124000, "quota": 220000, "errors": 8 },
    { "model": "deepseek-chat", "requests": 2010, "tokens_in": 540000, "tokens_out": 87000, "quota": 100000, "errors": 4 },
    { "model": "gemini-1.5-pro", "requests": 1704, "tokens_in": 720000, "tokens_out": 95000, "quota": 1380000, "errors": 7 }
  ]
}
```

- [ ] **Step 12: `stats-by-channel.json`(`{ rows: ByChannelRow[] }`)**

```json
{
  "rows": [
    { "channel_id": 1, "channel_name": "openai-main", "provider": "openai", "requests": 10340, "quota": 14000000, "errors": 26 },
    { "channel_id": 2, "channel_name": "anthropic-main", "provider": "anthropic", "requests": 4180, "quota": 9430000, "errors": 12 },
    { "channel_id": 3, "channel_name": "gemini-main", "provider": "gemini", "requests": 1704, "quota": 1380000, "errors": 7 },
    { "channel_id": 5, "channel_name": "deepseek-main", "provider": "deepseek", "requests": 2010, "quota": 100000, "errors": 4 }
  ]
}
```

- [ ] **Step 13: `stats-by-user.json`(`{ rows: ByUserRow[] }`)**

```json
{
  "rows": [
    { "user_id": 1001, "username": "demo_user", "requests": 12450, "quota": 18200000, "errors": 28 },
    { "user_id": 1002, "username": "test_user", "requests": 3120, "quota": 4700000, "errors": 11 },
    { "user_id": 1003, "username": "legacy_user", "requests": 2664, "quota": 1830000, "errors": 9 }
  ]
}
```

- [ ] **Step 14: `ledger.json`(对应 `LedgerEntry`,字段 `direction/amount_quota/amount_money/currency/ref_type/balance_after`)**

```json
[
  { "id": 5001, "wallet_id": 1001, "direction": "credit", "amount_quota": 10000000, "amount_money": 100.0, "currency": "USD", "ref_type": "manual", "ref_id": 9001, "balance_after": 10000000, "description": "初始充值", "created_at": "2026-01-01T00:00:00Z" },
  { "id": 5002, "wallet_id": 1001, "direction": "debit",  "amount_quota": 4200, "amount_money": 0.042, "currency": "USD", "ref_type": "usage", "ref_id": 70001, "balance_after": 9995800, "description": "gpt-4o 用量结算", "created_at": "2026-03-12T08:32:00Z" },
  { "id": 5003, "wallet_id": 1001, "direction": "debit",  "amount_quota": 18000, "amount_money": 0.18, "currency": "USD", "ref_type": "usage", "ref_id": 70002, "balance_after": 9977800, "description": "claude-3-5-sonnet 用量结算", "created_at": "2026-04-05T14:12:00Z" },
  { "id": 5004, "wallet_id": 1001, "direction": "credit", "amount_quota": 500000, "amount_money": 5.0, "currency": "USD", "ref_type": "redeem", "ref_id": 9100, "balance_after": 10477800, "description": "兑换码 DEMO500", "created_at": "2026-04-20T10:00:00Z" },
  { "id": 5005, "wallet_id": 1001, "direction": "debit",  "amount_quota": 1827800, "amount_money": 18.278, "currency": "USD", "ref_type": "usage", "ref_id": 70003, "balance_after": 8650000, "description": "近期累计用量", "created_at": "2026-05-25T08:00:00Z" }
]
```

- [ ] **Step 15: `usage.json`(对应 user `UsageStat`,handler 内会按 `range` query 参数返回同一份数据 — 静态 demo 不区分)**

```json
{
  "cost_usd": 1.34,
  "request_count": 218,
  "token_count": 96400,
  "range": "today"
}
```

- [ ] **Step 16: `notices.json`(对应 `Notice` 列表)**

```json
[
  { "id": 1, "title": "演示公告:欢迎试用", "content": "这是一个静态演示环境,所有写操作不会生效。", "level": "info", "target": "all", "status": 1, "publish_at": "2026-05-01T00:00:00Z", "expires_at": null, "pinned": true, "created_by": 1, "created_at": "2026-05-01T00:00:00Z", "updated_at": "2026-05-01T00:00:00Z" },
  { "id": 2, "title": "新增 Doubao / MiniMax 适配器", "content": "已上线 2 家新适配器,可在渠道列表中查看。", "level": "success", "target": "user", "status": 1, "publish_at": "2026-05-20T00:00:00Z", "expires_at": null, "pinned": false, "created_by": 1, "created_at": "2026-05-20T00:00:00Z", "updated_at": "2026-05-20T00:00:00Z" },
  { "id": 3, "title": "HuggingFace 渠道暂时熔断", "content": "上游连接超时,已自动跳过该渠道。", "level": "warning", "target": "admin", "status": 1, "publish_at": "2026-05-24T18:00:00Z", "expires_at": null, "pinned": false, "created_by": 1, "created_at": "2026-05-24T18:00:00Z", "updated_at": "2026-05-24T18:00:00Z" }
]
```

- [ ] **Step 17: `log-requests.json`(对应 `RequestLog`)**

```json
[
  { "id": 70001, "created_at": "2026-05-25T08:30:12Z", "user_id": 1001, "token_id": 101, "group_id": 1, "event_type": 0, "client_model": "gpt-4o", "upstream_model": "gpt-4o", "channel_id": 1, "protocol": "openai", "endpoint": "/v1/chat/completions", "ip": "127.0.0.1", "status_code": 200, "latency_ms": 1820, "ttft_ms": 412, "stream": true, "input_tokens": 420, "output_tokens": 188, "cached_tokens": 0, "reasoning_tokens": 0, "total_quota": 4920, "billing_input_ratio": 5.0, "billing_output_ratio": 15.0, "billing_group_ratio": 1.0, "error_code": 0, "error_msg": "", "trace_id": "tr_demo_001" },
  { "id": 70002, "created_at": "2026-05-25T08:35:44Z", "user_id": 1001, "token_id": 101, "group_id": 1, "event_type": 0, "client_model": "claude-3-5-sonnet", "upstream_model": "claude-3-5-sonnet-20241022", "channel_id": 2, "protocol": "anthropic", "endpoint": "/v1/messages", "ip": "127.0.0.1", "status_code": 200, "latency_ms": 2410, "ttft_ms": 580, "stream": false, "input_tokens": 1200, "output_tokens": 320, "cached_tokens": 0, "reasoning_tokens": 0, "total_quota": 8400, "billing_input_ratio": 3.0, "billing_output_ratio": 15.0, "billing_group_ratio": 1.0, "error_code": 0, "error_msg": "", "trace_id": "tr_demo_002" },
  { "id": 70003, "created_at": "2026-05-25T08:42:18Z", "user_id": 1002, "token_id": 103, "group_id": 1, "event_type": 0, "client_model": "gpt-4o-mini", "upstream_model": "gpt-4o-mini", "channel_id": 1, "protocol": "openai", "endpoint": "/v1/chat/completions", "ip": "10.0.0.5", "status_code": 429, "latency_ms": 50, "ttft_ms": 0, "stream": false, "input_tokens": 0, "output_tokens": 0, "cached_tokens": 0, "reasoning_tokens": 0, "total_quota": 0, "billing_input_ratio": 0.15, "billing_output_ratio": 0.6, "billing_group_ratio": 1.0, "error_code": 4290, "error_msg": "rate limited", "trace_id": "tr_demo_003" }
]
```

- [ ] **Step 18: `log-errors.json`(对应 `ErrorLog`)**

```json
[
  { "id": 80001, "created_at": "2026-05-24T18:00:00Z", "user_id": null, "token_id": null, "channel_id": 16, "error_code": 5021, "error_type": "channel.timeout", "stack": "Get \"https://api-inference.huggingface.co/...\": context deadline exceeded", "context": { "channel_name": "huggingface-main", "endpoint": "/v1/chat/completions" }, "trace_id": "tr_demo_err_001" },
  { "id": 80002, "created_at": "2026-05-25T08:42:18Z", "user_id": 1002, "token_id": 103, "channel_id": 1, "error_code": 4290, "error_type": "ratelimit", "stack": "token rpm limit reached", "context": { "limit": 30, "actual_rpm": 33 }, "trace_id": "tr_demo_003" }
]
```

- [ ] **Step 19: `log-audit.json`(对应 `AuditLog`)**

```json
[
  { "id": 90001, "created_at": "2026-05-25T07:00:00Z", "actor_id": 1, "actor_role": 3, "action": "channel.create", "target_type": "channel", "target_id": 16, "before": null, "after": { "name": "huggingface-main", "provider": "huggingface" }, "ip": "127.0.0.1" },
  { "id": 90002, "created_at": "2026-05-25T07:30:00Z", "actor_id": 1, "actor_role": 3, "action": "token.update", "target_type": "token", "target_id": 105, "before": { "status": 1 }, "after": { "status": 0 }, "ip": "127.0.0.1" }
]
```

- [ ] **Step 20: `admin-recharges.json`(对应 admin `ManualRecharge`)**

```json
[
  { "id": 9001, "user_id": 1001, "username": "demo_user", "amount_money": 700, "currency": "CNY", "amount_quota": 10000000, "status": 1, "applicant_note": "首次充值", "reviewer_id": 1, "review_note": "已确认转账", "reviewed_at": "2026-01-01T00:00:00Z", "created_at": "2025-12-31T18:00:00Z", "updated_at": "2026-01-01T00:00:00Z" },
  { "id": 9002, "user_id": 1002, "username": "test_user", "amount_money": 100, "currency": "CNY", "amount_quota": 0, "status": 0, "applicant_note": "测试充值", "reviewer_id": null, "review_note": "", "reviewed_at": null, "created_at": "2026-05-24T10:00:00Z", "updated_at": "2026-05-24T10:00:00Z" }
]
```

- [ ] **Step 21: `user-recharges.json`(对应 user `ManualRechargeOrder`,注意 `id: string`、`status` 字符串)**

```json
{
  "items": [
    { "id": "or_demo_a1", "amount_cny": 700, "remark": "首次充值", "status": "approved", "granted_usd": 100, "created_at": "2025-12-31T18:00:00Z", "updated_at": "2026-01-01T00:00:00Z" },
    { "id": "or_demo_a2", "amount_cny": 100, "remark": "补充额度", "status": "pending", "created_at": "2026-05-24T10:00:00Z", "updated_at": "2026-05-24T10:00:00Z" }
  ],
  "total": 2
}
```

- [ ] **Step 22: `oauth-bindings.json`(对应 `OAuthBinding[]`)**

```json
[
  { "provider": "github", "provider_user_id": "987654", "provider_login": "demo-octocat", "bound_at": "2026-01-15T08:00:00Z" }
]
```

- [ ] **Step 23: 验证 JSON 合法**

```bash
for f in web/shared/src/mock/data/*.json; do node -e "JSON.parse(require('fs').readFileSync(process.argv[1],'utf8')); console.log('ok',process.argv[1])" "$f"; done
```

Expected: 每个文件输出 `ok <path>`。

- [ ] **Step 24: 提交**

```bash
git add web/shared/src/mock/data/
git commit -m "feat(mock): 添加演示数据 fixtures(16 渠道/模型/统计/计费/日志等)"
```

---

## Task 3: 把路由表填实

**Files:**
- Modify: `web/shared/src/mock/routes.ts`
- Possibly modify: `web/shared/tsconfig.json`(开启 resolveJsonModule)

- [ ] **Step 1: 改写 `routes.ts`(URL 模式对照实际 api/*.ts 调用)**

```ts
import { paginate, ok, clone, type PageParams } from './helpers'

import adminUser from './data/admin-user.json'
import userProfile from './data/user-profile.json'
import adminWallet from './data/admin-wallet.json'
import userWallet from './data/user-wallet.json'
import channels from './data/channels.json'
import adminTokens from './data/admin-tokens.json'
import userTokens from './data/user-tokens.json'
import models from './data/models.json'
import statsOverview from './data/stats-overview.json'
import statsTimeseries from './data/stats-timeseries.json'
import statsByModel from './data/stats-by-model.json'
import statsByChannel from './data/stats-by-channel.json'
import statsByUser from './data/stats-by-user.json'
import ledger from './data/ledger.json'
import usage from './data/usage.json'
import notices from './data/notices.json'
import logRequests from './data/log-requests.json'
import logErrors from './data/log-errors.json'
import logAudit from './data/log-audit.json'
import adminRecharges from './data/admin-recharges.json'
import userRecharges from './data/user-recharges.json'
import oauthBindings from './data/oauth-bindings.json'

export type MockMethod = 'GET' | 'POST' | 'PATCH' | 'DELETE'
export type MockHandler = (method: MockMethod, url: string, params?: unknown) => unknown

export interface MockRoute {
  pattern: RegExp
  handler: MockHandler
  methods?: MockMethod[]
}

const writeOk: MockHandler = () => ok()

const channelById = (url: string) => {
  const m = url.match(/\/channels\/(\d+)/)
  const id = m ? Number(m[1]) : 0
  const found = (channels as any[]).find((c) => c.id === id)
  return found ? clone(found) : clone((channels as any[])[0])
}

const userTokenById = (url: string) => {
  const m = url.match(/\/tokens\/([^/?]+)/)
  const id = m ? m[1] : ''
  const found = (userTokens as any[]).find((t) => t.id === id)
  return found ? clone(found) : clone((userTokens as any[])[0])
}

export const routes: MockRoute[] = [
  // ============ admin =============
  { pattern: /^\/api\/admin\/auth\/login$/,                handler: () => ({ user: clone(adminUser), session: { id: 'demo-session', expires_at: '2099-12-31T23:59:59Z' } }) },
  { pattern: /^\/api\/admin\/auth\/logout$/,               handler: writeOk },
  { pattern: /^\/api\/admin\/auth\/me$/,                   handler: () => clone(adminUser) },

  { pattern: /^\/api\/admin\/stats\/overview$/,            handler: () => clone(statsOverview) },
  { pattern: /^\/api\/admin\/stats\/timeseries$/,          handler: () => clone(statsTimeseries) },
  { pattern: /^\/api\/admin\/stats\/by_model$/,            handler: () => clone(statsByModel) },
  { pattern: /^\/api\/admin\/stats\/by_channel$/,          handler: () => clone(statsByChannel) },
  { pattern: /^\/api\/admin\/stats\/by_user$/,             handler: () => clone(statsByUser) },

  { pattern: /^\/api\/admin\/channels\/\d+\/mappings$/,    handler: () => [] },
  { pattern: /^\/api\/admin\/channels\/\d+\/?$/,           handler: (m, u) => m === 'GET' ? channelById(u) : ok() },
  { pattern: /^\/api\/admin\/channels$/,                   handler: (m, _u, p) => m === 'GET' ? paginate(channels as any[], p as PageParams) : ok() },

  { pattern: /^\/api\/admin\/tokens\/\d+$/,                handler: (m, _u, _p) => m === 'GET' ? clone((adminTokens as any[])[0]) : ok() },
  { pattern: /^\/api\/admin\/tokens$/,                     handler: (_m, _u, p) => paginate(adminTokens as any[], p as PageParams) },

  { pattern: /^\/api\/admin\/model_catalogs\/\d+$/,        handler: (m, _u, _p) => m === 'GET' ? clone((models as any[])[0]) : ok() },
  { pattern: /^\/api\/admin\/model_catalogs$/,             handler: (m, _u, p) => m === 'GET' ? paginate(models as any[], p as PageParams) : ok() },

  { pattern: /^\/api\/admin\/logs\/requests$/,             handler: (_m, _u, p) => paginate(logRequests as any[], p as PageParams) },
  { pattern: /^\/api\/admin\/logs\/errors$/,               handler: (_m, _u, p) => paginate(logErrors as any[], p as PageParams) },
  { pattern: /^\/api\/admin\/logs\/audit$/,                handler: (_m, _u, p) => paginate(logAudit as any[], p as PageParams) },

  { pattern: /^\/api\/admin\/notices\/\d+\/(publish|unpublish)$/, handler: () => clone((notices as any[])[0]) },
  { pattern: /^\/api\/admin\/notices\/\d+$/,               handler: (m, _u, _p) => m === 'GET' ? clone((notices as any[])[0]) : ok() },
  { pattern: /^\/api\/admin\/notices$/,                    handler: (m, _u, p) => m === 'GET' ? paginate(notices as any[], p as PageParams) : ok() },

  { pattern: /^\/api\/admin\/payments\/manual_recharges\/\d+\/(approve|reject)$/, handler: () => clone((adminRecharges as any[])[0]) },
  { pattern: /^\/api\/admin\/payments\/manual_recharges\/\d+$/,                   handler: () => clone((adminRecharges as any[])[0]) },
  { pattern: /^\/api\/admin\/payments\/manual_recharges$/,                        handler: (_m, _u, p) => paginate(adminRecharges as any[], p as PageParams) },

  { pattern: /^\/api\/admin\/users\/\d+\/quota$/,          handler: () => ({ ok: true, balance_after: 8650000 }) },

  // 兜底:其他 admin 子模块返回空(分页或数组),避免白屏
  { pattern: /^\/api\/admin\/(pricing|ratelimit|groups?|users|settings|redeem|payments)/, handler: (_m, _u, p) => paginate([], p as PageParams) },

  // ============ user =============
  { pattern: /^\/api\/user\/profile$/,                     handler: (m, _u, _p) => m === 'GET' ? clone(userProfile) : clone(userProfile) },
  { pattern: /^\/api\/user\/password$/,                    handler: writeOk },
  { pattern: /^\/api\/user\/oauth\/bindings(\/[^/?]+)?$/,  handler: () => clone(oauthBindings) },
  { pattern: /^\/api\/auth\/oauth\/github\/start$/,        handler: () => ({ redirect_url: '#demo' }) },

  { pattern: /^\/api\/user\/wallet\/ledger$/,              handler: (_m, _u, p) => paginate(ledger as any[], p as PageParams) },
  { pattern: /^\/api\/user\/wallet$/,                      handler: () => clone(userWallet) },

  { pattern: /^\/api\/user\/tokens\/[^/?]+\/regenerate$/,  handler: () => ({ view: clone((userTokens as any[])[0]), plaintext_key: 'sk-prx-demo-regenerated-xxxxxxxxxxxx' }) },
  { pattern: /^\/api\/user\/tokens\/[^/?]+$/,              handler: (m, u, _p) => m === 'GET' ? userTokenById(u) : (m === 'POST' ? { view: clone((userTokens as any[])[0]), plaintext_key: 'sk-prx-demo-xxxxxxxxxxxx' } : ok()) },
  { pattern: /^\/api\/user\/tokens(\?.*)?$/,               handler: (m, _u, _p) => m === 'GET' ? clone({ items: userTokens, total: (userTokens as any[]).length }) : ({ view: clone((userTokens as any[])[0]), plaintext_key: 'sk-prx-demo-new-xxxxxxxxxxxx' }) },

  { pattern: /^\/api\/user\/usage(\?.*)?$/,                handler: () => clone(usage) },

  { pattern: /^\/api\/user\/payment\/manual\/[^/?]+\/cancel$/, handler: writeOk },
  { pattern: /^\/api\/user\/payment\/manual\/[^/?]+$/,         handler: () => clone((userRecharges as any).items[0]) },
  { pattern: /^\/api\/user\/payment\/manual(\?.*)?$/,          handler: (m) => m === 'GET' ? clone(userRecharges) : clone((userRecharges as any).items[0]) },

  { pattern: /^\/api\/user\/notices/,                      handler: () => clone(notices) },
]

export function routeMock(method: MockMethod, url: string, params?: unknown):
  { matched: boolean; data: unknown } {
  const path = url.split('?')[0]
  for (const r of routes) {
    if (r.methods && !r.methods.includes(method)) continue
    if (r.pattern.test(path) || r.pattern.test(url)) {
      return { matched: true, data: r.handler(method, url, params) }
    }
  }
  return { matched: false, data: null }
}

export type { PageParams }
```

注:user 的部分 URL 直接拼 `?page=&page_size=` 进 url 字符串(见 `user/api/token.ts` 和 `user/api/usage.ts`),所以正则末尾允许带 `(\?.*)?`,匹配时同时尝试 `path` 与原 `url`。

- [ ] **Step 2: 确保 shared 包 tsconfig 开启 `resolveJsonModule`**

检查 `web/shared/tsconfig.json`,在 `compilerOptions` 内补齐(若已存在则跳过):

```json
{
  "compilerOptions": {
    "resolveJsonModule": true,
    "esModuleInterop": true
  }
}
```

- [ ] **Step 3: typecheck**

Run: `pnpm -C web/shared typecheck`
Expected: 通过。

- [ ] **Step 4: 提交**

```bash
git add web/shared/src/mock/routes.ts web/shared/tsconfig.json
git commit -m "feat(mock): 路由表接入演示数据(对齐真实 API URL)"
```

---

## Task 4: admin `http.ts` 加 mock 分支

**Files:**
- Modify: `web/admin/src/api/http.ts`

- [ ] **Step 1: 修改 `web/admin/src/api/http.ts`**

在文件末尾的 `get/post/patch/del` 之前,加 MOCK 常量与 helper,然后改写 4 个导出函数:

```ts
const MOCK = import.meta.env.VITE_DEMO_MOCK === 'true'

async function mockOr<T>(method: 'GET' | 'POST' | 'PATCH' | 'DELETE', url: string, params?: unknown): Promise<T> {
  const { matchMock } = await import('@proapi/shared/mock')
  const { matched, data } = await matchMock<T>(method, url, params)
  if (!matched) console.warn(`[mock] no data for: ${method} ${url}`)
  return (data ?? ([] as unknown)) as T
}

export async function get<T>(url: string, params?: Record<string, unknown>, cfg?: AxiosRequestConfig) {
  if (MOCK) return mockOr<T>('GET', url, params)
  return http.get<T>(url, { params, ...cfg }).then((r) => r.data)
}
export async function post<T>(url: string, body?: unknown, cfg?: AxiosRequestConfig) {
  if (MOCK) return mockOr<T>('POST', url, body)
  return http.post<T>(url, body, cfg).then((r) => r.data)
}
export async function patch<T>(url: string, body?: unknown, cfg?: AxiosRequestConfig) {
  if (MOCK) return mockOr<T>('PATCH', url, body)
  return http.patch<T>(url, body, cfg).then((r) => r.data)
}
export async function del<T>(url: string, params?: Record<string, unknown>, cfg?: AxiosRequestConfig) {
  if (MOCK) return mockOr<T>('DELETE', url, params)
  return http.delete<T>(url, { params, ...cfg }).then((r) => r.data)
}
```

注意保留原 `Page<T>`、`Ok` 类型导出和 axios 实例 / 拦截器,**只改 4 个导出函数体**。

- [ ] **Step 2: 加 admin 的 vite-env.d.ts 声明(若需)**

检查 `web/admin/env.d.ts`,确保 `VITE_DEMO_MOCK` 被 TS 识别。若未声明,在文件末尾追加:

```ts
interface ImportMetaEnv {
  readonly VITE_DEMO_MOCK?: string
}
interface ImportMeta {
  readonly env: ImportMetaEnv
}
```

(`vite/client` 已提供 ImportMeta.env,此处只补声明的字段。若 env.d.ts 已 reference `vite/client`,加自定义 env interface 即可。)

- [ ] **Step 3: typecheck**

Run: `pnpm -C web/admin typecheck`
Expected: 通过。

- [ ] **Step 4: 提交**

```bash
git add web/admin/src/api/http.ts web/admin/env.d.ts
git commit -m "feat(admin): http.ts 加 VITE_DEMO_MOCK 分支"
```

---

## Task 5: admin demo 模式跳过 auth guard

**Files:**
- Modify: `web/admin/src/router/guard.ts`

`stores/user.ts` 不需要改动:mock 模式下 `authApi.me()` 已通过 `http.ts` 的 mock 分支返回 fake user,调用现有 `fetchMe()` 即可注入伪 session。

- [ ] **Step 1: 修改 `web/admin/src/router/guard.ts`**

整体改写:

```ts
import type { Router } from 'vue-router'
import { useUserStore } from '@/stores/user'

const MOCK = import.meta.env.VITE_DEMO_MOCK === 'true'

export function installGuards(router: Router) {
  router.beforeEach(async (to) => {
    if (MOCK) {
      const us = useUserStore()
      if (!us.fetched) {
        try { await us.fetchMe() } catch (_) { /* mock 不会失败 */ }
      }
      if (to.name === 'login') {
        return { name: 'dashboard' }
      }
      return true
    }

    const publicRoutes = ['login', 'forbidden', 'not-found']
    if (typeof to.name === 'string' && publicRoutes.includes(to.name)) return true

    const us = useUserStore()
    if (!us.fetched) {
      try { await us.fetchMe() } catch (_) { /* 401 handled by http interceptor */ }
    }

    if (!us.user) {
      return { name: 'login', query: { redirect: to.fullPath } }
    }

    const roles = (to.meta?.roles as number[] | undefined) ?? []
    if (roles.length > 0 && !roles.includes(us.user.role)) {
      return { name: 'forbidden' }
    }

    return true
  })

  router.afterEach((to) => {
    const title = (to.meta?.title as string | undefined) ?? 'proapi admin'
    if (typeof document !== 'undefined') {
      document.title = `${title} · proapi admin`
    }
  })
}
```

**关于路由名 `dashboard`**:已对照 `web/admin/src/router/routes.ts` 第 17 行确认 admin 首页路由 `name: 'dashboard'`(实施前可再次 grep `name: 'dashboard'` 验证)。

- [ ] **Step 2: typecheck**

Run: `pnpm -C web/admin typecheck`
Expected: 通过。

- [ ] **Step 3: 提交**

```bash
git add web/admin/src/router/guard.ts
git commit -m "feat(admin): demo 模式跳过 auth guard,login 自动跳首页"
```

---

## Task 6: admin router base + build:demo 脚本

**Files:**
- Modify: `web/admin/src/router/index.ts`
- Modify: `web/admin/package.json`

- [ ] **Step 1: 修改 `web/admin/src/router/index.ts`,history base 走 BASE_URL**

```ts
import { createRouter, createWebHistory } from 'vue-router'
import { routes } from './routes'
import { installGuards } from './guard'

const baseUrl = import.meta.env.VITE_DEMO_MOCK === 'true'
  ? import.meta.env.BASE_URL
  : '/admin/'

export const router = createRouter({
  history: createWebHistory(baseUrl),
  routes,
})

installGuards(router)
```

- [ ] **Step 2: 加 `build:demo` 脚本 + cross-env 依赖**

修改 `web/admin/package.json` 中的 `scripts` 和 `devDependencies`:

```json
{
  "scripts": {
    "dev": "vite --port 5173 --host 127.0.0.1",
    "dev:demo": "cross-env VITE_DEMO_MOCK=true vite --port 5173 --host 127.0.0.1",
    "build": "vue-tsc --noEmit && vite build",
    "build:demo": "cross-env VITE_DEMO_MOCK=true vue-tsc --noEmit && vite build --mode demo",
    "preview": "vite preview --port 5173",
    "typecheck": "vue-tsc --noEmit",
    "lint": "eslint . --ext .ts,.vue --max-warnings 0",
    "format": "prettier --write \"src/**/*.{ts,vue,css}\""
  }
}
```

并在 `devDependencies` 中追加 `"cross-env": "^7.0.3"`(保持已有项不变)。

- [ ] **Step 3: 安装新依赖**

Run: `pnpm install`
Expected: cross-env 出现在 lockfile 中。

- [ ] **Step 4: 验证 build:demo 跑通**

Run: `pnpm -C web/admin run build:demo -- --base=/admin-demo/`
Expected:
- 退出码 0
- `web/admin/dist/index.html` 存在
- `dist/index.html` 中的资源路径形如 `/admin-demo/assets/...`(用 `grep "/admin-demo/" web/admin/dist/index.html` 验证)

- [ ] **Step 5: 提交**

```bash
git add web/admin/src/router/index.ts web/admin/package.json pnpm-lock.yaml
git commit -m "feat(admin): build:demo 脚本 + router base 走 BASE_URL"
```

---

## Task 7: user `http.ts` 加 mock 分支

**Files:**
- Modify: `web/user/src/api/http.ts`

- [ ] **Step 1: 修改 `web/user/src/api/http.ts` 末尾 4 个导出函数**

在 `http.interceptors.response.use(...)` 之后,`export function get` 之前,插入:

```ts
const MOCK = import.meta.env.VITE_DEMO_MOCK === 'true'

async function mockOr<T>(method: 'GET' | 'POST' | 'PATCH' | 'DELETE', url: string, params?: unknown): Promise<T> {
  const { matchMock } = await import('@proapi/shared/mock')
  const { matched, data } = await matchMock<T>(method, url, params)
  if (!matched) console.warn(`[mock] no data for: ${method} ${url}`)
  return (data ?? ([] as unknown)) as T
}
```

然后改写 4 个导出函数(注意 user 的 http.ts 没有 params 参数,只接 config):

```ts
export function get<T = unknown>(url: string, config?: AxiosRequestConfig) {
  if (MOCK) return mockOr<T>('GET', url, config?.params)
  return http.get<T>(url, config).then(r => r.data)
}
export function post<T = unknown>(url: string, data?: unknown, config?: AxiosRequestConfig) {
  if (MOCK) return mockOr<T>('POST', url, data)
  return http.post<T>(url, data, config).then(r => r.data)
}
export function patch<T = unknown>(url: string, data?: unknown, config?: AxiosRequestConfig) {
  if (MOCK) return mockOr<T>('PATCH', url, data)
  return http.patch<T>(url, data, config).then(r => r.data)
}
export function del<T = unknown>(url: string, config?: AxiosRequestConfig) {
  if (MOCK) return mockOr<T>('DELETE', url, config?.params)
  return http.delete<T>(url, config).then(r => r.data)
}
```

- [ ] **Step 2: 加 user 的 env.d.ts VITE_DEMO_MOCK 声明**

参照 admin Task 4 Step 2,在 `web/user/env.d.ts` 中加 `VITE_DEMO_MOCK` 类型声明。

- [ ] **Step 3: typecheck**

Run: `pnpm -C web/user typecheck`
Expected: 通过。

- [ ] **Step 4: 提交**

```bash
git add web/user/src/api/http.ts web/user/env.d.ts
git commit -m "feat(user): http.ts 加 VITE_DEMO_MOCK 分支"
```

---

## Task 8: user router base + demo bypass + build:demo

**Files:**
- Modify: `web/user/src/router/index.ts`
- Modify: `web/user/package.json`

`stores/auth.ts` 不需要改动:mock 模式下 `profileApi.get()` 已通过 `http.ts` 的 mock 分支返回 fake profile,调用现有 `refresh()` 即可。

- [ ] **Step 1: 修改 `web/user/src/router/index.ts`,history base 走 BASE_URL + demo 跳过 auth**

把文件末尾的 `createRouter` 和 `beforeEach` 块替换为:

```ts
const MOCK = import.meta.env.VITE_DEMO_MOCK === 'true'
const baseUrl = MOCK ? import.meta.env.BASE_URL : '/'

export const router = createRouter({ history: createWebHistory(baseUrl), routes })

router.beforeEach(async (to) => {
  const auth = useAuthStore()

  if (MOCK) {
    if (!auth.user) {
      try { await auth.refresh() } catch { /* mock 不会失败 */ }
    }
    if (['login', 'register', 'forgot'].includes(to.name as string)) {
      return { name: 'home' }
    }
    return
  }

  if (to.meta.auth) {
    if (!auth.user) {
      try { await auth.refresh() } catch { /* fall through */ }
    }
    if (!auth.user) {
      return { name: 'login', query: { redirect: to.fullPath } }
    }
  }
  if (!to.meta.auth && auth.user && ['login', 'register', 'forgot'].includes(to.name as string)) {
    return { name: 'home' }
  }
})
```

保留原 `router.afterEach`、`import` 块、`routes` 数组不变。

- [ ] **Step 2: 加 `build:demo` 脚本**

修改 `web/user/package.json` `scripts`(devDependencies 加 cross-env):

```json
{
  "scripts": {
    "dev": "vite --port 5174 --host 127.0.0.1",
    "dev:demo": "cross-env VITE_DEMO_MOCK=true vite --port 5174 --host 127.0.0.1",
    "build": "vue-tsc --noEmit && vite build",
    "build:demo": "cross-env VITE_DEMO_MOCK=true vue-tsc --noEmit && vite build --mode demo",
    "preview": "vite preview --port 5174",
    "typecheck": "vue-tsc --noEmit",
    "lint": "eslint . --ext .ts,.vue --max-warnings 0",
    "format": "prettier --write \"src/**/*.{ts,vue,css}\""
  }
}
```

`devDependencies` 追加 `"cross-env": "^7.0.3"`。

- [ ] **Step 3: 安装依赖**

Run: `pnpm install`

- [ ] **Step 4: typecheck + build:demo**

```bash
pnpm -C web/user typecheck
pnpm -C web/user run build:demo -- --base=/user-demo/
```

Expected: 两者均退出码 0,`web/user/dist/index.html` 中资源前缀为 `/user-demo/`。

- [ ] **Step 5: 提交**

```bash
git add web/user/src/router/index.ts web/user/package.json pnpm-lock.yaml
git commit -m "feat(user): demo 模式跳过 auth + router base 走 BASE_URL + build:demo 脚本"
```

---

## Task 9: docs-site 构建脚本

**Files:**
- Create: `docs-site/scripts/build-demos.js`
- Modify: `docs-site/package.json`

- [ ] **Step 1: 创建 `docs-site/scripts/build-demos.js`**

```js
#!/usr/bin/env node
import { execSync } from 'node:child_process'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)
const rootDir = path.resolve(__dirname, '..', '..')
const publicDir = path.resolve(__dirname, '..', 'public')

const targets = [
  { name: 'admin', srcDir: path.join(rootDir, 'web', 'admin') },
  { name: 'user',  srcDir: path.join(rootDir, 'web', 'user')  },
]

for (const t of targets) {
  console.log(`[demo] Building ${t.name}...`)
  execSync(`pnpm run build:demo -- --base=/${t.name}-demo/`, {
    cwd: t.srcDir,
    stdio: 'inherit',
  })

  const dist = path.join(t.srcDir, 'dist')
  if (!fs.existsSync(dist)) {
    throw new Error(`Build output not found: ${dist}`)
  }

  const out = path.join(publicDir, `${t.name}-demo`)
  fs.rmSync(out, { recursive: true, force: true })
  fs.mkdirSync(out, { recursive: true })
  copyDir(dist, out)
  console.log(`[demo] ${t.name} → ${out}`)
}

function copyDir(src, dest) {
  for (const entry of fs.readdirSync(src, { withFileTypes: true })) {
    const s = path.join(src, entry.name)
    const d = path.join(dest, entry.name)
    if (entry.isDirectory()) {
      fs.mkdirSync(d, { recursive: true })
      copyDir(s, d)
    } else {
      fs.copyFileSync(s, d)
    }
  }
}
```

注意 docs-site `package.json` 是 `"type": "module"`,故用 ESM 写法。

- [ ] **Step 2: 修改 `docs-site/package.json`**

```json
{
  "name": "@proapi/docs-site",
  "version": "0.0.0",
  "private": true,
  "type": "module",
  "scripts": {
    "dev": "vitepress dev --port 5175 --host 127.0.0.1",
    "build": "vitepress build",
    "preview": "vitepress preview --port 5175",
    "demo:build": "node scripts/build-demos.js",
    "docs:build-with-demo": "pnpm run demo:build && pnpm run build"
  },
  "devDependencies": {
    "vitepress": "^1.4.1",
    "vue": "^3.5.12"
  }
}
```

- [ ] **Step 3: 跑通 `demo:build`**

Run: `pnpm -C docs-site run demo:build`
Expected:
- 退出码 0
- `docs-site/public/admin-demo/index.html` 存在
- `docs-site/public/user-demo/index.html` 存在

- [ ] **Step 4: 提交**

```bash
git add docs-site/scripts/build-demos.js docs-site/package.json
git commit -m "feat(docs-site): demo:build 脚本(同时构建 admin+user 演示)"
```

---

## Task 10: VitePress nav 加"在线演示"下拉

**Files:**
- Modify: `docs-site/.vitepress/config.ts`

- [ ] **Step 1: zh locale nav 加一条**

在 `docs-site/.vitepress/config.ts` 的 zh locale 的 `nav` 数组末尾(`{ text: '更新日志', link: '/zh/changelog' }` 之后)添加:

```ts
{
  text: '在线演示',
  items: [
    { text: '管理后台演示', link: '/admin-demo/index.html', target: '_blank' },
    { text: '用户中心演示', link: '/user-demo/index.html',  target: '_blank' },
  ],
},
```

- [ ] **Step 2: en locale nav 加一条**

en locale 的 `nav` 末尾(`{ text: 'Changelog', link: '/en/changelog' }` 之后)添加:

```ts
{
  text: 'Live Demo',
  items: [
    { text: 'Admin Demo', link: '/admin-demo/index.html', target: '_blank' },
    { text: 'User Demo',  link: '/user-demo/index.html',  target: '_blank' },
  ],
},
```

- [ ] **Step 3: 跑 docs:build-with-demo 验证**

```bash
pnpm -C docs-site run docs:build-with-demo
```

Expected: 退出码 0。

- [ ] **Step 4: preview 手工验证**

Run: `pnpm -C docs-site run preview`
Expected: 浏览器打开 `http://127.0.0.1:5175/zh/` 后:
- 顶部 nav 出现"在线演示"下拉
- 点"管理后台演示" → 新窗口 → 直接落到 admin dashboard(无 login)
- 点"用户中心演示" → 新窗口 → 直接落到 user 首页(无 login)

- [ ] **Step 5: 提交**

```bash
git add docs-site/.vitepress/config.ts
git commit -m "feat(docs-site): nav 添加'在线演示'下拉(zh+en)"
```

---

## Task 11: 在 introduction / quickstart 顶部加 banner

**Files:**
- Modify: `docs-site/zh/guide/introduction.md`
- Modify: `docs-site/zh/guide/quickstart.md`
- Modify: `docs-site/en/guide/introduction.md`
- Modify: `docs-site/en/guide/quickstart.md`

- [ ] **Step 1: `docs-site/zh/guide/introduction.md`**

在文件顶部 frontmatter(如有 `---` 块)之后、首个 H1 之前或之后,加 1 行 banner:

```md
> 💡 **想直接看看?** [管理后台演示](/admin-demo/index.html){target="_blank"} · [用户中心演示](/user-demo/index.html){target="_blank"}(纯前端 mock,无需部署后端)
```

- [ ] **Step 2: `docs-site/zh/guide/quickstart.md`**

同 Step 1。

- [ ] **Step 3: `docs-site/en/guide/introduction.md`**

```md
> 💡 **Want a quick look?** [Admin Demo](/admin-demo/index.html){target="_blank"} · [User Demo](/user-demo/index.html){target="_blank"} (front-end mock, no backend required)
```

- [ ] **Step 4: `docs-site/en/guide/quickstart.md`**

同 Step 3。

- [ ] **Step 5: 验证 markdown 渲染**

```bash
pnpm -C docs-site run dev
```

打开 `http://127.0.0.1:5175/zh/guide/introduction`(和 quickstart / en 对应页),确认 banner 显示正常,点击链接能跳转。

注:VitePress 不一定支持 `{target="_blank"}` 语法。如果 dev 检查发现链接没有 `target="_blank"` 属性,改为原生 HTML 写法:

```md
> 💡 **想直接看看?** <a href="/admin-demo/index.html" target="_blank">管理后台演示</a> · <a href="/user-demo/index.html" target="_blank">用户中心演示</a>(纯前端 mock,无需部署后端)
```

- [ ] **Step 6: 提交**

```bash
git add docs-site/zh/guide/introduction.md docs-site/zh/guide/quickstart.md docs-site/en/guide/introduction.md docs-site/en/guide/quickstart.md
git commit -m "docs: introduction/quickstart 顶部加演示链接 banner"
```

---

## Task 12: .gitignore 加 demo 产物

**Files:**
- Modify: `.gitignore`

- [ ] **Step 1: 修改 `.gitignore`**

`.gitignore` 已有 `web/*/dist/`、`docs-site/.vitepress/dist/` 等,只需追加 demo public 产物。在 "Vite / 前端产物" 段落末尾追加:

```
docs-site/public/admin-demo/
docs-site/public/user-demo/
```

- [ ] **Step 2: 验证忽略生效**

```bash
git status --short docs-site/public/
```

Expected: 输出只包含已跟踪的 `logo.svg` 等原有文件,**不**列出 `admin-demo/`、`user-demo/`。

- [ ] **Step 3: 提交**

```bash
git add .gitignore
git commit -m "chore: gitignore docs-site/public/{admin,user}-demo/"
```

---

## Task 13: 最终回归验证

**Files:** 无新增/修改

- [ ] **Step 1: 真实模式不回归**

```bash
pnpm -C web/admin run typecheck
pnpm -C web/admin run build
pnpm -C web/user run typecheck
pnpm -C web/user run build
```

Expected: 全部退出码 0。

- [ ] **Step 2: 真实 dev 启动验证**

```bash
make docker-up
make dev
```

打开 `http://127.0.0.1:5173`(admin):应该弹 login(因为真实模式下未登录)。
打开 `http://127.0.0.1:5174`(user):应该弹 login。

Expected: 行为与本次改动前一致。

Stop with Ctrl-C 后进入 Step 3。

- [ ] **Step 3: demo 模式完整端到端**

```bash
pnpm -C docs-site run docs:build-with-demo
pnpm -C docs-site run preview
```

打开 `http://127.0.0.1:5175/zh/`,在 nav "在线演示"下:

**admin 演示验证清单**:
- [ ] 直接落到 dashboard,不弹 login
- [ ] 渠道列表看到 16 条数据(openai-main / anthropic-main / ... / huggingface-main)
- [ ] 模型列表、stats dashboard、令牌列表、ledger 都有数据
- [ ] 点"新建渠道"提交,toast 显示成功(写操作 no-op,刷新后渠道列表不变)
- [ ] 刷新页面不被弹回 login

**user 演示验证清单**:
- [ ] 直接落到首页,不弹 login
- [ ] Wallet 余额、token 列表(2 条)、usage 等显示 mock 数据
- [ ] 刷新页面不被弹回 login

**控制台**:
- [ ] 浏览器 DevTools console 中若有 `[mock] no data for: ...` 警告,记录漏写的 URL 模式;非阻塞,但若是高频路径(影响显示)就在 `routes.ts` 补一条 + 补 fixture,然后回到 Task 9。

- [ ] **Step 4: 最终汇总提交(若 Step 3 修补了 mock 路由)**

```bash
git add -A
git commit -m "fix(mock): 补漏 routes/fixtures(回归发现)"
```

如无修补,跳过。

---

## Summary

完成上述 13 个任务后:
- 仓库:`web/shared/src/mock/` 提供完整 mock 数据 + 路由层;两个前端的 `http.ts` 在 `VITE_DEMO_MOCK=true` 时短路走 mock;router/store 配套跳过 auth
- 构建:`docs-site/scripts/build-demos.js` 同时构建 admin / user,产物落到 `docs-site/public/{admin,user}-demo/`
- 文档:VitePress nav 双语下拉 + introduction/quickstart 顶部 banner,访客可在 docs 站直接打开演示
- 真实模式行为零回归:`pnpm -C web/admin/user dev/build` 路径不变,axios 实例与拦截器不变
