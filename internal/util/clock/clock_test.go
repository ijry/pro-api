package clock

import (
	"sync"
	"testing"
	"time"
)

func TestReal_NowIsRealTime(t *testing.T) {
	before := time.Now()
	got := Real.Now()
	after := time.Now()
	if got.Before(before) || got.After(after) {
		t.Fatalf("Real.Now() = %v, want between %v and %v", got, before, after)
	}
}

func TestMock_NowDoesNotAdvance(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	m := NewMock(start)
	if got := m.Now(); !got.Equal(start) {
		t.Fatalf("want %v, got %v", start, got)
	}
	if got := m.Now(); !got.Equal(start) {
		t.Fatalf("Now() should not advance on its own, got %v", got)
	}
}

func TestMock_AdvanceMovesTime(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	m := NewMock(start)
	m.Advance(5 * time.Minute)
	want := start.Add(5 * time.Minute)
	if got := m.Now(); !got.Equal(want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestMock_Ticker_FiresAfterAdvance(t *testing.T) {
	m := NewMock(time.Now())
	tk := m.NewTicker(100 * time.Millisecond)
	defer tk.Stop()
	m.Advance(250 * time.Millisecond)
	fires := 0
	timeout := time.After(500 * time.Millisecond)
loop:
	for fires < 2 {
		select {
		case <-tk.C():
			fires++
		case <-timeout:
			break loop
		}
	}
	if fires < 2 {
		t.Fatalf("want >= 2 fires, got %d", fires)
	}
}

func TestMock_Sleep_UnblocksAfterAdvance(t *testing.T) {
	m := NewMock(time.Now())
	var wg sync.WaitGroup
	wg.Add(1)
	done := make(chan struct{})
	go func() {
		defer wg.Done()
		m.Sleep(100 * time.Millisecond)
		close(done)
	}()
	time.Sleep(20 * time.Millisecond)
	select {
	case <-done:
		t.Fatal("Sleep returned without Advance")
	default:
	}
	m.Advance(200 * time.Millisecond)
	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Sleep did not unblock after Advance")
	}
	wg.Wait()
}
