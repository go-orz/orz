package orz

import (
	"fmt"

	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Map = map[string]any

// Framework 框架结构
type Framework struct {
	app            *App
	configLoaders  []func(*App) error
	loggerMode     loggerMode
	customLogger   *zap.Logger
	enableDatabase bool
	enableHTTP     bool
	application    Application
	applicationSet bool
}

type loggerMode uint8

const (
	loggerModeUnset loggerMode = iota
	loggerModeCustom
	loggerModeFromConfig
)

// Option 框架配置选项
type Option func(*Framework) error

// WithConfig 设置配置文件路径
func WithConfig(configPath string) Option {
	return func(f *Framework) error {
		f.configLoaders = append(f.configLoaders, func(app *App) error {
			if err := app.LoadConfigFromFile(configPath); err != nil {
				return fmt.Errorf("load config failed: %w", err)
			}
			return nil
		})
		return nil
	}
}

// WithConfigBytes 从字节数据设置配置
func WithConfigBytes(data []byte) Option {
	return func(f *Framework) error {
		f.configLoaders = append(f.configLoaders, func(app *App) error {
			if err := app.LoadConfigFromBytes(data); err != nil {
				return fmt.Errorf("load config from bytes failed: %w", err)
			}
			return nil
		})
		return nil
	}
}

// WithConfigMap 从 map 设置配置
func WithConfigMap(configMap map[string]interface{}) Option {
	return func(f *Framework) error {
		f.configLoaders = append(f.configLoaders, func(app *App) error {
			if err := app.LoadConfigFromMap(configMap); err != nil {
				return fmt.Errorf("load config from map failed: %w", err)
			}
			return nil
		})
		return nil
	}
}

// WithLogger 设置日志器
func WithLogger(logger *zap.Logger) Option {
	return func(f *Framework) error {
		f.loggerMode = loggerModeCustom
		f.customLogger = logger
		return nil
	}
}

// WithLoggerFromConfig 根据配置启用日志器
func WithLoggerFromConfig() Option {
	return func(f *Framework) error {
		f.loggerMode = loggerModeFromConfig
		return nil
	}
}

// WithDatabase 启用数据库
func WithDatabase() Option {
	return func(f *Framework) error {
		f.enableDatabase = true
		return nil
	}
}

// WithHTTP 启用HTTP服务
func WithHTTP() Option {
	return func(f *Framework) error {
		f.enableHTTP = true
		return nil
	}
}

// WithApplication 设置应用实现
func WithApplication(application Application) Option {
	return func(f *Framework) error {
		f.application = application
		f.applicationSet = true
		return nil
	}
}

// NewFramework 创建新的框架实例
func NewFramework(options ...Option) (*Framework, error) {
	framework := &Framework{
		app: NewApp(),
	}

	// 应用所有选项
	for _, option := range options {
		if err := option(framework); err != nil {
			return nil, err
		}
	}

	if err := framework.initialize(); err != nil {
		return nil, err
	}

	return framework, nil
}

func (f *Framework) initialize() error {
	for _, loader := range f.configLoaders {
		if err := loader(f.app); err != nil {
			return err
		}
	}

	switch f.loggerMode {
	case loggerModeCustom:
		f.app.SetLogger(f.customLogger)
	case loggerModeFromConfig:
		if err := f.app.EnableLogger(); err != nil {
			return err
		}
	}

	if f.enableDatabase {
		if err := f.app.EnableDatabase(); err != nil {
			return err
		}
	}

	if f.enableHTTP {
		f.app.EnableHTTP()
	}

	if f.applicationSet && f.application != nil {
		if err := f.application.Configure(f.app); err != nil {
			return fmt.Errorf("failed to configure application: %w", err)
		}
	}

	return nil
}

// Run 运行应用
func (f *Framework) Run() error {
	return f.app.Run()
}

// App 获取应用实例
func (f *Framework) App() *App {
	return f.app
}

// 便捷方法

// GetDB 获取数据库实例
func (f *Framework) GetDB() *gorm.DB {
	return f.app.GetDatabase()
}

// GetEcho 获取Echo实例
func (f *Framework) GetEcho() *echo.Echo {
	return f.app.GetEcho()
}

// SimpleApp 简单应用实现
type SimpleApp struct {
	setupFn func(app *App) error
}

// NewSimpleApp 创建简单应用
func NewSimpleApp(setupFn func(app *App) error) *SimpleApp {
	return &SimpleApp{
		setupFn: setupFn,
	}
}

// Configure 配置应用
func (s *SimpleApp) Configure(a *App) error {
	// 执行自定义设置函数
	if s.setupFn != nil {
		return s.setupFn(a)
	}

	return nil
}

// Quick 快速启动函数
func Quick(configPath string, setupFn func(app *App) error) error {
	simpleApp := NewSimpleApp(setupFn)

	framework, err := NewFramework(
		WithConfig(configPath),
		WithLoggerFromConfig(),
		WithDatabase(),
		WithHTTP(),
		WithApplication(simpleApp),
	)
	if err != nil {
		return err
	}

	return framework.Run()
}
