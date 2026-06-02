# 计费集成 + 分组倍率 + 限流接线 实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把已有的 pricing / billing / ratelimit 三个模块接进 relay 请求链路,并支持 token 级和渠道级分组倍率过滤。

**Architecture:** 通过新的 GroupRatio 中间件将 `proapi:group_ratio` 注入 gin context;在 `/v1/` 路由链上依次挂 GroupRatio 中间件 → ratelimit 中间件;将 relay.Handler 改为 Deps 依赖模式,在 ChatCompletions 调用 relay.Chat 前后执行 pricing.EstimateMax → biller.Reserve → relay.Chat → pricing.Compute → biller.Commit → log.Write。

**Tech Stack:** Go 1.25 · Gin · GORM · gorm/driver/sqlite(测试) · `github.com/ijry/pro-api`

---

## 文件总览

| Op | 路径 | 职责 |
|----|------|------|
| 新增 | `migrations/mysql/000026_billing_integration.up.sql` | Token + Channel 加 group_id 列 |
| 新增 | `migrations/mysql/000026_billing_integration.down.sql` | 回滚 |
| 新增 | `migrations/postgres/000026_billing_integration.up.sql` | 同上 |
| 新增 | `migrations/postgres/000026_billing_integration.down.sql` | 同上 |
| 修改 | `internal/token/model.go` | Token + View 加 GroupID |
| 修改 | `internal/channel/channel.go` | Channel 加 GroupID |
| 修改 | `internal/channel/selector.go` | Select 按 hint.GroupID 过滤 |
| 新增 | `internal/server/middleware/groupratio.go` | GroupRatio 中间件 + CtxKeyBillingGroupRatio |
| 新增 | `internal/pricing/wire.go` | WirePricing(a) |
| 修改 | `internal/server/handler/relay/handler.go` | New(deps Deps)，ChatCompletions + streamChat 加 billing |
| 修改 | `cmd/proapi/main.go` | WirePricing 调用 + GroupRatio/ratelimit 中间件挂载 |

---

## Task 1: Migration 000026 — Token + Channel 加 group_id

**Files:**
- 新增: `migrations/mysql/000026_billing_integration.up.sql`
- 新增: `migrations/mysql/000026_billing_integration.down.sql`
- 新增: `migrations/postgres/000026_billing_integration.up.sql`
- 新增: `migrations/postgres/000026_billing_integration.down.sql`

- [ ] **Step 1: 检查当前最高迁移编号**

```bash
ls migrations/mysql/ | sort | tail -5
```

Expected: 最高是 000025_m2b_seed_settings.up.sql

- [ ] **Step 2: 写 MySQL up**

创建 `migrations/mysql/000026_billing_integration.up.sql`:

```sql
ALTER TABLE api_tokens  ADD COLUMN group_id BIGINT NOT NULL DEFAULT 0;
ALTER TABLE channels    ADD COLUMN group_id BIGINT NOT NULL DEFAULT 0;
```

- [ ] **Step 3: 写 MySQL down**

创建 `migrations/mysql/000026_billing_integration.down.sql`:

```sql
ALTER TABLE api_tokens  DROP COLUMN group_id;
ALTER TABLE channels    DROP COLUMN group_id;
```

- [ ] **Step 4: 写 PostgreSQL up**

创建 `migrations/postgres/000026_billing_integration.up.sql`:

```sql
ALTER TABLE api_tokens  ADD COLUMN group_id BIGINT NOT NULL DEFAULT 0;
ALTER TABLE channels    ADD COLUMN group_id BIGINT NOT NULL DEFAULT 0;
```

- [ ] **Step 5: 写 PostgreSQL down**

创建 `migrations/postgres/000026_billing_integration.down.sql`:

```sql
ALTER TABLE api_tokens  DROP COLUMN IF EXISTS group_id;
ALTER TABLE channels    DROP COLUMN IF EXISTS group_id;
```

- [ ] **Step 6: 验证文件存在**

```bash
ls migrations/mysql/000026* migrations/postgres/000026*
```

Expected: 4 个文件

- [ ] **Step 7: commit**

```bash
git add migrations/
git commit -m "feat(migration): 000026 — api_tokens/channels 加 group_id"
```

---

## Task 2: Token 模型加 GroupID

**Files:**
- 修改: `internal/token/model.go`

- [ ] **Step 1: 写失败测试**

在 `internal/token/` 查找现有测试文件。如果没有 `model_test.go`,创建它:

```go
package token

import "testing"

func TestTokenHasGroupID(t *testing.T) {
	var tok Token
	tok.GroupID = 42
	if tok.GroupID != 42 {
		t.Fatal("GroupID not settable")
	}
	var v View
	v.GroupID = 42
	if v.GroupID != 42 {
		t.Fatal("View.GroupID not settable")
	}
}
```

- [ ] **Step 2: 运行失败测试**

```bash
go test ./internal/token/ -run TestTokenHasGroupID -v
```

Expected: `FAIL` — `tok.GroupID undefined`

- [ ] **Step 3: 在 Token struct 加 GroupID**

在 `internal/token/model.go` 的 `Token` struct 中,在 `UserID` 行之后加:

```go
GroupID       int64           `gorm:"column:group_id;default:0"`
```

在 `View` struct 中同样在 `UserID` 之后加:

```go
GroupID int64
```

