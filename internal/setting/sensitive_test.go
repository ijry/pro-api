package setting

import "testing"

func TestIsSensitive_ExplicitWhitelist(t *testing.T) {
	keys := []string{
		"auth.github_oauth.client_secret",
		"auth.smtp.password",
		"payment.stripe.secret_key",
		"payment.stripe.webhook_secret",
	}
	for _, k := range keys {
		if !IsSensitive(k) {
			t.Errorf("IsSensitive(%q) = false, want true", k)
		}
	}
}

func TestIsSensitive_SuffixSecret(t *testing.T) {
	cases := []string{
		"foo.api_secret",
		"foo.bar.session_secret",
	}
	for _, k := range cases {
		if !IsSensitive(k) {
			t.Errorf("IsSensitive(%q) = false, want true", k)
		}
	}
}

func TestIsSensitive_SuffixPassword(t *testing.T) {
	cases := []string{
		"smtp.password",
		"db.master_password",
	}
	for _, k := range cases {
		if !IsSensitive(k) {
			t.Errorf("IsSensitive(%q) = false, want true", k)
		}
	}
}

func TestIsSensitive_SuffixKey(t *testing.T) {
	cases := []string{
		"openai.api_key",
		"foo.bar.private_key",
	}
	for _, k := range cases {
		if !IsSensitive(k) {
			t.Errorf("IsSensitive(%q) = false, want true", k)
		}
	}
}

func TestIsSensitive_NonSuffixNotMatched(t *testing.T) {
	cases := []string{
		"secret_path",       // 不在 "." 后,不应匹配
		"foo.secret_value",  // suffix 是 _value 不是 _secret/_password/_key
		"foo.key_size",
		"key_length",
	}
	for _, k := range cases {
		if IsSensitive(k) {
			t.Errorf("IsSensitive(%q) = true, want false", k)
		}
	}
}

func TestIsSensitive_NormalKey_False(t *testing.T) {
	cases := []string{
		"auth.allow_register",
		"notice.show_max",
		"billing.reserve_ttl_seconds",
	}
	for _, k := range cases {
		if IsSensitive(k) {
			t.Errorf("IsSensitive(%q) = true, want false", k)
		}
	}
}

func TestIsEncryptedValue_MatchesENC(t *testing.T) {
	cases := [][]byte{
		[]byte(`"ENC(v1,abc,def)"`),
		[]byte(` "ENC(v1,n,c)" `),
	}
	for _, c := range cases {
		if !IsEncryptedValue(c) {
			t.Errorf("IsEncryptedValue(%q) = false, want true", c)
		}
	}
}

func TestIsEncryptedValue_RejectsPlainString(t *testing.T) {
	cases := [][]byte{
		[]byte(`"hello"`),
		[]byte(`""`),
	}
	for _, c := range cases {
		if IsEncryptedValue(c) {
			t.Errorf("IsEncryptedValue(%q) = true, want false", c)
		}
	}
}

func TestIsEncryptedValue_RejectsNonString(t *testing.T) {
	cases := [][]byte{
		[]byte(`123`),
		[]byte(`true`),
		[]byte(`null`),
		[]byte(`{"k":"v"}`),
	}
	for _, c := range cases {
		if IsEncryptedValue(c) {
			t.Errorf("IsEncryptedValue(%q) = true, want false", c)
		}
	}
}

func TestIsPlaceholderValue_Exact(t *testing.T) {
	cases := [][]byte{
		[]byte(`"<encrypted>"`),
		[]byte(` "<encrypted>" `),
	}
	for _, c := range cases {
		if !IsPlaceholderValue(c) {
			t.Errorf("IsPlaceholderValue(%q) = false, want true", c)
		}
	}
}

func TestIsPlaceholderValue_RejectsOtherStrings(t *testing.T) {
	cases := [][]byte{
		[]byte(`"<encrypted_>"`),
		[]byte(`"encrypted"`),
		[]byte(`"foo"`),
		[]byte(`""`),
	}
	for _, c := range cases {
		if IsPlaceholderValue(c) {
			t.Errorf("IsPlaceholderValue(%q) = true, want false", c)
		}
	}
}
