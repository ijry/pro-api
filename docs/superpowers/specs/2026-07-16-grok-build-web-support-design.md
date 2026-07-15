# Grok Build / Grok Web 支持设计稿

- **日期:** 2026-07-16
- **范围:** 在现有 `internal/adapter` 体系中新增 `grok-build` 与 `grok-web` 两个 provider,分别覆盖 xAI 官方 OpenAI-compatible 接口与 Grok Web SSO 反代接口。

## 背景

当前 `pro-api` 的上游模型接入已经形成稳定分层:

- `adapter` 负责把 IR 请求翻译成厂商 HTTP 协议
- `relay` 负责按 channel 选择 provider 并做计费、日志、限流编排
- `account` 负责账号池、探测、刷新与失败标记
- `model_catalogs` 负责把前端可选模型和 provider 能力绑定起来

仓库里已经有 OpenAI 兼容适配器、Groq 适配器和一批 provider 注册样板,因此新增 Grok 更适合继续沿用“一个 provider 一个边界”的方式,而不是在现有 OpenAI adapter 上塞条件分支。

本次按已确认方案落地两条线:

- `grok-build`: 官方 API,走 API Key
- `grok-web`: Grok Web,走 SSO Cookie 反代

## 目标

1. `grok-build` 能通过现有 OpenAI-compatible 入口完成 chat / stream 调用。
2. `grok-web` 能通过现有 OpenAI-compatible 入口完成 chat / stream 调用,并把 Grok Web 的逐行 JSON 流稳定转换成 IR。
3. 两个 provider 都能在 `adapterreg` 注册,可被 channel/provider 选择器识别。
4. `model_catalogs` 增加 Grok 对应模型,让前端和 channel 配置可以直接选择。
5. 错误分类要和现有 relay / breaker 逻辑对齐,避免把临时网络错误误标成账号失效。
6. 账号池探测支持 `grok-build` 与 `grok-web`,用于手动探测和后台可用性检查。

## 非目标

- 不做 Grok 图像生成、图像编辑、视频生成、语音或文件上传链路。
- 不改现有 OpenAI / Anthropic / Gemini 等 provider 的行为。
- 不重做账号池 UI 或 channel UI。
- 不引入新的全局运行时配置项,除非实现时发现现有 `channel.BaseURL` 无法覆盖特殊场景。

## 架构

### 1. `grok-build`

`grok-build` 是一个薄封装 provider,实现方式与 `groq` 类似:

- 复用 `internal/adapter/openai` 的请求构造、响应解码和 SSE 读取逻辑
- 默认 base URL 为 `https://api.x.ai`
- `Adapter.Name()` 返回 `grok-build`
- 认证使用 `adapter.Credential.APIKey`
- `channel.Credential.BaseURL` 仍可覆盖默认 base URL,便于测试或自定义代理
- `Capabilities()` 只暴露 chat / stream,避免因为复用 OpenAI helper 而错误暴露 embedding / image / audio 能力

这个 provider 只负责“把 OpenAI-compatible 请求发给 xAI”,不引入 Grok Web 的 Cookie、headers 或 web-specific payload。

### 2. `grok-web`

`grok-web` 需要单独实现,因为它不是标准 OpenAI 协议:

- 默认 base URL 为 `https://grok.com`
- 请求路径固定为 `POST /rest/app-chat/conversations/new`
- 认证使用 `sso` / `sso-rw` Cookie
- 需要按 Grok Web 的 payload 结构发送 `message`、`modelName`、`modelMode`、`fileAttachments`、`toolOverrides`、`responseMetadata`
- 响应是逐行 JSON,不是标准 SSE chunk

实现上保留以下约束:

- `channel.Credential.APIKey` 存放裸 SSO token
- adapter 发请求时生成 `Cookie: sso=<token>; sso-rw=<token>`
- 通过 `Origin`、`Referer`、`User-Agent`、`x-statsig-id`、`x-xai-request-id` 维持 Grok Web 所需的基本浏览器上下文
- 只做 chat / stream,不把 Grok Web 的图片、视频、资产接口带进本次范围
- `Embed` 等非 chat 能力返回明确的 unsupported error,不静默降级到其他 provider 行为

