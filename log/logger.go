package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"strings"
	"time"
)

func NewLogger(level, filename string) *zap.Logger {
	return New(level, filename, "console")
}

func New(level, filename, encode string) *zap.Logger {
	consoleEncoderConfig := zap.NewProductionEncoderConfig()
	consoleEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	consoleEncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
	}

	fileEncoderConfig := zap.NewProductionEncoderConfig()
	fileEncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
	}

	var minLevel zapcore.Level

	switch strings.ToLower(level) {
	case "debug":
		minLevel = zapcore.DebugLevel
	case "info":
		minLevel = zapcore.InfoLevel
	case "warn", "warning":
		minLevel = zapcore.WarnLevel
	case "err", "error":
		minLevel = zapcore.ErrorLevel
	default:
		minLevel = zapcore.DebugLevel
	}

	var (
		consoleEncoder zapcore.Encoder
		fileEncoder    zapcore.Encoder
	)

	switch encode {
	case "console":
		consoleEncoder = zapcore.NewConsoleEncoder(consoleEncoderConfig)
		fileEncoder = zapcore.NewConsoleEncoder(fileEncoderConfig)
	case "json":
		consoleEncoder = zapcore.NewJSONEncoder(consoleEncoderConfig)
		fileEncoder = zapcore.NewJSONEncoder(fileEncoderConfig)
	default:
		consoleEncoder = zapcore.NewConsoleEncoder(consoleEncoderConfig)
		fileEncoder = zapcore.NewConsoleEncoder(fileEncoderConfig)
	}

	var cores = []zapcore.Core{
		zapcore.NewCore(
			consoleEncoder,
			zapcore.Lock(os.Stdout),
			zap.LevelEnablerFunc(func(level zapcore.Level) bool {
				return level >= minLevel
			}),
		),
		zapcore.NewCore(
			fileEncoder,
			zapcore.AddSync(&lumberjack.Logger{
				Filename:  filename, // ⽇志⽂件路径
				MaxSize:   100,      // 单位为MB,默认为512MB
				MaxAge:    5,        // 文件最多保存多少天
				LocalTime: true,     // 采用本地时间
				Compress:  false,    // 是否压缩日志
			}),
			zap.LevelEnablerFunc(func(level zapcore.Level) bool {
				return level >= minLevel
			}),
		),
	}

	return zap.New(zapcore.NewTee(cores...), zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
}
