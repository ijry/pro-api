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

// Anthropic 实现 Provider,使用 RFC 6749 refresh_token grant。
type Anthropic struct {
	cfg    Config
	client *http.Client
}

func NewAnthropic(cfg Config) *Anthropic {
	return &Anthropic{cfg: cfg, client: &http.Client{Timeout: 15 * time.Second}}
}

func (a *Anthropic) Start(_ context.Context, _ int64) (string, string, error) {
	return "", "", ErrNotImplemented
}

func (a *Anthropic) Callback(_ context.Context, _, _ string) (*account.Account, error) {
	return nil, ErrNotImplemented
}

func (a *Anthropic) ExchangeRefreshToken(ctx context.Context, refreshToken string) (*account.AccountCred, error) {
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)
	form.Set("client_id", a.cfg.ClientID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.cfg.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("anthropic token exchange: status %d", resp.StatusCode)
	}
	var body struct {
		AT        string `json:"access_token"`
		RT        string `json:"refresh_token"`
		ExpiresIn int64  `json:"expires_in"`
		TokenType string `json:"token_type"`
		Scope     string `json:"scope"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	return &account.AccountCred{
		AccessToken:  body.AT,
		RefreshToken: body.RT,
		TokenType:    body.TokenType,
		Scope:        body.Scope,
		ExpiresAt:    time.Now().UTC().Add(time.Duration(body.ExpiresIn) * time.Second),
	}, nil
}
