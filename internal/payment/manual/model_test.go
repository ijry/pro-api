package manual

import "testing"

func TestStatusName(t *testing.T) {
	cases := []struct {
		s    int8
		want string
	}{
		{StatusPending, "pending"},
		{StatusApproved, "approved"},
		{StatusRejected, "rejected"},
		{StatusCanceled, "canceled"},
		{99, "unknown"},
	}
	for _, c := range cases {
		got := StatusName(c.s)
		if got != c.want {
			t.Errorf("StatusName(%d) = %q, want %q", c.s, got, c.want)
		}
	}
}

func TestRecharge_TableName(t *testing.T) {
	if (Recharge{}).TableName() != "manual_recharges" {
		t.Fatalf("table name wrong")
	}
}
