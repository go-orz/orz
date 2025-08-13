package orz

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ConnectDatabase 连接数据库
func ConnectDatabase(cfg Database) (*gorm.DB, error) {
	return ConnectDatabaseWithLogger(cfg, nil)
}

// ConnectDatabaseWithLogger 连接数据库并指定日志器
func ConnectDatabaseWithLogger(cfg Database, zapLogger *zap.Logger) (*gorm.DB, error) {
	var wrapLogger logger.Interface
	if cfg.ShowSql && zapLogger != nil {
		wrapLogger = GormWrapLogger(zapLogger)
	} else if zapLogger != nil {
		wrapLogger = GormErrorLogger(zapLogger)
	} else {
		wrapLogger = logger.Default.LogMode(logger.Silent)
	}

	switch cfg.Type {
	case DatabaseMysql:
		return ConnectMysql(cfg.Mysql, wrapLogger)
	case DatabaseSqlite:
		return ConnectSqlite(cfg.Sqlite, wrapLogger)
	case DatabasePostgres, DatabasePostgresql:
		return ConnectPostgresql(cfg.Postgres, wrapLogger)
	default:
		return nil, fmt.Errorf("unknown database type: %s", cfg.Type)
	}
}

// ConnectMysql 连接MySQL数据库
func ConnectMysql(cfg MysqlCfg, logger logger.Interface) (*gorm.DB, error) {
	var dsn string
	if cfg.DSN != "" {
		dsn = cfg.DSN
	} else {
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.Username,
			cfg.Password,
			cfg.Hostname,
			cfg.Port,
			cfg.Database,
		)
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger,
	})
	if err != nil {
		return nil, fmt.Errorf("couldn't open database: %w", err)
	}
	return db, nil
}

// ConnectPostgresql 连接PostgreSQL数据库
func ConnectPostgresql(cfg PostgresCfg, logger logger.Interface) (*gorm.DB, error) {
	var dsn string
	if cfg.DSN != "" {
		dsn = cfg.DSN
	} else {
		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
			cfg.Hostname, cfg.Username, cfg.Password, cfg.Database, cfg.Port)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger,
	})
	if err != nil {
		return nil, fmt.Errorf("couldn't open database: %w", err)
	}
	return db, nil
}

// ConnectSqlite 连接SQLite数据库
func ConnectSqlite(cfg SqliteConfig, logger logger.Interface) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(cfg.Path), &gorm.Config{
		Logger: logger,
	})
	if err != nil {
		return nil, fmt.Errorf("couldn't open sqlite database: %w", err)
	}
	return db, nil
}

// GORM 日志适配器
type GormLogger struct {
	logger *zap.Logger
}

// GormWrapLogger 创建GORM日志包装器
func GormWrapLogger(logger *zap.Logger) logger.Interface {
	return &GormLogger{logger: logger}
}

// GormErrorLogger 创建只记录错误的GORM日志器
func GormErrorLogger(zapLogger *zap.Logger) logger.Interface {
	gormLogger := &GormLogger{logger: zapLogger}
	return gormLogger.LogMode(logger.Error) // Error level
}

func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := &GormLogger{logger: l.logger}
	return newLogger
}

func (l *GormLogger) Info(ctx context.Context, format string, args ...interface{}) {
	l.logger.Info(fmt.Sprintf(format, args...))
}

func (l *GormLogger) Warn(ctx context.Context, format string, args ...interface{}) {
	l.logger.Warn(fmt.Sprintf(format, args...))
}

func (l *GormLogger) Error(ctx context.Context, format string, args ...interface{}) {
	l.logger.Error(fmt.Sprintf(format, args...))
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	sql, rows := fc()
	elapsed := time.Since(begin)

	sugar := l.logger.Sugar() // 从 zap.Logger 得到 SugaredLogger

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		sugar.Errorf("query failed | sql=%s | rows=%d | elapsed=%s | err=%v",
			sql, rows, elapsed, err)
	} else {
		sugar.Debugf("query | sql=%s | rows=%d | elapsed=%s",
			sql, rows, elapsed)
	}
}
