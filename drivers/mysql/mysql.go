package mysql

import (
	"fmt"

	"github.com/go-orz/orz"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func init() {
	orz.RegisterDatabaseDriver(open, orz.DatabaseMysql)
}

func open(cfg orz.DatabaseConfig, logger gormlogger.Interface) (*gorm.DB, error) {
	db, err := gorm.Open(gormmysql.Open(buildDSN(cfg)), &gorm.Config{
		Logger: logger,
	})
	if err != nil {
		return nil, fmt.Errorf("couldn't open mysql database: %w", err)
	}

	return db, nil
}

func buildDSN(cfg orz.DatabaseConfig) string {
	if cfg.URL != "" {
		return cfg.URL
	}

	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True",
		cfg.Mysql.Username,
		cfg.Mysql.Password,
		cfg.Mysql.Hostname,
		cfg.Mysql.Port,
		cfg.Mysql.Database,
	)
}
