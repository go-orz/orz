package database

import (
	"fmt"
	"github.com/go-orz/orz/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func MustConnectMysql(cfg config.MysqlCfg, logger logger.Interface) (db *gorm.DB) {
	var err error
	if cfg.DSN == "" {
		cfg.DSN = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=60s",
			cfg.Username,
			cfg.Password,
			cfg.Hostname,
			cfg.Port,
			cfg.Database,
		)
	}

	db, err = gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{
		Logger: logger,
	})
	if err != nil {
		panic(fmt.Sprintf("couldn't open database: %v", err))
	}
	return db
}
