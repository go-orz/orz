package database

import (
	"github.com/go-orz/orz/config"
	"github.com/go-orz/orz/log"
	"github.com/go-orz/orz/z"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func MustConnectDatabase(cfg config.Database) (db *gorm.DB) {
	var wrapLogger logger.Interface
	if cfg.ShowSql {
		wrapLogger = z.GormWrapLogger(log.Z())
	} else {
		wrapLogger = z.GormErrorLogger(log.Z())
	}

	switch cfg.Type {
	case "mysql":
		return MustConnectMysql(cfg.Mysql, wrapLogger)
	case "clickhouse":
		return MustConnectClickHouse(cfg.ClickHouse, wrapLogger)
	case "sqlite":
		return MustConnectSqlite(cfg.Sqlite, wrapLogger)
	case "postgresql":
		return MustConnectPostgresql(cfg.Postgresql, wrapLogger)
	default:
		panic(`Unknown database type: ` + cfg.Type)
	}
	return db
}
