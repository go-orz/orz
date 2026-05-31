package orz

import (
	"net"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestRunWithoutHTTPBlocksUntilCancel(t *testing.T) {
	app := NewApp()
	app.SetLogger(zap.NewNop())

	done := make(chan error, 1)
	go func() {
		done <- app.Run()
	}()

	select {
	case err := <-done:
		t.Fatalf("Run returned early: %v", err)
	case <-time.After(100 * time.Millisecond):
	}

	app.cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run returned error: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Run did not exit after cancel")
	}
}

func TestRunReturnsHTTPStartError(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen returned error: %v", err)
	}
	defer listener.Close()

	app := NewApp()
	app.SetLogger(zap.NewNop())
	if err := app.LoadConfigFromMap(map[string]interface{}{
		"server": map[string]interface{}{
			"addr": listener.Addr().String(),
		},
	}); err != nil {
		t.Fatalf("LoadConfigFromMap returned error: %v", err)
	}
	app.EnableHTTP()

	err = app.Run()
	if err == nil {
		t.Fatal("expected Run to return server start error")
	}
}

func TestEnableHTTPUsesXForwardedForWithoutTrustedProxies(t *testing.T) {
	app := NewApp()
	app.SetLogger(zap.NewNop())
	if err := app.LoadConfigFromMap(map[string]interface{}{
		"server": map[string]interface{}{
			"ip_extractor": "x-forwarded-for",
		},
	}); err != nil {
		t.Fatalf("LoadConfigFromMap returned error: %v", err)
	}

	app.EnableHTTP()
	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.RemoteAddr = "10.0.0.2:1234"
	req.Header.Set("X-Forwarded-For", "203.0.113.9")

	ip := app.GetEcho().IPExtractor(req)
	if ip != "203.0.113.9" {
		t.Fatalf("expected forwarded IP extraction, got %q", ip)
	}
}

func TestEnableHTTPUsesTrustedProxyHeadersWhenConfigured(t *testing.T) {
	app := NewApp()
	app.SetLogger(zap.NewNop())
	if err := app.LoadConfigFromMap(map[string]interface{}{
		"server": map[string]interface{}{
			"ip_extractor":  "x-forwarded-for",
			"ip_trust_list": []string{"10.0.0.0/8"},
		},
	}); err != nil {
		t.Fatalf("LoadConfigFromMap returned error: %v", err)
	}

	app.EnableHTTP()
	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.RemoteAddr = "10.0.0.2:1234"
	req.Header.Set("X-Forwarded-For", "203.0.113.9, 10.0.0.2")

	ip := app.GetEcho().IPExtractor(req)
	if ip != "203.0.113.9" {
		t.Fatalf("expected forwarded IP extraction, got %q", ip)
	}
}

func TestEnableHTTPTrustsSingleProxyIP(t *testing.T) {
	app := NewApp()
	app.SetLogger(zap.NewNop())
	if err := app.LoadConfigFromMap(map[string]interface{}{
		"server": map[string]interface{}{
			"ip_extractor":  "x-forwarded-for",
			"ip_trust_list": []string{"127.0.0.1"},
		},
	}); err != nil {
		t.Fatalf("LoadConfigFromMap returned error: %v", err)
	}

	app.EnableHTTP()
	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	req.Header.Set("X-Forwarded-For", "203.0.113.9")

	ip := app.GetEcho().IPExtractor(req)
	if ip != "203.0.113.9" {
		t.Fatalf("expected forwarded IP extraction, got %q", ip)
	}
}

func TestEnableHTTPTrustsConfiguredProxyCIDR(t *testing.T) {
	app := NewApp()
	app.SetLogger(zap.NewNop())
	if err := app.LoadConfigFromMap(map[string]interface{}{
		"server": map[string]interface{}{
			"ip_extractor":  "x-forwarded-for",
			"ip_trust_list": []string{"172.16.0.0/12"},
		},
	}); err != nil {
		t.Fatalf("LoadConfigFromMap returned error: %v", err)
	}

	app.EnableHTTP()
	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.RemoteAddr = "172.18.0.1:1234"
	req.Header.Set("X-Forwarded-For", "198.51.100.77")

	ip := app.GetEcho().IPExtractor(req)
	if ip != "198.51.100.77" {
		t.Fatalf("expected forwarded IP extraction, got %q", ip)
	}
}

func TestEnableHTTPDoesNotTrustUnconfiguredPrivateProxyCIDR(t *testing.T) {
	app := NewApp()
	app.SetLogger(zap.NewNop())
	if err := app.LoadConfigFromMap(map[string]interface{}{
		"server": map[string]interface{}{
			"ip_extractor":  "x-forwarded-for",
			"ip_trust_list": []string{"10.0.0.0/8"},
		},
	}); err != nil {
		t.Fatalf("LoadConfigFromMap returned error: %v", err)
	}

	app.EnableHTTP()
	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.RemoteAddr = "172.18.0.1:1234"
	req.Header.Set("X-Forwarded-For", "198.51.100.77")

	ip := app.GetEcho().IPExtractor(req)
	if ip != "172.18.0.1" {
		t.Fatalf("expected direct IP extraction, got %q", ip)
	}
}

