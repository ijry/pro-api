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

// OpenAI 实现 Provider。除 access_token / refresh_token 外,还回填 id_token(OpenAI 特有)。
type OpenAI struct {
	cfg    Config
	client *http.Client
}

// NewOpenAI 构造 OpenAI OAuth 客户端。
func NewOpenAI(cfg Config) *OpenAI {
	return &OpenAI{cfg: cfg, client: &http.Client{Timeout: 15 * time.Second}}
}

// AuthCodeURL 拼出 OpenAI 授权码 + PKCE 授权跳转 URL。
func (o *OpenAI) AuthCodeURL(state, challenge string) string {
	return authCodeURL(o.cfg, state, challenge)
}

// ExchangeCode 用授权码 + code_verifier 换取凭证(grant_type=authorization_code)。
func (o *OpenAI) ExchangeCode(ctx context.Context, code, verifier string) (*account.AccountCred, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", o.cfg.RedirectURI)
	form.Set("client_id", o.cfg.ClientID)
	form.Set("code_verifier", verifier)
	return o.token(ctx, form)
}

// ExchangeRefreshToken 用 refresh_token 换取新凭证(grant_type=refresh_token)。
func (o *OpenAI) ExchangeRefreshToken(ctx context.Context, refreshToken string) (*account.AccountCred, error) {
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)
	form.Set("client_id", o.cfg.ClientID)
	return o.token(ctx, form)
}

func (o *OpenAI) token(ctx context.Context, form url.Values) (*account.AccountCred, error) {
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
