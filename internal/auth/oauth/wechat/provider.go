// Package wechat 提供微信网页授权/扫码登录 OAuth provider 实现。
//
// 流程:BuildAuthURL → (用户扫码授权)→ Exchange 用 code 换 access_token + openid/unionid,
// 再拉 /sns/userinfo 获取昵称与头像。
package wechat

import (
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
	AppID       string
	AppSecret   string
	HTTPTimeout time.Duration // 默认 10s
}

// provider 微信 OAuth 实现。
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
	return "wechat"
}

// BuildAuthURL 拼装微信扫码登录授权 URL。
// 格式: https://open.weixin.qq.com/connect/qrconnect?appid=...&...#wechat_redirect
func (p *provider) BuildAuthURL(_ context.Context, state, redirectURL string) (string, error) {
	if p.cfg.AppID == "" {
		return "", apierr.New(apierr.CodeForbidden, "微信登录未启用")
	}
	u, err := url.Parse("https://open.weixin.qq.com/connect/qrconnect")
	if err != nil {
		return "", fmt.Errorf("wechat oauth: parse auth url: %w", err)
	}
	q := u.Query()
	q.Set("appid", p.cfg.AppID)
	q.Set("redirect_uri", redirectURL)
	q.Set("response_type", "code")
	q.Set("scope", "snsapi_login")
	q.Set("state", state)
	u.RawQuery = q.Encode()
	return u.String() + "#wechat_redirect", nil
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	OpenID      string `json:"openid"`
	UnionID     string `json:"unionid"`
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
}

type userResponse struct {
	Nickname   string `json:"nickname"`
	HeadImgURL string `json:"headimgurl"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// Exchange 用 code 换 access_token + openid,再拉用户信息。
func (p *provider) Exchange(ctx context.Context, code, redirectURL string) (*oauth.UserInfo, string, error) {
	if p.cfg.AppID == "" {
		return nil, "", apierr.New(apierr.CodeForbidden, "微信登录未启用")
	}

	// Step 1: 换取 access_token
	tok, err := p.fetchToken(ctx, code)
	if err != nil {
		return nil, "", err
	}

	// Step 2: 拉取用户信息
	userRaw, u, err := p.fetchUserInfo(ctx, tok.AccessToken, tok.OpenID)
	if err != nil {
		return nil, "", err
	}

	// 优先使用 unionid 作为 ProviderUID,否则回退到 openid
	uid := tok.UnionID
	if uid == "" {
		uid = tok.OpenID
	}

	return &oauth.UserInfo{
		ProviderUID: uid,
		Name:        u.Nickname,
		Email:       "", // 微信不返回 email
		Avatar:      u.HeadImgURL,
		Raw:         userRaw,
	}, tok.AccessToken, nil
}

func (p *provider) fetchToken(ctx context.Context, code string) (*tokenResponse, error) {
	params := url.Values{}
	params.Set("appid", p.cfg.AppID)
	params.Set("secret", p.cfg.AppSecret)
	params.Set("code", code)
	params.Set("grant_type", "authorization_code")

	apiURL := "https://api.weixin.qq.com/sns/oauth2/access_token?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, apierr.Wrap(apierr.CodeInternal, "wechat token req build", err)
	}
	resp, err := p.http.Do(req)
	if err != nil {
		return nil, apierr.Wrap(apierr.CodeUpstreamUnavail, "wechat token http", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	var tok tokenResponse
	if err := json.Unmarshal(body, &tok); err != nil {
		return nil, apierr.Wrap(apierr.CodeUpstreamError, "wechat token decode", err)
	}
	if tok.ErrCode != 0 || tok.AccessToken == "" {
		return nil, apierr.New(apierr.CodeUpstreamError, fmt.Sprintf("wechat token error %d: %s", tok.ErrCode, tok.ErrMsg))
	}
	return &tok, nil
}

func (p *provider) fetchUserInfo(ctx context.Context, accessToken, openID string) ([]byte, *userResponse, error) {
	params := url.Values{}
	params.Set("access_token", accessToken)
	params.Set("openid", openID)
	params.Set("lang", "zh_CN")

	apiURL := "https://api.weixin.qq.com/sns/userinfo?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, nil, apierr.Wrap(apierr.CodeInternal, "wechat userinfo req", err)
	}
	resp, err := p.http.Do(req)
	if err != nil {
		return nil, nil, apierr.Wrap(apierr.CodeUpstreamUnavail, "wechat userinfo http", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode/100 != 2 {
		return nil, nil, apierr.New(apierr.CodeUpstreamError, fmt.Sprintf("wechat /userinfo http %d", resp.StatusCode))
	}
	body, _ := io.ReadAll(resp.Body)
	var u userResponse
	if err := json.Unmarshal(body, &u); err != nil {
		return nil, nil, apierr.Wrap(apierr.CodeUpstreamError, "wechat userinfo decode", err)
	}
	if u.ErrCode != 0 {
		return nil, nil, apierr.New(apierr.CodeUpstreamError, fmt.Sprintf("wechat userinfo error %d: %s", u.ErrCode, u.ErrMsg))
	}
	return body, &u, nil
}
