// Package discord 提供 Discord OAuth provider 实现。
//
// 流程:BuildAuthURL → (用户在 Discord 同意)→ Exchange 用 code 换 access_token,
// 再拉 /api/users/@me 获取用户信息。
package discord

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
	HTTPTimeout  time.Duration // 默认 10s
}

// provider Discord OAuth 实现。
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
	return "discord"
}

// BuildAuthURL 拼装 Discord OAuth 授权 URL。
func (p *provider) BuildAuthURL(_ context.Context, state, redirectURL string) (string, error) {
	if p.cfg.ClientID == "" {
		return "", apierr.New(apierr.CodeForbidden, "Discord 登录未启用")
	}
	u, err := url.Parse("https://discord.com/oauth2/authorize")
	if err != nil {
		return "", fmt.Errorf("discord oauth: parse auth url: %w", err)
	}
	q := u.Query()
	q.Set("client_id", p.cfg.ClientID)
	q.Set("redirect_uri", redirectURL)
	q.Set("response_type", "code")
	q.Set("scope", "identify email")
	q.Set("state", state)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	Error       string `json:"error"`
}

type userResponse struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	Email         string `json:"email"`
	Avatar        string `json:"avatar"`
	Discriminator string `json:"discriminator"`
}

// Exchange 用 code 换 access_token,然后调 /api/users/@me。
func (p *provider) Exchange(ctx context.Context, code, redirectURL string) (*oauth.UserInfo, string, error) {
	if p.cfg.ClientID == "" {
		return nil, "", apierr.New(apierr.CodeForbidden, "Discord 登录未启用")
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

	// 构建 avatar URL
	avatarURL := ""
	if u.Avatar != "" {
		avatarURL = fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.png", u.ID, u.Avatar)
	}

	return &oauth.UserInfo{
		ProviderUID: u.ID,
		Name:        u.Username,
		Email:       u.Email,
		Avatar:      avatarURL,
		Raw:         userRaw,
	}, tok.AccessToken, nil
}

func (p *provider) fetchToken(ctx context.Context, code, redirectURL string) (*tokenResponse, error) {
	form := url.Values{}
	form.Set("client_id", p.cfg.ClientID)
	form.Set("client_secret", p.cfg.ClientSecret)
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", redirectURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://discord.com/api/oauth2/token",
		strings.NewReader(form.Encode()))
	if err != nil {
		return nil, apierr.Wrap(apierr.CodeInternal, "discord token req build", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.http.Do(req)
	if err != nil {
		return nil, apierr.Wrap(apierr.CodeUpstreamUnavail, "discord token http", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var tok tokenResponse
	if err := json.Unmarshal(body, &tok); err != nil {
		return nil, apierr.Wrap(apierr.CodeUpstreamError, "discord token decode", err)
	}
	if tok.Error != "" || tok.AccessToken == "" {
		return nil, apierr.New(apierr.CodeUpstreamError, fmt.Sprintf("discord token error: %s", tok.Error))
	}
	return &tok, nil
}

func (p *provider) fetchUserInfo(ctx context.Context, accessToken string) ([]byte, *userResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://discord.com/api/users/@me", nil)
	if err != nil {
		return nil, nil, apierr.Wrap(apierr.CodeInternal, "discord userinfo req", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := p.http.Do(req)
	if err != nil {
		return nil, nil, apierr.Wrap(apierr.CodeUpstreamUnavail, "discord userinfo http", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return nil, nil, apierr.New(apierr.CodeUpstreamError, fmt.Sprintf("discord /users/@me http %d", resp.StatusCode))
	}
	body, _ := io.ReadAll(resp.Body)
	var u userResponse
	if err := json.Unmarshal(body, &u); err != nil {
		return nil, nil, apierr.Wrap(apierr.CodeUpstreamError, "discord userinfo decode", err)
	}
	return body, &u, nil
}
