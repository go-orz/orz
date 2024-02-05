package orz

import (
	"context"
	"gorm.io/gorm"
	"sync"
)

const _DB = "db"

var (
	_db     *gorm.DB
	_dbOnce sync.Once
)

func InjectDB(db *gorm.DB) {
	_dbOnce.Do(func() {
		_db = db
	})
}

func MustGetDB() *gorm.DB {
	if _db == nil {
		panic(`you must call orz.InjectDB(db *gorm.DB) first`)
	}
	return _db
}

type Service struct {
}

func (s *Service) InTransaction(ctx context.Context) bool {
	_, ok := ctx.Value(_DB).(*gorm.DB)
	return ok
}

func (s *Service) Transaction(ctx context.Context, f func(ctx context.Context) error) error {
	if !s.InTransaction(ctx) {
		db := MustGetDB()
		return db.Transaction(func(tx *gorm.DB) error {
			c := context.WithValue(ctx, _DB, tx)
			return f(c)
		})
	}
	return f(ctx)
}

func (s *Service) GetDB(ctx context.Context) *gorm.DB {
	val, ok := ctx.Value(_DB).(*gorm.DB)
	if !ok {
		return MustGetDB()
	}
	return val
}
