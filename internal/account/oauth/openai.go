package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ijry/pro-api/internal/account"
)

// OpenAI 实现 Provider。除了 access_token / refresh_token,还回填 id_token(OpenAI 特有)。
type OpenAI struct {
	cfg    Config
	client *http.Client
}

func NewOpenAI(cfg Config) *OpenAI {
	return &OpenAI{cfg: cfg, client: &http.Client{Timeout: 15 * time.Second}}
}

func (o *OpenAI) Start(_ context.Context, _ int64) (string, string, error) {
	return "", "", ErrNotImplemented
}

func (o *OpenAI) Callback(_ context.Context, _, _ string) (*account.Account, error) {
	return nil, ErrNotImplemented
}

func (o *OpenAI) ExchangeRefreshToken(ctx context.Context, refreshToken string) (*account.AccountCred, error) {
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)
	form.Set("client_id", o.cfg.ClientID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.cfg.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("openai token exchange: status %d", resp.StatusCode)
	}
	var body struct {
		AT        string `json:"access_token"`
		RT        string `json:"refresh_token"`
		IDToken   string `json:"id_token"`
		ExpiresIn int64  `json:"expires_in"`
		TokenType string `json:"token_type"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	return &account.AccountCred{
		AccessToken:  body.AT,
		RefreshToken: body.RT,
		IDToken:      body.IDToken,
		TokenType:    body.TokenType,
		ExpiresAt:    time.Now().UTC().Add(time.Duration(body.ExpiresIn) * time.Second),
	}, nil
}
