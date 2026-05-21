package token

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
	"testing"
)

func TestNewPlaintext_FormatAndLength(t *testing.T) {
	plaintext, hash, err := newPlaintext()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(plaintext, "pa-") {
		t.Fatalf("plaintext must start with 'pa-', got %q", plaintext)
	}
	// "pa-" + 48 字符 base64url(36 bytes → 48 chars)= 51
	if len(plaintext) != 3+48 {
		t.Fatalf("want length %d, got %d", 3+48, len(plaintext))
	}
	if len(hash) != 64 {
		t.Fatalf("want hash length 64, got %d", len(hash))
	}
}

func TestNewPlaintext_HashIsLowercaseHex64(t *testing.T) {
	plaintext, hash, err := newPlaintext()
	if err != nil {
		t.Fatal(err)
	}
	matched, _ := regexp.MatchString(`^[0-9a-f]{64}$`, hash)
	if !matched {
		t.Fatalf("hash is not 64 lowercase hex: %q", hash)
	}
	want := sha256.Sum256([]byte(plaintext))
	if hex.EncodeToString(want[:]) != hash {
		t.Fatal("hash must equal sha256(plaintext) lowercase hex")
	}
}

func TestNewPlaintext_Uniqueness(t *testing.T) {
	seen := map[string]struct{}{}
	for i := 0; i < 1000; i++ {
		p, _, err := newPlaintext()
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := seen[p]; ok {
			t.Fatalf("collision at iter %d: %s", i, p)
		}
		seen[p] = struct{}{}
	}
}

func TestPrefixForDisplay_DefaultShowLen8(t *testing.T) {
	plaintext := "pa-AbCdEf01XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX9876"
	got := prefixForDisplay(plaintext, 8)
	want := "pa-AbCdE****9876" // 前 8 char ("pa-AbCdE") + **** + 末 4
	if got != want {
		t.Fatalf("want %q, got %q", want, got)
	}
}

func TestPrefixForDisplay_CustomShowLen(t *testing.T) {
	plaintext := "pa-AbCdEfGhXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX9876"
	got := prefixForDisplay(plaintext, 12)
	want := plaintext[:12] + "****" + plaintext[len(plaintext)-4:]
	if got != want {
		t.Fatalf("want %q, got %q", want, got)
	}
}

func TestPrefixForDisplay_ShortKeyReturnsAsIs(t *testing.T) {
	plaintext := "pa-abc"
	got := prefixForDisplay(plaintext, 8)
	if got != plaintext {
		t.Fatalf("want passthrough %q, got %q", plaintext, got)
	}
}

func TestPrefixForDisplay_ZeroShowLenUsesDefault(t *testing.T) {
	plaintext := "pa-AbCdEf01XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX9876"
	got := prefixForDisplay(plaintext, 0)
	want := prefixForDisplay(plaintext, 8)
	if got != want {
		t.Fatalf("zero showLen should default to 8: want %q got %q", want, got)
	}
}
