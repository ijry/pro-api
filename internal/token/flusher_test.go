package token

import (
	"context"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"
)

// fakeSink 收集 flushLastUsed / flushQuota 的 payload。
type fakeSink struct {
	mu           sync.Mutex
	lastCalls    []map[int64]time.Time
	quotaCalls   []map[int64]int64
	failNextLast bool
	failNextQuot bool
}

func (s *fakeSink) flushLastUsed(_ context.Context, m map[int64]time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.failNextLast {
		s.failNextLast = false
		return assertErr("flush last_used failed")
	}
	cp := make(map[int64]time.Time, len(m))
	for k, v := range m {
		cp[k] = v
	}
	s.lastCalls = append(s.lastCalls, cp)
	return nil
}

func (s *fakeSink) flushQuota(_ context.Context, m map[int64]int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.failNextQuot {
		s.failNextQuot = false
		return assertErr("flush quota failed")
	}
	cp := make(map[int64]int64, len(m))
	for k, v := range m {
		cp[k] = v
	}
	s.quotaCalls = append(s.quotaCalls, cp)
	return nil
}

type assertErr string

func (e assertErr) Error() string { return string(e) }

// fakeNotifier 记录 PublishLastUsed 调用。
type fakeNotifier struct {
	mu    sync.Mutex
	calls []struct {
		id int64
		t  time.Time
	}
}

func (n *fakeNotifier) PublishLastUsed(_ context.Context, id int64, t time.Time) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.calls = append(n.calls, struct {
		id int64
		t  time.Time
	}{id, t})
}

func TestFlusher_TouchLastUsed_AggregatesMax(t *testing.T) {
	f := newFlusher(&fakeSink{}, nil, nil, zap.NewNop(), time.Hour, 100)
	now := time.Now()
	f.TouchLastUsed(1, now.Add(-time.Minute))
	f.TouchLastUsed(1, now)
	f.TouchLastUsed(1, now.Add(-time.Hour))
	if got := f.lastUsed[1]; !got.Equal(now) {
		t.Fatalf("want %v, got %v", now, got)
	}
}

func TestFlusher_IncrementUsage_AggregatesSum(t *testing.T) {
	f := newFlusher(&fakeSink{}, nil, nil, zap.NewNop(), time.Hour, 100)
	f.IncrementUsage(1, 10)
	f.IncrementUsage(1, 5)
	f.IncrementUsage(2, 3)
	if f.quotaDelta[1] != 15 {
		t.Fatalf("want 15, got %d", f.quotaDelta[1])
	}
	if f.quotaDelta[2] != 3 {
		t.Fatalf("want 3, got %d", f.quotaDelta[2])
	}
}

func TestFlusher_IncrementUsage_ZeroSkipped(t *testing.T) {
	f := newFlusher(&fakeSink{}, nil, nil, zap.NewNop(), time.Hour, 100)
	f.IncrementUsage(1, 0)
	if _, ok := f.quotaDelta[1]; ok {
		t.Fatal("0 delta should not allocate")
	}
}

func TestFlusher_FlushOnce_WritesAndClearsAccumulator(t *testing.T) {
	sink := &fakeSink{}
	f := newFlusher(sink, nil, nil, zap.NewNop(), time.Hour, 100)
	f.TouchLastUsed(1, time.Now())
	f.IncrementUsage(1, 50)
	f.flushOnce(context.Background())
	if len(sink.lastCalls) != 1 || len(sink.lastCalls[0]) != 1 {
		t.Fatalf("last_used not flushed: %+v", sink.lastCalls)
	}
	if len(sink.quotaCalls) != 1 || sink.quotaCalls[0][1] != 50 {
		t.Fatalf("quota not flushed: %+v", sink.quotaCalls)
	}
	// 累积器应被清空
	if len(f.lastUsed) != 0 || len(f.quotaDelta) != 0 {
		t.Fatal("accumulator not cleared after flush")
	}
}

func TestFlusher_FlushOnce_PublishesNotifier(t *testing.T) {
	sink := &fakeSink{}
	notifier := &fakeNotifier{}
	f := newFlusher(sink, notifier, nil, zap.NewNop(), time.Hour, 100)
	f.TouchLastUsed(1, time.Now())
	f.flushOnce(context.Background())
	if len(notifier.calls) != 1 {
		t.Fatalf("want 1 notifier call, got %d", len(notifier.calls))
	}
}

func TestFlusher_FlushOnce_SinkError_KeepsRunning(t *testing.T) {
	sink := &fakeSink{failNextLast: true}
	f := newFlusher(sink, nil, nil, zap.NewNop(), time.Hour, 100)
	f.TouchLastUsed(1, time.Now())
	f.flushOnce(context.Background())
	// 不应该 panic;累积器仍被清空(失败的批次丢弃)
	if len(f.lastUsed) != 0 {
		t.Fatal("snapshot should clear even on error")
	}
}

func TestFlusher_BatchChunking(t *testing.T) {
	sink := &fakeSink{}
	f := newFlusher(sink, nil, nil, zap.NewNop(), time.Hour, 2)
	for i := int64(1); i <= 5; i++ {
		f.IncrementUsage(i, i)
	}
	f.flushOnce(context.Background())
	total := 0
	for _, c := range sink.quotaCalls {
		total += len(c)
	}
	if total != 5 {
		t.Fatalf("want 5 ids across batches, got %d (chunks %v)", total, sink.quotaCalls)
	}
	// 至少有 3 个批次(2+2+1)
	if len(sink.quotaCalls) < 3 {
		t.Fatalf("expected >= 3 chunks for batch=2 with 5 items, got %d", len(sink.quotaCalls))
	}
}

func TestFlusher_FinalFlushOnClose(t *testing.T) {
	sink := &fakeSink{}
	f := newFlusher(sink, nil, nil, zap.NewNop(), 100*time.Millisecond, 100)
	f.start()
	f.IncrementUsage(1, 99)
	// Close 立刻触发 final flush
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	if len(sink.quotaCalls) == 0 {
		t.Fatal("expected final flush to write")
	}
	if sink.quotaCalls[len(sink.quotaCalls)-1][1] != 99 {
		t.Fatalf("final flush content wrong: %+v", sink.quotaCalls)
	}
}

func TestFlusher_Close_Idempotent(t *testing.T) {
	f := newFlusher(&fakeSink{}, nil, nil, zap.NewNop(), time.Hour, 100)
	f.start()
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}
