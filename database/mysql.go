package database

import (
	"fmt"
	"github.com/go-orz/orz/config"
	"github.com/go-orz/orz/log"
	"github.com/go-orz/orz/z"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func MustConnectMysql(cfg config.MysqlCfg) (db *gorm.DB) {
	var err error
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=60s",
		cfg.Username,
		cfg.Password,
		cfg.Hostname,
		cfg.Port,
		cfg.Database,
	)
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: z.LoggerWrap(log.Z()),
	})
	if err != nil {
		panic(fmt.Sprintf("couldn't open database: %v", err))
	}
	return db
}
