// Package password 提供 bcrypt 密码 hash + 强度校验。
package password

import (
	"errors"
	"fmt"
	"unicode"

	"github.com/ijry/pro-api/pkg/apierr"
	"golang.org/x/crypto/bcrypt"
)

// Cost 是 bcrypt cost,固定 10。
const Cost = 10

// ErrMismatch 表示密码与 hash 不匹配。
var ErrMismatch = errors.New("password mismatch")

// Hash 生成 bcrypt hash(cost=10)。
func Hash(plain string) (string, error) {
	if plain == "" {
		return "", errors.New("password: empty plaintext")
	}
	b, err := bcrypt.GenerateFromPassword([]byte(plain), Cost)
	if err != nil {
		return "", fmt.Errorf("password: hash: %w", err)
	}
	return string(b), nil
}

// Verify 校验 plain 与 hash;不匹配返回 ErrMismatch。
func Verify(hash, plain string) error {
	if hash == "" {
		return ErrMismatch
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrMismatch
		}
		return err
	}
	return nil
}

// CheckStrength 按 setting 校验密码;不满足返回 apierr.CodeInvalidParam 包装的错误。
//
//	minLen        密码最短长度(<=0 视为 8)
//	requireMixed  true 时要求至少包含字母 + 数字
func CheckStrength(plain string, minLen int, requireMixed bool) error {
	if minLen <= 0 {
		minLen = 8
	}
	if len(plain) < minLen {
		return apierr.New(apierr.CodeInvalidParam, fmt.Sprintf("密码长度至少 %d 位", minLen))
	}
	if !requireMixed {
		return nil
	}
	var hasLetter, hasDigit bool
	for _, r := range plain {
		switch {
		case unicode.IsLetter(r):
			hasLetter = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
		if hasLetter && hasDigit {
			return nil
		}
	}
	return apierr.New(apierr.CodeInvalidParam, "密码必须包含字母与数字")
}
