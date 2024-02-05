package log

import (
	"sync"

	"github.com/go-orz/orz/config"

	"go.uber.org/zap"
)

var (
	instance *zap.Logger
	once     sync.Once
)

func Z() *zap.Logger {
	once.Do(func() {
		conf := config.Conf().Log
		instance = NewLogger(conf.Level, conf.Filename)
	})
	return instance
}
