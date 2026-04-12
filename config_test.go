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
