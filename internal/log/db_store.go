package log

import (
	"context"
	"sync"
	"time"

	"github.com/ijry/pro-api/internal/util/clock"
	"github.com/ijry/pro-api/internal/util/idgen"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// dbStore implements Store using MySQL/PostgreSQL via GORM.
type dbStore struct {
	db          *gorm.DB
	driver      string
	clock       clock.Clock
	idgen       *idgen.Generator
	log         *zap.Logger
	flusher     *flusher
	partitioner partitioner
	cronStop    chan struct{}
	cronWG      sync.WaitGroup
}

func (s *dbStore) Write(ctx context.Context, e Event) {
	if e.ID == 0 {
		e.ID = s.idgen.Generate()
	}
	if e.CreatedAt.IsZero() {
		e.CreatedAt = s.clock.Now().UTC()
	}
	select {
	case s.flusher.chReq <- e:
	default:
		droppedTotal.WithLabelValues("request").Inc()
	}
}

func (s *dbStore) WriteError(ctx context.Context, e ErrorEvent) {
	if e.ID == 0 {
		e.ID = s.idgen.Generate()
	}
	if e.CreatedAt.IsZero() {
		e.CreatedAt = s.clock.Now().UTC()
	}
	select {
	case s.flusher.chErr <- e:
	default:
		droppedTotal.WithLabelValues("error").Inc()
	}
}

func (s *dbStore) Close() error {
	// Stop partition cron
	close(s.cronStop)
	s.cronWG.Wait()

	// Stop flusher with 30s timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return s.flusher.close(ctx)
}

// castToTextExpr returns the SQL expression to cast a column to text
// (MySQL uses CHAR, PG uses TEXT).
func (s *dbStore) castToTextExpr(col string) string {
	if s.driver == "postgres" {
		return "CAST(" + col + " AS TEXT)"
	}
	return "CAST(" + col + " AS CHAR)"
}
