package tokenize

import "testing"

type stubTokenizer struct{ name string }

func (s *stubTokenizer) Count(string, string) (int, error)             { return 42, nil }
func (s *stubTokenizer) CountMessages(string, []Message) (int, error)  { return 100, nil }
func (s *stubTokenizer) Name() string                                  { return s.name }

func TestRegistry_For_MatchesPattern_Exact(t *testing.T) {
	r := NewRegistry(NewApproximate(4))
	r.Register("gpt-4", &stubTokenizer{name: "gpt4"})
	if got := r.For("gpt-4"); got.Name() != "gpt4" {
		t.Fatalf("want gpt4, got %s", got.Name())
	}
}

func TestRegistry_For_MatchesPattern_Wildcard(t *testing.T) {
	r := NewRegistry(NewApproximate(4))
	r.Register("gpt-*", &stubTokenizer{name: "gptStar"})
	if got := r.For("gpt-4o"); got.Name() != "gptStar" {
		t.Fatalf("want gptStar, got %s", got.Name())
	}
	if got := r.For("gpt-3.5-turbo"); got.Name() != "gptStar" {
		t.Fatalf("want gptStar, got %s", got.Name())
	}
}

func TestRegistry_For_ExactBeatsWildcard(t *testing.T) {
	r := NewRegistry(NewApproximate(4))
	r.Register("gpt-*", &stubTokenizer{name: "gptStar"})
	r.Register("gpt-4o", &stubTokenizer{name: "gpt4o"})
	if got := r.For("gpt-4o"); got.Name() != "gpt4o" {
		t.Fatalf("want gpt4o (exact), got %s", got.Name())
	}
}

func TestRegistry_For_FallbackWhenNoMatch(t *testing.T) {
	fallback := NewApproximate(4)
	r := NewRegistry(fallback)
	if got := r.For("unknown-model"); got.Name() != fallback.Name() {
		t.Fatalf("want %s, got %s", fallback.Name(), got.Name())
	}
}

func TestNewDefaultRegistry_HasFallback(t *testing.T) {
	r, err := NewDefaultRegistry()
	if err != nil {
		t.Fatal(err)
	}
	if got := r.For("totally-unknown"); got.Name() != "approximate" {
		t.Fatalf("want approximate fallback, got %s", got.Name())
	}
}

func TestNewDefaultRegistry_RoutesGPT4(t *testing.T) {
	r, err := NewDefaultRegistry()
	if err != nil {
		t.Fatal(err)
	}
	tk := r.For("gpt-4o")
	if tk.Name() == "approximate" {
		t.Fatalf("gpt-4o should route to tiktoken, got %s", tk.Name())
	}
}
