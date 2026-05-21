package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ijry/pro-api/pkg/apierr"
)

// UserInfo 是 GitHub /user + /user/emails 拉到的归一化数据。
type UserInfo struct {
	ID     int64  // GitHub numeric id(stable)
	Login  string // username
	Name   string
	Email  string // primary verified email
	Avatar string
	Raw    []byte // /user 原始 JSON
}

// Provider 是 GitHub OAuth client 抽象。
type Provider interface {
	BuildAuthURL(ctx context.Context, state, redirectURL string) (string, error)
	Exchange(ctx context.Context, code, redirectURL string) (*UserInfo, string /*access_token*/, error)
}

// Config 是 New 的参数。
type Config struct {
	ClientID     string
	ClientSecret string
	Scopes       []string      // 默认 ["read:user", "user:email"]
	BaseURL      string        // 默认 "https://github.com"
	APIBaseURL   string        // 默认 "https://api.github.com"
	HTTPTimeout  time.Duration // 默认 10s
}

const (
	defaultBaseURL    = "https://github.com"
	defaultAPIBaseURL = "https://api.github.com"
)

var defaultScopes = []string{"read:user", "user:email"}

// provider 默认实现。
type provider struct {
	cfg  Config
	http *http.Client
}

// New 构造 Provider。
func New(cfg Config) Provider {
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultBaseURL
	}
	if cfg.APIBaseURL == "" {
		cfg.APIBaseURL = defaultAPIBaseURL
	}
	if cfg.HTTPTimeout <= 0 {
		cfg.HTTPTimeout = 10 * time.Second
	}
	if len(cfg.Scopes) == 0 {
		cfg.Scopes = defaultScopes
	}
	return &provider{cfg: cfg, http: &http.Client{Timeout: cfg.HTTPTimeout}}
}

// BuildAuthURL 拼装 OAuth 授权 URL。
func (p *provider) BuildAuthURL(_ context.Context, state, redirectURL string) (string, error) {
	if p.cfg.ClientID == "" {
		return "", apierr.New(apierr.CodeForbidden, "GitHub 登录未启用")
	}
	u, err := url.Parse(p.cfg.BaseURL + "/login/oauth/authorize")
	if err != nil {
		return "", fmt.Errorf("github oauth: parse base: %w", err)
	}
	q := u.Query()
	q.Set("client_id", p.cfg.ClientID)
	q.Set("redirect_uri", redirectURL)
	q.Set("scope", strings.Join(p.cfg.Scopes, " "))
	q.Set("state", state)
	q.Set("allow_signup", "true")
	u.RawQuery = q.Encode()
	return u.String(), nil
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	Error       string `json:"error"`
}

type userResponse struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

type emailEntry struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

// Exchange 用 code 换 access_token,然后调 /user + /user/emails。
func (p *provider) Exchange(ctx context.Context, code, redirectURL string) (*UserInfo, string, error) {
	if p.cfg.ClientID == "" {
		return nil, "", apierr.New(apierr.CodeForbidden, "GitHub 登录未启用")
	}
	form := url.Values{}
	form.Set("client_id", p.cfg.ClientID)
	form.Set("client_secret", p.cfg.ClientSecret)
	form.Set("code", code)
	if redirectURL != "" {
		form.Set("redirect_uri", redirectURL)
	}
	tokenURL := p.cfg.BaseURL + "/login/oauth/access_token"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, "", apierr.Wrap(apierr.CodeInternal, "github token req build", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := p.http.Do(req)
	if err != nil {
		return nil, "", apierr.Wrap(apierr.CodeUpstreamUnavail, "github token http", err)
	}
	defer resp.Body.Close()
	tokBody, _ := io.ReadAll(resp.Body)
	var tok tokenResponse
	if err := json.Unmarshal(tokBody, &tok); err != nil {
		return nil, "", apierr.Wrap(apierr.CodeUpstreamError, "github token decode", err)
	}
	if tok.Error != "" || tok.AccessToken == "" {
		return nil, "", apierr.New(apierr.CodeUpstreamError, fmt.Sprintf("github token error: %s", tok.Error))
	}

	// 拉 /user
	userRaw, user, err := p.fetchUser(ctx, tok.AccessToken)
	if err != nil {
		return nil, "", err
	}
	email := strings.TrimSpace(user.Email)
	if email == "" {
		// 兜底拉 /user/emails
		emails, err := p.fetchUserEmails(ctx, tok.AccessToken)
		if err == nil {
			for _, e := range emails {
				if e.Primary && e.Verified {
					email = e.Email
					break
				}
			}
			// 若没 primary,取任意 verified
			if email == "" {
				for _, e := range emails {
					if e.Verified {
						email = e.Email
						break
					}
				}
			}
		}
	}
	return &UserInfo{
		ID:     user.ID,
		Login:  user.Login,
		Name:   user.Name,
		Email:  email,
		Avatar: user.AvatarURL,
		Raw:    userRaw,
	}, tok.AccessToken, nil
}

func (p *provider) fetchUser(ctx context.Context, accessToken string) ([]byte, *userResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.cfg.APIBaseURL+"/user", nil)
	if err != nil {
		return nil, nil, apierr.Wrap(apierr.CodeInternal, "github user req", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := p.http.Do(req)
	if err != nil {
		return nil, nil, apierr.Wrap(apierr.CodeUpstreamUnavail, "github user http", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return nil, nil, apierr.New(apierr.CodeUpstreamError, fmt.Sprintf("github /user http %d", resp.StatusCode))
	}
	body, _ := io.ReadAll(resp.Body)
	var u userResponse
	if err := json.Unmarshal(body, &u); err != nil {
		return nil, nil, apierr.Wrap(apierr.CodeUpstreamError, "github user decode", err)
	}
	return body, &u, nil
}

func (p *provider) fetchUserEmails(ctx context.Context, accessToken string) ([]emailEntry, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.cfg.APIBaseURL+"/user/emails", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := p.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("github /user/emails http %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	var emails []emailEntry
	if err := json.Unmarshal(body, &emails); err != nil {
		return nil, err
	}
	return emails, nil
}