同时在 `Store` 接口中 `Decode` 方法(若存在)或 token service 的 `toView` 辅助函数中同步赋值 `GroupID`。搜索 toView/decode 模式:

```bash
grep -n "toView\|GroupID\|QuotaUsed" internal/token/service.go internal/token/repository.go 2>/dev/null | head -20
```

如找到 `toView` 函数:

```go
// 在赋值 UserID 之后添加:
GroupID: t.GroupID,
```

- [ ] **Step 4: 运行测试通过**

```bash
go test ./internal/token/ -run TestTokenHasGroupID -v
```

Expected: `PASS`

- [ ] **Step 5: 编译验证**

```bash
go build ./internal/token/...
```

Expected: 无报错

- [ ] **Step 6: commit**

```bash
git add internal/token/
git commit -m "feat(token): model + View 加 GroupID 字段"
```

---

## Task 3: Channel 模型加 GroupID

**Files:**
- 修改: `internal/channel/channel.go`

- [ ] **Step 1: 写失败测试**

创建 `internal/channel/group_test.go`(若 channel 包无该文件):

```go
package channel

import "testing"

func TestChannelHasGroupID(t *testing.T) {
	var ch Channel
	ch.GroupID = 5
	if ch.GroupID != 5 {
		t.Fatal("GroupID not settable on Channel")
	}
}
```

- [ ] **Step 2: 运行失败测试**

```bash
go test ./internal/channel/ -run TestChannelHasGroupID -v
```

Expected: `FAIL` — `ch.GroupID undefined`

- [ ] **Step 3: 加 GroupID 到 Channel struct**

在 `internal/channel/channel.go` 的 `Channel` struct 末尾、`CreatedAt` 之前加:

```go
GroupID     int64           `gorm:"column:group_id;default:0"         json:"group_id"`
```

- [ ] **Step 4: 运行通过**

```bash
go test ./internal/channel/ -run TestChannelHasGroupID -v
```

Expected: `PASS`

- [ ] **Step 5: commit**

```bash
git add internal/channel/channel.go internal/channel/group_test.go
git commit -m "feat(channel): Channel struct 加 GroupID 字段"
```

---

## Task 4: Channel Selector 按 GroupID 过滤

**Files:**
- 修改: `internal/channel/selector.go`
- Test: `internal/channel/selector_test.go`(新增 或 已存在则添加用例)

当前 `Select` 方法过滤链:status → Excluded → breaker → tags。
新增:若 `hint.GroupID > 0`,则只保留 `c.GroupID == 0 || c.GroupID == hint.GroupID` 的渠道。

- [ ] **Step 1: 写失败测试**

在 `internal/channel/` 找 selector 测试文件,如无则创建 `selector_group_test.go`:

```go
package channel

import (
	"context"
	"encoding/json"
	"testing"
)

func TestSelectorGroupIDFilter(t *testing.T) {
	chA := &Channel{ID: 1, Status: 0, Priority: 1, Weight: 1, GroupID: 0, Tags: json.RawMessage(`[]`)}   // 全局渠道
	chB := &Channel{ID: 2, Status: 0, Priority: 1, Weight: 1, GroupID: 5, Tags: json.RawMessage(`[]`)}   // group 5 专用
	chC := &Channel{ID: 3, Status: 0, Priority: 1, Weight: 1, GroupID: 9, Tags: json.RawMessage(`[]`)}   // group 9 专用

	// 注册模型 "m1" 到三个渠道
	cc := &channelCache{}
	cc.mu.Lock()
	cc.byModel = map[string][]*Channel{"m1": {chA, chB, chC}}
	cc.mu.Unlock()

	sel := newSelector(cc, &noopBreaker{}, 1)

	// hint.GroupID=5: 应只返回 chA(group=0) 或 chB(group=5),不返回 chC
	for i := 0; i < 20; i++ {
		ch, err := sel.Select(context.Background(), "m1", SelectHint{GroupID: 5})
		if err != nil {
			t.Fatal(err)
		}
		if ch.ID == 3 {
			t.Error("group 9 channel returned for group 5 hint")
		}
	}

	// hint.GroupID=0: 全部可用
	seen := map[int64]bool{}
	for i := 0; i < 60; i++ {
		ch, err := sel.Select(context.Background(), "m1", SelectHint{GroupID: 0})
		if err != nil {
			t.Fatal(err)
		}
		seen[ch.ID] = true
	}
	if !seen[3] {
		t.Error("group 9 channel not seen when GroupID hint is 0")
	}
}

// noopBreaker implements Breaker with no-op behavior for tests.
type noopBreaker struct{}

func (b *noopBreaker) State(id int64) BreakerState               { return StateClosed }
func (b *noopBreaker) RecordSuccess(id int64, _ interface{})      {}
func (b *noopBreaker) RecordFailure(id int64, _ error)            {}
```

查看 `Breaker` 接口实际签名:

```bash
grep -n "type Breaker interface\|State\|Record" internal/channel/breaker.go 2>/dev/null | head -10
```

根据实际签名调整 `noopBreaker` 方法。

- [ ] **Step 2: 运行失败测试**

```bash
go test ./internal/channel/ -run TestSelectorGroupIDFilter -v
```

Expected: `FAIL` — `chC` 在 group 5 hint 时仍会出现(因为 selector 没有 group 过滤)

