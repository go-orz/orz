package postgres

import (
	"fmt"

	"github.com/go-orz/orz"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func init() {
	orz.RegisterDatabaseDriver(open, orz.DatabasePostgres, orz.DatabasePostgresql)
}

func open(cfg orz.DatabaseConfig, logger gormlogger.Interface) (*gorm.DB, error) {
	db, err := gorm.Open(gormpostgres.Open(buildDSN(cfg)), &gorm.Config{
		Logger: logger,
	})
	if err != nil {
		return nil, fmt.Errorf("couldn't open postgres database: %w", err)
	}

	return db, nil
}

func buildDSN(cfg orz.DatabaseConfig) string {
	if cfg.URL != "" {
		return cfg.URL
	}

	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
		cfg.Postgres.Hostname,
		cfg.Postgres.Username,
		cfg.Postgres.Password,
		cfg.Postgres.Database,
		cfg.Postgres.Port,
	)
}
