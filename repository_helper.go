package orz

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// PageBuilder 分页查询构建器
type PageBuilder[T any, ID comparable] struct {
	repo      Repository[T, ID]
	pageIndex int
	pageSize  int
	sort      Sort
	matchers  []Matcher
	selectSQL string
	joins     []string
}

// NewPageBuilder 创建分页查询构建器
func NewPageBuilder[T any, ID comparable](repo Repository[T, ID]) *PageBuilder[T, ID] {
	return &PageBuilder[T, ID]{
		repo:      repo,
		pageIndex: 1,
		pageSize:  10,
	}
}

// PageIndex 设置页码
func (b *PageBuilder[T, ID]) PageIndex(pageIndex int) *PageBuilder[T, ID] {
	if pageIndex < 1 {
		pageIndex = 1
	}
	b.pageIndex = pageIndex
	return b
}

// PageSize 设置页大小
func (b *PageBuilder[T, ID]) PageSize(pageSize int) *PageBuilder[T, ID] {
	if pageSize < 1 {
		pageSize = 10
	}
	b.pageSize = pageSize
	return b
}

// Sort 设置排序
func (b *PageBuilder[T, ID]) Sort(sort Sort) *PageBuilder[T, ID] {
	b.sort = sort
	return b
}

// SortBy 设置升序排序
func (b *PageBuilder[T, ID]) SortBy(property string, allowedFields ...string) *PageBuilder[T, ID] {
	b.sort = NewSortBy(property, allowedFields...)
	return b
}

// SortByDesc 设置降序排序
func (b *PageBuilder[T, ID]) SortByDesc(property string, allowedFields ...string) *PageBuilder[T, ID] {
	b.sort = NewSortByDesc(property, allowedFields...)
	return b
}

func (b *PageBuilder[T, ID]) PageRequest(pr *PageRequest) *PageBuilder[T, ID] {
	b.PageIndex(pr.PageIndex)
	b.PageSize(pr.PageSize)
	b.Sort(NewSort(pr.SortOrder, pr.SortField, pr.SortAllowedFields...))
	return b
}

// Where 添加查询条件
func (b *PageBuilder[T, ID]) Where(matchers ...Matcher) *PageBuilder[T, ID] {
	b.matchers = append(b.matchers, matchers...)
	return b
}

// Equal 添加等值查询条件
func (b *PageBuilder[T, ID]) Equal(name string, value any) *PageBuilder[T, ID] {
	b.matchers = append(b.matchers, NewMatcher(name, value, MatcherEqual))
	return b
}

// Contains 添加模糊查询条件
func (b *PageBuilder[T, ID]) Contains(name string, value any) *PageBuilder[T, ID] {
	b.matchers = append(b.matchers, NewMatcher(name, value, MatcherContains))
	return b
}

// In 添加IN查询条件
func (b *PageBuilder[T, ID]) In(name string, values any) *PageBuilder[T, ID] {
	b.matchers = append(b.matchers, NewMatcher(name, values, MatcherIn))
	return b
}

// NotEqual 添加不等于查询条件
func (b *PageBuilder[T, ID]) NotEqual(name string, value any) *PageBuilder[T, ID] {
	b.matchers = append(b.matchers, NewMatcher(name, value, MatcherNotEqual))
	return b
}

// NotContains 添加不包含查询条件
func (b *PageBuilder[T, ID]) NotContains(name string, value any) *PageBuilder[T, ID] {
	b.matchers = append(b.matchers, NewMatcher(name, value, MatcherNotContains))
	return b
}

// NotIn 添加NOT IN查询条件
func (b *PageBuilder[T, ID]) NotIn(name string, values any) *PageBuilder[T, ID] {
	b.matchers = append(b.matchers, NewMatcher(name, values, MatcherNotIn))
	return b
}

// ContainsIgnoreCase 添加模糊查询条件（忽略大小写）
func (b *PageBuilder[T, ID]) ContainsIgnoreCase(name string, value any) *PageBuilder[T, ID] {
	b.matchers = append(b.matchers, NewMatcher(name, value, MatcherContainsIgnoreCase))
	return b
}

// NotContainsIgnoreCase 添加不包含查询条件（忽略大小写）
func (b *PageBuilder[T, ID]) NotContainsIgnoreCase(name string, value any) *PageBuilder[T, ID] {
	b.matchers = append(b.matchers, NewMatcher(name, value, MatcherNotContainsIgnoreCase))
	return b
}

// Tags 添加JSON数组标签查询条件
func (b *PageBuilder[T, ID]) Tags(name string, tags string) *PageBuilder[T, ID] {
	b.matchers = append(b.matchers, NewMatcher(name, tags, MatcherTags))
	return b
}

