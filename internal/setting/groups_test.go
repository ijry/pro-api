package setting

import "testing"

func TestGroupOf_KnownPrefixes(t *testing.T) {
	cases := map[string]Group{
		"auth.allow_register":             GroupAuth,
		"session.cookie_name":             GroupSession,
		"token.default_quota":             GroupToken,
		"pricing.usd_per_million":         GroupPricing,
		"billing.reserve_ttl_seconds":     GroupBilling,
		"channel.timeout":                 GroupChannel,
		"ratelimit.user_per_minute":       GroupRatelimit,
		"log.retention_days":              GroupLog,
		"notice.show_max":                 GroupNotice,
		"manual_recharge.min_amount":      GroupManualRecharge,
		"redeem.code_length":              GroupRedeem,
	}
	for key, want := range cases {
		if got := GroupOf(key); got != want {
			t.Errorf("GroupOf(%q): want %q, got %q", key, want, got)
		}
	}
}

func TestGroupOf_UnknownReturnsOther(t *testing.T) {
	cases := []string{
		"foo.bar",
		"xyz",
		"nope.something.deep",
		"",
	}
	for _, key := range cases {
		if got := GroupOf(key); got != GroupOther {
			t.Errorf("GroupOf(%q): want GroupOther, got %q", key, got)
		}
	}
}

func TestGroupOf_DeepKeyOnlyChecksHead(t *testing.T) {
	if got := GroupOf("auth.github_oauth.client_secret"); got != GroupAuth {
		t.Errorf("want GroupAuth, got %q", got)
	}
	if got := GroupOf("billing.reservation.ttl.seconds"); got != GroupBilling {
		t.Errorf("want GroupBilling, got %q", got)
	}
}

func TestAllGroups_StableOrder(t *testing.T) {
	expect := []Group{
		GroupAuth, GroupSession, GroupToken,
		GroupPricing, GroupBilling, GroupChannel,
		GroupRatelimit, GroupLog,
		GroupNotice, GroupManualRecharge, GroupRedeem,
		GroupOther,
	}
	got := AllGroups()
	if len(got) != len(expect) {
		t.Fatalf("len: want %d, got %d", len(expect), len(got))
	}
	for i := range expect {
		if got[i] != expect[i] {
			t.Errorf("index %d: want %q, got %q", i, expect[i], got[i])
		}
	}
}

func TestGroupLabel_AllGroupsHaveLabel(t *testing.T) {
	for _, g := range AllGroups() {
		if GroupLabel(g) == "" {
			t.Errorf("group %q has empty label", g)
		}
	}
}
