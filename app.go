package orz

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/labstack/echo/v5"
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

	if config := a.GetConfig(); config != nil {
		configureIPExtractor(e, config.Server, a.Logger())
	}

	ensureDirectIPExtractor(e)

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
		return a.runDaemon()
	}

	// 启动HTTP服务器
	return a.runHTTPServer(e)
}

// runHTTPServer 运行HTTP服务器
func (a *App) runHTTPServer(e *echo.Echo) error {
	signalCtx, stopSignals := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stopSignals()

	serverCtx, stopServer := context.WithCancel(context.Background())
	defer stopServer()

	var shutdownOnce sync.Once
	shutdown := func() {
		shutdownOnce.Do(func() {
			a.Logger().Info("shutting down server...")
			a.cancel()
			stopServer()
		})
	}

	go func() {
		select {
		case <-signalCtx.Done():
			shutdown()
		case <-a.ctx.Done():
			shutdown()
		case <-serverCtx.Done():
		}
	}()

	// 获取服务器配置
	config := a.GetConfig()
	addr := ":8080" // 默认端口
	if config != nil && config.Server.Addr != "" {
		addr = config.Server.Addr
	}

	a.Logger().Info("starting server", zap.String("addr", addr))

	startConfig := echo.StartConfig{
		Address:         addr,
		HideBanner:      true,
		HidePort:        true,
		GracefulTimeout: 10 * time.Second,
		OnShutdownError: func(err error) {
			if !errors.Is(err, http.ErrServerClosed) {
				a.Logger().Error("server forced to shutdown", zap.Error(err))
			}
		},
	}

	// 根据配置启动HTTP服务器
	var err = startConfig.Start(serverCtx, e)
	if err != nil {
		if errors.Is(err, http.ErrServerClosed) || errors.Is(err, context.Canceled) {
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
