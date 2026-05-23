package log

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type flusher struct {
	db        *gorm.DB
	log       *zap.Logger
	batchSize int
	interval  time.Duration

	chReq chan Event
	chErr chan ErrorEvent

	workersReq int
	workersErr int

	stopCh chan struct{}
	wg     sync.WaitGroup
}

func newFlusher(db *gorm.DB, log *zap.Logger, chCap, batchSize int, interval time.Duration, workersReq, workersErr int) *flusher {
	return &flusher{
		db:         db,
		log:        log,
		batchSize:  batchSize,
		interval:   interval,
		chReq:      make(chan Event, chCap),
		chErr:      make(chan ErrorEvent, chCap),
		workersReq: workersReq,
		workersErr: workersErr,
		stopCh:     make(chan struct{}),
	}
}

func (f *flusher) start() {
	for i := 0; i < f.workersReq; i++ {
		f.wg.Add(1)
		go f.workerReq(i)
	}
	for i := 0; i < f.workersErr; i++ {
		f.wg.Add(1)
		go f.workerErr(i)
	}
	f.wg.Add(1)
	go f.depthReporter()
}

func (f *flusher) workerReq(_ int) {
	defer f.wg.Done()
	batch := make([]Event, 0, f.batchSize)
	tick := time.NewTicker(f.interval)
	defer tick.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}
		f.doFlushReq(batch)
		batch = batch[:0]
	}

	for {
		select {
		case e, ok := <-f.chReq:
			if !ok {
				flush()
				return
			}
			batch = append(batch, e)
			if len(batch) >= f.batchSize {
				flush()
			}
		case <-tick.C:
			flush()
		case <-f.stopCh:
			for {
				select {
				case e, ok := <-f.chReq:
					if !ok {
						flush()
						return
					}
					batch = append(batch, e)
					if len(batch) >= f.batchSize {
						flush()
					}
				default:
					flush()
					return
				}
			}
		}
	}
}

func (f *flusher) workerErr(_ int) {
	defer f.wg.Done()
	batch := make([]ErrorEvent, 0, f.batchSize)
	tick := time.NewTicker(f.interval)
	defer tick.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}
		f.doFlushErr(batch)
		batch = batch[:0]
	}

	for {
		select {
		case e, ok := <-f.chErr:
			if !ok {
				flush()
				return
			}
			batch = append(batch, e)
			if len(batch) >= f.batchSize {
				flush()
			}
		case <-tick.C:
			flush()
		case <-f.stopCh:
			for {
				select {
				case e, ok := <-f.chErr:
					if !ok {
						flush()
						return
					}
					batch = append(batch, e)
				default:
					flush()
					return
				}
			}
		}
	}
}

func (f *flusher) depthReporter() {
	defer f.wg.Done()
	t := time.NewTicker(10 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			queueDepth.WithLabelValues("request").Set(float64(len(f.chReq)))
			queueDepth.WithLabelValues("error").Set(float64(len(f.chErr)))
		case <-f.stopCh:
			return
		}
	}
}

func (f *flusher) doFlushReq(rows []Event) {
	start := time.Now()
	defer func() {
		flushDurationSec.WithLabelValues("request").Observe(time.Since(start).Seconds())
	}()
	var err error
	for attempt := 0; attempt < 3; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err = f.db.WithContext(ctx).Create(&rows).Error
		cancel()
		if err == nil {
			return
		}
		time.Sleep(time.Duration(100*(1<<attempt)) * time.Millisecond)
	}
	writeFailuresTotal.WithLabelValues("request").Inc()
	f.log.Error("log: batch insert request_logs failed after 3 retries",
		zap.Error(err), zap.Int("batch_size", len(rows)))
}

func (f *flusher) doFlushErr(rows []ErrorEvent) {
	start := time.Now()
	defer func() {
		flushDurationSec.WithLabelValues("error").Observe(time.Since(start).Seconds())
	}()
	var err error
	for attempt := 0; attempt < 3; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err = f.db.WithContext(ctx).Create(&rows).Error
		cancel()
		if err == nil {
			return
		}
		time.Sleep(time.Duration(100*(1<<attempt)) * time.Millisecond)
	}
	writeFailuresTotal.WithLabelValues("error").Inc()
	f.log.Error("log: batch insert error_logs failed after 3 retries",
		zap.Error(err), zap.Int("batch_size", len(rows)))
}

func (f *flusher) close(ctx context.Context) error {
	close(f.chReq)
	close(f.chErr)
	close(f.stopCh)
	done := make(chan struct{})
	go func() {
		f.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		f.log.Warn("log: flusher close timeout, some events may be lost")
		return ctx.Err()
	}
}
