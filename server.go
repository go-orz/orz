package orz

import (
	"context"
	"errors"
	"github.com/go-orz/orz/config"
	"github.com/labstack/echo/v4"
	"github.com/spf13/cast"
	"go.uber.org/zap"
	"net"
	"net/http"
)

type Server interface {
	Start()
	Stop(ctx context.Context) error
	Echo() *echo.Echo
}

func NewServer(cfg config.Server, logger *zap.Logger) Server {
	return &server{
		e:      echo.New(),
		cfg:    cfg,
		logger: logger,
	}
}

type server struct {
	cfg    config.Server
	e      *echo.Echo
	logger *zap.Logger
}

func (r *server) Echo() *echo.Echo {
	return r.e
}

func (r *server) Start() {
	go func() {
		e := r.e
		cfg := r.cfg
		logger := r.logger

		addr := net.JoinHostPort(cfg.Host, cast.ToString(cfg.Port))

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

func (r *server) Stop(ctx context.Context) error {
	e := r.e
	return e.Shutdown(ctx)
}
