package x

import (
	"github.com/go-orz/orz/config"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"net"
	"strings"
)

func IPExtractor(logger *zap.Logger) echo.IPExtractor {
	ipExtractor := strings.ToLower(config.Conf().Server.IPExtractor)
	trustList := config.Conf().Server.IPTrustList

	var options = make([]echo.TrustOption, 0, len(trustList))
	for _, trust := range trustList {
		_, ipNet, err := net.ParseCIDR(trust)
		if err != nil {
			logger.Sugar().Fatalf("parse server.IPTrustList err: %v", err)
		}
		options = append(options, echo.TrustIPRange(ipNet))
		logger.Sugar().Infof("trust ip range: %s", trust)
	}

	switch ipExtractor {
	case "direct":
		logger.Sugar().Infof("extract ip direct")
		return echo.ExtractIPDirect()
	case "x-real-ip":
		logger.Sugar().Infof("extract ip from `X-Real-Ip` header")
		return echo.ExtractIPFromRealIPHeader(options...)
	case "x-forwarded-for":
		logger.Sugar().Infof("extract ip from `X-Forwarded-For` header")
		return echo.ExtractIPFromXFFHeader(options...)
	default:
		logger.Sugar().Infof("[default] extract ip direct")
		return echo.ExtractIPDirect()
	}
}
