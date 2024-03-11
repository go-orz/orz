package database

import (
	"fmt"
	"github.com/go-orz/orz/config"
	"gorm.io/driver/clickhouse"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func MustConnectClickHouse(cfg config.ClickHouseConfig, logger logger.Interface) (db *gorm.DB) {
	dsn := fmt.Sprintf("clickhouse://%s:%s@%s:%d/%s?dial_timeout=10s&read_timeout=20s",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database,
	)
	db, err := gorm.Open(clickhouse.New(clickhouse.Config{
		DSN:                          dsn,
		Conn:                         nil,      // initialize with existing database conn
		DisableDatetimePrecision:     true,     // disable datetime64 precision, not supported before clickhouse 20.4
		DontSupportRenameColumn:      true,     // rename column not supported before clickhouse 20.4
		DontSupportEmptyDefaultValue: false,    // do not consider empty strings as valid default values
		SkipInitializeWithVersion:    false,    // smart configure based on used version
		DefaultGranularity:           3,        // 1 granule = 8192 rows
		DefaultCompression:           "LZ4",    // default compression algorithm. LZ4 is lossless
		DefaultIndexType:             "minmax", // index stores extremes of the expression
		DefaultTableEngineOpts:       "ENGINE=MergeTree() ORDER BY tuple()",
	}), &gorm.Config{
		Logger:                 logger,
		SkipDefaultTransaction: false,
	})
	if err != nil {
		panic("failed to connect database")
	}
	return db
}
