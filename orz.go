package orz

import (
	"context"
	"errors"
	"github.com/go-orz/orz/config"
	"github.com/go-orz/orz/database"
	"github.com/go-orz/orz/log"
	"github.com/go-orz/orz/x"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"net/http"
	"os"
	"os/signal"
)

func New(configPath string) *Orz {
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
	// 初始化web服务器
	e := echo.New()
	e.IPExtractor = x.IPExtractor()

	return &Orz{
		Config:   conf,
		Database: _db,
		Logger:   log.Z(),
		echo:     e,
	}
}

type Orz struct {
	Config   *config.Config
	Database *gorm.DB
	Logger   *zap.Logger
	echo     *echo.Echo
}

func (r *Orz) Start() {
	go func() {
		e := r.echo
		cfg := r.Config.Server

		addr := cfg.Addr
		logger := r.Logger
		logger.Sugar().Infof("http server start at: %v", addr)

		var err error
		if cfg.TLS.Enabled {
			if cfg.TLS.Auto {
				err = e.StartAutoTLS(addr)
			} else {
				err = e.StartTLS(addr, cfg.TLS.Cert, cfg.TLS.Key)
			}
		} else {
			err = e.Start(addr)
		}

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Sugar().Errorf("shutting down server err: %v", err)
		}
	}()
}

func (r *Orz) Wait() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
}

func (r *Orz) Stop(ctx context.Context) error {
	return r.echo.Shutdown(ctx)
}

func (r *Orz) Echo() *echo.Echo {
	return r.echo
}
