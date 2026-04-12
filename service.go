package orz

import (
	"context"

	"gorm.io/gorm"
)

// contextKey 上下文键类型
type contextKey string

// 上下文键定义
const (
	dbContextKey contextKey = "db"
)

// Service 基础服务类，提供事务管理和数据库访问能力
// 用户自定义的service可以继承它，专注于业务逻辑
type Service struct {
	db *gorm.DB // 持有数据库连接
}

// NewService 创建服务实例
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// InTransaction 检查是否在事务中
func (s *Service) InTransaction(ctx context.Context) bool {
	_, ok := ctx.Value(dbContextKey).(*gorm.DB)
	return ok
}

// Transaction 在事务中执行业务逻辑
func (s *Service) Transaction(ctx context.Context, f func(ctx context.Context) error) error {
	if !s.InTransaction(ctx) {
		// 使用 Service 持有的数据库连接
		return s.db.Transaction(func(tx *gorm.DB) error {
			c := context.WithValue(ctx, dbContextKey, tx)
			return f(c)
		})
	}
	return f(ctx)
}

// WithTx 将事务放入上下文
func WithTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, dbContextKey, tx)
}

// WithDB 将数据库连接放入上下文（用于非事务场景）
func WithDB(ctx context.Context, db *gorm.DB) context.Context {
	return context.WithValue(ctx, dbContextKey, db)
}
