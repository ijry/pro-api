package log

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type pgPartitioner struct {
	db     *gorm.DB
	parent string
}

// pgPartitionName returns the partition table name, e.g. "request_logs_2026_05".
func pgPartitionName(parent string, t time.Time) string {
	return fmt.Sprintf("%s_%04d_%02d", parent, t.Year(), int(t.Month()))
}

func (p *pgPartitioner) Ensure(ctx context.Context, month time.Time) error {
	name := pgPartitionName(p.parent, month)
	next := month.AddDate(0, 1, 0)

	exists, err := p.partitionExists(ctx, name)
	if err != nil {
		partitionEnsureTotal.WithLabelValues("error").Inc()
		return err
	}
	if exists {
		partitionEnsureTotal.WithLabelValues("exists").Inc()
		return nil
	}

	sql := fmt.Sprintf(
		"CREATE TABLE IF NOT EXISTS %s PARTITION OF %s "+
			"FOR VALUES FROM ('%s') TO ('%s')",
		name, p.parent,
		month.Format("2006-01-02"),
		next.Format("2006-01-02"),
	)
	if err := p.db.WithContext(ctx).Exec(sql).Error; err != nil {
		partitionEnsureTotal.WithLabelValues("error").Inc()
		return fmt.Errorf("create partition %s: %w", name, err)
	}
	partitionEnsureTotal.WithLabelValues("created").Inc()
	return nil
}

func (p *pgPartitioner) partitionExists(ctx context.Context, name string) (bool, error) {
	var cnt int64
	err := p.db.WithContext(ctx).Raw(
		`SELECT COUNT(*) FROM pg_class WHERE relname = ? AND relkind = 'r'`,
		name,
	).Scan(&cnt).Error
	return cnt > 0, err
}
