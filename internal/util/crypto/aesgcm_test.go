package crypto

import (
	"strings"
	"testing"
)

func TestNew_RejectsBadKeyLength(t *testing.T) {
	if _, err := NewAESGCM(make([]byte, 16)); err == nil {
		t.Fatal("want error for 16-byte key")
	}
	if _, err := NewAESGCM(make([]byte, 32)); err != nil {
		t.Fatalf("want ok for 32-byte key: %v", err)
	}
}

func TestEncrypt_Decrypt_RoundTrip(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	c, _ := NewAESGCM(key)
	plain := "the quick brown fox 中文也行"
	cipher, err := c.Encrypt(plain)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(cipher, "ENC(v1,") {
		t.Fatalf("want ENC(v1,...) prefix, got %q", cipher)
	}
	got, err := c.Decrypt(cipher)
	if err != nil {
		t.Fatal(err)
	}
	if got != plain {
		t.Fatalf("want %q, got %q", plain, got)
	}
}

func TestEncrypt_RandomNonce(t *testing.T) {
	c, _ := NewAESGCM(make([]byte, 32))
	a, _ := c.Encrypt("same")
	b, _ := c.Encrypt("same")
	if a == b {
		t.Fatal("two encrypts of same plaintext should produce different ciphertexts")
	}
}

func TestDecrypt_RejectsTampered(t *testing.T) {
	c, _ := NewAESGCM(make([]byte, 32))
	cipher, _ := c.Encrypt("hello")
	tampered := cipher[:len(cipher)-2] + "XX"
	if _, err := c.Decrypt(tampered); err == nil {
		t.Fatal("want decrypt error on tampered ciphertext")
	}
}

func TestDecrypt_RejectsWrongFormat(t *testing.T) {
	c, _ := NewAESGCM(make([]byte, 32))
	if _, err := c.Decrypt("plain text"); err == nil {
		t.Fatal("want error for non-ENC string")
	}
	if _, err := c.Decrypt("ENC(v2,a,b)"); err == nil {
		t.Fatal("want error for unsupported version")
	}
}