- [ ] **Step 3: 在 selector.Select 加 GroupID 过滤**

在 `internal/channel/selector.go` 的 `for _, c := range candidates` 循环中,在现有过滤条件之后加:

```go
if hint.GroupID > 0 && c.GroupID != 0 && c.GroupID != hint.GroupID {
    continue
}
```

完整过滤块变成:

```go
for _, c := range candidates {
    if c.Status != 0 {
        continue
    }
    if inInt64Slice(hint.Excluded, c.ID) {
        continue
    }
    if s.breaker.State(c.ID) == StateOpen {
        continue
    }
    if !tagsMatch(c.Tags, hint.Tags) {
        continue
    }
    if hint.GroupID > 0 && c.GroupID != 0 && c.GroupID != hint.GroupID {
        continue
    }
    filtered = append(filtered, c)
}
```

- [ ] **Step 4: 运行通过**

```bash
go test ./internal/channel/ -run TestSelectorGroupIDFilter -v
```

Expected: `PASS`

- [ ] **Step 5: 运行 channel 包全量测试**

```bash
go test ./internal/channel/... -v 2>&1 | tail -20
```

Expected: 全部 `PASS`

- [ ] **Step 6: commit**

```bash
git add internal/channel/
git commit -m "feat(channel): Selector 按 hint.GroupID 过滤渠道"
```

---

## Task 5: GroupRatio 中间件

**Files:**
- 新增: `internal/server/middleware/groupratio.go`

此中间件做两件事:
1. 如果 `token.View.GroupID > 0`,用 token 级分组覆盖 `proapi:group_id`(覆盖 user 级分组)
2. 读取 groupID → 调用 ratioLookup → 写入 `proapi:group_ratio`(供 ratelimit 中间件) 和 `proapi:billing_group_ratio`(供 relay handler)

- [ ] **Step 1: 写测试文件**

新建 `internal/server/middleware/groupratio_test.go`:

```go
package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ijry/pro-api/internal/token"
)

func TestGroupRatioMiddleware_SetsRatio(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	lookup := func(_ context.Context, gid int64) (float64, error) {
		if gid == 5 {
			return 2.0, nil
		}
		return 1.0, nil
	}

	r.GET("/test", func(c *gin.Context) {
		// 模拟 TokenAuth 设置了 group_id=5
		c.Set(token.CtxKeyGroupID, int64(5))
	}, GroupRatioMiddleware(lookup), func(c *gin.Context) {
		ratio, _ := c.Get(CtxKeyBillingGroupRatio)
		if ratio != 2.0 {
			t.Errorf("billing ratio: want 2.0, got %v", ratio)
		}
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)
}

func TestGroupRatioMiddleware_TokenGroupOverridesUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	lookup := func(_ context.Context, gid int64) (float64, error) {
		return float64(gid) * 0.5, nil // ratio = gid * 0.5 for test
	}

	r.GET("/test",
		func(c *gin.Context) {
			// TokenAuth 设置了 user 级 group_id=3
			c.Set(token.CtxKeyGroupID, int64(3))
			// Token 本身有 group_id=7
			c.Set(token.CtxKeyToken, &token.View{GroupID: 7})
		},
		GroupRatioMiddleware(lookup),
		func(c *gin.Context) {
			gid, _ := c.Get(token.CtxKeyGroupID)
			if gid != int64(7) {
				t.Errorf("want group_id=7 (token override), got %v", gid)
			}
		},
	)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)
}
```

- [ ] **Step 2: 运行失败**

```bash
go test ./internal/server/middleware/ -run "TestGroupRatio" -v
```

Expected: `FAIL` — `GroupRatioMiddleware undefined`

- [ ] **Step 3: 实现**

新建 `internal/server/middleware/groupratio.go`:

```go
package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/ijry/pro-api/internal/ratelimit"
	"github.com/ijry/pro-api/internal/token"
)

const (
	// CtxKeyBillingGroupRatio 是 relay handler 读取的分组倍率(用于 pricing.EstimateInput)。
	CtxKeyBillingGroupRatio = "proapi:billing_group_ratio"
)

// GroupRatioLookup 返回 groupID 对应的消耗倍率。
type GroupRatioLookup func(ctx context.Context, groupID int64) (float64, error)

// GroupRatioMiddleware 解析当前请求的分组倍率并注入 context。
//
// 执行顺序:
//  1. 若 token.View.GroupID > 0,用 token 级分组覆盖 proapi:group_id
//  2. 读取 groupID,调用 lookup 查倍率
//  3. 写入 proapi:group_ratio(供 ratelimit)和 proapi:billing_group_ratio(供 relay handler)
func GroupRatioMiddleware(lookup GroupRatioLookup) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Token 级分组覆盖
		if tv, ok := c.Get(token.CtxKeyToken); ok {
			if view, ok := tv.(*token.View); ok && view != nil && view.GroupID > 0 {
				c.Set(token.CtxKeyGroupID, view.GroupID)
			}
		}

		// 2. 查倍率
		gid := token.GroupIDFromContext(c)
		if gid > 0 && lookup != nil {
			if ratio, err := lookup(c.Request.Context(), gid); err == nil {
				c.Set(ratelimit.CtxKeyGroupRatio, ratio)
				c.Set(CtxKeyBillingGroupRatio, ratio)
			}
		}

		c.Next()
	}
}
```

- [ ] **Step 4: 运行通过**

