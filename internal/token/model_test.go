package token

import "testing"

func TestTokenHasGroupID(t *testing.T) {
	var tok Token
	tok.GroupID = 42
	if tok.GroupID != 42 {
		t.Fatal("GroupID not settable")
	}
	var v View
	v.GroupID = 42
	if v.GroupID != 42 {
		t.Fatal("View.GroupID not settable")
	}
}
