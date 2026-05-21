package tokenize

import "testing"

func TestApproximate_Name(t *testing.T) {
	a := NewApproximate(4)
	if a.Name() != "approximate" {
		t.Fatalf("want approximate, got %s", a.Name())
	}
}

func TestApproximate_CountIsBytesDiv4(t *testing.T) {
	a := NewApproximate(4)
	cases := []struct {
		text string
		want int
	}{
		{"", 0},
		{"abc", 1},
		{"abcd", 1},
		{"abcde", 2},
		{"中文也行", 3},
	}
	for _, c := range cases {
		got, err := a.Count("any-model", c.text)
		if err != nil {
			t.Fatalf("Count(%q) returned err %v", c.text, err)
		}
		if got != c.want {
			t.Errorf("Count(%q) = %d, want %d", c.text, got, c.want)
		}
	}
}

func TestApproximate_RejectsZeroDivisor(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("want panic for divisor=0")
		}
	}()
	_ = NewApproximate(0)
}

func TestApproximate_CountMessages_SumsContentAndRoleOverhead(t *testing.T) {
	a := NewApproximate(4)
	msgs := []Message{
		{Role: "system", Content: "abcd"},
		{Role: "user", Content: "abcdefgh"},
	}
	got, _ := a.CountMessages("any-model", msgs)
	if got != 11 {
		t.Fatalf("want 11, got %d", got)
	}
}
