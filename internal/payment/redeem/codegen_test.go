package redeem

import (
	"strings"
	"testing"
)

func TestGeneratePlaintext_LengthIs16(t *testing.T) {
	p, err := generatePlaintext()
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(p) != codeLen {
		t.Fatalf("want len %d, got %d (%q)", codeLen, len(p), p)
	}
}

func TestGeneratePlaintext_OnlyAlphabetChars(t *testing.T) {
	for i := 0; i < 100; i++ {
		p, err := generatePlaintext()
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		for _, r := range p {
			if !strings.ContainsRune(codeAlphabet, r) {
				t.Fatalf("illegal char %q in %q", r, p)
			}
		}
	}
}

func TestGeneratePlaintext_ExcludesAmbiguous(t *testing.T) {
	// alphabet must not contain 0/O/I/L/1
	for _, c := range "0OIL1" {
		if strings.ContainsRune(codeAlphabet, c) {
			t.Fatalf("alphabet contains ambiguous char %q", c)
		}
	}
}

func TestHashCode_Stable(t *testing.T) {
	h1 := hashCode("ABCD2345")
	h2 := hashCode("ABCD2345")
	if h1 != h2 {
		t.Fatalf("hash not stable: %s vs %s", h1, h2)
	}
	if len(h1) != 64 {
		t.Fatalf("want 64 hex chars, got %d", len(h1))
	}
}

func TestHashCode_DifferentInputsProduceDifferentHashes(t *testing.T) {
	if hashCode("ABCD") == hashCode("ABCE") {
		t.Fatal("collision on different inputs")
	}
}

func TestFormat_AddsHyphens(t *testing.T) {
	got := format("ABCD2345EFGH6789")
	want := "ABCD-2345-EFGH-6789"
	if got != want {
		t.Fatalf("want %q, got %q", want, got)
	}
}

func TestFormat_WrongLength_ReturnsAsIs(t *testing.T) {
	if got := format("ABCD"); got != "ABCD" {
		t.Fatalf("want pass-through, got %q", got)
	}
}

func TestNormalize_AlreadyNormalized_OK(t *testing.T) {
	got, ok := normalize("ABCD2345EFGH6789")
	if !ok {
		t.Fatal("want ok")
	}
	if got != "ABCD2345EFGH6789" {
		t.Fatalf("got %q", got)
	}
}

func TestNormalize_LowerCase_OK(t *testing.T) {
	got, ok := normalize("abcd2345efgh6789")
	if !ok {
		t.Fatal("want ok")
	}
	if got != "ABCD2345EFGH6789" {
		t.Fatalf("got %q", got)
	}
}

func TestNormalize_WithHyphens_OK(t *testing.T) {
	got, ok := normalize("ABCD-2345-EFGH-6789")
	if !ok {
		t.Fatal("want ok")
	}
	if got != "ABCD2345EFGH6789" {
		t.Fatalf("got %q", got)
	}
}

func TestNormalize_WithSpaces_OK(t *testing.T) {
	got, ok := normalize("ABCD 2345\tEFGH\n6789")
	if !ok {
		t.Fatal("want ok")
	}
	if got != "ABCD2345EFGH6789" {
		t.Fatalf("got %q", got)
	}
}

func TestNormalize_WithIllegalChar_Fail(t *testing.T) {
	// 'O' 是歧义字符,被排除
	if _, ok := normalize("ABCD2345EFGH678O"); ok {
		t.Fatal("want fail on illegal char O")
	}
	// 1 也被排除
	if _, ok := normalize("ABCD2345EFGH6781"); ok {
		t.Fatal("want fail on illegal char 1")
	}
}

func TestNormalize_WrongLength_Fail(t *testing.T) {
	if _, ok := normalize("ABCD"); ok {
		t.Fatal("want fail on short")
	}
	if _, ok := normalize("ABCD2345EFGH6789X"); ok {
		t.Fatal("want fail on long")
	}
}

func TestNormalize_NoAutoAmbiguousFix(t *testing.T) {
	// 'O' 不会被自动纠正为 '0'
	if _, ok := normalize("ABCD2345EFGHO789"); ok {
		t.Fatal("must not auto-fix O→0")
	}
}

func TestPrefix_FirstFour(t *testing.T) {
	if prefix("ABCD2345EFGH6789") != "ABCD" {
		t.Fatal("prefix wrong")
	}
}

func TestPrefix_Short_ReturnsAsIs(t *testing.T) {
	if prefix("AB") != "AB" {
		t.Fatal("short prefix should pass through")
	}
}

func TestAutoBatchNo_FormatLikeB_YyyymmddHHMMSS(t *testing.T) {
	got := autoBatchNo(mustParseTime(t, "2026-05-21T08:30:45Z"))
	// 长度 = 1 ('B') + 14 (timestamp) + 4 (random) = 19
	if len(got) != 19 {
		t.Fatalf("want 19, got %d (%q)", len(got), got)
	}
	if got[0] != 'B' {
		t.Fatalf("must start with B, got %c", got[0])
	}
	// 中间 14 位时间戳
	if got[1:15] != "20260521083045" {
		t.Fatalf("timestamp mismatch: %q", got[1:15])
	}
}
