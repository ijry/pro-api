// Package feishu 提供飞书 OAuth provider 实现。
//
// 流程:BuildAuthURL → (用户在飞书同意)→ Exchange 用 code 换 access_token,
// 再拉 /authen/v1/user_info 获取用户信息。
package feishu

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
	AppID       string
	AppSecret   string
	HTTPTimeout time.Duration // 默认 10s
}

// provider 飞书 OAuth 实现。
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
	return "feishu"
}

// BuildAuthURL 拼装飞书 OAuth 授权 URL。
func (p *provider) BuildAuthURL(_ context.Context, state, redirectURL string) (string, error) {
	if p.cfg.AppID == "" {
		return "", apierr.New(apierr.CodeForbidden, "飞书登录未启用")
	}
	u, err := url.Parse("https://open.feishu.cn/open-apis/authen/v1/authorize")
	if err != nil {
		return "", fmt.Errorf("feishu oauth: parse auth url: %w", err)
	}
	q := u.Query()
	q.Set("app_id", p.cfg.AppID)
	q.Set("redirect_uri", redirectURL)
	q.Set("scope", "contact:user.base:readonly")
	q.Set("state", state)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

type tokenRequest struct {
	GrantType string `json:"grant_type"`
	Code      string `json:"code"`
}

type tokenData struct {
	AccessToken string `json:"access_token"`
	OpenID      string `json:"open_id"`
	UnionID     string `json:"union_id"`
}

type tokenResponse struct {
	Code int       `json:"code"`
	Msg  string    `json:"msg"`
	Data tokenData `json:"data"`
}

type avatarInfo struct {
	Avatar72 string `json:"avatar_72"`
}

type userInfoData struct {
	Name            string     `json:"name"`
	EnterpriseEmail string     `json:"enterprise_email"`
	Avatar          avatarInfo `json:"avatar"`
}

type userInfoResponse struct {
	Code int          `json:"code"`
	Msg  string       `json:"msg"`
	Data userInfoData `json:"data"`
}

// Exchange 用 code 换 access_token,然后调 /authen/v1/user_info。
func (p *provider) Exchange(ctx context.Context, code, redirectURL string) (*oauth.UserInfo, string, error) {
	if p.cfg.AppID == "" {
		return nil, "", apierr.New(apierr.CodeForbidden, "飞书登录未启用")
	}

	// Step 1: 换取 access_token
	tok, err := p.fetchToken(ctx, code)
	if err != nil {
		return nil, "", err
	}

	// Step 2: 拉取用户信息
	userRaw, u, err := p.fetchUserInfo(ctx, tok.Data.AccessToken)
	if err != nil {
		return nil, "", err
	}

	// 优先使用 union_id 作为 ProviderUID,否则回退到 open_id
	uid := tok.Data.UnionID
	if uid == "" {
		uid = tok.Data.OpenID
	}

	return &oauth.UserInfo{
		ProviderUID: uid,
		Name:        u.Data.Name,
		Email:       u.Data.EnterpriseEmail,
		Avatar:      u.Data.Avatar.Avatar72,
		Raw:         userRaw,
	}, tok.Data.AccessToken, nil
}

func (p *provider) fetchToken(ctx context.Context, code string) (*tokenResponse, error) {
	body, err := json.Marshal(tokenRequest{GrantType: "authorization_code", Code: code})
	if err != nil {
		return nil, apierr.Wrap(apierr.CodeInternal, "feishu token req encode", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://open.feishu.cn/open-apis/authen/v1/oidc/access_token",
		bytes.NewReader(body))
	if err != nil {
		return nil, apierr.Wrap(apierr.CodeInternal, "feishu token req build", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(p.cfg.AppID, p.cfg.AppSecret)

	resp, err := p.http.Do(req)
	if err != nil {
		return nil, apierr.Wrap(apierr.CodeUpstreamUnavail, "feishu token http", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	var tok tokenResponse
	if err := json.Unmarshal(respBody, &tok); err != nil {
		return nil, apierr.Wrap(apierr.CodeUpstreamError, "feishu token decode", err)
	}
	if tok.Code != 0 || tok.Data.AccessToken == "" {
		return nil, apierr.New(apierr.CodeUpstreamError, fmt.Sprintf("feishu token error %d: %s", tok.Code, tok.Msg))
	}
	return &tok, nil
}

func (p *provider) fetchUserInfo(ctx context.Context, accessToken string) ([]byte, *userInfoResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://open.feishu.cn/open-apis/authen/v1/user_info", nil)
	if err != nil {
		return nil, nil, apierr.Wrap(apierr.CodeInternal, "feishu userinfo req", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := p.http.Do(req)
	if err != nil {
		return nil, nil, apierr.Wrap(apierr.CodeUpstreamUnavail, "feishu userinfo http", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return nil, nil, apierr.New(apierr.CodeUpstreamError, fmt.Sprintf("feishu /user_info http %d", resp.StatusCode))
	}
	body, _ := io.ReadAll(resp.Body)
	var u userInfoResponse
	if err := json.Unmarshal(body, &u); err != nil {
		return nil, nil, apierr.Wrap(apierr.CodeUpstreamError, "feishu userinfo decode", err)
	}
	if u.Code != 0 {
		return nil, nil, apierr.New(apierr.CodeUpstreamError, fmt.Sprintf("feishu userinfo error %d: %s", u.Code, u.Msg))
	}
	return body, &u, nil
}
