package orz

import (
	"go.uber.org/zap"
	"sync"
)

var (
	_logger     *zap.Logger
	_loggerOnce sync.Once
)

func InjectLogger(logger *zap.Logger) {
	_loggerOnce.Do(func() {
		_logger = logger
	})
}

func MustGetLogger() *zap.Logger {
	if _logger == nil {
		panic(`you must call orz.InjectLogger(logger *zap.Logger) first`)
	}
	return _logger
}
