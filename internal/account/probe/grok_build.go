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

const grokBuildDefaultBase = "https://api.x.ai"

// GrokBuild probes xAI official API credentials using GET /v1/models.
type GrokBuild struct {
	base   string
	client *http.Client
}

func NewGrokBuild(base string) *GrokBuild {
	if base == "" {
		base = grokBuildDefaultBase
	}
	return &GrokBuild{base: strings.TrimRight(base, "/"), client: &http.Client{Timeout: 3 * time.Second}}
}

func (g *GrokBuild) Probe(ctx context.Context, cred account.AccountCred) (http.Header, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, g.base+"/v1/models", nil)
	if err != nil {
		return nil, err
	}
	tok := cred.APIKey
	if tok == "" {
		tok = cred.AccessToken
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode >= 400 {
		return resp.Header, fmt.Errorf("grok-build probe: status %d", resp.StatusCode)
	}
	return resp.Header, nil
}
