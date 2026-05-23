package log

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// partitioner defines the interface for creating DB partitions.
type partitioner interface {
	Ensure(ctx context.Context, month time.Time) error
}

func (s *dbStore) EnsurePartitions(ctx context.Context, start time.Time, months int) error {
	startMonth := time.Date(start.UTC().Year(), start.UTC().Month(), 1, 0, 0, 0, 0, time.UTC)
	var firstErr error
	for i := 0; i <= months; i++ {
		m := startMonth.AddDate(0, i, 0)
		if err := s.partitioner.Ensure(ctx, m); err != nil {
			s.log.Error("log: ensure partition failed",
				zap.Time("month", m), zap.Error(err))
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

// startPartitionCron starts a daily cron at 04:00 UTC to ensure partitions.
func (s *dbStore) startPartitionCron(ctx context.Context) {
	s.cronWG.Add(1)
	go func() {
		defer s.cronWG.Done()
		for {
			next := nextCronAt(s.clock.Now(), 4, 0)
			wait := next.Sub(s.clock.Now())
			select {
			case <-time.After(wait):
				ensureCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
				_ = s.EnsurePartitions(ensureCtx, s.clock.Now(), 12)
				cancel()
			case <-s.cronStop:
				return
			}
		}
	}()
}

// nextCronAt calculates the next occurrence of hh:mm UTC (today or tomorrow).
func nextCronAt(now time.Time, hour, minute int) time.Time {
	t := time.Date(now.UTC().Year(), now.UTC().Month(), now.UTC().Day(),
		hour, minute, 0, 0, time.UTC)
	if !t.After(now.UTC()) {
		t = t.AddDate(0, 0, 1)
	}
	return t
}
