package orz

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
)

const (
	ipExtractorDirect        = "direct"
	ipExtractorXForwardedFor = "x-forwarded-for"
	ipExtractorXRealIP       = "x-real-ip"
)

func configureIPExtractor(e *echo.Echo, server ServerConfig, logger *zap.Logger) {
	trustOptions, trustedProxies := buildTrustedProxyOptions(server.IPTrustList, logger)
	if len(trustedProxies) > 0 {
		logger.Info("trusted proxy IP/CIDR list configured", zap.Strings("trustedProxies", trustedProxies))
	}

	switch extractor := normalizeIPExtractor(server.IPExtractor); strings.ToLower(extractor) {
	case ipExtractorXForwardedFor:
		e.IPExtractor = echo.ExtractIPFromXFFHeader(trustOptions...)
	case ipExtractorXRealIP:
		e.IPExtractor = echo.ExtractIPFromRealIPHeader(trustOptions...)
	case ipExtractorDirect:
		e.IPExtractor = echo.ExtractIPDirect()
	default:
		e.IPExtractor = extractIPFromHeader(extractor, trustOptions...)
	}
}

func buildTrustedProxyOptions(ipTrustList []string, logger *zap.Logger) ([]echo.TrustOption, []string) {
	if len(ipTrustList) == 0 {
		return nil, nil
	}

	options := []echo.TrustOption{
		echo.TrustLoopback(false),
		echo.TrustLinkLocal(false),
		echo.TrustPrivateNet(false),
	}

	trustedProxies := make([]string, 0, len(ipTrustList))
	for _, value := range ipTrustList {
		ipRange, err := parseTrustedProxyIPRange(value)
		if err != nil {
			logger.Warn("failed to parse trusted proxy IP/CIDR", zap.String("ip", value), zap.Error(err))
			continue
		}

		options = append(options, echo.TrustIPRange(ipRange))
		trustedProxies = append(trustedProxies, strings.TrimSpace(value))
	}

	return options, trustedProxies
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
