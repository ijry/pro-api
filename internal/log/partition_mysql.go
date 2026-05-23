package log

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type mysqlPartitioner struct {
	db    *gorm.DB
	table string
}

// mysqlPartitionName returns the partition name, e.g. "p_2026_05".
func mysqlPartitionName(t time.Time) string {
	return fmt.Sprintf("p_%04d_%02d", t.Year(), int(t.Month()))
}

func (p *mysqlPartitioner) Ensure(ctx context.Context, month time.Time) error {
	name := mysqlPartitionName(month)
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

	// REORGANIZE p_max to insert new partition before MAXVALUE
	sql := fmt.Sprintf(
		"ALTER TABLE %s REORGANIZE PARTITION p_max INTO ("+
			"PARTITION %s VALUES LESS THAN (TO_DAYS('%s')), "+
			"PARTITION p_max VALUES LESS THAN MAXVALUE)",
		p.table, name, next.Format("2006-01-02"),
	)
	if err := p.db.WithContext(ctx).Exec(sql).Error; err != nil {
		partitionEnsureTotal.WithLabelValues("error").Inc()
		return fmt.Errorf("create partition %s: %w", name, err)
	}
	partitionEnsureTotal.WithLabelValues("created").Inc()
	return nil
}

func (p *mysqlPartitioner) partitionExists(ctx context.Context, name string) (bool, error) {
	var cnt int64
	err := p.db.WithContext(ctx).Raw(
		`SELECT COUNT(*) FROM information_schema.PARTITIONS
		 WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND PARTITION_NAME = ?`,
		p.table, name,
	).Scan(&cnt).Error
	return cnt > 0, err
}