### 3. 注册与路由

`internal/adapterreg/WireAdapters` 增加两个注册项:

- `grok-build`
- `grok-web`

`relay` 不需要新增路由,仍然沿用现有 OpenAI-compatible 入口:

- `POST /v1/chat/completions`
- `POST /v1/completions`
- `POST /v1/embeddings`

Grok 只是在 provider 层被选中,不新增单独的 `/v1/grok/...` 路径。

## 数据流

### `grok-build`

`OpenAI JSON -> IR -> relay -> grok-build adapter -> xAI API -> OpenAI-compatible JSON`

这个路径基本沿用现有 `openai` adapter,主要差异只有 provider 名和 base URL。

### `grok-web`

`OpenAI JSON -> IR -> relay -> grok-web adapter -> Grok Web REST -> 逐行 JSON -> IR -> OpenAI SSE`

关键转换点:

1. `ir.ChatRequest.Messages` 需要在 adapter 内合并成 Grok Web 需要的单条 `message`
2. `model` 需要映射到 Grok Web 的 `modelName` / `modelMode`
3. 流式响应要把 `result.response.token` 转成 `ir.ChatChunk`
4. 最终 `modelResponse.message` 或流末尾内容要拼成完整 `ir.ChatResponse`

## 模型与目录

### `grok-build`

初始模型目录只放 chat 系模型,以 xAI 官方 API 当前可用为准:

- `grok-4`
- `grok-3`
- `grok-3-mini`
- `grok-3-mini-fast`

### `grok-web`

按 `grok2api` 的 Grok Web 模型目录落 seed:

- `grok-3`
- `grok-3-mini`
- `grok-3-thinking`
- `grok-4`
- `grok-4-mini`
- `grok-4-thinking`
- `grok-4-heavy`
- `grok-4.1-mini`
- `grok-4.1-fast`
- `grok-4.1-expert`
- `grok-4.1-thinking`

adapter 内部维护模型映射表:

| Client model | Grok `modelName` | Grok `modelMode` |
|---|---|---|
| `grok-3` | `grok-3` | `MODEL_MODE_GROK_3` |
| `grok-3-mini` | `grok-3` | `MODEL_MODE_GROK_3_MINI_THINKING` |
| `grok-3-thinking` | `grok-3` | `MODEL_MODE_GROK_3_THINKING` |
| `grok-4` | `grok-4` | `MODEL_MODE_GROK_4` |
| `grok-4-mini` | `grok-4-mini` | `MODEL_MODE_GROK_4_MINI_THINKING` |
| `grok-4-thinking` | `grok-4` | `MODEL_MODE_GROK_4_THINKING` |
| `grok-4-heavy` | `grok-4` | `MODEL_MODE_HEAVY` |
| `grok-4.1-mini` | `grok-4-1-thinking-1129` | `MODEL_MODE_GROK_4_1_MINI_THINKING` |
| `grok-4.1-fast` | `grok-4-1-thinking-1129` | `MODEL_MODE_FAST` |
| `grok-4.1-expert` | `grok-4-1-thinking-1129` | `MODEL_MODE_EXPERT` |
| `grok-4.1-thinking` | `grok-4-1-thinking-1129` | `MODEL_MODE_GROK_4_1_THINKING` |

### `model_catalogs` 更新

`migrations/*/000023_seed_model_catalogs.up.sql` 追加 Grok 相关行,`owned_by` 分别使用:

- `grok-build`
- `grok-web`

capabilities 先只写 chat / stream,避免把未验证的接口能力提前暴露给前端和路由。

## 错误处理

### `grok-build`

沿用现有 OpenAI-compatible 错误分类:

- `401/403` -> 凭证错误
- `429` -> 限流
- `5xx` / 网络失败 -> 上游错误或超时

