package token

import (
	"errors"
	"testing"

	"github.com/ijry/pro-api/pkg/apierr"
)

// makeView 返回一个最小化的 View,只填白名单字段。
func makeView(ips, models []string) *View {
	return &View{
		AllowedIPs:    ips,
		AllowedModels: models,
	}
}

// === IP 白名单 ===

func TestAssertIPAllowed_EmptyList_AllowsAll(t *testing.T) {
	if err := AssertIPAllowed(makeView(nil, nil), "1.2.3.4"); err != nil {
		t.Fatal(err)
	}
}

func TestAssertIPAllowed_ExactIPv4(t *testing.T) {
	v := makeView([]string{"1.2.3.4"}, nil)
	if err := AssertIPAllowed(v, "1.2.3.4"); err != nil {
		t.Fatalf("want allow, got %v", err)
	}
	if err := AssertIPAllowed(v, "1.2.3.5"); err == nil {
		t.Fatal("want reject")
	}
}

func TestAssertIPAllowed_ExactIPv6(t *testing.T) {
	v := makeView([]string{"::1"}, nil)
	if err := AssertIPAllowed(v, "::1"); err != nil {
		t.Fatalf("want allow, got %v", err)
	}
}

func TestAssertIPAllowed_CIDRv4(t *testing.T) {
	v := makeView([]string{"10.0.0.0/8"}, nil)
	if err := AssertIPAllowed(v, "10.1.2.3"); err != nil {
		t.Fatalf("want allow, got %v", err)
	}
	if err := AssertIPAllowed(v, "11.0.0.1"); err == nil {
		t.Fatal("want reject")
	}
}

func TestAssertIPAllowed_CIDRv6(t *testing.T) {
	v := makeView([]string{"fd00::/8"}, nil)
	if err := AssertIPAllowed(v, "fd00::1"); err != nil {
		t.Fatalf("want allow, got %v", err)
	}
	if err := AssertIPAllowed(v, "fe80::1"); err == nil {
		t.Fatal("want reject")
	}
}

func TestAssertIPAllowed_PortStripped(t *testing.T) {
	v := makeView([]string{"1.2.3.4/32"}, nil)
	if err := AssertIPAllowed(v, "1.2.3.4:1234"); err != nil {
		t.Fatalf("want allow, got %v", err)
	}
	v6 := makeView([]string{"::1"}, nil)
	if err := AssertIPAllowed(v6, "[::1]:8080"); err != nil {
		t.Fatalf("want allow, got %v", err)
	}
}

func TestAssertIPAllowed_BadCIDR_Skipped(t *testing.T) {
	// 第一条规则错误,第二条正确;应当匹配第二条
	v := makeView([]string{"not-a-cidr", "10.0.0.0/8"}, nil)
	if err := AssertIPAllowed(v, "10.1.1.1"); err != nil {
		t.Fatalf("want allow, got %v", err)
	}
}

func TestAssertIPAllowed_InvalidIP_Rejected(t *testing.T) {
	v := makeView([]string{"10.0.0.0/8"}, nil)
	err := AssertIPAllowed(v, "garbage-string")
	if err == nil {
		t.Fatal("want reject")
	}
	var ae *apierr.Error
	if !errors.As(err, &ae) || ae.Code != apierr.CodeIPNotAllowed {
		t.Fatalf("want CodeIPNotAllowed, got %v", err)
	}
}

func TestAssertIPAllowed_NoMatch_ReturnsIPNotAllowed(t *testing.T) {
	v := makeView([]string{"10.0.0.0/8"}, nil)
	err := AssertIPAllowed(v, "1.2.3.4")
	if err == nil {
		t.Fatal("want reject")
	}
	var ae *apierr.Error
	if !errors.As(err, &ae) || ae.Code != apierr.CodeIPNotAllowed {
		t.Fatalf("want CodeIPNotAllowed, got %v", err)
	}
}

// === 模型白名单 ===

func TestAssertModelAllowed_EmptyList_AllowsAll(t *testing.T) {
	if err := AssertModelAllowed(makeView(nil, nil), "anything"); err != nil {
		t.Fatal(err)
	}
}

func TestAssertModelAllowed_ExactMatch(t *testing.T) {
	v := makeView(nil, []string{"gpt-4o"})
	if err := AssertModelAllowed(v, "gpt-4o"); err != nil {
		t.Fatalf("want allow, got %v", err)
	}
	if err := AssertModelAllowed(v, "gpt-4"); err == nil {
		t.Fatal("want reject")
	}
}

func TestAssertModelAllowed_WildcardSuffix(t *testing.T) {
	v := makeView(nil, []string{"gpt-4*"})
	if err := AssertModelAllowed(v, "gpt-4o"); err != nil {
		t.Fatalf("want allow, got %v", err)
	}
	if err := AssertModelAllowed(v, "gpt-4-turbo"); err != nil {
		t.Fatalf("want allow, got %v", err)
	}
}

func TestAssertModelAllowed_WildcardNonMatch(t *testing.T) {
	v := makeView(nil, []string{"gpt-4*"})
	if err := AssertModelAllowed(v, "claude-3"); err == nil {
		t.Fatal("want reject")
	}
}

func TestAssertModelAllowed_ExactPrecedesWildcard(t *testing.T) {
	v := makeView(nil, []string{"gpt-4o", "gpt-4*"})
	if err := AssertModelAllowed(v, "gpt-4o"); err != nil {
		t.Fatalf("want allow, got %v", err)
	}
	if err := AssertModelAllowed(v, "gpt-4-turbo"); err != nil {
		t.Fatalf("want allow (via wildcard), got %v", err)
	}
}

func TestAssertModelAllowed_CaseSensitive(t *testing.T) {
	v := makeView(nil, []string{"gpt-4*"})
	if err := AssertModelAllowed(v, "GPT-4o"); err == nil {
		t.Fatal("want reject on case mismatch")
	}
}

func TestAssertModelAllowed_OnlySuffixWildcard(t *testing.T) {
	// 仅末尾 * 是通配,中间或开头的 * 不被视为通配字符
	v := makeView(nil, []string{"*-turbo"})
	if err := AssertModelAllowed(v, "gpt-4-turbo"); err == nil {
		t.Fatal("want reject — 开头通配不支持")
	}
}

func TestAssertModelAllowed_ReturnsCode(t *testing.T) {
	v := makeView(nil, []string{"gpt-4o"})
	err := AssertModelAllowed(v, "claude-3")
	var ae *apierr.Error
	if !errors.As(err, &ae) || ae.Code != apierr.CodeModelNotAllowed {
		t.Fatalf("want CodeModelNotAllowed, got %v", err)
	}
}
