package orz

import (
	"context"
	"github.com/go-orz/orz/config"
	"github.com/go-orz/orz/database"
	"github.com/go-orz/orz/log"
	"github.com/kardianos/service"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"time"
)

type BeforeBootFunc func(e *echo.Echo)
type AfterBootFunc func(e *echo.Echo)

func NewApp(configPath string) *App {
	// 初始化配置
	config.MustInit(configPath)
	conf := config.Conf()
	// 初始化日志
	InjectLogger(log.Z())

	// 初始化数据库
	if conf.Database.Enabled {
		db := database.MustConnectDatabase(conf.Database)
		InjectDB(db)
	}

	log.Z().Debug("config", zap.Any("conf", conf))
	return &App{
		server: NewServer(conf.Server, log.Z()),
	}
}

type App struct {
	server Server

	bbf []BeforeBootFunc
	abf []AfterBootFunc
}

func (r *App) Start() {
	for _, f := range r.bbf {
		f(r.server.Echo())
	}
	r.server.Start()
	for _, f := range r.abf {
		f(r.server.Echo())
	}
}

func (r *App) Wait() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
}

func (r *App) Stop(ctx context.Context) error {
	return r.server.Stop(ctx)
}

func (r *App) AddBeforeBootFunc(f BeforeBootFunc) {
	r.bbf = append(r.bbf, f)
}

func (r *App) AddAfterBootFunc(f AfterBootFunc) {
	r.abf = append(r.abf, f)
}

func NewSystemService(app *App) *SystemService {
	return &SystemService{
		app: app,
	}
}

var _ service.Interface = (*SystemService)(nil)

type SystemService struct {
	app *App
}

func (r SystemService) Start(s service.Service) error {
	logger := MustGetLogger()
	if service.Interactive() {
		logger.Info("Running in terminal.")
	} else {
		logger.Info("Running under service manager.")
	}
	r.app.Start()
	return nil
}

func (r SystemService) Stop(s service.Service) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	return r.app.Stop(ctx)
}
