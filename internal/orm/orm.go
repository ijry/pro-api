// Package orm 初始化 GORM,支持 MySQL 与 PostgreSQL。
package orm

import (
	"fmt"
	"time"

	"github.com/ijry/pro-api/internal/app/config"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Open 根据 cfg.Driver 打开数据库连接。
func Open(cfg config.DatabaseConfig) (*gorm.DB, error) {
	gormCfg := &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Warn),
		DisableForeignKeyConstraintWhenMigrating: true,
		NowFunc: func() time.Time { return time.Now().UTC() },
	}
	var dial gorm.Dialector
	switch cfg.Driver {
	case "mysql":
		dial = mysql.Open(cfg.DSN)
	case "postgres":
		dial = postgres.Open(cfg.DSN)
	default:
		return nil, fmt.Errorf("orm: unsupported driver %q", cfg.Driver)
	}
	db, err := gorm.Open(dial, gormCfg)
	if err != nil {
		return nil, fmt.Errorf("orm: open %s: %w", cfg.Driver, err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("orm: db.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifeSec) * time.Second)
	return db, nil
}
