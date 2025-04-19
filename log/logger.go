package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"strings"
)

type Config struct {
	Level      string // debug, info, warn, error
	Filename   string // 日志文件路径
	Encode     string // console, json
	Console    bool   // 是否输出到控制台
	MaxSize    int    // 日志文件最大大小(MB)
	MaxAge     int    // 日志保留天数
	Compress   bool   // 是否压缩日志
	LocalTime  bool   // 是否使用本地时间
	Caller     bool   // 是否记录调用者信息
	Stacktrace bool   // 是否记录堆栈信息
}

var defaultConfig = Config{
	Level:      "debug",
	Encode:     "console",
	Console:    true,
	MaxSize:    100,
	MaxAge:     5,
	Compress:   false,
	LocalTime:  true,
	Caller:     true,
	Stacktrace: true,
}

func NewLogger(level, filename string) *zap.Logger {
	cfg := defaultConfig
	cfg.Level = level
	cfg.Filename = filename
	return NewWithConfig(cfg)
}

func New(level, filename, encode string) *zap.Logger {
	cfg := defaultConfig
	cfg.Level = level
	cfg.Filename = filename
	cfg.Encode = encode
	return NewWithConfig(cfg)
}

func NewWithConfig(cfg Config) *zap.Logger {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"),
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var minLevel zapcore.Level
	switch strings.ToLower(cfg.Level) {
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

	var encoder zapcore.Encoder
	switch cfg.Encode {
	case "json":
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	default:
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	var cores []zapcore.Core

	if cfg.Console {
		cores = append(cores, zapcore.NewCore(
			encoder,
			zapcore.Lock(os.Stdout),
			zap.LevelEnablerFunc(func(level zapcore.Level) bool {
				return level >= minLevel
			}),
		))
	}

	if cfg.Filename != "" {
		cores = append(cores, zapcore.NewCore(
			encoder,
			zapcore.AddSync(&lumberjack.Logger{
				Filename:  cfg.Filename,
				MaxSize:   cfg.MaxSize,
				MaxAge:    cfg.MaxAge,
				LocalTime: cfg.LocalTime,
				Compress:  cfg.Compress,
			}),
			zap.LevelEnablerFunc(func(level zapcore.Level) bool {
				return level >= minLevel
			}),
		))
	}

	var options []zap.Option
	if cfg.Caller {
		options = append(options, zap.AddCaller())
	}
	if cfg.Stacktrace {
		options = append(options, zap.AddStacktrace(zapcore.ErrorLevel))
	}

	return zap.New(zapcore.NewTee(cores...), options...)
}
