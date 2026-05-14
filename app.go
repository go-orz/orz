package orz

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Application 应用配置接口
type Application interface {
	Configure(app *App) error
}

// App 应用容器
type App struct {
	logger        *zap.Logger
	database      *gorm.DB
	echo          *echo.Echo
	configManager *ConfigManager
	ctx           context.Context
	cancel        context.CancelFunc
}

// NewApp 创建新的应用
func NewApp() *App {
	ctx, cancel := context.WithCancel(context.Background())
	return &App{
		configManager: NewConfigManager(),
		ctx:           ctx,
		cancel:        cancel,
	}
}

// EnableLogger 根据配置初始化日志器
func (a *App) EnableLogger() error {
	config := a.GetConfig()
	if config == nil {
		return fmt.Errorf("config not loaded")
	}

	logger := NewLoggerFromConfig(config.Log)
	a.SetLogger(logger)
	return nil
}

// EnableDatabase 启用数据库
func (a *App) EnableDatabase() error {
	config := a.GetConfig()
	if config == nil || !config.Database.Enabled {
		return fmt.Errorf("database not enabled in config")
	}

	log := a.Logger()
	db, err := ConnectDatabaseWithLogger(config.Database, log)
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	a.SetDatabase(db)
	return nil
}

// EnableHTTP 启用HTTP服务
func (a *App) EnableHTTP() {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// 应用服务器配置
	config := a.GetConfig()
	if config != nil {
		ipTrustList := config.Server.IPTrustList
		options := []echo.TrustOption{
			echo.TrustLoopback(false),
			echo.TrustLinkLocal(false),
			echo.TrustPrivateNet(false),
		}
		trustedProxyCount := 0
		for _, ip := range ipTrustList {
			ipNet, err := parseTrustedProxyIPRange(ip)
			if err != nil {
				a.Logger().Warn("failed to parse trusted proxy IP", zap.String("ip", ip), zap.Error(err))
				continue
			}
			options = append(options, echo.TrustIPRange(ipNet))
			trustedProxyCount++
		}

		if trustedProxyCount > 0 {
			a.Logger().Info("trusted proxy IPs configured", zap.Strings("trustedProxies", ipTrustList))
		}

		ipExtractor := strings.ToLower(strings.TrimSpace(config.Server.IPExtractor))
		if ipExtractor == "" {
			ipExtractor = "direct"
		}

		switch ipExtractor {
		case "x-forwarded-for":
			if trustedProxyCount == 0 {
				a.Logger().Warn("x-forwarded-for extractor requested without trusted proxies; falling back to direct")
				e.IPExtractor = echo.ExtractIPDirect()
				break
			}
			e.IPExtractor = echo.ExtractIPFromXFFHeader(options...)
		case "x-real-ip":
			if trustedProxyCount == 0 {
				a.Logger().Warn("x-real-ip extractor requested without trusted proxies; falling back to direct")
				e.IPExtractor = echo.ExtractIPDirect()
				break
			}
			e.IPExtractor = echo.ExtractIPFromRealIPHeader(options...)
		case "direct":
			e.IPExtractor = echo.ExtractIPDirect()
		default:
			a.Logger().Warn("unknown IP extractor; falling back to direct", zap.String("ipExtractor", config.Server.IPExtractor))
			e.IPExtractor = echo.ExtractIPDirect()
		}
	}

	if e.IPExtractor == nil {
		e.IPExtractor = echo.ExtractIPDirect()
	}

	a.SetEcho(e)
}

