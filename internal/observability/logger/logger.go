// Package logger 封装 zap 初始化,集中决定级别、编码、字段。
package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New 创建一个新 logger。
//
//	level: debug | info | warn | error
//	format: json | console
//	sink: 输出目标;nil 等价于 stderr。
func New(level, format string, sink zapcore.WriteSyncer) (*zap.Logger, error) {
	lvl, err := parseLevel(level)
	if err != nil {
		return nil, err
	}
	encCfg := zap.NewProductionEncoderConfig()
	encCfg.TimeKey = "ts"
	encCfg.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	encCfg.MessageKey = "msg"
	encCfg.LevelKey = "level"
	encCfg.CallerKey = "caller"
	encCfg.EncodeLevel = zapcore.LowercaseLevelEncoder

	var enc zapcore.Encoder
	switch format {
	case "json":
		enc = zapcore.NewJSONEncoder(encCfg)
	case "console":
		encCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		enc = zapcore.NewConsoleEncoder(encCfg)
	default:
		return nil, fmt.Errorf("logger: unknown format %q", format)
	}
	if sink == nil {
		sink = zapcore.AddSync(os.Stderr)
	}
	core := zapcore.NewCore(enc, sink, lvl)
	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel)), nil
}

func parseLevel(s string) (zapcore.Level, error) {
	switch s {
	case "debug":
		return zapcore.DebugLevel, nil
	case "info":
		return zapcore.InfoLevel, nil
	case "warn":
		return zapcore.WarnLevel, nil
	case "error":
		return zapcore.ErrorLevel, nil
	default:
		return 0, fmt.Errorf("logger: unknown level %q", s)
	}
}
