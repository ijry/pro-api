package channel

import "testing"

func TestChannelHasGroupID(t *testing.T) {
	var ch Channel
	ch.GroupID = 5
	if ch.GroupID != 5 {
		t.Fatal("GroupID not settable on Channel")
	}
}
