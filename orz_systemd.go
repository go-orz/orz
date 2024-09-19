package orz

import (
	"context"
	"github.com/kardianos/service"
	"time"
)

var _ service.Interface = (*SystemService)(nil)

func NewSystemService(app *Orz) *SystemService {
	return &SystemService{
		app: app,
	}
}

type SystemService struct {
	app *Orz
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
