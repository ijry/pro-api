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
// Order of operations:
//  1. If token.View.GroupID > 0, override proapi:group_id with the token-level group
//  2. Look up ratio for the resolved groupID
//  3. Set proapi:group_ratio (for ratelimit) and proapi:billing_group_ratio (for relay handler)
func GroupRatioMiddleware(lookup GroupRatioLookup) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Token-level group overrides user-level group
		if tv, ok := c.Get(ctxKeyToken); ok {
			// Try to extract GroupID from the token object.
			// It should be *token.View with a GroupID field > 0.
			type tokenWithGroup interface {
				GetGroupID() int64
			}
			if twg, ok := tv.(tokenWithGroup); ok && twg.GetGroupID() > 0 {
				c.Set(ctxKeyGroupID, twg.GetGroupID())
			}
		}

		// 2. Look up ratio
		gid := groupIDFromContext(c)
		if gid > 0 && lookup != nil {
			if ratio, err := lookup(c.Request.Context(), gid); err == nil {
				c.Set(CtxKeyGroupRatio, ratio)
				c.Set(CtxKeyBillingGroupRatio, ratio)
			}
		}

		c.Next()
	}
}

// groupIDFromContext extracts the group id from the gin context.
// This duplicates token.GroupIDFromContext to avoid circular imports.
func groupIDFromContext(c *gin.Context) int64 {
	if v, ok := c.Get(ctxKeyGroupID); ok {
		if id, ok := v.(int64); ok {
			return id
		}
	}
	return 0
}
