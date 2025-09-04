package orz

import (
	"context"
	"fmt"
	"os"
	"os/signal"
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

	logger, err := NewLoggerFromConfig(config.Log)
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}

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
		// 配置IP提取器
		if config.Server.IPExtractor != "" {
			switch config.Server.IPExtractor {
			case "X-Forwarded-For":
				e.IPExtractor = echo.ExtractIPFromXFFHeader()
			case "X-Real-IP":
				e.IPExtractor = echo.ExtractIPFromRealIPHeader()
			}
		}

		// 配置信任的IP列表
		if len(config.Server.IPTrustList) > 0 {
			// 这里可以添加IP白名单中间件的逻辑
			a.Logger().Info("IP trust list configured", zap.Strings("trustedIPs", config.Server.IPTrustList))
		}
	}

	a.SetEcho(e)
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
		return nil
	}

	// 启动HTTP服务器
	return a.runHTTPServer(e)
}

// runHTTPServer 运行HTTP服务器
func (a *App) runHTTPServer(e *echo.Echo) error {
	// 优雅关闭
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		a.Logger().Info("shutting down server...")
		a.cancel()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := e.Shutdown(ctx); err != nil {
			a.Logger().Error("server forced to shutdown", zap.Error(err))
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
	if config != nil && config.Server.TLS.Enabled {
		if config.Server.TLS.Auto {
			a.Logger().Info("starting HTTPS server with auto TLS")
			if err := e.StartAutoTLS(addr); err != nil {
				a.Logger().Info("server stopped")
			}
		} else if config.Server.TLS.Cert != "" && config.Server.TLS.Key != "" {
			a.Logger().Info("starting HTTPS server with custom TLS",
				zap.String("cert", config.Server.TLS.Cert),
				zap.String("key", config.Server.TLS.Key))
			if err := e.StartTLS(addr, config.Server.TLS.Cert, config.Server.TLS.Key); err != nil {
				a.Logger().Info("server stopped")
			}
		} else {
			a.Logger().Error("TLS enabled but cert/key not provided, falling back to HTTP")
			if err := e.Start(addr); err != nil {
				a.Logger().Info("server stopped")
			}
		}
	} else {
		if err := e.Start(addr); err != nil {
			a.Logger().Info("server stopped")
		}
	}

	return nil
}

// runDaemon 以守护进程模式运行
func (a *App) runDaemon() error {
	a.Logger().Info("running in daemon mode")

	// 等待信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	a.Logger().Info("shutting down daemon...")
	a.cancel()
	return nil
}
