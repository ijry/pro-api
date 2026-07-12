package probe

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ijry/pro-api/internal/account"
)

// anthropicDefaultBase 与 adapter/anthropic 的 defaultBaseURL 一致,
// 用于 base 未配置时兜底,避免请求打到相对路径 "/v1/..." 直接失败。
const anthropicDefaultBase = "https://api.anthropic.com"

// Anthropic 调用 /v1/messages/count_tokens 做轻量探测,仅取响应头。
type Anthropic struct {
	base   string
	client *http.Client
}

func NewAnthropic(base string) *Anthropic {
	if base == "" {
		base = anthropicDefaultBase
	}
	return &Anthropic{base: base, client: &http.Client{Timeout: 3 * time.Second}}
}

func (a *Anthropic) Probe(ctx context.Context, cred account.AccountCred) (http.Header, error) {
	body := strings.NewReader(`{"model":"claude-3-5-haiku-20241022","messages":[{"role":"user","content":"ping"}]}`)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.base+"/v1/messages/count_tokens", body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("anthropic-version", "2023-06-01")
	// 鉴权头必须与真实转发一致(见 adapter/anthropic:x-api-key),否则会把有效的
	// api_key 账号探成 401 并误标记失效。api_key 账号走 x-api-key;OAuth 订阅账号
	// (只有 access_token)走 Authorization: Bearer + oauth-beta 头。
	if cred.APIKey != "" {
		req.Header.Set("x-api-key", cred.APIKey)
	} else {
		req.Header.Set("Authorization", "Bearer "+cred.AccessToken)
		req.Header.Set("anthropic-beta", "oauth-2025-04-20")
	}
	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode >= 400 {
		return resp.Header, fmt.Errorf("anthropic probe: status %d", resp.StatusCode)
	}
	return resp.Header, nil
}
