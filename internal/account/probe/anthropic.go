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

// Anthropic 调用 /v1/messages/count_tokens 做轻量探测,仅取响应头。
type Anthropic struct {
	base   string
	client *http.Client
}

func NewAnthropic(base string) *Anthropic {
	return &Anthropic{base: base, client: &http.Client{Timeout: 3 * time.Second}}
}

func (a *Anthropic) Probe(ctx context.Context, cred account.AccountCred) (http.Header, error) {
	body := strings.NewReader(`{"model":"claude-3-5-haiku-20241022","messages":[{"role":"user","content":"ping"}]}`)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.base+"/v1/messages/count_tokens", body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+cred.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("anthropic-version", "2023-06-01")
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