```bash
go test ./internal/server/middleware/ -run "TestGroupRatio" -v
```

Expected: `PASS`

- [ ] **Step 5: 编译**

```bash
go build ./internal/server/middleware/...
```

- [ ] **Step 6: commit**

```bash
git add internal/server/middleware/groupratio.go internal/server/middleware/groupratio_test.go
git commit -m "feat(middleware): GroupRatioMiddleware — token 级分组覆盖 + 倍率注入"
```

---

## Task 6: Wire Pricing Service

**Files:**
- 新增: `internal/pricing/wire.go`
- 修改: `internal/app/application.go`(如果还没有 `Biller any` 字段,确认后补)
- 修改: `cmd/proapi/main.go`

目前 `a.Biller` 已经由 `billing.WireBilling(application)` 在 main.go 第 112 行填好。
`a.PricingSvc` 是 `any`、始终 nil。需要新建 `pricing.WirePricing(a)` 函数。

- [ ] **Step 1: 检查 pricing.Config 所需依赖**

```bash
grep -n "type Config struct\|IDGenerator\|CatalogInfo\|GroupRatioLookup" internal/pricing/pricing.go | head -20
grep -n "type IDGenerator interface\|type CatalogInfo" internal/pricing/pricing.go | head -10
```

Expected: 找到 `IDGenerator` 是接口(Generate() int64),`CatalogInfo` 是接口或结构。

- [ ] **Step 2: 查 ModelCatalog 接口在哪**

```bash
grep -rn "CatalogInfo\|type.*Catalog" internal/pricing/ --include="*.go" | grep -v _test | head -10
```

Expected: 找到 `CatalogInfo` 的定义及其 `ModelInfo` 方法等。

- [ ] **Step 3: 新建 internal/pricing/wire.go**

根据 Step 1-2 的结果,`Config.Catalog` 和 `Config.GroupRatio` 都是可 nil 的。写:

```go
package pricing

import (
	"context"
	"fmt"

	"github.com/ijry/pro-api/internal/app"
	"github.com/ijry/pro-api/internal/group"
)

// WirePricing 装配 pricing.Service 并注入到 app.PricingSvc。
// GroupRatio lookup 从 app.GroupSvc(group.Service) 中取 RatioFor。
func WirePricing(ctx context.Context, a *app.Application) error {
	if a == nil {
		return fmt.Errorf("WirePricing: app is nil")
	}

	var groupRatio GroupRatioLookup
	if gs, ok := a.GroupSvc.(group.Service); ok {
		groupRatio = func(ctx context.Context, id int64) float64 {
			r, _ := gs.RatioFor(ctx, id)
			return r
		}
	}

	svc, err := New(ctx, Config{
		DB:           a.DB,
		Cache:        a.Cache,
		Log:          a.Log,
		Clock:        a.Clock,
		IDGen:        a.IDGen,
		Audit:        a.Audit,
		GroupRatio:   groupRatio,
	})
	if err != nil {
		return fmt.Errorf("WirePricing: %w", err)
	}
	a.PricingSvc = svc
	a.AddCloser("pricing", svc.Close)
	return nil
}

// PricingFrom 从 app.PricingSvc 取回 Pricing 接口。nil 表示未装配。
func PricingFrom(a *app.Application) Pricing {
	if a == nil {
		return nil
	}
	p, _ := a.PricingSvc.(Pricing)
	return p
}
```

注意: `IDGen *idgen.Generator` 在 `Config` 中是 `IDGenerator` 接口,`a.IDGen` 是 `*idgen.Generator`,它实现了 `Generate() int64`。只需直接传即可。

- [ ] **Step 4: 编译**

```bash
go build ./internal/pricing/...
```

Expected: 无报错

- [ ] **Step 5: 在 main.go 调用 WirePricing**

在 `cmd/proapi/main.go` 找到 `billing.WireBilling(application)` 附近(约第 112 行),在其之后加:

```go
if err := pricing.WirePricing(ctx, application); err != nil {
    log.Warn("pricing wire failed, billing estimates will be zero", zap.Error(err))
    // 不 fatal — pricing 不可用时 relay 仍可运行,只是不做预扣
}
```

同时在 import 块加 `"github.com/ijry/pro-api/internal/pricing"`(如果尚未存在)。

- [ ] **Step 6: 全量编译**

```bash
go build ./cmd/proapi/...
```

Expected: 无报错

- [ ] **Step 7: commit**

```bash
git add internal/pricing/wire.go cmd/proapi/main.go
git commit -m "feat(pricing): WirePricing + PricingFrom + 在 main.go 装配"
```

---

## Task 7: Relay Handler 加 Billing

**Files:**
- 修改: `internal/server/handler/relay/handler.go`

这是核心改动。将 `Handler` 改为 Deps 模式,在 `ChatCompletions` 和 `streamChat` 中分别加入 billing 流程。

当前:
- `Handler struct { relay *relaySvc.Service }`
- `New(r *relaySvc.Service) *Handler`

改后:
- `Handler struct{ deps Deps }`
- `Deps struct{ Relay, Pricing, Biller, LogSvc 均可 nil }`
- `New(deps Deps) *Handler`
- `NewSimple(r *relaySvc.Service) *Handler` — 兼容现有调用点(在 main.go 还没改之前)

