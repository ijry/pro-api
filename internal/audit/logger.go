package audit

import (
	"context"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// IDGenerator 是 audit 模块对 idgen 的最小依赖(便于测试)。
type IDGenerator interface {
	Generate() int64
}

// Logger 写审计日志。
type Logger interface {
	Log(ctx context.Context, entry Entry) error
}

// dbLogger 把审计写入 audit_logs 表。失败时只打日志,不向上返回错误。
type dbLogger struct {
	db    *gorm.DB
	log   *zap.Logger
	idgen IDGenerator
}

// NewDB 构造一个写 DB 的 Logger。
func NewDB(db *gorm.DB, log *zap.Logger, idg IDGenerator) *dbLogger {
	if log == nil {
		log = zap.NewNop()
	}
	return &dbLogger{db: db, log: log, idgen: idg}
}

// Log 写入一条 audit。entry 中的 ID/CreatedAt 会被自动填充。
func (l *dbLogger) Log(ctx context.Context, e Entry) error {
	e.ID = l.idgen.Generate()
	e.CreatedAt = time.Now().UTC()
	if err := l.db.WithContext(ctx).Create(&e).Error; err != nil {
		l.log.Error("audit: write failed",
			zap.Error(err),
			zap.String("action", e.Action),
			zap.String("target_type", e.TargetType),
		)
		return nil // 故意吞掉,避免审计失败阻塞业务
	}
	return nil
}

// noopLogger 用于测试或禁用审计的场景。
type noopLogger struct{}

func (noopLogger) Log(context.Context, Entry) error { return nil }

// NewNoop 返回一个不做任何事的 Logger。
func NewNoop() Logger { return noopLogger{} }
