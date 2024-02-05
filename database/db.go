package database

import (
	"github.com/go-orz/orz/config"
	"gorm.io/gorm"
)

func MustConnectDatabase(cfg config.Database) (db *gorm.DB) {
	switch cfg.Type {
	case "mysql":
		return MustConnectMysql(cfg.Mysql)
	case "clickhouse":
		return MustConnectClickHouse(cfg.ClickHouse)
	case "sqlite":
		return MustConnectSqlite(cfg.Sqlite)
	default:
		panic(`Unknown database type: ` + cfg.Type)
	}
	return db
}
