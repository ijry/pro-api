package ratelimit

import "testing"

func TestCanonicalIP_IPv4(t *testing.T) {
	if got := CanonicalIP("1.2.3.4"); got != "1.2.3.0/24" {
		t.Fatalf("want 1.2.3.0/24, got %q", got)
	}
}

func TestCanonicalIP_IPv4WithPort(t *testing.T) {
	if got := CanonicalIP("1.2.3.4:5678"); got != "1.2.3.0/24" {
		t.Fatalf("want 1.2.3.0/24, got %q", got)
	}
}

func TestCanonicalIP_IPv6(t *testing.T) {
	if got := CanonicalIP("2001:db8::1"); got != "2001:db8::/64" {
		t.Fatalf("want 2001:db8::/64, got %q", got)
	}
}

func TestCanonicalIP_IPv6WithPort(t *testing.T) {
	if got := CanonicalIP("[2001:db8::1]:80"); got != "2001:db8::/64" {
		t.Fatalf("want 2001:db8::/64, got %q", got)
	}
}

func TestCanonicalIP_IPv4MappedV6(t *testing.T) {
	// "::ffff:1.2.3.4" should be treated as IPv4 → /24
	if got := CanonicalIP("::ffff:1.2.3.4"); got != "1.2.3.0/24" {
		t.Fatalf("want 1.2.3.0/24, got %q", got)
	}
}

func TestCanonicalIP_Loopback(t *testing.T) {
	if got := CanonicalIP("127.0.0.1"); got != "127.0.0.0/24" {
		t.Fatalf("want 127.0.0.0/24, got %q", got)
	}
}

func TestCanonicalIP_IPv6Loopback(t *testing.T) {
	if got := CanonicalIP("::1"); got != "::/64" {
		t.Fatalf("want ::/64, got %q", got)
	}
}

func TestCanonicalIP_InvalidString_ReturnsRaw(t *testing.T) {
	if got := CanonicalIP("localhost"); got != "localhost" {
		t.Fatalf("want localhost, got %q", got)
	}
}

func TestCanonicalIP_EmptyString_ReturnsEmpty(t *testing.T) {
	if got := CanonicalIP(""); got != "" {
		t.Fatalf("want empty, got %q", got)
	}
}
