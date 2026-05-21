// Package crypto 提供应用层加密原语,目前只有 AES-256-GCM。
//
// 加密输出格式:ENC(v1,nonce_b64,ciphertext_b64)
// 未来轮换版本时可引入 ENC(v2,...) 并保留对 v1 的解密能力。
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
)

// AESGCM 是一个无状态、并发安全的加密器。
type AESGCM struct {
	aead cipher.AEAD
}

// NewAESGCM 用 32 字节 key 构造加密器。
func NewAESGCM(key []byte) (*AESGCM, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("crypto: key must be 32 bytes (AES-256), got %d", len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("crypto: aes.NewCipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("crypto: cipher.NewGCM: %w", err)
	}
	return &AESGCM{aead: aead}, nil
}

// Encrypt 返回 ENC(v1,nonce_b64,ciphertext_b64) 格式字符串。
func (c *AESGCM) Encrypt(plain string) (string, error) {
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("crypto: rand nonce: %w", err)
	}
	ct := c.aead.Seal(nil, nonce, []byte(plain), nil)
	return fmt.Sprintf("ENC(v1,%s,%s)",
		base64.RawStdEncoding.EncodeToString(nonce),
		base64.RawStdEncoding.EncodeToString(ct),
	), nil
}

// Decrypt 解码 ENC(v1,...) 字符串。
func (c *AESGCM) Decrypt(s string) (string, error) {
	if !strings.HasPrefix(s, "ENC(") || !strings.HasSuffix(s, ")") {
		return "", errors.New("crypto: not an ENC payload")
	}
	body := s[len("ENC(") : len(s)-1]
	parts := strings.SplitN(body, ",", 3)
	if len(parts) != 3 {
		return "", errors.New("crypto: malformed ENC payload")
	}
	if parts[0] != "v1" {
		return "", fmt.Errorf("crypto: unsupported ENC version %q", parts[0])
	}
	nonce, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("crypto: nonce decode: %w", err)
	}
	ct, err := base64.RawStdEncoding.DecodeString(parts[2])
	if err != nil {
		return "", fmt.Errorf("crypto: ciphertext decode: %w", err)
	}
	plain, err := c.aead.Open(nil, nonce, ct, nil)
	if err != nil {
		return "", fmt.Errorf("crypto: aead.Open: %w", err)
	}
	return string(plain), nil
}
