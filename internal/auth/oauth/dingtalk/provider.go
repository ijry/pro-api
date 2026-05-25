// Package dingtalk 提供钉钉 OAuth provider 实现。
//
// 流程:BuildAuthURL → (用户在钉钉同意)→ Exchange 用 code 换 access_token,
// 再拉 /v1.0/contact/users/me 获取用户信息。
package dingtalk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/ijry/pro-api/internal/auth/oauth"
	"github.com/ijry/pro-api/pkg/apierr"
)

// Config 是 New 的参数。
type Config struct {
	ClientID     string
	ClientSecret string
	HTTPTimeout  time.Duration // 默认 10s
}

// provider 钉钉 OAuth 实现。
type provider struct {
	cfg  Config
	http *http.Client
}

// New 构造 Provider。
func New(cfg Config) oauth.Provider {
	if cfg.HTTPTimeout <= 0 {
		cfg.HTTPTimeout = 10 * time.Second
	}
	return &provider{cfg: cfg, http: &http.Client{Timeout: cfg.HTTPTimeout}}
}

// Name 返回 provider 标识符。
func (p *provider) Name() string {
	return "dingtalk"
}

// BuildAuthURL 拼装钉钉 OAuth 授权 URL。
func (p *provider) BuildAuthURL(_ context.Context, state, redirectURL string) (string, error) {
	if p.cfg.ClientID == "" {
		return "", apierr.New(apierr.CodeForbidden, "钉钉登录未启用")
	}
	u, err := url.Parse("https://login.dingtalk.com/oauth2/auth")
	if err != nil {
		return "", fmt.Errorf("dingtalk oauth: parse auth url: %w", err)
	}
	q := u.Query()
	q.Set("client_id", p.cfg.ClientID)
	q.Set("redirect_uri", redirectURL)
	q.Set("response_type", "code")
	q.Set("scope", "openid")
	q.Set("state", state)
	q.Set("prompt", "consent")
	u.RawQuery = q.Encode()
	return u.String(), nil
}

type tokenRequest struct {
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	Code         string `json:"code"`
	RedirectURI  string `json:"redirectUri"`
	GrantType    string `json:"grantType"`
}

type tokenResponse struct {
	AccessToken string `json:"accessToken"`
	Code        string `json:"code"`
	Message     string `json:"message"`
}

type userResponse struct {
	OpenID    string `json:"openId"`
	UnionID   string `json:"unionId"`
	Nick      string `json:"nick"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatarUrl"`
}

// Exchange 用 code 换 access_token,然后调 /v1.0/contact/users/me。
func (p *provider) Exchange(ctx context.Context, code, redirectURL string) (*oauth.UserInfo, string, error) {
	if p.cfg.ClientID == "" {
		return nil, "", apierr.New(apierr.CodeForbidden, "钉钉登录未启用")
	}

	// Step 1: 换取 access_token
	tok, err := p.fetchToken(ctx, code, redirectURL)
	if err != nil {
		return nil, "", err
	}

	// Step 2: 拉取用户信息
	userRaw, u, err := p.fetchUserInfo(ctx, tok.AccessToken)
	if err != nil {
		return nil, "", err
	}

	// 优先使用 unionId 作为 ProviderUID,否则回退到 openId
	uid := u.UnionID
	if uid == "" {
		uid = u.OpenID
	}

	return &oauth.UserInfo{
		ProviderUID: uid,
		Name:        u.Nick,
		Email:       u.Email,
		Avatar:      u.AvatarURL,
		Raw:         userRaw,
	}, tok.AccessToken, nil
}

func (p *provider) fetchToken(ctx context.Context, code, redirectURL string) (*tokenResponse, error) {
	body, err := json.Marshal(tokenRequest{
		ClientID:     p.cfg.ClientID,
		ClientSecret: p.cfg.ClientSecret,
		Code:         code,
		RedirectURI:  redirectURL,
		GrantType:    "authorization_code",
	})
	if err != nil {
		return nil, apierr.Wrap(apierr.CodeInternal, "dingtalk token req encode", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.dingtalk.com/v1.0/oauth2/userAccessToken",
		bytes.NewReader(body))
	if err != nil {
		return nil, apierr.Wrap(apierr.CodeInternal, "dingtalk token req build", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.http.Do(req)
	if err != nil {
		return nil, apierr.Wrap(apierr.CodeUpstreamUnavail, "dingtalk token http", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	var tok tokenResponse
	if err := json.Unmarshal(respBody, &tok); err != nil {
		return nil, apierr.Wrap(apierr.CodeUpstreamError, "dingtalk token decode", err)
	}
	if tok.AccessToken == "" {
		return nil, apierr.New(apierr.CodeUpstreamError, fmt.Sprintf("dingtalk token error: %s %s", tok.Code, tok.Message))
	}
	return &tok, nil
}

func (p *provider) fetchUserInfo(ctx context.Context, accessToken string) ([]byte, *userResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://api.dingtalk.com/v1.0/contact/users/me", nil)
	if err != nil {
		return nil, nil, apierr.Wrap(apierr.CodeInternal, "dingtalk userinfo req", err)
	}
	req.Header.Set("x-acs-dingtalk-access-token", accessToken)

	resp, err := p.http.Do(req)
	if err != nil {
		return nil, nil, apierr.Wrap(apierr.CodeUpstreamUnavail, "dingtalk userinfo http", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return nil, nil, apierr.New(apierr.CodeUpstreamError, fmt.Sprintf("dingtalk /users/me http %d", resp.StatusCode))
	}
	body, _ := io.ReadAll(resp.Body)
	var u userResponse
	if err := json.Unmarshal(body, &u); err != nil {
		return nil, nil, apierr.Wrap(apierr.CodeUpstreamError, "dingtalk userinfo decode", err)
	}
	return body, &u, nil
}
