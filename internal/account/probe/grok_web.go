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

const grokWebDefaultBase = "https://grok.com"

// GrokWeb probes Grok Web SSO credentials using GET /rest/rate-limits.
type GrokWeb struct {
	base   string
	client *http.Client
}

func NewGrokWeb(base string) *GrokWeb {
	if base == "" {
		base = grokWebDefaultBase
	}
	return &GrokWeb{base: strings.TrimRight(base, "/"), client: &http.Client{Timeout: 3 * time.Second}}
}

func (g *GrokWeb) Probe(ctx context.Context, cred account.AccountCred) (http.Header, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, g.base+"/rest/rate-limits", nil)
	if err != nil {
		return nil, err
	}
	tok := cred.APIKey
	if tok == "" {
		tok = cred.AccessToken
	}
	tok = strings.TrimPrefix(strings.TrimSpace(tok), "sso=")
	req.Header.Set("Cookie", "sso="+tok+"; sso-rw="+tok)
	req.Header.Set("Origin", "https://grok.com")
	req.Header.Set("Referer", "https://grok.com/")
	resp, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode >= 400 {
		return resp.Header, fmt.Errorf("grok-web probe: status %d", resp.StatusCode)
	}
	return resp.Header, nil
}
