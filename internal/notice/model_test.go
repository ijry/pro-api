package notice

import "testing"

func TestIsValidLevel(t *testing.T) {
	for _, s := range []string{"info", "warning", "danger", "success"} {
		if !IsValidLevel(s) {
			t.Errorf("IsValidLevel(%q) = false, want true", s)
		}
	}
	for _, s := range []string{"", "foo", "INFO"} {
		if IsValidLevel(s) {
			t.Errorf("IsValidLevel(%q) = true, want false", s)
		}
	}
}

func TestIsValidTarget(t *testing.T) {
	for _, s := range []string{"all", "user", "admin"} {
		if !IsValidTarget(s) {
			t.Errorf("IsValidTarget(%q) = false, want true", s)
		}
	}
	for _, s := range []string{"", "foo", "ALL"} {
		if IsValidTarget(s) {
			t.Errorf("IsValidTarget(%q) = true, want false", s)
		}
	}
}

func TestNotice_TableName(t *testing.T) {
	if (Notice{}).TableName() != "notices" {
		t.Fatal("TableName")
	}
}

func TestToUserNotice_CopiesFields(t *testing.T) {
	n := &Notice{ID: 1, Title: "t", Content: "c", Level: "info", Target: "all"}
	u := ToUserNotice(n, true)
	if u.ID != 1 || u.Title != "t" || !u.IsRead {
		t.Fatalf("got %+v", u)
	}
}
