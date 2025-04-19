package database

import (
	"github.com/go-orz/orz/config"
	"github.com/go-orz/orz/log"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func MustConnectDatabase(cfg config.Database) (db *gorm.DB) {
	var wrapLogger logger.Interface
	if cfg.ShowSql {
		wrapLogger = GormWrapLogger(log.Z())
	} else {
		wrapLogger = GormErrorLogger(log.Z())
	}

	switch cfg.Type {
	case config.DatabaseMysql:
		return MustConnectMysql(cfg.Mysql, wrapLogger)
	case config.DatabaseClickhouse:
		return MustConnectClickHouse(cfg.ClickHouse, wrapLogger)
	case config.DatabaseSqlite:
		return MustConnectSqlite(cfg.Sqlite, wrapLogger)
	case config.DatabasePostgres, config.DatabasePostgresql:
		return MustConnectPostgresql(cfg.Postgres, wrapLogger)
	default:
		panic(`Unknown database type: ` + cfg.Type)
	}
	return db
}
