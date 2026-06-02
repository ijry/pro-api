# 计费集成 + 分组倍率 + 限流 设计稿

- **日期:** 2026-06-02
- **范围:** 连接 pricing / billing / ratelimit 三个已实现的模块,打通 relay handler 请求流的完整计费链路;同时支持 new-api 风格的分组倍率和 token 分组绑定。

## 背景

目前 `pricing`(倍率计算)、`billing`(预扣/实扣)、`ratelimit`(按组限流)三个模块的代码都已经实现,但**没有任何一个被接进请求流**:

- relay handler 调用 `relay.Chat()` 后直接返回响应,不做任何记账
- `application.PricingSvc` 始终 nil
- `CtxKeyGroupRatio` 始终无人写入
- `ratelimit.Middleware` 未挂到任何路由上

本设计把这三条传动轴全部接上,同时补齐 token-level 分组绑定和 channel-level 分组过滤。

## 数据流(改造后)

```
请求 → TokenAuth → GroupRatio 中间件 → ratelimit 中间件 → Relay Handler
                                                              │
                                                 pricing.EstimateMax
                                                      ↓
                                                 biller.Reserve(预扣)
                                                      ↓
                                                 relay.Chat()
                                                      ↓
                                              pricing.Compute(实际)
                                                      ↓
                                                biller.Commit(实扣)
                                                      ↓
                                                log.Event(写入)
                                                      ↓
                                                响应回客户端
```

## 文件清单

| Op | 路径 | 内容 |
|----|------|------|
| 修改 | `internal/token/model.go` | Token 加 GroupID 字段 |
| 修改 | `internal/channel/channel.go` | Channel 加 GroupID 字段 |
| 修改 | `internal/channel/selector.go` | Select 按 SelectHint.GroupID 过滤 |
| 新增 | `internal/server/middleware/groupratio.go` | 读 groupID → 查倍率 → 写入 context |
| 修改 | `internal/ratelimit/middleware.go` | 读取 CtxKeyGroupRatio |
| 新建 | `internal/server/handler/relay/deps.go` | Relay Handler 扩展依赖(可选) |
| 修改 | `internal/server/handler/relay/handler.go` | Chat/streamChat 后加记账 |
| 修改 | `internal/server/middleware/ratelimit.go` | 确认 CtxKey/PlanInput 对齐 |
| 修改 | `cmd/proapi/main.go` | wire pricing/group/biller + 挂中间件 |
| 新增 | `migrations/mysql/000026_billing_integration.up.sql` | Token 加 group_id,Channel 加 group_id |
| 新增 | `migrations/postgres/000026_billing_integration.up.sql` | 同上 |

## 各组件细节

### 1. Token 加 GroupID

`api_tokens` 表新增一列:

```sql
ALTER TABLE api_tokens ADD COLUMN group_id BIGINT NOT NULL DEFAULT 0;
```

`internal/token/model.go` Token struct 加:
```go
GroupID int64 `gorm:"column:group_id;default:0"`
```

**解析逻辑(TokenAuth 或新中间件):** 如果 token.group_id > 0,优先使用 token.group_id;否则 fallback 到 token 的 owner(user) 的 group_id。GroupRatio 中间件从 `CtxKeyGroupID` 读值后查倍率。

### 2. Channel 加 GroupID + 按组过滤

`channels` 表新增一列:

```sql
ALTER TABLE channels ADD COLUMN group_id BIGINT NOT NULL DEFAULT 0;
```

`internal/channel/channel.go` Channel struct 加:
```go
GroupID int64 `gorm:"column:group_id;default:0"`
```

`Selector.Select(ctx, hint)` 过滤逻辑:
- 若 `hint.GroupID > 0`:只返回 `channel.group_id == 0 || channel.group_id == hint.GroupID` 的渠道
- 若 `hint.GroupID == 0`:返回所有(现有行为不变)

Hint 的 GroupID 由 relay handler 在调用 selector 前从上下文中读取 `CtxKeyGroupID` 填入。

### 3. GroupRatio 中间件

新建 `internal/server/middleware/groupratio.go`:

