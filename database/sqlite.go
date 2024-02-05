package database

import (
	"fmt"
	"github.com/go-orz/orz/config"
	"github.com/go-orz/orz/log"
	"github.com/go-orz/orz/z"
	"gorm.io/gorm"
	"path/filepath"

	"github.com/glebarez/sqlite"
)

func MustConnectSqlite(cfg config.SqliteConfig) (db *gorm.DB) {
	dir := filepath.Dir(cfg.Path)
	if err := z.MkdirIfNotExists(dir); err != nil {
		panic(fmt.Sprintf("couldn't create sqlite db dir: %v", err))
	}
	var err error
	db, err = gorm.Open(sqlite.Open(cfg.Path), &gorm.Config{
		Logger: z.LoggerWrap(log.Z()),
	})
	if err != nil {
		panic(fmt.Sprintf("couldn't open sqlite database: %v", err))
	}
	return db
}
