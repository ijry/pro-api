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

// Anthropic 实现 Provider,使用 RFC 6749 授权码 + PKCE 与 refresh_token grant。
type Anthropic struct {
	cfg    Config
	client *http.Client
}

// NewAnthropic 构造 Anthropic OAuth 客户端。
func NewAnthropic(cfg Config) *Anthropic {
	return &Anthropic{cfg: cfg, client: &http.Client{Timeout: 15 * time.Second}}
}

// AuthCodeURL 拼出 Anthropic 授权码 + PKCE 授权跳转 URL。
func (a *Anthropic) AuthCodeURL(state, challenge string) string {
	return authCodeURL(a.cfg, state, challenge)
}

// ExchangeCode 用授权码 + code_verifier 换取凭证(grant_type=authorization_code)。
func (a *Anthropic) ExchangeCode(ctx context.Context, code, verifier string) (*account.AccountCred, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", a.cfg.RedirectURI)
	form.Set("client_id", a.cfg.ClientID)
	form.Set("code_verifier", verifier)
	return a.token(ctx, form)
}

// ExchangeRefreshToken 用 refresh_token 换取新凭证(grant_type=refresh_token)。
func (a *Anthropic) ExchangeRefreshToken(ctx context.Context, refreshToken string) (*account.AccountCred, error) {
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)
	form.Set("client_id", a.cfg.ClientID)
	return a.token(ctx, form)
}

func (a *Anthropic) token(ctx context.Context, form url.Values) (*account.AccountCred, error) {
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
