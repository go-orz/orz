package orz

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// DatabaseOpener 根据配置打开数据库连接。
type DatabaseOpener func(cfg DatabaseConfig, logger gormlogger.Interface) (*gorm.DB, error)

var (
	databaseDriversMu         sync.RWMutex
	databaseDrivers           = map[DatabaseType]DatabaseOpener{}
	databaseDriverImportHints = map[DatabaseType]string{
		DatabaseSqlite:     "github.com/go-orz/orz/drivers/sqlite",
		DatabaseMysql:      "github.com/go-orz/orz/drivers/mysql",
		DatabasePostgres:   "github.com/go-orz/orz/drivers/postgres",
		DatabasePostgresql: "github.com/go-orz/orz/drivers/postgres",
	}
)

// RegisterDatabaseDriver 注册数据库驱动。
// 通常在独立驱动包的 init 中调用，例如 drivers/sqlite。
func RegisterDatabaseDriver(opener DatabaseOpener, types ...DatabaseType) {
	if opener == nil {
		panic("orz: database opener is nil")
	}
	if len(types) == 0 {
		panic("orz: no database types provided")
	}

	databaseDriversMu.Lock()
	defer databaseDriversMu.Unlock()

	for _, databaseType := range types {
		if databaseType == "" {
			panic("orz: database type is empty")
		}
		if _, exists := databaseDrivers[databaseType]; exists {
			panic(fmt.Sprintf("orz: database driver already registered for type %q", databaseType))
		}
		databaseDrivers[databaseType] = opener
	}
}

// RegisteredDatabaseDrivers 返回当前已注册的数据库驱动类型。
func RegisteredDatabaseDrivers() []DatabaseType {
	databaseDriversMu.RLock()
	defer databaseDriversMu.RUnlock()

	drivers := make([]DatabaseType, 0, len(databaseDrivers))
	for databaseType := range databaseDrivers {
		drivers = append(drivers, databaseType)
	}

	sort.Slice(drivers, func(i, j int) bool {
		return drivers[i] < drivers[j]
	})

	return drivers
}

func openDatabase(cfg DatabaseConfig, logger gormlogger.Interface) (*gorm.DB, error) {
	if cfg.Type == "" {
		return nil, fmt.Errorf("database type is empty")
	}

	databaseDriversMu.RLock()
	opener, ok := databaseDrivers[cfg.Type]
	databaseDriversMu.RUnlock()
	if !ok {
		registered := RegisteredDatabaseDrivers()
		if hint, exists := databaseDriverImportHints[cfg.Type]; exists {
			return nil, fmt.Errorf("database driver %q is not registered; import %s", cfg.Type, hint)
		}
		if len(registered) == 0 {
			return nil, fmt.Errorf("database driver %q is not registered; import a driver package such as github.com/go-orz/orz/drivers/sqlite", cfg.Type)
		}

		available := make([]string, 0, len(registered))
		for _, driver := range registered {
			available = append(available, string(driver))
		}
		return nil, fmt.Errorf("database driver %q is not registered; registered drivers: %s", cfg.Type, strings.Join(available, ", "))
	}

	return opener(cfg, logger)
}
