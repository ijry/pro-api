package redeem

import "testing"

func TestStatusName(t *testing.T) {
	cases := []struct {
		s    int8
		want string
	}{
		{StatusUnused, "unused"},
		{StatusUsed, "used"},
		{StatusDisabled, "disabled"},
		{99, "unknown"},
	}
	for _, c := range cases {
		got := StatusName(c.s)
		if got != c.want {
			t.Errorf("StatusName(%d) = %q, want %q", c.s, got, c.want)
		}
	}
}

func TestCode_TableName(t *testing.T) {
	if (Code{}).TableName() != "redeem_codes" {
		t.Fatalf("table name wrong")
	}
}