func TestEnableHTTPUsesCustomIPHeader(t *testing.T) {
	app := NewApp()
	app.SetLogger(zap.NewNop())
	if err := app.LoadConfigFromMap(map[string]interface{}{
		"server": map[string]interface{}{
			"ip_extractor": "X-Client-IP",
		},
	}); err != nil {
		t.Fatalf("LoadConfigFromMap returned error: %v", err)
	}

	app.EnableHTTP()
	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.RemoteAddr = "10.0.0.2:1234"
	req.Header.Set("X-Client-IP", "203.0.113.9")

	ip := app.GetEcho().IPExtractor(req)
	if ip != "203.0.113.9" {
		t.Fatalf("expected custom header IP extraction, got %q", ip)
	}
}

func TestEnableHTTPCustomIPHeaderFallsBackToDirectWhenHeaderMissing(t *testing.T) {
	app := NewApp()
	app.SetLogger(zap.NewNop())
	if err := app.LoadConfigFromMap(map[string]interface{}{
		"server": map[string]interface{}{
			"ip_extractor": "X-Client-IP",
		},
	}); err != nil {
		t.Fatalf("LoadConfigFromMap returned error: %v", err)
	}

	app.EnableHTTP()
	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.RemoteAddr = "10.0.0.2:1234"

	ip := app.GetEcho().IPExtractor(req)
	if ip != "10.0.0.2" {
		t.Fatalf("expected direct IP extraction, got %q", ip)
	}
}

func TestEnableHTTPUsesDirectIPExtractionWhenExtractorIsBlank(t *testing.T) {
	app := NewApp()
	app.SetLogger(zap.NewNop())
	if err := app.LoadConfigFromMap(map[string]interface{}{
		"server": map[string]interface{}{
			"ip_extractor": " ",
		},
	}); err != nil {
		t.Fatalf("LoadConfigFromMap returned error: %v", err)
	}

	app.EnableHTTP()
	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.RemoteAddr = "10.0.0.2:1234"
	req.Header.Set("X-Forwarded-For", "203.0.113.9")

	ip := app.GetEcho().IPExtractor(req)
	if ip != "10.0.0.2" {
		t.Fatalf("expected direct IP extraction, got %q", ip)
	}
}

func TestNewIPExtractorUsesExtractorAndTrustList(t *testing.T) {
	extractor := NewIPExtractor("x-forwarded-for", []string{"10.0.0.0/8"})

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.RemoteAddr = "10.0.0.2:1234"
	req.Header.Set("X-Forwarded-For", "203.0.113.9, 10.0.0.2")

	ip := extractor(req)
	if ip != "203.0.113.9" {
		t.Fatalf("expected forwarded IP extraction, got %q", ip)
	}
}

func TestNewIPExtractorUsesCustomHeader(t *testing.T) {
	extractor := NewIPExtractor("X-Client-IP", nil)

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.RemoteAddr = "10.0.0.2:1234"
	req.Header.Set("X-Client-IP", "203.0.113.9")

	ip := extractor(req)
	if ip != "203.0.113.9" {
		t.Fatalf("expected custom header IP extraction, got %q", ip)
	}
}

func TestExtractClientIPs(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.RemoteAddr = "10.0.0.2:1234"
	req.Header.Set("X-Real-IP", "198.51.100.77")
	req.Header.Set("X-Forwarded-For", "203.0.113.9, 10.0.0.2")

	ips := ExtractClientIPs(req, []string{"10.0.0.0/8", "198.51.100.0/24"})

	if ips.Direct != "10.0.0.2" {
		t.Fatalf("expected direct IP extraction, got %q", ips.Direct)
	}
	if ips.XRealIP != "198.51.100.77" {
		t.Fatalf("expected X-Real-IP extraction, got %q", ips.XRealIP)
	}
	if ips.XForwardedFor != "203.0.113.9" {
		t.Fatalf("expected X-Forwarded-For extraction, got %q", ips.XForwardedFor)
	}
}

func TestExtractClientIPMap(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.RemoteAddr = "10.0.0.2:1234"
	req.Header.Set("X-Real-IP", "198.51.100.77")
	req.Header.Set("X-Forwarded-For", "203.0.113.9, 10.0.0.2")

	ips := ExtractClientIPMap(req, []string{"10.0.0.0/8", "198.51.100.0/24"})

	if ips["direct"] != "10.0.0.2" {
		t.Fatalf("expected direct IP extraction, got %q", ips["direct"])
	}
	if ips["x-real-ip"] != "198.51.100.77" {
		t.Fatalf("expected X-Real-IP extraction, got %q", ips["x-real-ip"])
	}
	if ips["x-forwarded-for"] != "203.0.113.9" {
		t.Fatalf("expected X-Forwarded-For extraction, got %q", ips["x-forwarded-for"])
	}
}