读取上下文帮助函数:
- `token.GroupIDFromContext(c)` → groupID
- `token.UserIDFromContext(c)` → userID  
- `token.FromContext(c.Request.Context())` → (*View, bool) → tokenID
- `c.GetFloat64(middleware.CtxKeyBillingGroupRatio)` → billingGroupRatio

- [ ] **Step 1: 查看 billing.Biller 接口具体签名**

```bash
grep -n "Reserve\|Commit\|Refund" internal/billing/biller.go | head -10
```

Expected:
```
Reserve(ctx context.Context, userID, tokenID, estCost int64) (reservationID string, err error)
Commit(ctx context.Context, reservationID string, actualCost int64) error
Refund(ctx context.Context, reservationID string) error
```

- [ ] **Step 2: 查 BillerFrom 模式**

```bash
grep -n "BillerFrom\|Biller\b" internal/billing/wire.go | head -10
```

如果没有 `BillerFrom` 函数则跳过;我们在 relay handler 的 deps 里直接传 billing.Biller 接口。

- [ ] **Step 3: 写 relay handler 测试(billing 路径)**

在 `internal/server/handler/relay/` 创建 `billing_test.go`:

```go
package relay

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ijry/pro-api/internal/billing"
	"github.com/ijry/pro-api/internal/pricing"
	"github.com/ijry/pro-api/internal/protocol/ir"
	relaySvc "github.com/ijry/pro-api/internal/relay"
	"github.com/ijry/pro-api/internal/token"
)

// fakePricing is a stub pricing.Pricing.
type fakePricing struct{ estCost int64 }

func (f *fakePricing) Compute(_ context.Context, _ pricing.ComputeInput) pricing.ComputeResult {
	return pricing.ComputeResult{Quota: f.estCost, Ratios: pricing.Ratios{Group: 1.0}}
}
func (f *fakePricing) RatioFor(_ context.Context, _ string, _ int64, _ pricing.ChannelInfo) pricing.Ratios {
	return pricing.Ratios{}
}
func (f *fakePricing) EstimateMax(_ context.Context, _ string, _ pricing.EstimateInput) int64 {
	return f.estCost
}
func (f *fakePricing) DefaultMaxOut(_ context.Context, _ string) int { return 4096 }

// fakeBiller is a stub billing.Biller.
type fakeBiller struct {
	reserved  bool
	committed bool
	refunded  bool
	lastCost  int64
}

func (b *fakeBiller) Reserve(_ context.Context, _, _, _ int64) (string, error) {
	b.reserved = true
	return "rsv-1", nil
}
func (b *fakeBiller) Commit(_ context.Context, _ string, cost int64) error {
	b.committed = true
	b.lastCost = cost
	return nil
}
func (b *fakePricing) Refund(_ context.Context, _ string) error { return nil }
func (b *fakeBiller) Refund(_ context.Context, _ string) error  { b.refunded = true; return nil }
func (b *fakeBiller) Close() error                              { return nil }

func buildTestRelayRouter(t *testing.T, relaySvcMock *relaySvc.Service, fp *fakePricing, fb *fakeBiller) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := New(Deps{
		Relay:   relaySvcMock,
		Pricing: fp,
		Biller:  fb,
	})
	r.POST("/v1/chat/completions", func(c *gin.Context) {
		c.Set(token.CtxKeyUserID, int64(1))
		c.Set(token.CtxKeyToken, &token.View{ID: 10, GroupID: 0})
		c.Set(token.CtxKeyGroupID, int64(0))
	}, h.ChatCompletions)
	return r
}

func TestChatCompletions_BillingCommit(t *testing.T) {
	// This test verifies that biller.Reserve and biller.Commit are called.
	// Since relay.Service requires real dependencies, we skip if nil.
	t.Skip("integration test — run manually with a wired relay service")
}

func TestDeps_NilBillerSkipsBilling(t *testing.T) {
	// With nil Biller, ChatCompletions should still work (no billing crash).
	// The actual relay call will fail with missing deps — tested via compilation only.
	h := New(Deps{})
	if h == nil {
		t.Fatal("New with empty Deps should return non-nil Handler")
	}
}
```

注意: relay.Service 需要完整依赖,无法在单元测试中 mock。本测试只验证 Deps 构造和 nil-safety。

- [ ] **Step 4: 运行测试(期望 FAIL 因为 New 签名变了)**

```bash
go test ./internal/server/handler/relay/ -run "TestDeps" -v
```

Expected: `FAIL` — `New` does not take `Deps`

- [ ] **Step 5: 实现 Deps 模式 + billing 流程**

将 `internal/server/handler/relay/handler.go` 的顶部结构改为:

```go
package relay

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ijry/pro-api/internal/adapter"
	"github.com/ijry/pro-api/internal/billing"
	"github.com/ijry/pro-api/internal/channel"
	ilog "github.com/ijry/pro-api/internal/log"
	"github.com/ijry/pro-api/internal/pricing"
	"github.com/ijry/pro-api/internal/protocol/anthropic"
	"github.com/ijry/pro-api/internal/protocol/gemini"
	"github.com/ijry/pro-api/internal/protocol/ir"
	"github.com/ijry/pro-api/internal/protocol/openai"
	relaySvc "github.com/ijry/pro-api/internal/relay"
	"github.com/ijry/pro-api/internal/server/middleware"
	"github.com/ijry/pro-api/internal/token"
	"github.com/ijry/pro-api/pkg/apierr"
)

// Deps holds relay handler dependencies.
// Pricing, Biller, and LogSvc are optional; nil means skip billing/logging.
type Deps struct {
	Relay   *relaySvc.Service
	Pricing pricing.Pricing
	Biller  billing.Biller
	LogSvc  ilog.Store
}

// Handler handles relay (proxy) endpoints.
type Handler struct{ deps Deps }

// New constructs a relay Handler with full dependency injection.
func New(deps Deps) *Handler { return &Handler{deps: deps} }
```

