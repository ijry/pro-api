package middleware

const (
	// CtxKeyUserID 是注入到 gin.Context 的 user id key。
	CtxKeyUserID = "proapi:user_id"
	// CtxKeyRole 是当前会话的 user role。
	CtxKeyRole = "proapi:role"
	// CtxKeySessionID 是当前 session id。
	CtxKeySessionID = "proapi:session_id"

	// CookieSession 是 session cookie 名。
	CookieSession = "proapi_session"
	// CookieCSRF 是 CSRF cookie 名(非 HttpOnly,前端可读)。
	CookieCSRF = "proapi_csrf"
	// HeaderSession 是 fallback 的 session header。
	HeaderSession = "X-Session-Token"
	// HeaderCSRF 是 CSRF 校验 header。
	HeaderCSRF = "X-CSRF-Token"
)
