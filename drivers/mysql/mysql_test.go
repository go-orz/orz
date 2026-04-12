package mysql

import (
	"strings"
	"testing"

	"github.com/go-orz/orz"
)

func TestBuildDSNDoesNotForceLocalTimezone(t *testing.T) {
	dsn := buildDSN(orz.DatabaseConfig{
		Mysql: orz.MysqlCfg{
			Hostname: "localhost",
			Port:     3306,
			Username: "root",
			Password: "secret",
			Database: "app",
		},
	})

	if strings.Contains(strings.ToLower(dsn), "loc=local") {
		t.Fatalf("dsn should not force local timezone: %s", dsn)
	}
}