将 `ChatCompletions` 改为:

```go
// ChatCompletions handles POST /v1/chat/completions
func (h *Handler) ChatCompletions(c *gin.Context) {
	req, err := openai.DecodeChat(c.Request.Body, openai.DecodeOptions{AllowUnknownFields: true})
	if err != nil {
		middleware.SetErr(c, apierr.New(apierr.CodeInvalidParam, err.Error()))
		return
	}
	ch := channelFromContext(c)

	if req.Stream {
		h.streamChat(c, req, ch)
		return
	}

	ctx := c.Request.Context()
	userID := token.UserIDFromContext(c)
	groupID := token.GroupIDFromContext(c)
	var tokenID int64
	if tv, ok := token.FromContext(ctx); ok {
		tokenID = tv.ID
	}

	// Pre-deduct
	var reservationID string
	if h.deps.Pricing != nil && h.deps.Biller != nil {
		billingRatio, _ := c.Get(middleware.CtxKeyBillingGroupRatio)
		var br float64
		if v, ok := billingRatio.(float64); ok {
			br = v
		}
		estCost := h.deps.Pricing.EstimateMax(ctx, req.Model, pricing.EstimateInput{
			InputTokens:       countInputTokens(req),
			MaxOutTokens:      h.deps.Pricing.DefaultMaxOut(ctx, req.Model),
			Stream:            false,
			BillingGroupRatio: br,
		})
		if estCost > 0 {
			reservationID, err = h.deps.Biller.Reserve(ctx, userID, tokenID, estCost)
			if err != nil {
				middleware.SetErr(c, apierr.New(apierr.CodeInsufficientQuota, "quota reserve failed"))
				return
			}
		}
	}

	start := time.Now()
	resp, _, err := h.deps.Relay.Chat(ctx, req, ch)
	latencyMS := int(time.Since(start).Milliseconds())

	if err != nil {
		if reservationID != "" && h.deps.Biller != nil {
			_ = h.deps.Biller.Refund(ctx, reservationID)
		}
		middleware.SetErr(c, mapErr(err))
		return
	}

	// Commit actual cost
	var actualCost int64
	var ratios pricing.Ratios
	if h.deps.Pricing != nil && resp.Usage.PromptTokens+resp.Usage.CompletionTokens > 0 {
		result := h.deps.Pricing.Compute(ctx, pricing.ComputeInput{
			Model:        req.Model,
			GroupID:      groupID,
			InputTokens:  resp.Usage.PromptTokens,
			OutputTokens: resp.Usage.CompletionTokens,
			CachedTokens: resp.Usage.CachedTokens,
		})
		actualCost = result.Quota
		ratios = result.Ratios
		if reservationID != "" && h.deps.Biller != nil {
			_ = h.deps.Biller.Commit(ctx, reservationID, actualCost)
		}
	} else if reservationID != "" && h.deps.Biller != nil {
		_ = h.deps.Biller.Refund(ctx, reservationID)
	}

	// Write request log
	h.writeLog(ctx, c, req.Model, groupID, tokenID, latencyMS, false,
		resp.Usage.PromptTokens, resp.Usage.CompletionTokens, resp.Usage.CachedTokens,
		actualCost, ratios, http.StatusOK, "")

	c.JSON(http.StatusOK, openai.EncodeChat(resp))
}
```

将 `streamChat` 改为:

