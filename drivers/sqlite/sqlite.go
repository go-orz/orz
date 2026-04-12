package sqlite

import (
	"fmt"

	gormsqlite "github.com/glebarez/sqlite"
	"github.com/go-orz/orz"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func init() {
	orz.RegisterDatabaseDriver(open, orz.DatabaseSqlite)
}

func open(cfg orz.DatabaseConfig, logger gormlogger.Interface) (*gorm.DB, error) {
	dsn := cfg.URL
	if dsn == "" {
		dsn = cfg.Sqlite.Path
	}

	db, err := gorm.Open(gormsqlite.Open(dsn), &gorm.Config{
		Logger: logger,
	})
	if err != nil {
		return nil, fmt.Errorf("couldn't open sqlite database: %w", err)
	}

	return db, nil
}
