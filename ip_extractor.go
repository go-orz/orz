package orz

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"
)

const (
	ipExtractorDirect        = "direct"
	ipExtractorXForwardedFor = "x-forwarded-for"
	ipExtractorXRealIP       = "x-real-ip"
)

// ClientIPs contains client IP values extracted with each built-in strategy.
type ClientIPs struct {
	Direct        string `json:"direct"`
	XRealIP       string `json:"x-real-ip"`
	XForwardedFor string `json:"x-forwarded-for"`
}

// NewIPExtractor returns an Echo IP extractor from the configured extractor name and trusted proxy list.
func NewIPExtractor(ipExtractor string, ipTrustList []string) echo.IPExtractor {
	trustOptions := buildTrustedProxyOptions(ipTrustList)

	switch extractor := normalizeIPExtractor(ipExtractor); strings.ToLower(extractor) {
	case ipExtractorXForwardedFor:
		return echo.ExtractIPFromXFFHeader(trustOptions...)
	case ipExtractorXRealIP:
		return echo.ExtractIPFromRealIPHeader(trustOptions...)
	case ipExtractorDirect:
		return echo.ExtractIPDirect()
	default:
		return extractIPFromHeader(extractor, trustOptions...)
	}
}

// ExtractClientIPs extracts direct, X-Real-IP, and X-Forwarded-For IP values using the same trusted proxy list.
func ExtractClientIPs(req *http.Request, ipTrustList []string) ClientIPs {
	trustOptions := buildTrustedProxyOptions(ipTrustList)

	return ClientIPs{
		Direct:        echo.ExtractIPDirect()(req),
		XRealIP:       echo.ExtractIPFromRealIPHeader(trustOptions...)(req),
		XForwardedFor: echo.ExtractIPFromXFFHeader(trustOptions...)(req),
	}
}

// ExtractClientIPMap extracts direct, X-Real-IP, and X-Forwarded-For IP values using map keys suitable for JSON responses.
func ExtractClientIPMap(req *http.Request, ipTrustList []string) map[string]string {
	ips := ExtractClientIPs(req, ipTrustList)
	return map[string]string{
		ipExtractorDirect:        ips.Direct,
		ipExtractorXRealIP:       ips.XRealIP,
		ipExtractorXForwardedFor: ips.XForwardedFor,
	}
}

func buildTrustedProxyOptions(ipTrustList []string) []echo.TrustOption {
	if len(ipTrustList) == 0 {
		return nil
	}

	options := []echo.TrustOption{
		echo.TrustLoopback(false),
		echo.TrustLinkLocal(false),
		echo.TrustPrivateNet(false),
	}

	for _, value := range ipTrustList {
		ipRange, err := parseTrustedProxyIPRange(value)
		if err != nil {
			continue
		}

		options = append(options, echo.TrustIPRange(ipRange))
	}

	return options
}

func normalizeIPExtractor(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ipExtractorDirect
	}
	return value
}

func ensureDirectIPExtractor(e *echo.Echo) {
	if e.IPExtractor == nil {
		e.IPExtractor = echo.ExtractIPDirect()
	}
}

func extractIPFromHeader(header string, options ...echo.TrustOption) echo.IPExtractor {
	header = http.CanonicalHeaderKey(strings.TrimSpace(header))
	directExtractor := echo.ExtractIPDirect()
	xffExtractor := echo.ExtractIPFromXFFHeader(options...)

	return func(req *http.Request) string {
		values := req.Header.Values(header)
		if len(values) == 0 {
			return directExtractor(req)
		}

		headerRequest := new(http.Request)
		*headerRequest = *req
		headerRequest.Header = http.Header{
			echo.HeaderXForwardedFor: values,
		}
		return xffExtractor(headerRequest)
	}
}

func parseTrustedProxyIPRange(value string) (*net.IPNet, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, fmt.Errorf("empty IP range")
	}

	if strings.Contains(value, "/") {
		_, ipNet, err := net.ParseCIDR(value)
		if err != nil {
			return nil, fmt.Errorf("invalid CIDR: %w", err)
		}
		return ipNet, nil
	}

	ip := net.ParseIP(value)
	if ip == nil {
		return nil, fmt.Errorf("invalid IP or CIDR")
	}

	if ip4 := ip.To4(); ip4 != nil {
		return &net.IPNet{
			IP:   ip4,
			Mask: net.CIDRMask(32, 32),
		}, nil
	}

	return &net.IPNet{
		IP:   ip,
		Mask: net.CIDRMask(128, 128),
	}, nil
}
