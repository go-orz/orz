package postgres

import (
	"strings"
	"testing"

	"github.com/go-orz/orz"
)

func TestBuildDSNDoesNotForceTimezone(t *testing.T) {
	dsn := buildDSN(orz.DatabaseConfig{
		Postgres: orz.PostgresCfg{
			Hostname: "localhost",
			Port:     5432,
			Username: "postgres",
			Password: "secret",
			Database: "app",
		},
	})

	if strings.Contains(strings.ToLower(dsn), "timezone=") {
		t.Fatalf("dsn should not force timezone: %s", dsn)
	}
}
