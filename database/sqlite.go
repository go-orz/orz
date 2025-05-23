package database

import (
	"fmt"
	"github.com/go-orz/orz/config"
	"github.com/go-orz/orz/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"path/filepath"

	"github.com/glebarez/sqlite"
)

func MustConnectSqlite(cfg config.SqliteConfig, logger logger.Interface) (db *gorm.DB) {
	dir := filepath.Dir(cfg.Path)
	if err := utils.MkdirIfNotExists(dir); err != nil {
		panic(fmt.Sprintf("couldn't create sqlite db dir: %v", err))
	}
	var err error
	db, err = gorm.Open(sqlite.Open(cfg.Path), &gorm.Config{
		Logger: logger,
	})
	if err != nil {
		panic(fmt.Sprintf("couldn't open sqlite database: %v", err))
	}
	return db
}
