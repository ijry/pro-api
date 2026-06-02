package channel

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestSelectorGroupIDFilter(t *testing.T) {
	chA := &Channel{ID: 1, Status: 0, Priority: 1, Weight: 1, GroupID: 0, Tags: json.RawMessage(`[]`)}
	chB := &Channel{ID: 2, Status: 0, Priority: 1, Weight: 1, GroupID: 5, Tags: json.RawMessage(`[]`)}
	chC := &Channel{ID: 3, Status: 0, Priority: 1, Weight: 1, GroupID: 9, Tags: json.RawMessage(`[]`)}

	cc := &channelCache{}
	cc.mu.Lock()
	cc.channelsByModel = map[string][]*Channel{"m1": {chA, chB, chC}}
	cc.mu.Unlock()

	sel := newSelector(cc, &noopBreaker{}, 1)

	// GroupID=5 hint: must NOT return chC (group=9)
	for i := 0; i < 30; i++ {
		ch, err := sel.Select(context.Background(), "m1", SelectHint{GroupID: 5})
		if err != nil {
			t.Fatal(err)
		}
		if ch.ID == 3 {
			t.Error("group 9 channel returned for group 5 hint")
		}
	}

	// GroupID=0 hint: all channels available
	seen := map[int64]bool{}
	for i := 0; i < 100; i++ {
		ch, err := sel.Select(context.Background(), "m1", SelectHint{GroupID: 0})
		if err != nil {
			t.Fatal(err)
		}
		seen[ch.ID] = true
	}
	if !seen[3] {
		t.Error("group 9 channel not seen when GroupID hint is 0")
	}
}

type noopBreaker struct{}

func (n *noopBreaker) State(_ int64) BreakerState                                { return StateClosed }
func (n *noopBreaker) Snapshot(_ int64) HealthSnapshot                           { return HealthSnapshot{} }
func (n *noopBreaker) RecordSuccess(_ int64, _ time.Duration)                    {}
func (n *noopBreaker) RecordFailure(_ int64, _ error)                            {}
func (n *noopBreaker) AcquireProbe(_ context.Context, _ int64, _ time.Duration) (bool, func(), error) {
	return false, nil, nil
}
func (n *noopBreaker) ForceTransition(_ int64, _ BreakerState) error { return nil }
func (n *noopBreaker) Close() error                                  { return nil }
