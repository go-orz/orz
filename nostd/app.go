package nostd

import (
	"context"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type Application interface {
	Name() string
	AutoMigrate(db *gorm.DB) error
	Routing(e *echo.Group) error
	Boot(ctx context.Context) error
}
