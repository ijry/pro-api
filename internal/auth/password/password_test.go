package password

import (
	"errors"
	"testing"

	"github.com/ijry/pro-api/pkg/apierr"
	"golang.org/x/crypto/bcrypt"
)

func TestHash_VerifyRoundTrip(t *testing.T) {
	h, err := Hash("P@ssw0rd!")
	if err != nil {
		t.Fatal(err)
	}
	if err := Verify(h, "P@ssw0rd!"); err != nil {
		t.Fatalf("verify failed: %v", err)
	}
}

func TestVerify_WrongPasswordMismatch(t *testing.T) {
	h, _ := Hash("a")
	if err := Verify(h, "b"); !errors.Is(err, ErrMismatch) {
		t.Fatalf("want ErrMismatch, got %v", err)
	}
}

func TestVerify_EmptyHash(t *testing.T) {
	if err := Verify("", "x"); !errors.Is(err, ErrMismatch) {
		t.Fatalf("want ErrMismatch, got %v", err)
	}
}

func TestHash_CostIs10(t *testing.T) {
	h, _ := Hash("x")
	c, _ := bcrypt.Cost([]byte(h))
	if c != Cost {
		t.Fatalf("want cost %d, got %d", Cost, c)
	}
}

func TestHash_EmptyError(t *testing.T) {
	if _, err := Hash(""); err == nil {
		t.Fatal("want error")
	}
}

func TestCheckStrength_MinLength(t *testing.T) {
	if err := CheckStrength("short", 8, false); err == nil {
		t.Fatal("want error")
	}
	if err := CheckStrength("longenough", 8, false); err != nil {
		t.Fatalf("want ok, got %v", err)
	}
}

func TestCheckStrength_RequireMixed(t *testing.T) {
	if err := CheckStrength("onlyletters", 8, true); err == nil {
		t.Fatal("want error for letters-only")
	}
	if err := CheckStrength("12345678", 8, true); err == nil {
		t.Fatal("want error for digits-only")
	}
	if err := CheckStrength("ab12cd34", 8, true); err != nil {
		t.Fatalf("want ok, got %v", err)
	}
}

func TestCheckStrength_DefaultMinIs8(t *testing.T) {
	err := CheckStrength("1234567", 0, false)
	var ae *apierr.Error
	if !errors.As(err, &ae) || ae.Code != apierr.CodeInvalidParam {
		t.Fatalf("want apierr CodeInvalidParam, got %v", err)
	}
}
