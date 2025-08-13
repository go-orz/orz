package orz

import (
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func NewLoggerFromConfig(cfg Log) (*zap.Logger, error) {
	// 解析日志级别
	level := parseLogLevel(cfg.Level)

	// 基础 encoder 配置（无颜色，给文件用）
	baseEncoderConfig := zap.NewProductionEncoderConfig()
	baseEncoderConfig.TimeKey = "time"
	baseEncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000")
	baseEncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	// 控制台 encoder 配置（带颜色）
	consoleEncoderConfig := baseEncoderConfig
	consoleEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	var cores []zapcore.Core

	// 文件输出（无颜色）
	if cfg.Filename != "" {
		if err := ensureDir(cfg.Filename); err != nil {
			return nil, err
		}

		rotateWriter := &lumberjack.Logger{
			Filename:  cfg.Filename,
			MaxSize:   getMaxSize(cfg.MaxSize),
			MaxAge:    getMaxAge(cfg.MaxAge),
			Compress:  cfg.Compress,
			LocalTime: true,
		}

		fileEncoder := zapcore.NewConsoleEncoder(baseEncoderConfig)
		if strings.ToLower(cfg.Encode) == "json" {
			fileEncoder = zapcore.NewJSONEncoder(baseEncoderConfig)
		}

		fileCore := zapcore.NewCore(fileEncoder, zapcore.AddSync(rotateWriter), level)
		cores = append(cores, fileCore)
	}

	// 控制台输出（彩色）
	if cfg.Console {
		consoleEncoder := zapcore.NewConsoleEncoder(consoleEncoderConfig)
		consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level)
		cores = append(cores, consoleCore)
	}

	// 如果都没设置，默认输出到彩色控制台
	if len(cores) == 0 {
		consoleEncoder := zapcore.NewConsoleEncoder(consoleEncoderConfig)
		consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level)
		cores = append(cores, consoleCore)
	}

	// 合并 core
	core := zapcore.NewTee(cores...)
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	return logger, nil
}

// parseLogLevel 解析日志级别
func parseLogLevel(levelStr string) zapcore.Level {
	switch strings.ToLower(levelStr) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	case "panic":
		return zapcore.PanicLevel
	default:
		return zapcore.InfoLevel
	}
}

// getMaxSize 获取日志文件最大大小，默认100MB
func getMaxSize(size int) int {
	if size <= 0 {
		return 100
	}
	return size
}

// getMaxAge 获取日志保留天数，默认7天
func getMaxAge(age int) int {
	if age <= 0 {
		return 7
	}
	return age
}

// ensureDir 确保目录存在
func ensureDir(filename string) error {
	dir := filepath.Dir(filename)
	if dir == "." {
		return nil
	}
	return os.MkdirAll(dir, 0755)
}
