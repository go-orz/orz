package orz

import (
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// NewLoggerFromConfig 根据配置创建logger
func NewLoggerFromConfig(cfg Log) (*zap.Logger, error) {
	// 解析日志级别
	level := parseLogLevel(cfg.Level)

	// 配置编码器
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "time"
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000")
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	// 选择编码器类型
	var encoder zapcore.Encoder
	if strings.ToLower(cfg.Encode) == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	var cores []zapcore.Core

	// 文件输出
	if cfg.Filename != "" {
		if err := ensureDir(cfg.Filename); err != nil {
			return nil, err
		}

		// 使用 lumberjack 进行日志轮转
		rotateWriter := &lumberjack.Logger{
			Filename:  cfg.Filename,
			MaxSize:   getMaxSize(cfg.MaxSize), // MB
			MaxAge:    getMaxAge(cfg.MaxAge),   // 天数
			Compress:  cfg.Compress,            // 是否压缩
			LocalTime: true,                    // 使用本地时间
		}

		fileCore := zapcore.NewCore(encoder, zapcore.AddSync(rotateWriter), level)
		cores = append(cores, fileCore)
	}

	// 控制台输出
	if cfg.Console {
		consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
		consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level)
		cores = append(cores, consoleCore)
	}

	// 如果没有任何输出，默认输出到控制台
	if len(cores) == 0 {
		consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
		consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level)
		cores = append(cores, consoleCore)
	}

	// 合并所有核心
	var core zapcore.Core
	if len(cores) == 1 {
		core = cores[0]
	} else {
		core = zapcore.NewTee(cores...)
	}

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