// LeftJoin 添加左连接
func (b *PageBuilder[T, ID]) LeftJoin(table, condition string) *PageBuilder[T, ID] {
	joinSQL := fmt.Sprintf("LEFT JOIN %s ON %s", table, condition)
	b.joins = append(b.joins, joinSQL)
	return b
}

// RightJoin 添加右连接
func (b *PageBuilder[T, ID]) RightJoin(table, condition string) *PageBuilder[T, ID] {
	joinSQL := fmt.Sprintf("RIGHT JOIN %s ON %s", table, condition)
	b.joins = append(b.joins, joinSQL)
	return b
}

// InnerJoin 添加内连接
func (b *PageBuilder[T, ID]) InnerJoin(table, condition string) *PageBuilder[T, ID] {
	joinSQL := fmt.Sprintf("INNER JOIN %s ON %s", table, condition)
	b.joins = append(b.joins, joinSQL)
	return b
}

// Select 设置查询字段
func (b *PageBuilder[T, ID]) Select(fields string) *PageBuilder[T, ID] {
	b.selectSQL = fields
	return b
}

// Execute 执行分页查询（返回原类型）
func (b *PageBuilder[T, ID]) Execute(ctx context.Context) (*PageResult[T], error) {
	db := b.repo.GetDB(ctx)

	// 验证排序字段安全性
	if err := b.sort.Validate(); err != nil {
		return nil, fmt.Errorf("sort validation failed: %w", err)
	}

	// 应用 SELECT 字段
	if b.selectSQL != "" {
		db = db.Select(b.selectSQL)
	}

	// 应用 JOIN 语句
	for _, join := range b.joins {
		db = db.Joins(join)
	}

	// 应用查询匹配器
	var err error
	db, err = ApplyMatchers(db, b.matchers, b.repo.GetTableName(), nil)
	if err != nil {
		return nil, fmt.Errorf("apply matchers failed: %w", err)
	}

	// 应用排序
	if !b.sort.IsEmpty() {
		db = db.Order(fmt.Sprintf("%s %s", b.sort.Field, b.sort.Order))
	}

	// 查询总数
	var total int64
	countDB := db.Session(&gorm.Session{})
	if err := countDB.Table(b.repo.GetTableName()).Count(&total).Error; err != nil {
		return nil, fmt.Errorf("count failed: %w", err)
	}

	// 查询数据
	var items []T
	dataDB := db.Offset((b.pageIndex - 1) * b.pageSize).Limit(b.pageSize)
	if err := dataDB.Table(b.repo.GetTableName()).Find(&items).Error; err != nil {
		return nil, fmt.Errorf("find page items failed: %w", err)
	}

	return &PageResult[T]{
		Items: items,
		Total: total,
	}, nil
}

// ExecuteAsTyped 执行泛型分页查询（返回强类型结果）
// 这个方法通过直接操作数据库来实现不同类型的查询
func ExecuteAsTyped[R any, T any, ID comparable](ctx context.Context, builder *PageBuilder[T, ID]) (*PageResult[R], error) {
	db := builder.repo.GetDB(ctx)

	// 验证排序字段安全性
	if err := builder.sort.Validate(); err != nil {
		return nil, fmt.Errorf("sort validation failed: %w", err)
	}

	// 应用 SELECT 字段
	if builder.selectSQL != "" {
		db = db.Select(builder.selectSQL)
	}

	// 应用 JOIN 语句
	for _, join := range builder.joins {
		db = db.Joins(join)
	}

	// 应用查询匹配器
	var err error
	db, err = ApplyMatchers(db, builder.matchers, builder.repo.GetTableName(), nil)
	if err != nil {
		return nil, fmt.Errorf("apply matchers failed: %w", err)
	}

	// 应用排序
	if !builder.sort.IsEmpty() {
		db = db.Order(fmt.Sprintf("%s %s", builder.sort.Field, builder.sort.Order))
	}

	// 查询总数
	var total int64
	countDB := db.Session(&gorm.Session{})
	if err := countDB.Table(builder.repo.GetTableName()).Count(&total).Error; err != nil {
		return nil, fmt.Errorf("count failed: %w", err)
	}

	// 查询数据
	var items []R
	dataDB := db.Offset((builder.pageIndex - 1) * builder.pageSize).Limit(builder.pageSize)
	if err := dataDB.Table(builder.repo.GetTableName()).Find(&items).Error; err != nil {
		return nil, fmt.Errorf("find page items failed: %w", err)
	}

	return &PageResult[R]{
		Items: items,
		Total: total,
	}, nil
}

// 便捷的链式查询方法
// 为 Repository 添加扩展方法

// Query 创建分页查询构建器
func Query[T any, ID comparable](repo Repository[T, ID]) *PageBuilder[T, ID] {
	return NewPageBuilder(repo)
}
