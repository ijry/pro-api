package log

import (
	"errors"

	"go.uber.org/zap"
)

// ClickHouseConfig — M3 实现需要的字段，M1 仅占位。
type ClickHouseConfig struct {
	DSN             string
	Database        string
	BatchSize       int
	FlushIntervalMS int
}

// NewClickHouse 工厂，M1 直接返回 ErrCHNotImplemented。
func NewClickHouse(cfg ClickHouseConfig, log *zap.Logger) (Store, error) {
	return nil, ErrCHNotImplemented
}

// ErrCHNotImplemented is returned when ClickHouse backend is requested in M1.
var ErrCHNotImplemented = errors.New("log: clickhouse backend not implemented in M1; available in M3")
