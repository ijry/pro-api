package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
)

const (
	// CtxKeyGroupRatio is read by the ratelimit middleware.
	CtxKeyGroupRatio = "proapi:group_ratio"
	// CtxKeyBillingGroupRatio is read by the relay handler for pricing.EstimateInput.
	CtxKeyBillingGroupRatio = "proapi:billing_group_ratio"

	// These constants are defined here to avoid circular imports.
	// They duplicate definitions in token/middleware.go.
	ctxKeyToken   = "proapi:token"
	ctxKeyGroupID = "proapi:group_id"
)

// TokenView is a minimal copy of token.View for type checking.
// The full struct definition is in the token package.
type TokenView interface {
	// GroupID is the group id associated with this token.
	GetGroupID() int64
}

// GroupRatioLookup returns the consumption ratio for a groupID.
type GroupRatioLookup func(ctx context.Context, groupID int64) (float64, error)

// GroupRatioMiddleware resolves the current request's group ratio and injects it into context.
//
// Group is taken exclusively from token.GroupID. If the token has no group (GroupID == 0),
// no ratio is applied — there is no user-level fallback.
func GroupRatioMiddleware(lookup GroupRatioLookup) gin.HandlerFunc {
	return func(c *gin.Context) {
		var gid int64
		if tv, ok := c.Get(ctxKeyToken); ok {
			type tokenWithGroup interface {
				GetGroupID() int64
			}
			if twg, ok := tv.(tokenWithGroup); ok {
				gid = twg.GetGroupID()
			}
		}
		if gid > 0 {
			c.Set(ctxKeyGroupID, gid)
			c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), ctxKeyGroupID, gid))
			if lookup != nil {
				if ratio, err := lookup(c.Request.Context(), gid); err == nil {
					c.Set(CtxKeyGroupRatio, ratio)
					c.Set(CtxKeyBillingGroupRatio, ratio)
				}
			}
		}

		c.Next()
	}
}
