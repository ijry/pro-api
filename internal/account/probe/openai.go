package probe

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ijry/pro-api/internal/account"
)

// openaiDefaultBase 与 adapter/openai 的 defaultBaseURL 一致,
// 用于 base 未配置时兜底,避免请求打到相对路径 "/v1/..." 直接失败。
const openaiDefaultBase = "https://api.openai.com"

// OpenAI 调用 GET /v1/models 做轻量探测,取响应头(主要是 x-ratelimit-*)。
type OpenAI struct {
	base   string
	client *http.Client
}

func NewOpenAI(base string) *OpenAI {
	if base == "" {
		base = openaiDefaultBase
	}
	return &OpenAI{base: base, client: &http.Client{Timeout: 3 * time.Second}}
}

func (o *OpenAI) Probe(ctx context.Context, cred account.AccountCred) (http.Header, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, o.base+"/v1/models", nil)
	if err != nil {
		return nil, err
	}
	// 选凭证优先级必须与真实转发(relay.resolveCred)一致:APIKey 优先,其次
	// AccessToken。否则会出现"探测测 access_token、转发用 api_key"的错配 ——
	// 一个 api_key 有效但 access_token 过期的 Codex 账号会被误标记为 Expired。
	tok := cred.APIKey
	if tok == "" {
		tok = cred.AccessToken
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode >= 400 {
		return resp.Header, fmt.Errorf("openai probe: status %d", resp.StatusCode)
	}
	return resp.Header, nil
}
