package orz

import (
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"strings"
	"testing"
)

func TestConnectDatabaseRequiresRegisteredDriver(t *testing.T) {
	withIsolatedDatabaseDrivers(t)

	_, err := ConnectDatabase(DatabaseConfig{Type: DatabaseSqlite})
	if err == nil {
		t.Fatal("expected error when no database driver is registered")
	}
	if !strings.Contains(err.Error(), "not registered") {
		t.Fatalf("expected not registered error, got %v", err)
	}
}

func TestConnectDatabaseUsesRegisteredDriver(t *testing.T) {
	withIsolatedDatabaseDrivers(t)

	called := false
	RegisterDatabaseDriver(func(cfg DatabaseConfig, logger gormlogger.Interface) (*gorm.DB, error) {
		called = true
		if cfg.Type != DatabaseMysql {
			t.Fatalf("unexpected database type: %s", cfg.Type)
		}
		return &gorm.DB{}, nil
	}, DatabaseMysql)

	db, err := ConnectDatabase(DatabaseConfig{Type: DatabaseMysql})
	if err != nil {
		t.Fatalf("ConnectDatabase returned error: %v", err)
	}
	if db == nil {
		t.Fatal("expected non-nil db")
	}
	if !called {
		t.Fatal("expected registered opener to be called")
	}
}

func withIsolatedDatabaseDrivers(t *testing.T) {
	t.Helper()

	databaseDriversMu.Lock()
	saved := databaseDrivers
	databaseDrivers = map[DatabaseType]DatabaseOpener{}
	databaseDriversMu.Unlock()

	t.Cleanup(func() {
		databaseDriversMu.Lock()
		databaseDrivers = saved
		databaseDriversMu.Unlock()
	})
}
