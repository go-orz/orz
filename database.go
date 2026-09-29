package orz

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// ConnectDatabase 连接数据库
func ConnectDatabase(cfg DatabaseConfig) (*gorm.DB, error) {
	return connectDatabaseWithGormLogger(cfg, nil)
}

// ConnectDatabaseWithLogger 连接数据库并指定日志器
func ConnectDatabaseWithLogger(cfg DatabaseConfig, zapLogger *zap.Logger) (*gorm.DB, error) {
	var wrapLogger gormlogger.Interface
	if cfg.ShowSql && zapLogger != nil {
		wrapLogger = GormWrapLogger(zapLogger)
	} else if zapLogger != nil {
		wrapLogger = GormErrorLogger(zapLogger)
	} else {
		wrapLogger = gormlogger.Default.LogMode(gormlogger.Silent)
	}

	return connectDatabaseWithGormLogger(cfg, wrapLogger)
}

// ConnectMysql 连接MySQL数据库
// 需要先导入对应驱动包，例如：
// import _ "github.com/go-orz/orz/drivers/mysql"
func ConnectMysql(url string, cfg MysqlCfg, logger gormlogger.Interface) (*gorm.DB, error) {
	return connectDatabaseWithGormLogger(DatabaseConfig{
		Type:  DatabaseMysql,
		URL:   url,
		Mysql: cfg,
	}, logger)
}

// ConnectPostgresql 连接PostgreSQL数据库
// 需要先导入对应驱动包，例如：
// import _ "github.com/go-orz/orz/drivers/postgres"
func ConnectPostgresql(url string, cfg PostgresCfg, logger gormlogger.Interface) (*gorm.DB, error) {
	return connectDatabaseWithGormLogger(DatabaseConfig{
		Type:     DatabasePostgres,
		URL:      url,
		Postgres: cfg,
	}, logger)
}

// ConnectSqlite 连接SQLite数据库
// 需要先导入对应驱动包，例如：
// import _ "github.com/go-orz/orz/drivers/sqlite"
func ConnectSqlite(url string, cfg SqliteConfig, logger gormlogger.Interface) (*gorm.DB, error) {
	return connectDatabaseWithGormLogger(DatabaseConfig{
		Type:   DatabaseSqlite,
		URL:    url,
		Sqlite: cfg,
	}, logger)
}

func connectDatabaseWithGormLogger(cfg DatabaseConfig, logger gormlogger.Interface) (*gorm.DB, error) {
	pool, err := resolveDatabasePool(cfg.Pool)
	if err != nil {
		return nil, err
	}
	db, err := openDatabase(cfg, logger)
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get database connection pool: %w", err)
	}
	sqlDB.SetMaxOpenConns(pool[0])
	sqlDB.SetMaxIdleConns(pool[1])
	sqlDB.SetConnMaxLifetime(time.Duration(pool[2]) * time.Second)
	sqlDB.SetConnMaxIdleTime(time.Duration(pool[3]) * time.Second)
	return db, nil
}

func resolveDatabasePool(cfg DatabasePoolConfig) ([4]int, error) {
	values := [4]int{30, 10, 3600, 300}
	configured := []int{cfg.MaxOpenConns, cfg.MaxIdleConns, cfg.ConnMaxLifetimeSeconds, cfg.ConnMaxIdleTimeSeconds}
	names := []string{"max_open_conns", "max_idle_conns", "conn_max_lifetime_seconds", "conn_max_idle_time_seconds"}
	for index, value := range configured {
		if value == 0 {
			continue
		}
		if value < 0 || (index >= 2 && int64(value) > int64(time.Duration(1<<63-1)/time.Second)) {
			return values, fmt.Errorf("invalid database.pool.%s: %d", names[index], value)
		}
		values[index] = value
	}
	if values[0] > 0 && values[1] > values[0] {
		values[1] = values[0]
	}
	return values, nil
}

// GormLogger GORM 日志适配器
type GormLogger struct {
	logger *zap.Logger
}

// GormWrapLogger 创建GORM日志包装器
func GormWrapLogger(logger *zap.Logger) gormlogger.Interface {
	return &GormLogger{logger: logger}
}

// GormErrorLogger 创建只记录错误的GORM日志器
func GormErrorLogger(zapLogger *zap.Logger) gormlogger.Interface {
	gormLogger := &GormLogger{logger: zapLogger}
	return gormLogger.LogMode(gormlogger.Error)
}

func (l *GormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	switch level {
	case gormlogger.Silent:
		l.logger = l.logger.WithOptions(zap.IncreaseLevel(zapcore.FatalLevel))
	case gormlogger.Error:
		l.logger = l.logger.WithOptions(zap.IncreaseLevel(zapcore.ErrorLevel))
	case gormlogger.Warn:
		l.logger = l.logger.WithOptions(zap.IncreaseLevel(zapcore.WarnLevel))
	case gormlogger.Info:
		l.logger = l.logger.WithOptions(zap.IncreaseLevel(zapcore.InfoLevel))
	}

	return l
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

	sugar := l.logger.Sugar()

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		sugar.Errorf("query failed | sql=%s | rows=%d | elapsed=%s | err=%v",
			sql, rows, elapsed, err)
	} else {
		sugar.Debugf("query | sql=%s | rows=%d | elapsed=%s",
			sql, rows, elapsed)
	}
}
