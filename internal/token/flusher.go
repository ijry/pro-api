package token

import (
	"context"
	"sync"
	"time"

	"github.com/ijry/pro-api/internal/util/clock"
	"go.uber.org/zap"
)

// flushSink 是 flusher 写库 + 发布广播的抽象,便于测试替换。
type flushSink interface {
	flushLastUsed(ctx context.Context, m map[int64]time.Time) error
	flushQuota(ctx context.Context, m map[int64]int64) error
}

// cacheNotifier 是 flusher 同步 last_used 给其他实例的接口(由 tokenCache 实现)。
type cacheNotifier interface {
	PublishLastUsed(ctx context.Context, tokenID int64, t time.Time)
}

// flusher 把 last_used_at / quota_used 写库聚合成 30s 一批,减少 DB QPS。
//
//   - TouchLastUsed: 仅保留 max(已记录, 新 t)
//   - IncrementUsage: 累加 delta(允许 0,实际跳过)
//   - run: 内部 goroutine 周期 flush
//   - Close: 关闭 stop,等 final flush 完成(同步)
type flusher struct {
	mu         sync.Mutex
	lastUsed   map[int64]time.Time
	quotaDelta map[int64]int64

	interval time.Duration
	batch    int

	sink     flushSink
	notifier cacheNotifier
	clk      clock.Clock
	log      *zap.Logger

	stopCh chan struct{}
	doneCh chan struct{}
	once   sync.Once
	closed bool
}

// newFlusher 构造 flusher 但不启动。调用 start() 后 goroutine 才运行。
func newFlusher(sink flushSink, notifier cacheNotifier, clk clock.Clock, log *zap.Logger, interval time.Duration, batch int) *flusher {
	if interval <= 0 {
		interval = defaultFlushInterval
	}
	if batch <= 0 {
		batch = defaultFlushBatchSize
	}
	if log == nil {
		log = zap.NewNop()
	}
	if clk == nil {
		clk = clock.Real
	}
	return &flusher{
		lastUsed:   make(map[int64]time.Time, 64),
		quotaDelta: make(map[int64]int64, 64),
		interval:   interval,
		batch:      batch,
		sink:       sink,
		notifier:   notifier,
		clk:        clk,
		log:        log,
		stopCh:     make(chan struct{}),
		doneCh:     make(chan struct{}),
	}
}

// TouchLastUsed 记录 token 使用时间,只保留 max。
func (f *flusher) TouchLastUsed(tokenID int64, t time.Time) {
	if tokenID == 0 {
		return
	}
	f.mu.Lock()
	if cur, ok := f.lastUsed[tokenID]; !ok || t.After(cur) {
		f.lastUsed[tokenID] = t
	}
	f.mu.Unlock()
}

// IncrementUsage 累加 quota delta。
func (f *flusher) IncrementUsage(tokenID int64, delta int64) {
	if tokenID == 0 || delta == 0 {
		return
	}
	f.mu.Lock()
	f.quotaDelta[tokenID] += delta
	f.mu.Unlock()
}

// start 启动后台 goroutine。多次调用安全(once)。
func (f *flusher) start() {
	f.once.Do(func() {
		go f.run()
	})
}

// run 周期 flush;收到 stop 后做 final flush 然后关 doneCh。
func (f *flusher) run() {
	defer close(f.doneCh)
	ticker := f.clk.NewTicker(f.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C():
			f.flushOnce(context.Background())
		case <-f.stopCh:
			// final flush 使用独立 context 避免被取消
			f.flushOnce(context.Background())
			return
		}
	}
}

// snapshot 取走当前累积的两张 map(原子替换为空 map)。
func (f *flusher) snapshot() (map[int64]time.Time, map[int64]int64) {
	f.mu.Lock()
	lu := f.lastUsed
	qu := f.quotaDelta
	f.lastUsed = make(map[int64]time.Time, 64)
	f.quotaDelta = make(map[int64]int64, 64)
	f.mu.Unlock()
	return lu, qu
}

// flushOnce 把当前累积写库;按 batch 分片避免单批过大。
func (f *flusher) flushOnce(ctx context.Context) {
	lu, qu := f.snapshot()
	if len(lu) > 0 {
		for chunk := range chunkLastUsed(lu, f.batch) {
			if err := f.sink.flushLastUsed(ctx, chunk); err != nil {
				f.log.Warn("token: flush last_used failed", zap.Error(err), zap.Int("count", len(chunk)))
				continue
			}
			if f.notifier != nil {
				for id, t := range chunk {
					f.notifier.PublishLastUsed(ctx, id, t)
				}
			}
		}
	}
	if len(qu) > 0 {
		for chunk := range chunkQuota(qu, f.batch) {
			if err := f.sink.flushQuota(ctx, chunk); err != nil {
				f.log.Warn("token: flush quota failed", zap.Error(err), zap.Int("count", len(chunk)))
			}
		}
	}
}

// Close 信号 stop,同步等 final flush 结束。可重复调用。
func (f *flusher) Close() error {
	f.mu.Lock()
	if f.closed {
		f.mu.Unlock()
		return nil
	}
	f.closed = true
	f.mu.Unlock()
	close(f.stopCh)
	<-f.doneCh
	return nil
}

// chunkLastUsed 按 batch 分片 map[int64]time.Time。
func chunkLastUsed(m map[int64]time.Time, batch int) <-chan map[int64]time.Time {
	out := make(chan map[int64]time.Time)
	go func() {
		defer close(out)
		cur := make(map[int64]time.Time, batch)
		for k, v := range m {
			cur[k] = v
			if len(cur) >= batch {
				out <- cur
				cur = make(map[int64]time.Time, batch)
			}
		}
		if len(cur) > 0 {
			out <- cur
		}
	}()
	return out
}

// chunkQuota 按 batch 分片 map[int64]int64。
func chunkQuota(m map[int64]int64, batch int) <-chan map[int64]int64 {
	out := make(chan map[int64]int64)
	go func() {
		defer close(out)
		cur := make(map[int64]int64, batch)
		for k, v := range m {
			cur[k] = v
			if len(cur) >= batch {
				out <- cur
				cur = make(map[int64]int64, batch)
			}
		}
		if len(cur) > 0 {
			out <- cur
		}
	}()
	return out
}
