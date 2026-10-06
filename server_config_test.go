package orz

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
)

type serverLimits struct {
	readHeader, read, write, idle time.Duration
	maxHeaderBytes                int
}

func limitsOf(s *http.Server) serverLimits {
	return serverLimits{s.ReadHeaderTimeout, s.ReadTimeout, s.WriteTimeout, s.IdleTimeout, s.MaxHeaderBytes}
}

// Capture defaults from Echo itself, rather than duplicating its defaults here.
func echoServerLimits(t *testing.T, config *ServerConfig) serverLimits {
	t.Helper()
	var limits serverLimits
	stop := errors.New("captured server configuration")
	start := echo.StartConfig{
		Address: "127.0.0.1:0", HideBanner: true, HidePort: true,
		BeforeServeFunc: func(s *http.Server) error {
			if config != nil {
				config.configureHTTPServer(s)
			}
			limits = limitsOf(s)
			return stop
		},
	}
	if err := start.Start(context.Background(), echo.New()); !errors.Is(err, stop) {
		t.Fatalf("Start: %v", err)
	}
	return limits
}

func TestServerLimitsConfiguration(t *testing.T) {
	defaults := echoServerLimits(t, nil)
	for _, tc := range []struct {
		name     string
		values   map[string]interface{}
		expected serverLimits
	}{
		{"omitted", map[string]interface{}{}, defaults},
		{"zero preserves Echo defaults", map[string]interface{}{
			"read_header_timeout": 0, "read_timeout": 0,
			"write_timeout": 0, "idle_timeout": 0, "max_header_bytes": 0,
		}, defaults},
		{"negative preserves Echo defaults", map[string]interface{}{
			"read_header_timeout": -1, "read_timeout": -1,
			"write_timeout": -1, "idle_timeout": -1, "max_header_bytes": -1,
		}, defaults},
		{"configured", map[string]interface{}{
			"read_header_timeout": "5s", "read_timeout": "2m",
			"write_timeout": "1m", "idle_timeout": "90s", "max_header_bytes": 65536,
		}, serverLimits{5 * time.Second, 120 * time.Second, 60 * time.Second, 90 * time.Second, 65536}},
		{"read only preserves other defaults", map[string]interface{}{"read_timeout": "2m"},
			serverLimits{defaults.readHeader, 120 * time.Second, defaults.write, defaults.idle, defaults.maxHeaderBytes}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			app := NewApp()
			defer app.cancel()
			if err := app.LoadConfigFromMap(map[string]interface{}{"server": tc.values}); err != nil {
				t.Fatal(err)
			}
			cfg := app.GetConfig().Server
			got := echoServerLimits(t, &cfg)
			if !reflect.DeepEqual(got, tc.expected) {
				t.Fatalf("limits = %+v, want %+v", got, tc.expected)
			}
		})
	}
}

func TestServerDurationYAML(t *testing.T) {
	app := NewApp()
	defer app.cancel()
	if err := app.LoadConfigFromBytes([]byte(`server:
  read_header_timeout: 500ms
  read_timeout: 2m
  write_timeout: 1m30s
  idle_timeout: 0s
  max_header_bytes: 65536
`)); err != nil {
		t.Fatal(err)
	}
	cfg := app.GetConfig()
	if cfg == nil {
		t.Fatal("duration configuration could not be decoded")
	}
	want := serverLimits{500 * time.Millisecond, 2 * time.Minute, 90 * time.Second, 0, 65536}
	s := &http.Server{}
	cfg.Server.configureHTTPServer(s)
	if got := limitsOf(s); got != want {
		t.Fatalf("limits = %+v, want %+v", got, want)
	}
}
