package zhipu

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// GenerateJWT 生成智谱 AI 所需的 JWT token。
//
// API Key 格式："id.secret"，用 "." 分割。
// JWT header: {"alg":"HS256","typ":"JWT","sign_type":"SIGN"}
// JWT payload: {"api_key": id, "exp": unix_ms, "timestamp": unix_ms}
// 签名：HMAC-SHA256(base64url(header).base64url(payload), secret)
func GenerateJWT(apiKey string, ttl time.Duration) (string, error) {
	parts := strings.SplitN(apiKey, ".", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("zhipu: invalid api_key format, expected 'id.secret'")
	}
	id, secret := parts[0], parts[1]

	now := time.Now()
	expMs := now.Add(ttl).UnixMilli()
	tsMs := now.UnixMilli()

	header := map[string]string{
		"alg":       "HS256",
		"typ":       "JWT",
		"sign_type": "SIGN",
	}
	payload := map[string]any{
		"api_key":   id,
		"exp":       expMs,
		"timestamp": tsMs,
	}

	headerBytes, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	headerEncoded := base64.RawURLEncoding.EncodeToString(headerBytes)
	payloadEncoded := base64.RawURLEncoding.EncodeToString(payloadBytes)
	signingInput := headerEncoded + "." + payloadEncoded

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signingInput))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return signingInput + "." + sig, nil
}
