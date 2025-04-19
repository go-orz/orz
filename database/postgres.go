package database

import (
	"fmt"
	"github.com/go-orz/orz/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func MustConnectPostgresql(cfg config.PostgresCfg, logger logger.Interface) (db *gorm.DB) {
	var err error
	if cfg.DSN == "" {
		cfg.DSN = fmt.Sprintf("user=%s password=%s host=%s port=%d dbname=%s sslmode=disable",
			cfg.Username,
			cfg.Password,
			cfg.Hostname,
			cfg.Port,
			cfg.Database,
		)
	}

	db, err = gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{
		Logger: logger,
	})
	if err != nil {
		panic(fmt.Sprintf("couldn't open database: %v", err))
	}
	return db
}
