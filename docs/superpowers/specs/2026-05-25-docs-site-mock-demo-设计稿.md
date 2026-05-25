# docs-site Mock 演示模式 — 设计稿

- 日期:2026-05-25
- 参考:`/Users/admin/Documents/Repos/xyito/open/lyshop/docs-site/`(已落地的同模式实现)
- 状态:已与用户对齐,等待实现计划

## 1. 背景与目标

proapi 是 LLM API 网关,仓库中已有:

- `docs-site/`:VitePress 双语文档站(zh / en),sidebar 完整。
- `web/admin/`(port 5173, Vue 3 + Naive UI):管理后台,登录后展示渠道、令牌、计费、ledger 等。
- `web/user/`(port 5174, Vue 3):用户中心,展示个人 wallet / token / usage。
- 后端 Go 网关,16 家 adapter 已接入,**无 mock adapter**。

访客阅读文档时,无法直接看到 admin / user 长什么样;部署本地需要 Go + DB + Redis,门槛高。

**目标**:复制 lyshop docs-site 的"前端打包 + mock 数据"演示模式,让访客在文档站直接打开 admin / user 的可交互演示,**不需要部署后端**。

## 2. 核心决策

| 决策 | 选择 |
|---|---|
| 演示哪些前端入口 | **admin + user 两个都演示** |
| 是否做悬浮窗(iframe) | 否(admin / user 都是 PC 应用,新窗口打开更合适) |
| Mock 数据放哪里 | `web/shared/src/mock/`(admin / user 共用) |
| Login 怎么处理 | demo 模式直接 bypass,落到 dashboard,不弹登录页 |
| 后端是否参与 | 否,纯前端 mock,后端零改动 |

## 3. 仓库布局

```
proapi/
├── web/
│   ├── shared/src/
│   │   └── mock/                       ← 新建
│   │       ├── index.ts                ← 导出 matchMock(method, url, params)
│   │       ├── routes.ts               ← URL pattern → handler 路由表
│   │       ├── helpers.ts              ← 分页 / 筛选 / clone 工具
│   │       └── data/
│   │           ├── user-profile.json
│   │           ├── admin-user.json
│   │           ├── channels.json       ← 16 个真实 adapter 名
│   │           ├── tokens.json
│   │           ├── models.json
│   │           ├── stats-dashboard.json
│   │           ├── ledger.json
│   │           ├── usage.json
│   │           ├── notices.json
│   │           └── logs.json
│   ├── admin/
│   │   ├── package.json                ← 新增 build:demo 脚本
│   │   └── src/
│   │       ├── api/http.ts             ← 修改:VITE_DEMO_MOCK 分支
│   │       ├── router/index.ts         ← 修改:demo 模式跳过 auth guard
│   │       └── stores/user.ts          ← 修改:demo 模式注入 fake session
│   └── user/                           ← 同 admin
├── docs-site/
│   ├── package.json                    ← 新增 demo:build / docs:build-with-demo
│   ├── scripts/
│   │   └── build-demos.js              ← 新建
│   ├── .vitepress/config.ts            ← 修改:nav 加 "在线演示" 下拉
│   ├── zh/guide/introduction.md        ← 修改:顶部加演示链接提示
│   ├── zh/guide/quickstart.md          ← 同上
│   ├── en/guide/introduction.md        ← 同上
│   ├── en/guide/quickstart.md          ← 同上
│   └── public/
│       ├── admin-demo/                 ← build 产物 (gitignore)
│       └── user-demo/                  ← build 产物 (gitignore)
└── .gitignore                          ← 新增 demo build 产物条目
```

## 4. 数据流

```
浏览器 → admin/user SPA
       → src/api/<biz>.ts(业务调用)
       → src/api/http.ts get/post/...
            ├─ VITE_DEMO_MOCK=true → @proapi/shared/mock matchMock() → JSON
            └─ 默认 → axios → /api/...
```

mock 命中行为:
- `100~300ms` 随机延迟(模拟真实网络)
- 返回值与 axios 同形:业务层 `.then(r => r.data)` 不需改
- 路径未命中:`console.warn('[mock] no data for: GET /xxx')` + 兜底返回空数组 / 空对象

mock 不命中**不抛错**,只是空数据,确保 demo 不会因为我们漏写一条 mock 数据就白屏。

## 5. 关键组件设计

### 5.1 Mock 层(`web/shared/src/mock/`)

**`index.ts`** 暴露唯一入口:

```ts
export async function matchMock<T>(
  method: 'GET' | 'POST' | 'PATCH' | 'DELETE',
  url: string,
  params?: unknown,
): Promise<{ matched: boolean; data: T | null }> {
  const delay = 100 + Math.random() * 200
  await new Promise(r => setTimeout(r, delay))
  return routeMock(method, url, params)
}
```

**`routes.ts`** 是一张按 URL 前缀匹配的路由表:

