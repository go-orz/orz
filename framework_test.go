package orz

import (
	"testing"

	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type frameworkOrderApp struct {
	configured bool
	hasConfig  bool
	hasDB      bool
	hasHTTP    bool
}

func (a *frameworkOrderApp) Configure(app *App) error {
	a.configured = true
	a.hasConfig = app.GetConfig() != nil
	a.hasDB = app.GetDatabase() != nil
	a.hasHTTP = app.GetEcho() != nil
	return nil
}

func TestNewFrameworkInitializesDependenciesBeforeApplication(t *testing.T) {
	withIsolatedDatabaseDrivers(t)
	RegisterDatabaseDriver(func(cfg DatabaseConfig, logger gormlogger.Interface) (*gorm.DB, error) {
		return &gorm.DB{}, nil
	}, DatabaseType("stub"))

	application := &frameworkOrderApp{}
	framework, err := NewFramework(
		WithApplication(application),
		WithHTTP(),
		WithDatabase(),
		WithLoggerFromConfig(),
		WithConfigMap(map[string]interface{}{
			"database": map[string]interface{}{
				"enabled": true,
				"type":    "stub",
			},
			"server": map[string]interface{}{
				"addr": ":0",
			},
		}),
	)
	if err != nil {
		t.Fatalf("NewFramework returned error: %v", err)
	}
	if framework == nil {
		t.Fatal("expected framework")
	}
	if !application.configured || !application.hasConfig || !application.hasDB || !application.hasHTTP {
		t.Fatalf("application was not initialized with all dependencies: %+v", application)
	}
}
