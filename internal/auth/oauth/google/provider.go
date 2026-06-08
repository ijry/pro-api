// Package google 提供 Google OAuth2 provider 实现。
//
// 流程:BuildAuthURL → (用户在 Google 同意)→ Exchange 用 code 换 token + 拉
// /oauth2/v2/userinfo。
package google

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ijry/pro-api/internal/auth/oauth"
	"github.com/ijry/pro-api/pkg/apierr"
)

// Config 是 New 的参数。
type Config struct {
	ClientID     string
	ClientSecret string
	Scopes       []string      // 默认 ["openid","email","profile"]
	HTTPTimeout  time.Duration // 默认 10s
}

var defaultScopes = []string{"openid", "email", "profile"}

// provider 默认实现。
type provider struct {
	cfg  Config
	http *http.Client
}

// New 构造 Provider。
func New(cfg Config) oauth.Provider {
	if cfg.HTTPTimeout <= 0 {
		cfg.HTTPTimeout = 10 * time.Second
	}
	if len(cfg.Scopes) == 0 {
		cfg.Scopes = defaultScopes
	}
	return &provider{cfg: cfg, http: &http.Client{Timeout: cfg.HTTPTimeout}}
}

// Name 返回 provider 标识符。
func (p *provider) Name() string {
	return "google"
}

// BuildAuthURL 拼装 Google OAuth 授权 URL。
func (p *provider) BuildAuthURL(_ context.Context, state, redirectURL string) (string, error) {
	if p.cfg.ClientID == "" {
		return "", apierr.New(apierr.CodeForbidden, "Google 登录未启用")
	}
	u, err := url.Parse("https://accounts.google.com/o/oauth2/v2/auth")
	if err != nil {
		return "", fmt.Errorf("google oauth: parse auth url: %w", err)
	}
	q := u.Query()
	q.Set("client_id", p.cfg.ClientID)
	q.Set("redirect_uri", redirectURL)
	q.Set("response_type", "code")
	q.Set("scope", strings.Join(p.cfg.Scopes, " "))
	q.Set("state", state)
	q.Set("access_type", "offline")
	u.RawQuery = q.Encode()
	return u.String(), nil
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	IDToken     string `json:"id_token"`
	Error       string `json:"error"`
}

type userResponse struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Picture string `json:"picture"`
}

// Exchange 用 code 换 access_token,然后调 /oauth2/v2/userinfo。
func (p *provider) Exchange(ctx context.Context, code, redirectURL string) (*oauth.UserInfo, string, error) {
	if p.cfg.ClientID == "" {
		return nil, "", apierr.New(apierr.CodeForbidden, "Google 登录未启用")
	}
	form := url.Values{}
	form.Set("code", code)
	form.Set("client_id", p.cfg.ClientID)
	form.Set("client_secret", p.cfg.ClientSecret)
	form.Set("redirect_uri", redirectURL)
	form.Set("grant_type", "authorization_code")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://oauth2.googleapis.com/token", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, "", apierr.Wrap(apierr.CodeInternal, "google token req build", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := p.http.Do(req)
	if err != nil {
		return nil, "", apierr.Wrap(apierr.CodeUpstreamUnavail, "google token http", err)
	}
	defer func() { _ = resp.Body.Close() }()
	tokBody, _ := io.ReadAll(resp.Body)
	var tok tokenResponse
	if err := json.Unmarshal(tokBody, &tok); err != nil {
		return nil, "", apierr.Wrap(apierr.CodeUpstreamError, "google token decode", err)
	}
	if tok.Error != "" || tok.AccessToken == "" {
		return nil, "", apierr.New(apierr.CodeUpstreamError, fmt.Sprintf("google token error: %s", tok.Error))
	}

	// 拉 /oauth2/v2/userinfo
	userRaw, u, err := p.fetchUserInfo(ctx, tok.AccessToken)
	if err != nil {
		return nil, "", err
	}
	return &oauth.UserInfo{
		ProviderUID: u.ID,
		Name:        u.Name,
		Email:       u.Email,
		Avatar:      u.Picture,
		Raw:         userRaw,
	}, tok.AccessToken, nil
}

func (p *provider) fetchUserInfo(ctx context.Context, accessToken string) ([]byte, *userResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, nil, apierr.Wrap(apierr.CodeInternal, "google userinfo req", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := p.http.Do(req)
	if err != nil {
		return nil, nil, apierr.Wrap(apierr.CodeUpstreamUnavail, "google userinfo http", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode/100 != 2 {
		return nil, nil, apierr.New(apierr.CodeUpstreamError, fmt.Sprintf("google /userinfo http %d", resp.StatusCode))
	}
	body, _ := io.ReadAll(resp.Body)
	var u userResponse
	if err := json.Unmarshal(body, &u); err != nil {
		return nil, nil, apierr.Wrap(apierr.CodeUpstreamError, "google userinfo decode", err)
	}
	return body, &u, nil
}
