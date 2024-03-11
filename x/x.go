package x

import (
	"github.com/go-orz/orz/config"
	"github.com/labstack/echo/v4"
)

func IPExtractor() echo.IPExtractor {
	switch config.Conf().Server.IPExtractor {
	case "direct":
		return echo.ExtractIPDirect()
	case "x-real-ip":
		return echo.ExtractIPFromRealIPHeader()
	case "x-forwarded-for":
		return echo.ExtractIPFromXFFHeader()
	default:
		return echo.ExtractIPDirect()
	}
}
