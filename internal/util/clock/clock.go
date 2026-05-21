// Package clock 提供可注入的时钟,便于测试。
package clock

import (
	"sync"
	"time"
)

// Clock 是时钟抽象。生产用 Real,测试用 NewMock。
type Clock interface {
	Now() time.Time
	Sleep(d time.Duration)
	NewTicker(d time.Duration) Ticker
}

// Ticker 是 time.Ticker 的接口化版本。
type Ticker interface {
	C() <-chan time.Time
	Stop()
}

// Real 是包装 time.* 的真实时钟。
var Real Clock = realClock{}

type realClock struct{}

func (realClock) Now() time.Time        { return time.Now() }
func (realClock) Sleep(d time.Duration) { time.Sleep(d) }
func (realClock) NewTicker(d time.Duration) Ticker {
	return realTicker{t: time.NewTicker(d)}
}

type realTicker struct{ t *time.Ticker }

func (r realTicker) C() <-chan time.Time { return r.t.C }
func (r realTicker) Stop()               { r.t.Stop() }

// Mock 是可控的时钟。时间不会自动走,必须显式 Advance。
type Mock struct {
	mu      sync.Mutex
	now     time.Time
	sleeps  []sleepWaiter
	tickers []*mockTicker
}

type sleepWaiter struct {
	wakeAt time.Time
	done   chan struct{}
}

// NewMock 构造 Mock,起始时间 start。
func NewMock(start time.Time) *Mock {
	return &Mock{now: start}
}

// Now 返回当前模拟时间。
func (m *Mock) Now() time.Time {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.now
}

// Sleep 阻塞调用者,直到 Advance(d) 或更多。
func (m *Mock) Sleep(d time.Duration) {
	m.mu.Lock()
	w := sleepWaiter{wakeAt: m.now.Add(d), done: make(chan struct{})}
	m.sleeps = append(m.sleeps, w)
	m.mu.Unlock()
	<-w.done
}

// NewTicker 返回一个 mock ticker。
func (m *Mock) NewTicker(d time.Duration) Ticker {
	m.mu.Lock()
	defer m.mu.Unlock()
	t := &mockTicker{
		interval: d,
		next:     m.now.Add(d),
		ch:       make(chan time.Time, 16),
	}
	m.tickers = append(m.tickers, t)
	return t
}

// Advance 推进时间 d,并唤醒所有应该到点的 Sleep 与 Ticker。
func (m *Mock) Advance(d time.Duration) {
	m.mu.Lock()
	target := m.now.Add(d)
	m.now = target

	// 唤醒 sleep
	remaining := m.sleeps[:0]
	for _, w := range m.sleeps {
		if !w.wakeAt.After(target) {
			close(w.done)
		} else {
			remaining = append(remaining, w)
		}
	}
	m.sleeps = remaining

	// 推进 ticker
	for _, t := range m.tickers {
		for !t.next.After(target) {
			select {
			case t.ch <- t.next:
			default:
			}
			t.next = t.next.Add(t.interval)
		}
	}
	m.mu.Unlock()
}

type mockTicker struct {
	interval time.Duration
	next     time.Time
	ch       chan time.Time
	stopped  bool
}

func (m *mockTicker) C() <-chan time.Time { return m.ch }
func (m *mockTicker) Stop()               { m.stopped = true }
