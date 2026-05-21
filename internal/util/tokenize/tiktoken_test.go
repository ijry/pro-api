package tokenize

import "testing"

func TestTiktoken_New_KnownEncoding(t *testing.T) {
	tk, err := NewTiktoken("cl100k_base")
	if err != nil {
		t.Fatal(err)
	}
	if tk.Name() != "tiktoken/cl100k_base" {
		t.Fatalf("got %s", tk.Name())
	}
}

func TestTiktoken_New_RejectsUnknownEncoding(t *testing.T) {
	if _, err := NewTiktoken("not_a_real_encoding"); err == nil {
		t.Fatal("want error for unknown encoding")
	}
}

func TestTiktoken_Count_Empty(t *testing.T) {
	tk, _ := NewTiktoken("cl100k_base")
	n, err := tk.Count("gpt-4", "")
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("want 0, got %d", n)
	}
}

func TestTiktoken_Count_KnownInputs(t *testing.T) {
	tk, _ := NewTiktoken("cl100k_base")
	n, _ := tk.Count("gpt-4", "Hello world")
	if n < 1 || n > 4 {
		t.Errorf("Hello world: got %d, want around 2", n)
	}
	n2, _ := tk.Count("gpt-4", "The quick brown fox jumps over the lazy dog")
	if n2 < 5 {
		t.Errorf("sentence: got %d, want >= 5", n2)
	}
}

func TestTiktoken_CountMessages_AddsRoleOverhead(t *testing.T) {
	tk, _ := NewTiktoken("cl100k_base")
	msgs := []Message{
		{Role: "system", Content: "You are helpful"},
		{Role: "user", Content: "Hi"},
	}
	n, err := tk.CountMessages("gpt-4", msgs)
	if err != nil {
		t.Fatal(err)
	}
	if n < 6 {
		t.Errorf("want >= 6, got %d", n)
	}
}
