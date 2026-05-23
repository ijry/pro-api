package redeem

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	mathrand "math/rand"
	"strings"
	"time"
)

// codeAlphabet 是去歧义的 31 字符集(从 A-Z + 0-9 中排除 0/O/I/L/1)。
//
// 排除原因:
//
//	'0' 与 'O' / 'D' / 'Q' 易混
//	'I' 与 '1' / 'L' / '|' 易混
//
// 31^16 ≈ 2.4 × 10^23,生日碰撞远在安全范围。
const codeAlphabet = "23456789ABCDEFGHJKMNPQRSTUVWXYZ"

// codeLen 是无分隔的明文长度。
const codeLen = 16

// generatePlaintext 用 crypto/rand 生成 16 位明文(无分隔)。
func generatePlaintext() (string, error) {
	n := big.NewInt(int64(len(codeAlphabet)))
	b := make([]byte, codeLen)
	for i := range b {
		r, err := rand.Int(rand.Reader, n)
		if err != nil {
			return "", err
		}
		b[i] = codeAlphabet[r.Int64()]
	}
	return string(b), nil
}

// hashCode 对已 normalize 的明文做 sha256,返回 64 字符 hex(小写)。
func hashCode(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}

// prefix 取明文前 4 位作为展示前缀(无分隔)。明文短于 4 位时原样返回。
func prefix(plain string) string {
	if len(plain) < 4 {
		return plain
	}
	return plain[:4]
}

// format 把无分隔 16 位明文转 XXXX-XXXX-XXXX-XXXX,返回给管理员展示用。
// 长度不为 16 时原样返回(防御性)。
func format(plain string) string {
	if len(plain) != codeLen {
		return plain
	}
	return plain[0:4] + "-" + plain[4:8] + "-" + plain[8:12] + "-" + plain[12:16]
}

// normalize 把用户输入(可能带分隔 / 空白 / 小写)转无分隔大写 16 位明文。
//
//	去掉所有 '-' / ' ' / '\t' / '\r' / '\n';转大写;长度必须等于 codeLen;
//	含非字母表字符 → 返回 false。
//
// 不做"歧义字符自动纠正"(例如 'O' → '0') — 这样用户最坏只是"码错了",
// 不会出现"我以为兑了 A 单结果兑了 B 单"的认知偏差。
func normalize(input string) (string, bool) {
	cleaned := strings.Map(func(r rune) rune {
		switch r {
		case '-', ' ', '\t', '\r', '\n':
			return -1
		}
		if r >= 'a' && r <= 'z' {
			return r - 32
		}
		return r
	}, input)
	if len(cleaned) != codeLen {
		return "", false
	}
	for _, r := range cleaned {
		if !strings.ContainsRune(codeAlphabet, r) {
			return "", false
		}
	}
	return cleaned, true
}

// autoBatchNo 生成 "B" + yyyymmddHHMMSS + 4 位随机数 字符串。
//
// 当管理员未提供 batch_no 时由后端自动生成。注意:批次号不是密钥,
// 仅做组织 / 检索;math/rand 足够。
func autoBatchNo(t time.Time) string {
	// 不使用全局 math/rand 的 Seed 状态(避免被多 goroutine 调用降级到 0):
	// Go 1.20+ math/rand 顶层函数已使用 race-safe 自动种子。
	return "B" + t.UTC().Format("20060102150405") +
		fmt.Sprintf("%04d", mathrand.Intn(10000))
}
