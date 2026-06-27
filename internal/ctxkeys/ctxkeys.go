// Package ctxkeys defines typed keys for values stored on a raw
// context.Context. Using a dedicated key type (rather than a built-in string)
// avoids cross-package collisions and satisfies staticcheck's SA1029.
//
// These keys are shared across packages because values flow between them:
// token.TokenAuth writes Token/UserID, middleware.GroupRatioMiddleware writes
// GroupID, and token.*FromContext reads all three back. Gin's own
// c.Set/c.Get use string keys (see the CtxKey* constants in the token and
// middleware packages, derived from these) and are intentionally separate.
package ctxkeys

// Key is the type for pro-api context keys.
type Key string

// String returns the underlying string, for the rare caller that needs the
// raw value (e.g. logging).
func (k Key) String() string { return string(k) }

const (
	// Token holds the *token.View for the current request.
	Token Key = "proapi:token"
	// UserID holds the authenticated user id (int64).
	UserID Key = "proapi:user_id"
	// GroupID holds the token's explicit group id (int64).
	GroupID Key = "proapi:group_id"
)