```go
func (h *Handler) streamChat(c *gin.Context, req *ir.ChatRequest, ch *channel.Channel) {
	ctx := c.Request.Context()
	userID := token.UserIDFromContext(c)
	groupID := token.GroupIDFromContext(c)
	var tokenID int64
	if tv, ok := token.FromContext(ctx); ok {
		tokenID = tv.ID
	}

	// Pre-deduct
	var reservationID string
	if h.deps.Pricing != nil && h.deps.Biller != nil {
		billingRatio, _ := c.Get(middleware.CtxKeyBillingGroupRatio)
		var br float64
		if v, ok := billingRatio.(float64); ok {
			br = v
		}
		estCost := h.deps.Pricing.EstimateMax(ctx, req.Model, pricing.EstimateInput{
			InputTokens:       countInputTokens(req),
			MaxOutTokens:      h.deps.Pricing.DefaultMaxOut(ctx, req.Model),
			Stream:            true,
			BillingGroupRatio: br,
		})
		if estCost > 0 {
			var err error
			reservationID, err = h.deps.Biller.Reserve(ctx, userID, tokenID, estCost)
			if err != nil {
				middleware.SetErr(c, apierr.New(apierr.CodeInsufficientQuota, "quota reserve failed"))
				return
			}
		}
	}

	start := time.Now()
	reader, _, err := h.deps.Relay.ChatStream(ctx, req, ch)
	if err != nil {
		if reservationID != "" && h.deps.Biller != nil {
			_ = h.deps.Biller.Refund(ctx, reservationID)
		}
		middleware.SetErr(c, mapErr(err))
		return
	}
	defer reader.Close()

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("X-Accel-Buffering", "no")

	flusher, hasFlusher := c.Writer.(http.Flusher)
	var flushFn func()
	if hasFlusher {
		flushFn = flusher.Flush
	}
	sw := openai.NewStreamWriter(c.Writer, flushFn)

	var finalUsage ir.Usage
	for {
		chunk, err := reader.Next(ctx)
		if err != nil {
			if err == io.EOF {
				_ = sw.WriteDone()
				break
			}
			_, _ = fmt.Fprintf(c.Writer, "data: {\"error\":\"stream error: %s\"}\n\n", err.Error())
			if hasFlusher {
				flusher.Flush()
			}
			break
		}
		if chunk.Usage != nil {
			finalUsage = *chunk.Usage
		}
		if err := sw.WriteChunk(chunk); err != nil {
			break
		}
	}

	latencyMS := int(time.Since(start).Milliseconds())

	// Commit after stream ends
	var actualCost int64
	var ratios pricing.Ratios
	if h.deps.Pricing != nil && finalUsage.PromptTokens+finalUsage.CompletionTokens > 0 {
		result := h.deps.Pricing.Compute(ctx, pricing.ComputeInput{
			Model:        req.Model,
			GroupID:      groupID,
			InputTokens:  finalUsage.PromptTokens,
			OutputTokens: finalUsage.CompletionTokens,
			CachedTokens: finalUsage.CachedTokens,
		})
		actualCost = result.Quota
		ratios = result.Ratios
		if reservationID != "" && h.deps.Biller != nil {
			_ = h.deps.Biller.Commit(ctx, reservationID, actualCost)
		}
	} else if reservationID != "" && h.deps.Biller != nil {
		_ = h.deps.Biller.Refund(ctx, reservationID)
	}

	h.writeLog(ctx, c, req.Model, groupID, tokenID, latencyMS, true,
		finalUsage.PromptTokens, finalUsage.CompletionTokens, finalUsage.CachedTokens,
		actualCost, ratios, http.StatusOK, "")
}
```

添加 helper 函数到 handler.go:

```go
// countInputTokens 估算请求的输入 token 数(粗略:每条 message 文本长度 / 4)。
// ir.Message.Content 是 []ir.ContentPart,每个 ContentPart.Text 是文本内容。
func countInputTokens(req *ir.ChatRequest) int {
	n := 0
	for _, m := range req.Messages {
		for _, p := range m.Content {
			if p.Text != "" {
				n += len(p.Text) / 4
			}
		}
	}
	if n < 10 {
		n = 10
	}
	return n
}

// writeLog 非阻塞写入一条请求日志。
func (h *Handler) writeLog(ctx context.Context, c *gin.Context, model string,
	groupID, tokenID int64, latencyMS int, stream bool,
	in, out, cached int, cost int64, ratios pricing.Ratios,
	statusCode int, errMsg string,
) {
	if h.deps.LogSvc == nil {
		return
	}
	userID := token.UserIDFromContext(c)
	gid := groupID
	e := ilog.Event{
		UserID:             userID,
		TokenID:            tokenID,
		GroupID:            &gid,
		ClientModel:        model,
		Protocol:           "openai",
		Endpoint:           c.FullPath(),
		IP:                 c.ClientIP(),
		UserAgent:          c.Request.UserAgent(),
		StatusCode:         statusCode,
		LatencyMS:          latencyMS,
		Stream:             stream,
		InputTokens:        in,
		OutputTokens:       out,
		CachedTokens:       cached,
		TotalQuota:         cost,
		BillingInputRatio:  ratios.Input,
		BillingOutputRatio: ratios.Output,
		BillingGroupRatio:  ratios.Group,
		ErrorMsg:           errMsg,
	}
	h.deps.LogSvc.Write(ctx, e)
}
```

注意: `ir.ChatRequest.Messages` 是 `[]ir.Message`,每条 Message 有 `Content []ir.ContentPart`,每个 `ContentPart` 有 `Text string`。`countInputTokens` 已按此字段名编写。

- [ ] **Step 6: 运行测试**

```bash
go test ./internal/server/handler/relay/ -run "TestDeps" -v
```

Expected: `PASS`

- [ ] **Step 7: 全量编译**

```bash
go build ./internal/server/handler/relay/...
```

- [ ] **Step 8: commit**

```bash
git add internal/server/handler/relay/
git commit -m "feat(relay): Handler 改 Deps 模式 + ChatCompletions/streamChat billing 接线"
```

---

## Task 8: main.go 接线 — 挂中间件 + 更新 relay Handler 构造

**Files:**
- 修改: `cmd/proapi/main.go`

当前 relay handler 的构造方式(在 `wireRoutes` 函数中,约第 242 行):

```go
if relaySvc, ok := a.Relay.(*relay.Service); ok {
    relayH := relayhdr.New(relaySvc)
    relayH.Register(v1)
    ...
}
```

改为:

```go
if relaySvc, ok := a.Relay.(*relay.Service); ok {
    relayH := relayhdr.New(relayhdr.Deps{
        Relay:   relaySvc,
        Pricing: pricing.PricingFrom(a),
        Biller:  billerFrom(a),
        LogSvc:  ilog.StoreFrom(a),
    })
    relayH.Register(v1)
    ...
}
```

