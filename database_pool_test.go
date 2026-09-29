package orz

import (
	"testing"
)

func TestResolveDatabasePool(t *testing.T) {
	tests := []struct {
		name      string
		config    DatabasePoolConfig
		want      [4]int
		wantError bool
	}{
		{name: "defaults", want: [4]int{30, 10, 3600, 300}},
		{name: "idle capped", config: DatabasePoolConfig{MaxOpenConns: 2}, want: [4]int{2, 2, 3600, 300}},
		{name: "custom", config: DatabasePoolConfig{MaxOpenConns: 50, MaxIdleConns: 5, ConnMaxLifetimeSeconds: 60, ConnMaxIdleTimeSeconds: 15}, want: [4]int{50, 5, 60, 15}},
		{name: "negative open", config: DatabasePoolConfig{MaxOpenConns: -1}, wantError: true},
		{name: "negative idle", config: DatabasePoolConfig{MaxIdleConns: -1}, wantError: true},
		{name: "negative lifetime", config: DatabasePoolConfig{ConnMaxLifetimeSeconds: -1}, wantError: true},
		{name: "negative idle time", config: DatabasePoolConfig{ConnMaxIdleTimeSeconds: -1}, wantError: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := resolveDatabasePool(test.config)
			if (err != nil) != test.wantError {
				t.Fatalf("unexpected error: %v", err)
			}
			if err == nil && got != test.want {
				t.Fatalf("got %v, want %v", got, test.want)
			}
		})
	}
}

func TestDatabasePoolConfigDecoding(t *testing.T) {
	app := NewApp()
	if err := app.LoadConfigFromBytes([]byte(`
Database:
  Pool:
    MaxOpenConns: 12
    MaxIdleConns: 0
    ConnMaxLifetimeSeconds: 60
    ConnMaxIdleTimeSeconds: 15
`)); err != nil {
		t.Fatal(err)
	}
	got, err := resolveDatabasePool(app.GetConfig().Database.Pool)
	if err != nil || got != [4]int{12, 10, 60, 15} {
		t.Fatalf("unexpected pool config: %v, %v", got, err)
	}
}

func TestDatabasePoolOldConfigDefaults(t *testing.T) {
	app := NewApp()
	if err := app.LoadConfigFromBytes([]byte("Database:\n  Type: postgres\n")); err != nil {
		t.Fatal(err)
	}
	got, err := resolveDatabasePool(app.GetConfig().Database.Pool)
	if err != nil || got != [4]int{30, 10, 3600, 300} {
		t.Fatalf("unexpected defaults for old config: %v, %v", got, err)
	}
}
