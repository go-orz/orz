package database

import (
	"context"
	"fmt"
	"time"

	"github.com/go-errors/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type Logger struct {
	xLogger                   *zap.Logger
	LogLevel                  gormlogger.LogLevel
	SlowThreshold             time.Duration
	IgnoreRecordNotFoundError bool
}

func GormWrapLogger(xLogger *zap.Logger) *Logger {
	var logLevel gormlogger.LogLevel
	switch xLogger.Level() {
	case zap.DebugLevel:
		logLevel = gormlogger.Info
	case zap.WarnLevel:
		logLevel = gormlogger.Warn
	case zap.ErrorLevel:
		logLevel = gormlogger.Error
	case zap.FatalLevel:
		logLevel = gormlogger.Silent
	}
	return &Logger{
		xLogger:                   xLogger,
		LogLevel:                  logLevel,
		SlowThreshold:             200 * time.Millisecond,
		IgnoreRecordNotFoundError: true,
	}
}

func GormErrorLogger(xLogger *zap.Logger) *Logger {
	return &Logger{
		xLogger:                   xLogger,
		LogLevel:                  gormlogger.Error,
		SlowThreshold:             200 * time.Millisecond,
		IgnoreRecordNotFoundError: true,
	}
}

func (l Logger) SetAsDefault() {
	gormlogger.Default = l
}

func (l Logger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	return Logger{
		xLogger:                   l.xLogger,
		SlowThreshold:             l.SlowThreshold,
		LogLevel:                  level,
		IgnoreRecordNotFoundError: l.IgnoreRecordNotFoundError,
	}
}

func (l Logger) Info(ctx context.Context, str string, args ...interface{}) {
	if l.LogLevel < gormlogger.Info {
		return
	}
	l.xLogger.Sugar().Infof(str, args)
}

func (l Logger) Warn(ctx context.Context, str string, args ...interface{}) {
	if l.LogLevel < gormlogger.Warn {
		return
	}
	l.xLogger.Sugar().Warnf(str, args...)
}

func (l Logger) Error(ctx context.Context, str string, args ...interface{}) {
	if l.LogLevel < gormlogger.Error {
		return
	}
	l.xLogger.Sugar().Errorf(str, args...)
}

func (l Logger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= 0 {
		return
	}
	elapsed := time.Since(begin)
	logger := l.xLogger
	switch {
	case err != nil && l.LogLevel >= gormlogger.Error && (!l.IgnoreRecordNotFoundError || !errors.Is(err, gorm.ErrRecordNotFound)):
		sql, rows := fc()
		logger.Sugar().Errorf("err:%v, elapsed:%v, rows:%v, sql:%v", err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
	case l.SlowThreshold != 0 && elapsed > l.SlowThreshold && l.LogLevel >= gormlogger.Warn:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		logger.Sugar().Warnf("%v elapsed:%v, rows:%v, sql:%v", slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql)
	case l.LogLevel >= gormlogger.Info:
		sql, rows := fc()
		logger.Sugar().Debugf("elapsed:%v, rows:%v, sql:%v", float64(elapsed.Nanoseconds())/1e6, rows, sql)
	}
}
