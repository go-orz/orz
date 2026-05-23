package orz

import "testing"

func TestLoadConfigFromMapSupportsSnakeCaseFields(t *testing.T) {
	app := NewApp()
	if err := app.LoadConfigFromMap(map[string]interface{}{
		"database": map[string]interface{}{
			"enabled":  true,
			"type":     "sqlite",
			"show_sql": true,
			"sqlite": map[string]interface{}{
				"path": ":memory:",
			},
		},
		"server": map[string]interface{}{
			"ip_extractor":  "x-forwarded-for",
			"ip_trust_list": []string{"10.0.0.0/8"},
		},
	}); err != nil {
		t.Fatalf("LoadConfigFromMap returned error: %v", err)
	}

	cfg := app.GetConfig()
	if cfg == nil {
		t.Fatal("expected config")
	}
	if !cfg.Database.ShowSql {
		t.Fatal("expected database.show_sql to be loaded")
	}
	if cfg.Server.IPExtractor != "x-forwarded-for" {
		t.Fatalf("expected ip_extractor to be loaded, got %q", cfg.Server.IPExtractor)
	}
	if len(cfg.Server.IPTrustList) != 1 || cfg.Server.IPTrustList[0] != "10.0.0.0/8" {
		t.Fatalf("expected ip_trust_list to be loaded, got %v", cfg.Server.IPTrustList)
	}
}

func TestLoadConfigFromBytesSupportsFieldNameKeys(t *testing.T) {
	app := NewApp()
	err := app.LoadConfigFromBytes([]byte(`
DATABASE:
  Enabled: true
  Type: sqlite
  ShowSql: true
  URL: file:test.db
SERVER:
  Addr: :9090
  TLS:
    Enabled: true
    Auto: true
  IPExtractor: x-real-ip
  IPTrustList:
    - 192.168.0.0/16
LOG:
  MaxSize: 42
APP:
  feature_flag: enabled
`))
	if err != nil {
		t.Fatalf("LoadConfigFromBytes returned error: %v", err)
	}

	cfg := app.GetConfig()
	if cfg == nil {
		t.Fatal("expected config")
	}
	if !cfg.Database.ShowSql {
		t.Fatal("expected Database.ShowSql to be loaded")
	}
	if cfg.Database.URL != "file:test.db" {
		t.Fatalf("expected Database.URL to be loaded, got %q", cfg.Database.URL)
	}
	if cfg.Server.Addr != ":9090" {
		t.Fatalf("expected Server.Addr to be loaded, got %q", cfg.Server.Addr)
	}
	if !cfg.Server.TLS.Enabled || !cfg.Server.TLS.Auto {
		t.Fatalf("expected Server.TLS to be loaded, got %+v", cfg.Server.TLS)
	}
	if cfg.Server.IPExtractor != "x-real-ip" {
		t.Fatalf("expected Server.IPExtractor to be loaded, got %q", cfg.Server.IPExtractor)
	}
	if len(cfg.Server.IPTrustList) != 1 || cfg.Server.IPTrustList[0] != "192.168.0.0/16" {
		t.Fatalf("expected Server.IPTrustList to be loaded, got %v", cfg.Server.IPTrustList)
	}
	if cfg.Log.MaxSize != 42 {
		t.Fatalf("expected Log.MaxSize to be loaded, got %d", cfg.Log.MaxSize)
	}
	if cfg.App["feature_flag"] != "enabled" {
		t.Fatalf("expected App.feature_flag key to be preserved, got %v", cfg.App)
	}
}

func TestLoadConfigFromBytesSupportsUppercaseSnakeCaseKeys(t *testing.T) {
	app := NewApp()
	err := app.LoadConfigFromBytes([]byte(`
DATABASE:
  ENABLED: true
  TYPE: sqlite
  SHOW_SQL: true
SERVER:
  IP_EXTRACTOR: x-forwarded-for
  IP_TRUST_LIST:
    - 10.0.0.0/8
LOG:
  MAX_SIZE: 64
`))
	if err != nil {
		t.Fatalf("LoadConfigFromBytes returned error: %v", err)
	}

	cfg := app.GetConfig()
	if cfg == nil {
		t.Fatal("expected config")
	}
	if !cfg.Database.ShowSql {
		t.Fatal("expected DATABASE.SHOW_SQL to be loaded")
	}
	if cfg.Server.IPExtractor != "x-forwarded-for" {
		t.Fatalf("expected SERVER.IP_EXTRACTOR to be loaded, got %q", cfg.Server.IPExtractor)
	}
	if len(cfg.Server.IPTrustList) != 1 || cfg.Server.IPTrustList[0] != "10.0.0.0/8" {
		t.Fatalf("expected SERVER.IP_TRUST_LIST to be loaded, got %v", cfg.Server.IPTrustList)
	}
	if cfg.Log.MaxSize != 64 {
		t.Fatalf("expected LOG.MAX_SIZE to be loaded, got %d", cfg.Log.MaxSize)
	}
}

func TestLoadConfigFromBytesSupportsUppercaseFieldNameKeys(t *testing.T) {
	app := NewApp()
	err := app.LoadConfigFromBytes([]byte(`
DATABASE:
  ENABLED: true
  TYPE: sqlite
  SHOWSQL: true
SERVER:
  IPEXTRACTOR: x-real-ip
  IPTRUSTLIST:
    - 172.16.0.0/12
LOG:
  MAXSIZE: 128
`))
	if err != nil {
		t.Fatalf("LoadConfigFromBytes returned error: %v", err)
	}

	cfg := app.GetConfig()
	if cfg == nil {
		t.Fatal("expected config")
	}
	if !cfg.Database.ShowSql {
		t.Fatal("expected DATABASE.SHOWSQL to be loaded")
	}
	if cfg.Server.IPExtractor != "x-real-ip" {
		t.Fatalf("expected SERVER.IPEXTRACTOR to be loaded, got %q", cfg.Server.IPExtractor)
	}
	if len(cfg.Server.IPTrustList) != 1 || cfg.Server.IPTrustList[0] != "172.16.0.0/12" {
		t.Fatalf("expected SERVER.IPTRUSTLIST to be loaded, got %v", cfg.Server.IPTrustList)
	}
	if cfg.Log.MaxSize != 128 {
		t.Fatalf("expected LOG.MAXSIZE to be loaded, got %d", cfg.Log.MaxSize)
	}
}