func parseTrustedProxyIPRange(value string) (*net.IPNet, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, fmt.Errorf("empty IP range")
	}

	if _, ipNet, err := net.ParseCIDR(value); err == nil {
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

// SetDatabase 设置数据库连接
func (a *App) SetDatabase(db *gorm.DB) {
	a.database = db
}

// SetEcho 设置 Echo 实例
func (a *App) SetEcho(e *echo.Echo) {
	a.echo = e
}

// Logger 获取日志器
func (a *App) Logger() *zap.Logger {
	if a.logger == nil {
		// 如果没有设置日志器，返回默认的
		a.logger, _ = zap.NewDevelopment()
	}
	return a.logger
}

// SetLogger 设置日志器
func (a *App) SetLogger(logger *zap.Logger) {
	a.logger = logger
}

// GetDatabase 获取数据库连接
func (a *App) GetDatabase() *gorm.DB {
	return a.database
}

// GetEcho 获取Echo实例
func (a *App) GetEcho() *echo.Echo {
	return a.echo
}

// Context 获取应用上下文
func (a *App) Context() context.Context {
	return a.ctx
}

// LoadConfigFromFile 从文件加载配置
func (a *App) LoadConfigFromFile(configPath string) error {
	return a.configManager.LoadFromFile(configPath)
}

// LoadConfigFromBytes 从字节数组加载配置
func (a *App) LoadConfigFromBytes(data []byte) error {
	return a.configManager.LoadFromBytes(data)
}

// LoadConfigFromMap 从Map加载配置
func (a *App) LoadConfigFromMap(data map[string]interface{}) error {
	return a.configManager.LoadFromMap(data)
}

// GetConfig 获取配置
func (a *App) GetConfig() *Config {
	return a.configManager.GetConfig()
}

// Run 运行应用
func (a *App) Run() error {
	// 获取Echo实例
	e := a.GetEcho()
	if e == nil {
		a.Logger().Info("no HTTP server configured, running in daemon mode")
		return a.runDaemon()
	}

	// 启动HTTP服务器
	return a.runHTTPServer(e)
}

// runHTTPServer 运行HTTP服务器
func (a *App) runHTTPServer(e *echo.Echo) error {
	signalCtx, stopSignals := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stopSignals()

	watchCtx, stopWatch := context.WithCancel(context.Background())
	defer stopWatch()

	var shutdownOnce sync.Once
	shutdown := func() {
		shutdownOnce.Do(func() {
			a.Logger().Info("shutting down server...")
			a.cancel()

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := e.Shutdown(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
				a.Logger().Error("server forced to shutdown", zap.Error(err))
			}
		})
	}

	go func() {
		select {
		case <-signalCtx.Done():
			shutdown()
		case <-a.ctx.Done():
			shutdown()
		case <-watchCtx.Done():
		}
	}()

	// 获取服务器配置
	config := a.GetConfig()
	addr := ":8080" // 默认端口
	if config != nil && config.Server.Addr != "" {
		addr = config.Server.Addr
	}

	a.Logger().Info("starting server", zap.String("addr", addr))

	// 根据配置启动HTTP或HTTPS服务器
	var err error
	if config != nil && config.Server.TLS.Enabled {
		if config.Server.TLS.Auto {
			a.Logger().Info("starting HTTPS server with auto TLSConfig")
			err = e.StartAutoTLS(addr)
		} else if config.Server.TLS.Cert != "" && config.Server.TLS.Key != "" {
			a.Logger().Info("starting HTTPS server with custom TLSConfig",
				zap.String("cert", config.Server.TLS.Cert),
				zap.String("key", config.Server.TLS.Key))
			err = e.StartTLS(addr, config.Server.TLS.Cert, config.Server.TLS.Key)
		} else {
			a.Logger().Error("TLSConfig enabled but cert/key not provided, falling back to HTTP")
			err = e.Start(addr)
		}
	} else {
		err = e.Start(addr)
	}

	stopWatch()
	if err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			a.Logger().Info("server stopped")
			return nil
		}
		return err
	}

	return nil
}

// runDaemon 以守护进程模式运行
func (a *App) runDaemon() error {
	a.Logger().Info("running in daemon mode")

	signalCtx, stopSignals := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stopSignals()

	select {
	case <-signalCtx.Done():
		a.Logger().Info("shutting down daemon...")
		a.cancel()
	case <-a.ctx.Done():
		a.Logger().Info("shutting down daemon...")
	}

	return nil
}
