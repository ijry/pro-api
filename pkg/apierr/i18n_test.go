package apierr

import "testing"

func TestMessage_KnownCodeZH(t *testing.T) {
	got := Message(LangZH, CodeBalanceInsufficient)
	if got == "" {
		t.Fatal("expected non-empty zh message")
	}
}

func TestMessage_KnownCodeEN(t *testing.T) {
	got := Message(LangEN, CodeBalanceInsufficient)
	if got == "" {
		t.Fatal("expected non-empty en message")
	}
}

func TestMessage_UnknownLangFallsBackToEN(t *testing.T) {
	got := Message(Lang("fr"), CodeBalanceInsufficient)
	if got != Message(LangEN, CodeBalanceInsufficient) {
		t.Fatalf("expected fallback to en, got %q", got)
	}
}

func TestMessage_UnknownCodeReturnsEmpty(t *testing.T) {
	got := Message(LangZH, Code(99999))
	if got != "" {
		t.Fatalf("expected empty for unknown code, got %q", got)
	}
}

func TestLocalized_BuildsErrorWithLocalizedMessage(t *testing.T) {
	e := Localized(LangZH, CodeBalanceInsufficient)
	if e.Code != CodeBalanceInsufficient {
		t.Fatalf("code mismatch")
	}
	if e.Message == "" {
		t.Fatal("expected non-empty message")
	}
}

func TestM1NewCodes_AllHaveZHAndEN(t *testing.T) {
	newCodes := []Code{
		CodeEmailNotVerified, CodeCaptchaInvalid,
		CodeUpstreamRateLimit, CodeUpstreamQuota,
		CodeReservationNotFound, CodeReservationCommitted,
		CodeChannelDisabled, CodeChannelMisconfig,
	}
	for _, c := range newCodes {
		if Message(LangZH, c) == "" {
			t.Errorf("code %d missing ZH message", c)
		}
		if Message(LangEN, c) == "" {
			t.Errorf("code %d missing EN message", c)
		}
	}
}

func TestM1NewCodes_HTTPStatus(t *testing.T) {
	cases := map[Code]int{
		CodeEmailNotVerified:     401,
		CodeCaptchaInvalid:       401,
		CodeUpstreamRateLimit:    429,
		CodeUpstreamQuota:        429,
		CodeReservationNotFound:  410,
		CodeReservationCommitted: 409,
		CodeChannelDisabled:      503,
		CodeChannelMisconfig:     503,
	}
	for code, want := range cases {
		if got := New(code, "").HTTPStatus; got != want {
			t.Errorf("code %d want HTTP %d, got %d", code, want, got)
		}
	}
}
