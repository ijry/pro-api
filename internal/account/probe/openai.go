package probe

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/ijry/pro-api/internal/account"
)

// OpenAI 调用 GET /v1/models 做轻量探测,取响应头(主要是 x-ratelimit-*)。
type OpenAI struct {
	base   string
	client *http.Client
}

func NewOpenAI(base string) *OpenAI {
	return &OpenAI{base: base, client: &http.Client{Timeout: 3 * time.Second}}
}

func (o *OpenAI) Probe(ctx context.Context, cred account.AccountCred) (http.Header, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, o.base+"/v1/models", nil)
	if err != nil {
		return nil, err
	}
	tok := cred.AccessToken
	if tok == "" {
		tok = cred.APIKey
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	return resp.Header, nil
}
