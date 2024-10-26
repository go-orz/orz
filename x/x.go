package x

import (
	"fmt"
	"github.com/go-orz/orz/config"
	"github.com/labstack/echo/v4"
	"net"
	"strings"
)

func IPExtractor() echo.IPExtractor {
	ipExtractor := strings.ToLower(config.Conf().Server.IPExtractor)
	trustList := config.Conf().Server.IPTrustList

	var options = make([]echo.TrustOption, 0, len(trustList))
	for _, trust := range trustList {
		_, ipNet, err := net.ParseCIDR(trust)
		if err != nil {
			panic(fmt.Errorf("invalid trust option: %s", trust))
		}
		options = append(options, echo.TrustIPRange(ipNet))
	}

	switch ipExtractor {
	case "direct":
		return echo.ExtractIPDirect()
	case "x-real-ip":
		return echo.ExtractIPFromRealIPHeader(options...)
	case "x-forwarded-for":
		return echo.ExtractIPFromXFFHeader(options...)
	default:
		return echo.ExtractIPDirect()
	}
}
