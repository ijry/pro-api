package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

const (
	// keyByteLen 是明文 key 的随机字节数。36 bytes → base64url 48 字符。
	keyByteLen = 36
	// keyPrefixLit 是明文 key 的固定前缀,用于客户端识别。
	keyPrefixLit = "pa-"
	// defaultShowLen 是 prefix 展示时前缀的默认明文字符数(含 "pa-")。
	defaultShowLen = 8
)

// newPlaintext 生成一个新的明文 key 与其 sha256 hex 哈希。
//
//	明文形式 : "pa-" + 48 字符 base64url(36 字节随机)
//	hash 形式: hex.EncodeToString(sha256(plaintext)) — 小写 64 字符
func newPlaintext() (plaintext string, hash string, err error) {
	buf := make([]byte, keyByteLen)
	if _, err := rand.Read(buf); err != nil {
		return "", "", fmt.Errorf("token: read random: %w", err)
	}
	body := base64.RawURLEncoding.EncodeToString(buf)
	plaintext = keyPrefixLit + body
	h := sha256.Sum256([]byte(plaintext))
	hash = hex.EncodeToString(h[:])
	return plaintext, hash, nil
}

// hashPlaintext 计算 plaintext 的 sha256 hex,用于对外校验路径。
func hashPlaintext(plaintext string) string {
	h := sha256.Sum256([]byte(plaintext))
	return hex.EncodeToString(h[:])
}

// prefixForDisplay 把明文 key 转成展示形态 "pa-xxxxxxxx****xxxx"。
//
//   - showLen <= 0 视为默认值 8
//   - 明文长度 <= showLen + 4 时直接返回明文(不脱敏)
//   - 否则返回 "前 showLen 字符 + **** + 末 4 字符"
func prefixForDisplay(plaintext string, showLen int) string {
	if showLen <= 0 {
		showLen = defaultShowLen
	}
	if len(plaintext) <= showLen+4 {
		return plaintext
	}
	return plaintext[:showLen] + "****" + plaintext[len(plaintext)-4:]
}