同时:
1. `v1` 路由组加入 GroupRatio 中间件和 ratelimit 中间件
2. 添加 `billerFrom` 辅助函数

- [ ] **Step 1: 检查 BillerFrom 是否已存在**

```bash
grep -n "BillerFrom\|a\.Biller\b" internal/billing/wire.go cmd/proapi/main.go | head -10
```

如无 `BillerFrom`,则在 `main.go` 中内联 type assertion:

```go
func billerFrom(a *app.Application) billing.Biller {
    b, _ := a.Biller.(billing.Biller)
    return b
}
```

- [ ] **Step 2: 查 group 服务的 RatioFor 闭包构造**

我们需要把 group.Service.RatioFor 包装成 `middleware.GroupRatioLookup`:

```go
var groupRatioLookup middleware.GroupRatioLookup
if gs, ok := a.GroupSvc.(group.Service); ok {
    groupRatioLookup = func(ctx context.Context, id int64) (float64, error) {
        return gs.RatioFor(ctx, id)
    }
}
```

- [ ] **Step 3: 修改 wireRoutes 中的 v1 路由组**

找到:

```go
v1 := eng.Group("/v1", middleware.ErrorResponse("openai"))
```

改为:

```go
var groupRatioLookup mw.GroupRatioLookup
if gs, ok := a.GroupSvc.(group.Service); ok {
    groupRatioLookup = func(ctx context.Context, id int64) (float64, error) {
        return gs.RatioFor(ctx, id)
    }
}

v1 := eng.Group("/v1", middleware.ErrorResponse("openai"))
// TokenAuth 已在 authhwire.RegisterRoutes 中挂到 v1 — 不重复挂
// GroupRatio 和 ratelimit 挂在已有 TokenAuth 之后:
v1.Use(mw.GroupRatioMiddleware(groupRatioLookup))
if limiter, ok := a.Limiter.(ratelimit.Limiter); ok {
    planner := ratelimit.PlannerFrom(a) // 已在 WireRateLimit 中注册
    if planner != nil {
        v1.Use(ratelimit.Middleware(limiter, planner, a.Setting, log))
    }
}

- [ ] **Step 4: 修改 relay handler 构造**

找到:
```go
relayH := relayhdr.New(relaySvc)
```

改为:
```go
relayH := relayhdr.New(relayhdr.Deps{
    Relay:   relaySvc,
    Pricing: pricingFrom(a),
    Biller:  billerFrom(a),
    LogSvc:  ilog.StoreFrom(a),
})
```

其中 `pricingFrom` 包装 pricing.PricingFrom:

```go
func pricingFrom(a *app.Application) pricing.Pricing {
    return pricing.PricingFrom(a)
}
```

或直接内联 `pricing.PricingFrom(a)`。

- [ ] **Step 5: 全量编译**

```bash
go build ./cmd/proapi/...
```

如有编译错误,逐一修复:
- import 缺失 → 加到 import 块
- 类型不匹配 → 查实际接口并调整

- [ ] **Step 6: 全量测试**

```bash
go test ./internal/... 2>&1 | tail -30
```

Expected: 全部 PASS(relay handler 的 integration test 跳过)

- [ ] **Step 7: commit**

```bash
git add cmd/proapi/main.go
git commit -m "feat(server): relay 路由挂 GroupRatio/ratelimit 中间件 + Biller/Pricing/Log 依赖注入"
```

---

## Task 9: 全量验证

**Files:** 无新增

- [ ] **Step 1: 全量编译**

```bash
go build ./...
```

Expected: 零错误

- [ ] **Step 2: 运行全量测试**

```bash
go test ./... 2>&1 | grep -E "FAIL|ok" | head -40
```

Expected: 无 `FAIL`

- [ ] **Step 3: 确认关键路径**

```bash
# GroupID 字段存在
grep -n "GroupID" internal/token/model.go internal/channel/channel.go

# Selector 过滤存在
grep -n "hint.GroupID" internal/channel/selector.go

# GroupRatio 中间件存在
grep -n "GroupRatioMiddleware\|CtxKeyBillingGroupRatio" internal/server/middleware/groupratio.go

# Pricing wired
grep -n "WirePricing\|PricingFrom" internal/pricing/wire.go

# Relay Deps 模式
grep -n "type Deps struct\|func New" internal/server/handler/relay/handler.go
```

- [ ] **Step 4: 最终 commit(如有遗留改动)**

```bash
git status
# 如无未提交内容则跳过
```

---

## 自查:Spec 覆盖验证

| Spec 要求 | 计划 Task |
|-----------|-----------|
| Token 加 group_id | Task 1(migration) + Task 2(model) |
| Channel 加 group_id | Task 1(migration) + Task 3(model) |
| Channel selector 按组过滤 | Task 4 |
| GroupRatio 中间件 | Task 5 |
| Pricing service 装配 | Task 6 |
| Relay handler billing 流程 | Task 7 |
| ratelimit 挂载 | Task 8 |
| main.go 完整接线 | Task 8 |
| Migration 文件 | Task 1 |
| CtxKeyBillingGroupRatio | Task 5 |
| 流式路径 billing | Task 7(streamChat) |
| 请求日志写入 | Task 7(writeLog) |