```ts
type Handler = (method: string, url: string, params?: any) => unknown

const routes: Array<{ pattern: RegExp; handler: Handler }> = [
  // 公共
  { pattern: /^\/api\/auth\/me$/,        handler: () => userProfile },
  { pattern: /^\/api\/notices/,          handler: () => notices },

  // admin
  { pattern: /^\/api\/admin\/stats/,     handler: () => statsDashboard },
  { pattern: /^\/api\/admin\/channels/,  handler: (m, u, p) => paginate(channels, p) },
  { pattern: /^\/api\/admin\/tokens/,    handler: (m, u, p) => paginate(tokens, p) },
  { pattern: /^\/api\/admin\/models/,    handler: () => models },
  { pattern: /^\/api\/admin\/ledger/,    handler: (m, u, p) => paginate(ledger, p) },
  { pattern: /^\/api\/admin\/logs/,      handler: (m, u, p) => paginate(logs, p) },

  // user
  { pattern: /^\/api\/user\/wallet/,     handler: () => userWallet },
  { pattern: /^\/api\/user\/tokens/,     handler: (m, u, p) => paginate(userTokens, p) },
  { pattern: /^\/api\/user\/usage/,      handler: () => usage },
  { pattern: /^\/api\/user\/recharge/,   handler: (m, u, p) => paginate(rechargeList, p) },
]
```

**写操作**(POST/PATCH/DELETE):统一返回 `{ ok: true }` + console.warn `[mock] write ignored`,不改 mock store。刷新后所有"创建"恢复原状,符合静态 demo 预期。

**Helpers**(`helpers.ts`):
- `paginate(items, params)`:根据 `?page=&size=` 切片,返回 `{ items, total, page, size }`(与 `http.ts` 中 `Page<T>` 类型一致)
- `clone()`:深拷贝,防 mock data 被业务层修改污染

### 5.2 `http.ts` 改造(admin / user 同形)

底层 `get/post/patch/del` 加 mock 分支,axios 实例和拦截器**完全保留**:

```ts
const MOCK = import.meta.env.VITE_DEMO_MOCK === 'true'

async function mockOr<T>(method: HttpMethod, url: string, params?: unknown): Promise<T> {
  const { matchMock } = await import('@proapi/shared/mock')
  const { matched, data } = await matchMock<T>(method, url, params)
  if (!matched) console.warn(`[mock] no data for: ${method} ${url}`)
  return (data ?? ([] as unknown)) as T   // 数组业务接列表,对象业务接详情;空数组兼容多数路径
}

export async function get<T>(url: string, params?: Record<string, unknown>, cfg?: AxiosRequestConfig) {
  if (MOCK) return mockOr<T>('GET', url, params)
  return http.get<T>(url, { params, ...cfg }).then(r => r.data)
}
// post / patch / del 同形(写操作 mockOr 返回 { ok: true })
```

非 MOCK 模式下完全等价于现有实现,**零回归风险**。

### 5.3 Login bypass

**`src/stores/user.ts`** 在 store hydrate 时:

```ts
const MOCK = import.meta.env.VITE_DEMO_MOCK === 'true'

if (MOCK) {
  state.user = mockUserProfile        // 来自 @proapi/shared/mock/data
  state.loggedIn = true
  state.csrfToken = 'demo-csrf'       // 满足 http.ts unsafe method 头校验
}
```

**`src/router/index.ts`** beforeEach 短路:

```ts
router.beforeEach((to, from, next) => {
  if (MOCK) {
    if (to.name === 'login') return next({ name: 'dashboard' })  // login 强制跳走
    return next()                                                 // 其他页一律放行
  }
  // 真实模式原有逻辑保留(查 token / 跳 login 等)
  ...
})
```

效果:
- 直接打开 `/admin-demo/` → 跳过 login → 落到 dashboard
- 刷新任意页面 → 不会被弹回 login
- 真实 `pnpm run dev` 行为完全不变

### 5.4 admin / user 的 build:demo 脚本

`web/admin/package.json`:

```json
{
  "scripts": {
    "build:demo": "cross-env VITE_DEMO_MOCK=true vue-tsc --noEmit && vite build --mode demo"
  },
  "devDependencies": {
    "cross-env": "^7.0.3"
  }
}
```

`web/user/package.json` 同形。

- `cross-env`:让脚本在 Windows / mac / linux 一致传环境变量(lyshop 已是这种用法)
- `--mode demo`:Vite 加载 `.env.demo`(后续如要加 `VITE_DEMO_BANNER=true` 等可走这里)
- `--base` 不在脚本里写死,由 `docs-site/scripts/build-demos.js` 命令行参数传入,避免演示路径耦合

### 5.5 docs-site 构建脚本

**`docs-site/scripts/build-demos.js`**:

```js
#!/usr/bin/env node
const { execSync } = require('child_process')
const fs = require('fs')
const path = require('path')

const rootDir = path.resolve(__dirname, '..', '..')

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
  const out  = path.join(__dirname, '..', 'public', `${t.name}-demo`)
  fs.rmSync(out, { recursive: true, force: true })
  fs.mkdirSync(out, { recursive: true })
  copyDir(dist, out)
  console.log(`[demo] ${t.name} → ${out}`)
}

function copyDir(src, dest) {
  for (const e of fs.readdirSync(src, { withFileTypes: true })) {
    const s = path.join(src, e.name), d = path.join(dest, e.name)
    if (e.isDirectory()) { fs.mkdirSync(d, { recursive: true }); copyDir(s, d) }
    else fs.copyFileSync(s, d)
  }
}
```

