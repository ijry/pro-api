package ratelimit

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ijry/pro-api/internal/server/middleware"
	"github.com/ijry/pro-api/internal/setting"
	"github.com/ijry/pro-api/internal/token"
	"github.com/ijry/pro-api/pkg/apierr"
	"go.uber.org/zap"
)

// Ctx keys for middleware integration. 与 token 包共用 user_id / token_id 字段。
const (
	// CtxKeyTotalTokens 是 handler 在 c.Next() 期间写入的 token 数(供 TPM 计数)。
	CtxKeyTotalTokens = "proapi:total_tokens"
	// CtxKeyClientModel 是上游 ModelAllowlist / handler 写入的客户端 model 名。
	CtxKeyClientModel = "proapi:client_model"
	// CtxKeyGroupID 与 token.CtxKeyGroupID 同值。
	CtxKeyGroupID = token.CtxKeyGroupID
)

// CtxKeyGroupRatio is imported from middleware package to avoid circular imports.
var CtxKeyGroupRatio = middleware.CtxKeyGroupRatio

// contextPlanResolver 是从 gin.Context 派生 PlanInput 的钩子。
// 默认实现走 token / middleware 的 ctx keys;测试可替换。
var contextPlanResolver = defaultPlanResolver

// Middleware 构造代理 API 限流中间件。
//
// 依赖前置:TokenAuth 已把 token / user / group 信息写入 ctx;ModelAllowlist 已写 client_model。
//
// 顺序:
//  1. 全局开关 ratelimit.enabled
//  2. PlanInput → PlanRPM → AllowMulti
//  3. 被拒:WriteHeaders(denied) + SetErr → ErrorResponse 渲染
//  4. 通过:WriteHeaders(allowed) + 暂存 TPM checks → c.Next()
//  5. c.Next() 后读 total_tokens → FillTPMCost → ConsumeTPM(用 detached ctx,客户端断连不丢)
func Middleware(l Limiter, p *Planner, sett setting.Store, log *zap.Logger) gin.HandlerFunc {
	if log == nil {
		log = zap.NewNop()
	}
	return func(c *gin.Context) {
		reqCtx := c.Request.Context()
		if sett != nil && !sett.GetBool(reqCtx, "ratelimit.enabled", true) {
			c.Next()
			return
		}
		in := contextPlanResolver(c)
		rpmChecks := p.PlanRPM(reqCtx, in)
		decision := l.AllowMulti(reqCtx, rpmChecks)
		if !decision.Allowed {
			WriteHeaders(c, decision)
			middleware.SetErr(c, apierr.New(codeForDimension(decision.Dimension), messageForDimension(decision.Dimension)))
			return
		}
		WriteHeaders(c, decision)
		// 暂存 TPM checks,等 c.Next() 后用
		tpmChecks := p.PlanTPM(reqCtx, in)
		c.Next()
		// c.Next() 后:扣 TPM
		total := readIntCtx(c, CtxKeyTotalTokens)
		if total <= 0 || len(tpmChecks) == 0 {
			return
		}
		filled := FillTPMCost(tpmChecks, total)
		// 用 detached context,避免客户端断连导致 ConsumeTPM 失败
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := l.ConsumeTPM(ctx, filled); err != nil {
			log.Warn("ratelimit: tpm consume returned error",
				zap.Error(err),
				zap.Int64("user_id", in.UserID),
				zap.Int("total_tokens", total),
			)
		}
	}
}

// defaultPlanResolver 从 gin.Context 拼装 PlanInput。
func defaultPlanResolver(c *gin.Context) PlanInput {
	in := PlanInput{
		UserID:  readInt64Ctx(c, token.CtxKeyUserID),
		GroupID: readInt64Ctx(c, token.CtxKeyGroupID),
		Model:   readStringCtx(c, CtxKeyClientModel),
		IP:      c.ClientIP(),
	}
	// token id / RPM / TPM 来自 token.View
	if v, ok := c.Get(token.CtxKeyToken); ok {
		if view, ok := v.(*token.View); ok && view != nil {
			in.TokenID = view.ID
			in.TokenRPMOverride = view.RPMLimit
			in.TokenTPMOverride = view.TPMLimit
		}
	}
	// group ratio 是可选 — 若中间件链上有人放进 ctx,就用
	if v, ok := c.Get(CtxKeyGroupRatio); ok {
		switch x := v.(type) {
		case float64:
			in.GroupRatio = x
		case float32:
			in.GroupRatio = float64(x)
		case int:
			in.GroupRatio = float64(x)
		}
	}
	return in
}

// codeForDimension 把限流维度映射到 apierr 错误码。
func codeForDimension(d Dimension) apierr.Code {
	switch d {
	case DimUserRPM, DimUserTPM:
		return apierr.CodeRateLimitUser
	case DimTokenRPM, DimTokenTPM:
		return apierr.CodeRateLimitToken
	case DimIPRPM:
		return apierr.CodeRateLimitIP
	case DimModelRPM, DimModelTPM:
		return apierr.CodeRateLimitGlobal // M1 沿用 global 表达 model 维度
	}
	return apierr.CodeRateLimitGlobal
}

// messageForDimension 提供面向用户的 message。
func messageForDimension(d Dimension) string {
	switch d {
	case DimUserRPM, DimUserTPM:
		return "请求过于频繁,请稍后再试(用户维度)"
	case DimTokenRPM, DimTokenTPM:
		return "令牌请求过于频繁,请稍后再试"
	case DimIPRPM:
		return "IP 请求过于频繁,请稍后再试"
	case DimModelRPM, DimModelTPM:
		return "模型请求过于频繁,请稍后再试"
	}
	return "请求过于频繁,请稍后再试"
}

// ===== ctx helpers =====

func readIntCtx(c *gin.Context, k string) int {
	return int(readInt64Ctx(c, k))
}

func readInt64Ctx(c *gin.Context, k string) int64 {
	v, ok := c.Get(k)
	if !ok {
		return 0
	}
	switch x := v.(type) {
	case int64:
		return x
	case int:
		return int64(x)
	case float64:
		return int64(x)
	case float32:
		return int64(x)
	}
	return 0
}

func readStringCtx(c *gin.Context, k string) string {
	v, ok := c.Get(k)
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}