### `grok-web`

错误分类要偏保守:

- `401/403` 视为 SSO token 失效或权限不足
- `429` 视为限流或风控
- `5xx` / 连接失败 / 读取失败视为暂时性上游故障
- JSON 行解析失败直接跳过,不要因此中断整条流

响应里若出现 `xai:tool_usage_card` 或其他 Grok 特殊标签,已知结构化标签做过滤/降噪;未知标签按普通文本保留,避免解析器因上游新增标签而中断。

## 账号探测

`grok-build` 新增轻量 probe:

- 默认 base URL 为 `https://api.x.ai`
- 请求 `GET /v1/models`
- 认证头为 `Authorization: Bearer <api_key>`

`grok-web` 新增轻量 probe:

- 默认 base URL 为 `https://grok.com`
- 请求 `GET /rest/rate-limits`
- 认证头为 `Cookie: sso=<token>; sso-rw=<token>`

本次只判断探测请求是否成功,不解析配额。401 / 403 / 429 的状态分类继续交给现有 breaker 逻辑处理。

## 测试

### 单测

- `grok-build`:
  - adapter 名称和 base URL
  - OpenAI 请求复用是否正确
  - 默认模型列表是否注册
- `grok-web`:
  - Cookie 生成
  - payload 组装
  - 非流式响应收集
  - 流式逐行 token 解析
  - 401 / 429 / 5xx 分类
  - 空行、坏行、缺字段时的容错
- `adapterreg`:
  - 注册后可按 name 取到两个 provider
- `account/probe`:
  - `grok-build` probe 使用 Bearer token 请求 `/v1/models`
  - `grok-web` probe 使用 SSO Cookie 请求 `/rest/rate-limits`

### 结构测试

- `go test ./internal/adapter/...`
- `go test ./internal/server/...`
- `go test ./internal/account/...`
- `go test ./...` 的整体编译通过

## 文件清单

| 类型 | 路径 | 内容 |
|---|---|---|
| 新增 | `internal/adapter/grokbuild/*` | xAI 官方 OpenAI-compatible provider |
| 新增 | `internal/adapter/grokweb/*` | Grok Web 反代 provider |
| 新增 | `internal/account/probe/grok_build.go` | `grok-build` 账号可用性探测 |
| 新增 | `internal/account/probe/grok_web.go` | `grok-web` 账号可用性探测 |
| 修改 | `internal/account/wire/wire.go` | 注册 Grok probe |
| 修改 | `internal/adapterreg/wire.go` | 注册 `grok-build` / `grok-web` |
| 修改 | `migrations/mysql/000023_seed_model_catalogs.up.sql` | 追加 Grok 模型 |
| 修改 | `migrations/mysql/000023_seed_model_catalogs.down.sql` | 删除 Grok seed 模型 |
| 修改 | `migrations/postgres/000023_seed_model_catalogs.up.sql` | 追加 Grok 模型 |
| 修改 | `migrations/postgres/000023_seed_model_catalogs.down.sql` | 删除 Grok seed 模型 |
| 修改 | `docs-site/zh/architecture/adapter-layer.md` | 更新支持列表 |
| 修改 | `docs-site/en/architecture/adapter-layer.md` | 同步英文说明 |
| 修改 | `README.md` / `README_zh.md` | 补充 Grok 支持说明 |

## 验收标准

1. `grok-build` 和 `grok-web` 都能在 `adapterreg` 中被正确注册和取用。
2. `grok-build` 能通过现有 OpenAI 入口完成 chat / stream。
3. `grok-web` 能通过现有 OpenAI 入口完成 chat / stream,并稳定处理 Grok Web 的逐行 JSON。
4. Grok 相关模型能出现在 `model_catalogs` 中并在前后端可选。
5. `grok-build` 和 `grok-web` 账号能通过 probe 做轻量可用性检查。
6. 新增测试通过,且不影响现有 provider 的测试结果。
7. `go test ./...` 通过,并且没有引入额外运行时配置依赖。
