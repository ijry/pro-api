package oauth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"net/url"
	"strings"
)

// newVerifier 生成 RFC 7636 PKCE code_verifier(32 字节随机 → base64url 无填充,43 字符)。
func newVerifier() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// challengeS256 由 code_verifier 计算 S256 code_challenge。
func challengeS256(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

// newState 生成防 CSRF 的随机 state(16 字节 → base64url)。
func newState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// authCodeURL 按 OAuth2 授权码 + PKCE(S256)拼出授权跳转 URL。
// 各 provider 的 AuthCodeURL 共用此实现。
func authCodeURL(cfg Config, state, challenge string) string {
	q := url.Values{}
	q.Set("response_type", "code")
	q.Set("client_id", cfg.ClientID)
	if cfg.RedirectURI != "" {
		q.Set("redirect_uri", cfg.RedirectURI)
	}
	if len(cfg.Scopes) > 0 {
		q.Set("scope", strings.Join(cfg.Scopes, " "))
	}
	q.Set("state", state)
	q.Set("code_challenge", challenge)
	q.Set("code_challenge_method", "S256")
	return cfg.AuthURL + "?" + q.Encode()
}
