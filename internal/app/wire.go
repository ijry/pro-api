package app

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/ijry/pro-api/internal/app/config"
	"github.com/ijry/pro-api/internal/audit"
	"github.com/ijry/pro-api/internal/cache"
	"github.com/ijry/pro-api/internal/orm"
	"github.com/ijry/pro-api/internal/setting"
	"github.com/ijry/pro-api/internal/util/clock"
	"github.com/ijry/pro-api/internal/util/crypto"
	"github.com/ijry/pro-api/internal/util/idgen"
	"github.com/ijry/pro-api/internal/util/tokenize"
	"go.uber.org/zap"
)

// SetupBasic 初始化"地基"层(M1-01 范围)。
//
// 顺序:
//
//	1. crypto from master_key
//	2. idgen from node_id
//	3. orm.Open
//	4. cache.NewClient
//	5. clock.Real
//	6. tokenize.NewDefaultRegistry
//	7. setting.New(订阅 Pub/Sub)
//	8. audit.NewDB
func SetupBasic(ctx context.Context, cfg *config.Config, log *zap.Logger) (*Application, error) {
	app := &Application{Config: cfg, Log: log}

	masterKey, err := decodeMasterKey(cfg.MasterKey)
	if err != nil {
		return nil, fmt.Errorf("crypto: decode master_key: %w", err)
	}
	c, err := crypto.NewAESGCM(masterKey)
	if err != nil {
		return nil, fmt.Errorf("crypto: %w", err)
	}
	app.Crypto = c

	idg, err := idgen.New(cfg.NodeID)
	if err != nil {
		return nil, fmt.Errorf("idgen: %w", err)
	}
	app.IDGen = idg

	db, err := orm.Open(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("orm: %w", err)
	}
	app.DB = db
	app.AddCloser("db", func() error {
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	})

	rdb, err := cache.NewClient(ctx, cfg.Redis)
	if err != nil {
		return nil, fmt.Errorf("redis: %w", err)
	}
	app.Cache = rdb
	app.AddCloser("redis", rdb.Close)

	app.Clock = clock.Real

	tr, err := tokenize.NewDefaultRegistry()
	if err != nil {
		return nil, fmt.Errorf("tokenize: %w", err)
	}
	app.Tokenize = tr

	sett, err := setting.New(ctx, setting.Config{
		DB:    db,
		Cache: rdb,
		Log:   log,
	})
	if err != nil {
		return nil, fmt.Errorf("setting: %w", err)
	}
	app.Setting = sett
	app.AddCloser("setting", sett.Close)

	app.Audit = audit.NewDB(db, log, idg)

	log.Info("application basic setup complete",
		zap.Int("node_id", cfg.NodeID),
		zap.String("db_driver", cfg.Database.Driver),
		zap.String("redis_addr", cfg.Redis.Addr),
	)
	return app, nil
}

// decodeMasterKey 接受 32 字节 raw / 64 字符 hex / base64 编码。
func decodeMasterKey(s string) ([]byte, error) {
	if len(s) == 64 {
		if b, err := hexDecode(s); err == nil {
			return b, nil
		}
	}
	for _, enc := range []*base64.Encoding{
		base64.StdEncoding,
		base64.RawStdEncoding,
		base64.URLEncoding,
		base64.RawURLEncoding,
	} {
		if b, err := enc.DecodeString(s); err == nil && len(b) == 32 {
			return b, nil
		}
	}
	if len(s) == 32 {
		return []byte(s), nil
	}
	return nil, fmt.Errorf("master_key must be 32 raw bytes / 64 hex / base64-of-32 bytes")
}

func hexDecode(s string) ([]byte, error) {
	s = strings.ToLower(s)
	out := make([]byte, len(s)/2)
	for i := 0; i < len(out); i++ {
		hi, err := hexDigit(s[i*2])
		if err != nil {
			return nil, err
		}
		lo, err := hexDigit(s[i*2+1])
		if err != nil {
			return nil, err
		}
		out[i] = byte(hi<<4 | lo)
	}
	return out, nil
}

func hexDigit(b byte) (int, error) {
	switch {
	case b >= '0' && b <= '9':
		return int(b - '0'), nil
	case b >= 'a' && b <= 'f':
		return int(b-'a') + 10, nil
	default:
		return 0, fmt.Errorf("invalid hex digit %q", b)
	}
}