`--base=/admin-demo/` 让 build 产物里的资源路径形如 `/admin-demo/assets/xxx.js`,无论 docs 部署到 `docs.foo.com/` 还是 `foo.com/proapi/` 都正确(VitePress 自身也是同样机制)。

**`docs-site/package.json` 新增**:

```json
{
  "scripts": {
    "demo:build": "node scripts/build-demos.js",
    "docs:build-with-demo": "pnpm run demo:build && pnpm run build"
  }
}
```

本地写文档继续用 `pnpm run dev` / `pnpm run build`,不强制构建 demo(慢)。CI / 发布走 `docs:build-with-demo`。

### 5.6 VitePress nav 配置

`docs-site/.vitepress/config.ts` 中,两个 locale 各加一条 nav:

```ts
// zh
nav: [
  ...原有项,
  {
    text: '在线演示',
    items: [
      { text: '管理后台演示', link: '/admin-demo/index.html', target: '_blank' },
      { text: '用户中心演示', link: '/user-demo/index.html',  target: '_blank' },
    ],
  },
]

// en 同形,文案 "Live Demo" / "Admin Demo" / "User Demo"
```

并在 `zh/guide/introduction.md`、`zh/guide/quickstart.md`(及对应英文)顶部加 1 行 banner:

```md
> 💡 **想直接看看?** [管理后台演示](/admin-demo/index.html) · [用户中心演示](/user-demo/index.html)(纯前端 mock,无需部署后端)
```

### 5.7 .gitignore 增量

```
docs-site/public/admin-demo/
docs-site/public/user-demo/
docs-site/.vitepress/dist/
docs-site/.vitepress/cache/
web/admin/dist/
web/user/dist/
```

(如已存在的条目跳过)

## 6. 范围

### 6.1 范围内(M1 一次落地)

| # | 项 | 文件估算 |
|---|---|---|
| 1 | `web/shared/src/mock/`(routes + 10 个 JSON + helpers + index) | ~14 |
| 2 | `web/admin/src/api/http.ts` + router/store mock 分支 | 3 |
| 3 | `web/user/src/api/http.ts` + router/store mock 分支 | 3 |
| 4 | `web/{admin,user}/package.json` 加 `build:demo` + cross-env 依赖 | 2 |
| 5 | `docs-site/scripts/build-demos.js` | 1 |
| 6 | `docs-site/package.json` 加脚本 | 1 |
| 7 | `docs-site/.vitepress/config.ts` 加 nav | 1 |
| 8 | `docs-site/{zh,en}/guide/{introduction,quickstart}.md` 加演示提示 | 4 |
| 9 | `.gitignore` 增量 | 1 |

### 6.2 范围外(YAGNI,不做)

- ❌ 后端 mock adapter / 后端"演示模式" —— 演示完全前端,后端不动
- ❌ Playwright 截图 showcase / 首页 hero 改造 —— 待营销首页阶段再加
- ❌ 手机壳浮窗 iframe —— PC 应用不合适
- ❌ Mock 写操作持久化(localStorage 模拟"真改") —— 复杂度高,M1 写操作 no-op
- ❌ E2E 自动化 —— 手工跑 `pnpm run dev:demo` 验证即可

## 7. 验收标准

1. `cd docs-site && pnpm run docs:build-with-demo` 成功,产出包含 `public/admin-demo/` 和 `public/user-demo/`
2. `pnpm run preview` 后,docs 首页可见"在线演示"下拉,两个链接均可打开并**直接落到 dashboard,不弹 login**
3. **admin demo**:
   - 渠道列表展示 16 条真实 adapter 名
   - 模型列表、stats dashboard、令牌列表、ledger 都有 mock 数据
   - 点"新建渠道"提交后 toast 成功,但刷新后渠道列表不变(写操作 no-op)
4. **user demo**:
   - Wallet 余额、token 列表(2 条)、usage 图表均显示 mock 数据
5. **真实模式不回归**:`pnpm run dev`(admin / user)行为完全不变,所有真实 axios 请求路径不受影响
6. `web/admin && pnpm run typecheck` / `web/user && pnpm run typecheck` 通过

## 8. 不确定与后续

- **shared 包导入路径**:`@proapi/shared/mock` 是否需要在 shared `package.json` 的 `exports` 里加一个子路径导出?实现时验证。若不便,退化为 `@proapi/shared` 单一导出 + 命名空间。
- **路由名**:`dashboard` 是占位,实现时按 admin / user 实际首页路由名确定。
- **写操作 toast 文案**:`[mock] write ignored` 是 console 日志,不影响用户;若 demo 要更"假装真实",可后续加 toast"演示模式下不会真正保存"。M1 不做。
