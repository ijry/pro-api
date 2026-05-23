package log

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/ijry/pro-api/internal/setting"
	"github.com/ijry/pro-api/internal/util/clock"
	"github.com/ijry/pro-api/internal/util/idgen"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Store 是日志读写主入口。
type Store interface {
	// Write 非阻塞写入一条 request log。channel 满时静默丢弃并 counter++。
	Write(ctx context.Context, e Event)

	// WriteError 非阻塞写入一条 error log。
	WriteError(ctx context.Context, e ErrorEvent)

	// Query 多条件分页查询 request_logs。必须传入 From/To，否则返回 ErrTimeRangeRequired。
	Query(ctx context.Context, q Query) (QueryResult, error)

	// QueryErrors 查询 error_logs。
	QueryErrors(ctx context.Context, q ErrorQuery) (ErrorQueryResult, error)

	// QueryAudits 查询 audit_logs（只读）。
	QueryAudits(ctx context.Context, q AuditQuery) (AuditQueryResult, error)

	// StatsOverview 聚合概览统计。
	StatsOverview(ctx context.Context, q OverviewQuery) (Overview, error)

	// StatsTimeseries 按时间桶聚合。
	StatsTimeseries(ctx context.Context, q TimeseriesQuery) (Timeseries, error)

	// StatsByModel 按模型聚合。
	StatsByModel(ctx context.Context, q GroupQuery) (GroupResult, error)

	// StatsByChannel 按渠道聚合。
	StatsByChannel(ctx context.Context, q GroupQuery) (GroupResult, error)

	// StatsByUser 按用户聚合。
	StatsByUser(ctx context.Context, q GroupQuery) (GroupResult, error)

	// ExportCSV 流式导出 request_logs。
	ExportCSV(ctx context.Context, q Query, w io.Writer) (rowsWritten int64, err error)

	// EnsurePartitions 检查并创建分区。幂等。
	EnsurePartitions(ctx context.Context, start time.Time, months int) error

	// Close 优雅关停。
	Close() error
}

// Backend 选择存储实现。
type Backend string

const (
	BackendDB         Backend = "db"
	BackendClickHouse Backend = "clickhouse"
)

// Config 装配参数。
type Config struct {
	Backend    Backend
	DB         *gorm.DB
	Clock      clock.Clock
	IDGen      *idgen.Generator
	Log        *zap.Logger
	Setting    setting.Store
}

// New 是装配入口。
func New(ctx context.Context, cfg Config) (Store, error) {
	if cfg.Backend == BackendClickHouse {
		return nil, ErrCHNotImplemented
	}
	if cfg.DB == nil {
		return nil, errors.New("log: db is required for backend=db")
	}

	// 检测数据库驱动
	driver := ""
	if sqlDB, err := cfg.DB.DB(); err == nil {
		// 通过 driver name 判断
		_ = sqlDB
	}
	// 从 gorm dialector 名判断
	dname := cfg.DB.Name()
	if dname == "postgres" || dname == "pgx" {
		driver = "postgres"
	} else {
		driver = "mysql"
	}

	log := cfg.Log
	if log == nil {
		log = zap.NewNop()
	}
	clk := cfg.Clock
	if clk == nil {
		clk = clock.Real
	}
	idg := cfg.IDGen
	if idg == nil {
		return nil, errors.New("log: idgen is required")
	}

	// batch size / interval from settings
	batchSize := 100
	flushIntervalMS := 1000
	chCap := 4096
	workersReq := 2
	workersErr := 2

	if cfg.Setting != nil {
		if v := cfg.Setting.GetInt(ctx, "log.flush_batch_size", 0); v > 0 {
			batchSize = v
		}
		if v := cfg.Setting.GetInt(ctx, "log.flush_interval_ms", 0); v > 0 {
			flushIntervalMS = v
		}
		if v := cfg.Setting.GetInt(ctx, "log.channel_buffer_size", 0); v > 0 {
			chCap = v
		}
		if v := cfg.Setting.GetInt(ctx, "log.workers_request", 0); v > 0 {
			workersReq = v
		}
		if v := cfg.Setting.GetInt(ctx, "log.workers_error", 0); v > 0 {
			workersErr = v
		}
	}

	flushInterval := time.Duration(flushIntervalMS) * time.Millisecond

	s := &dbStore{
		db:     cfg.DB,
		driver: driver,
		clock:  clk,
		idgen:  idg,
		log:    log,
	}

	// 选择分区实现
	if driver == "postgres" {
		s.partitioner = &pgPartitioner{db: cfg.DB, parent: "request_logs"}
	} else {
		s.partitioner = &mysqlPartitioner{db: cfg.DB, table: "request_logs"}
	}

	// 启动 flusher
	f := newFlusher(cfg.DB, log, chCap, batchSize, flushInterval, workersReq, workersErr)
	s.flusher = f
	f.start()

	// 启动分区 cron
	s.cronStop = make(chan struct{})
	s.startPartitionCron(ctx)

	// 启动时建分区
	go func() {
		ensureCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		months := 12
		if cfg.Setting != nil {
			if v := cfg.Setting.GetInt(ensureCtx, "log.partition_ensure_months", 0); v > 0 {
				months = v
			}
		}
		if err := s.EnsurePartitions(ensureCtx, clk.Now(), months); err != nil {
			log.Warn("log: initial partition ensure failed", zap.Error(err))
		} else {
			log.Info("log: store initialized", zap.String("backend", "db"),
				zap.String("driver", driver))
		}
	}()

	return s, nil
}

// Sentinel errors
var (
	ErrTimeRangeRequired = errors.New("log: time range required")
	ErrTimeRangeTooWide  = errors.New("log: time range too wide (max 31 days)")
)

const (
	maxQueryRangeDays = 31
	defaultPageSize   = 50
	maxPageSize       = 200
)