```go
type GroupRatio func(ctx context.Context, groupID int64) float64

func GroupRatioMiddleware(lookup GroupRatio) gin.HandlerFunc {
    return func(c *gin.Context) {
        gid := token.GroupIDFromContext(c)
        if gid > 0 && lookup != nil {
            ratio := lookup(c.Request.Context(), gid)
            c.Set(ratelimit.CtxKeyGroupRatio, ratio)
            c.Set(CtxKeyBillingGroupRatio, ratio)
        }
        c.Next()
    }
}
```

`CtxKeyBillingGroupRatio` 定义在同一个文件中,用于 pricing.EstimateInput.BillingGroupRatio。

### 4. 挂载 ratelimit + GroupRatio 中间件

在 `cmd/proapi/main.go` 的 relay 路由注册处(目前只有 TokenAuth):

```go
v1 := eng.Group("/v1",
    middleware.ErrorResponse("openai"),
    middleware.TokenAuth(tokenStore, userLookup),   // 已有
    middleware.GroupRatioMiddleware(groupRatioLookup), // 新加
    ratelimit.Middleware(limiter, planner, sett, log), // 新加
)
```

### 5. Relay Handler 改用 Deps 依赖模式

Handler 改造成 Deps 注入:

```go
// Deps holds relay handler dependencies (all optional; nil = skip billing/logging).
type Deps struct {
    Relay   *relay.Service
    Pricing pricing.Pricing
    Biller  billing.Biller
    LogSvc  log.Store
}

type Handler struct{ deps Deps }

func New(deps Deps) *Handler
```

`Chat()` 方法改造:

```
handler.Chat(c):
  1. 解析请求,获取 model,userID,tokenID
  2. 若 h.Pricing 和 h.Biller 都非 nil:
     a. 从 context 读 BillingGroupRatio
     b. pricing.EstimateMax() → 预估值
     c. biller.Reserve(ctx, userID, tokenID, estCost)
  3. relay.Chat()
  4. 若 h.Pricing 非 nil:
     a. pricing.Compute(ctx, {Model, GroupID, InputTokens, OutputTokens, ...})
  5. 若 h.Biller 非 nil:
     a. biller.Commit(ctx, reservationID, actual.Quota)
  6. 若 h.LogSvc 非 nil:
     a. 构建 log.Event(含 GroupID, BillingGroupRatio, 用量等)
     b. logSvc.Create()
  7. c.JSON(...)
```

**流式路径(streamChat):** 流式比非流式复杂,因为 token 是在 SSE 发送过程中逐步到达的。处理方式:
- 在 Chat 开始前 Reserve
- 在读取 SSE 块的循环中累积 input_tokens + output_tokens
- 循环结束后 Compute + Commit
- 若客户端中途断开(EOF),Refund 剩余部分

### 6. 请求日志(Event)

`internal/log/event.go` 已有字段:
```go
GroupID          *int64
BillingGroupRatio float64
```

修改 relay handler 在 Chat 返回后填充这些字段并写入。

### 7. Migration

文件 `000026_billing_integration`:

```sql
-- MySQL
ALTER TABLE api_tokens ADD COLUMN group_id BIGINT NOT NULL DEFAULT 0;
ALTER TABLE channels ADD COLUMN group_id BIGINT NOT NULL DEFAULT 0;
```

## 不做项

- Account 池按组分配(号池与分组无直接关系)
- Admin 前端分组 CRUD 页面(已有分组 handler 见 admin handler)
- 按组倍率反向计算(折扣场景等特例)
- 全量集成测试(只做单元测试 + 编译验证)

## 验收标准

1. `go build ./...` 通过
2. 全部已有测试通过
3. TokenAuth 后 groupID = token.group_id || user.group_id
4. GroupRatio 中间件写入 CtxKeyGroupRatio 和 CtxKeyBillingGroupRatio
5. relay handler Chat() 后调用 pricing.Compute → biller.Commit
6. ratelimit 中间件收到正确的 PlanInput.GroupRatio
7. 流式路径结束时正确计算并 Commit
