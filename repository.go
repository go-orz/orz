package orz

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/spf13/cast"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Repository 通用仓库接口
type Repository[T any, ID comparable] interface {
	Create(ctx context.Context, entity *T) error
	CreateInBatches(ctx context.Context, entities []T, batchSize int) error
	CreateOrUpdate(ctx context.Context, entity *T) error

	FindById(ctx context.Context, id ID) (T, error)
	FindByIdExists(ctx context.Context, id ID) (T, bool, error)
	FindByIdIn(ctx context.Context, ids []ID) ([]T, error)
	FindAll(ctx context.Context) ([]T, error)
	ExistsById(ctx context.Context, id ID) (bool, error)

	UpdateById(ctx context.Context, entity *T) error
	UpdateColumnsById(ctx context.Context, id ID, columns map[string]interface{}) error
	Save(ctx context.Context, entity *T) error

	DeleteById(ctx context.Context, id ID) error
	DeleteByIdIn(ctx context.Context, ids []ID) error

	Count(ctx context.Context) (int64, error)

	Find(ctx context.Context, matchers []Matcher, sort Sort) ([]T, error)

	FindOne(ctx context.Context, matchers []Matcher) (T, error)
	Exists(ctx context.Context, matchers []Matcher) (bool, error)
	CountByMatchers(ctx context.Context, matchers []Matcher) (int64, error)

	GetDB(ctx context.Context) *gorm.DB
	GetTableName() string
}

// 分页查询相关类型定义

type SortOrder string

const (
	ASC  SortOrder = "asc"  // 正序
	DESC SortOrder = "desc" // 倒序
)

// Sort 排序配置
type Sort struct {
	Order         SortOrder // 排序方式
	Field         string    // 排序字段（数据库字段名，蛇形命名）
	AllowedFields []string  // 允许排序的字段白名单
	validated     bool      // 是否已验证
}

// NewSort 创建排序配置
func NewSort(order SortOrder, field string, allowedFields ...string) Sort {
	if order == "" {
		order = DESC
	}
	return Sort{
		Order:         order,
		Field:         CamelToSnake(field),     // 自动转换为蛇形命名
		AllowedFields: allowedFields,           // 允许排序的字段白名单
		validated:     len(allowedFields) == 0, // 如果没有白名单则默认为有效
	}
}

// NewSortBy 创建升序排序配置
func NewSortBy(property string, allowedFields ...string) Sort {
	return NewSort(ASC, property, allowedFields...)
}

// NewSortByDesc 创建降序排序配置
func NewSortByDesc(property string, allowedFields ...string) Sort {
	return NewSort(DESC, property, allowedFields...)
}

// IsEmpty 检查是否为空排序
func (s *Sort) IsEmpty() bool {
	return s.Field == ""
}

// IsValid 检查排序字段是否在白名单中
func (s *Sort) IsValid() bool {
	if s.IsEmpty() {
		return true // 空排序总是有效的
	}

	if len(s.AllowedFields) == 0 {
		return false // 没有白名单则无效
	}

	// 检查原始字段名和蛇形字段名
	originalProperty := s.Field
	for _, field := range s.AllowedFields {
		if field == originalProperty || CamelToSnake(field) == originalProperty {
			return true
		}
	}
	return false
}

// Validate 验证排序配置（用于Repository内部）
func (s *Sort) Validate() error {
	if s.validated {
		return nil
	}

	if !s.IsValid() {
		if s.IsEmpty() {
			return nil // 空排序不需要验证
		}
		return fmt.Errorf("invalid sort field '%s', allowed fields: %v", s.Field, s.AllowedFields)
	}

	s.validated = true
	return nil
}

// PageResult 分页结果
type PageResult[T any] struct {
	Items []T   `json:"items"` // 当前页数据
	Total int64 `json:"total"` // 总记录数
}

// NewPageResult 创建分页结果
func NewPageResult[T any](items []T, total int64) *PageResult[T] {
	return &PageResult[T]{
		Items: items,
		Total: total,
	}
}

// 查询匹配器相关类型

type MatcherMode string

const (
	MatcherContains           MatcherMode = "contains"             // 模糊查询
	MatcherContainsIgnoreCase MatcherMode = "contains-ignore-case" // 模糊查询忽略大小写
	MatcherEqual              MatcherMode = "equal"                // = 比较
	MatcherIn                 MatcherMode = "in"                   // in 查询

	MatcherNotContains           MatcherMode = "not-contains"
	MatcherNotContainsIgnoreCase MatcherMode = "not-contains-ignore-case"
	MatcherNotEqual              MatcherMode = "not-equal"
	MatcherNotIn                 MatcherMode = "not-in"

	MatcherTags    MatcherMode = "tags"    // JSON数组标签查询
	MatcherKeyword MatcherMode = "keyword" // 关键词查询（多字段OR查询）
)

// Empty 空值标记
var Empty = struct{}{}

// Matcher 查询匹配器
type Matcher struct {
	Name        string
	Value       any
	Mode        MatcherMode
	CustomTable bool // 自定义表
}

// KeywordMatcher 关键词匹配器（多字段OR查询）
type KeywordMatcher struct {
	Names       []string // 搜索的字段名列表
	Value       any      // 搜索值
	CustomTable bool     // 自定义表
}

// NewMatcher 创建查询匹配器
func NewMatcher(name string, value any, mode MatcherMode) Matcher {
	return Matcher{
		Name:  name,
		Value: value,
		Mode:  mode,
	}
}

// NewKeywordMatcher 创建关键词匹配器
func NewKeywordMatcher(names []string, value any) *KeywordMatcher {
	return &KeywordMatcher{
		Names: names,
		Value: value,
	}
}

// NewKeywordMatcherWithCustomTable 创建自定义表的关键词匹配器
func NewKeywordMatcherWithCustomTable(names []string, value any, customTable bool) *KeywordMatcher {
	return &KeywordMatcher{
		Names:       names,
		Value:       value,
		CustomTable: customTable,
	}
}

// SnakeName 获取蛇形命名的字段名
func (m *Matcher) SnakeName() string {
	return CamelToSnake(m.Name)
}

// CamelToSnake 驼峰转蛇形命名
func CamelToSnake(s string) string {
	if s == "" {
		return ""
	}

	var result []byte
	for i, r := range s {
		if i > 0 && 'A' <= r && r <= 'Z' {
			result = append(result, '_')
		}
		if 'A' <= r && r <= 'Z' {
			result = append(result, byte(r+'a'-'A'))
		} else {
			result = append(result, byte(r))
		}
	}
	return string(result)
}

// BaseRepository 基础仓库实现
type BaseRepository[T any, ID comparable] struct {
	tableName string
	db        *gorm.DB
	getDB     func(ctx context.Context) *gorm.DB
}

// NewRepository 创建新的仓库实例，直接传入数据库连接
func NewRepository[T any, ID comparable](db *gorm.DB) Repository[T, ID] {
	return &BaseRepository[T, ID]{
		db: db,
	}
}

// NewRepositoryWithGetter 创建仓库实例，使用函数获取数据库连接（支持事务）
func NewRepositoryWithGetter[T any, ID comparable](getDB func(ctx context.Context) *gorm.DB) Repository[T, ID] {
	return &BaseRepository[T, ID]{
		getDB: getDB,
	}
}

// NewRepositoryFromApp 从应用容器创建仓库实例
func NewRepositoryFromApp[T any, ID comparable](app interface {
	GetDatabase() (*gorm.DB, error)
	GetConfig() *Config
}) (Repository[T, ID], error) {
	db, err := app.GetDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to get database from app: %w", err)
	}
	repo := NewRepository[T, ID](db)
	return repo, nil
}

// GetDB 获取数据库实例
func (r *BaseRepository[T, ID]) GetDB(ctx context.Context) *gorm.DB {
	// 优先从 context 中获取事务连接
	if r.getDB != nil {
		return r.getDB(ctx)
	}

	// 检查 context 中是否有事务（使用统一的上下文键）
	if tx, ok := ctx.Value(dbContextKey).(*gorm.DB); ok {
		return tx
	}
	return r.db
}

// GetTableName 获取表名
func (r *BaseRepository[T, ID]) GetTableName() string {
	if r.tableName == "" {
		var t T
		typeOf := reflect.TypeOf(t)
		if typeOf.Kind() == reflect.Ptr {
			typeOf = typeOf.Elem()
		}

		// 尝试调用 TableName 方法
		instance := reflect.New(typeOf)
		if method := instance.MethodByName("TableName"); method.IsValid() {
			if ret := method.Call(nil); len(ret) > 0 {
				r.tableName = ret[0].String()
			}
		}

		// 如果没有 TableName 方法，使用类型名的复数形式
		if r.tableName == "" {
			r.tableName = CamelToSnake(typeOf.Name()) + "s"
		}
	}
	return r.tableName
}

// Create 创建实体
func (r *BaseRepository[T, ID]) Create(ctx context.Context, entity *T) error {
	db := r.GetDB(ctx)
	return db.Create(entity).Error
}

// CreateInBatches 批量创建实体
func (r *BaseRepository[T, ID]) CreateInBatches(ctx context.Context, entities []T, batchSize int) error {
	db := r.GetDB(ctx)
	return db.CreateInBatches(entities, batchSize).Error
}

// CreateOrUpdate 创建或更新实体
func (r *BaseRepository[T, ID]) CreateOrUpdate(ctx context.Context, entity *T) error {
	db := r.GetDB(ctx)
	return db.Clauses(clause.OnConflict{UpdateAll: true}).Create(entity).Error
}

// FindById 根据ID查找实体
func (r *BaseRepository[T, ID]) FindById(ctx context.Context, id ID) (T, error) {
	var entity T
	db := r.GetDB(ctx)
	err := db.Table(r.GetTableName()).Where("id = ?", id).First(&entity).Error
	return entity, err
}

// FindByIdExists 根据ID查找实体，返回是否存在
func (r *BaseRepository[T, ID]) FindByIdExists(ctx context.Context, id ID) (T, bool, error) {
	entity, err := r.FindById(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			var zero T
			return zero, false, nil
		}
		var zero T
		return zero, false, err
	}
	return entity, true, nil
}

// FindByIdIn 根据ID列表查找实体
func (r *BaseRepository[T, ID]) FindByIdIn(ctx context.Context, ids []ID) ([]T, error) {
	var entities []T
	if len(ids) == 0 {
		return entities, nil
	}
	db := r.GetDB(ctx)
	err := db.Table(r.GetTableName()).Where("id in ?", ids).Find(&entities).Error
	return entities, err
}

// FindAll 查找所有实体
func (r *BaseRepository[T, ID]) FindAll(ctx context.Context) ([]T, error) {
	var entities []T
	db := r.GetDB(ctx)
	err := db.Table(r.GetTableName()).Find(&entities).Error
	return entities, err
}

// ExistsById 检查实体是否存在
func (r *BaseRepository[T, ID]) ExistsById(ctx context.Context, id ID) (bool, error) {
	var count int64
	db := r.GetDB(ctx)
	err := db.Table(r.GetTableName()).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// UpdateById 根据ID更新实体
func (r *BaseRepository[T, ID]) UpdateById(ctx context.Context, entity *T) error {
	db := r.GetDB(ctx)
	return db.Table(r.GetTableName()).Save(entity).Error
}

// UpdateColumnsById 根据ID更新指定列
func (r *BaseRepository[T, ID]) UpdateColumnsById(ctx context.Context, id ID, columns map[string]interface{}) error {
	db := r.GetDB(ctx)
	return db.Table(r.GetTableName()).Where("id", id).UpdateColumns(columns).Error
}

// Save 保存实体（包含零值）
func (r *BaseRepository[T, ID]) Save(ctx context.Context, entity *T) error {
	db := r.GetDB(ctx)
	return db.Table(r.GetTableName()).Save(entity).Error
}

// DeleteById 根据ID删除实体
func (r *BaseRepository[T, ID]) DeleteById(ctx context.Context, id ID) error {
	db := r.GetDB(ctx)
	return db.Table(r.GetTableName()).Where("id = ?", id).Delete(nil).Error
}

// DeleteByIdIn 根据ID列表删除实体
func (r *BaseRepository[T, ID]) DeleteByIdIn(ctx context.Context, ids []ID) error {
	db := r.GetDB(ctx)
	return db.Table(r.GetTableName()).Where("id in ?", ids).Delete(nil).Error
}

// Count 统计实体数量
func (r *BaseRepository[T, ID]) Count(ctx context.Context) (int64, error) {
	db := r.GetDB(ctx)
	var total int64
	err := db.Table(r.GetTableName()).Count(&total).Error
	return total, err
}

// Find 条件查询
func (r *BaseRepository[T, ID]) Find(ctx context.Context, matchers []Matcher, sort Sort) ([]T, error) {
	var items []T
	db := r.GetDB(ctx)

	// 验证排序字段安全性
	sortCopy := sort // 创建副本避免修改原始对象
	if err := sortCopy.Validate(); err != nil {
		return nil, fmt.Errorf("sort validation failed: %w", err)
	}

	db, err := r.match(db, matchers)
	if err != nil {
		return nil, err
	}

	if !sort.IsEmpty() {
		// 使用安全的字段名（已通过白名单验证）
		db = db.Order(fmt.Sprintf(`%s %s`, r.wrap(sort.Field), sort.Order))
	}

	err = db.Table(r.GetTableName()).Find(&items).Error
	return items, err
}

// wrap 包装字段名为表名.字段名格式
func (r *BaseRepository[T, ID]) wrap(property string) string {
	return fmt.Sprintf("%s.%s", r.GetTableName(), property)
}

// wrapQuery 包装查询字段名
func (r *BaseRepository[T, ID]) wrapQuery(m Matcher) string {
	if m.CustomTable {
		return m.SnakeName()
	}
	return fmt.Sprintf("%s.%s", r.GetTableName(), m.SnakeName())
}

// likeValue 包装LIKE查询的值
func (r *BaseRepository[T, ID]) likeValue(v any) string {
	return "%" + cast.ToString(v) + "%"
}

// ApplyKeywordMatcher 应用关键词匹配器到查询
func ApplyKeywordMatcher(db *gorm.DB, keywordMatcher *KeywordMatcher, tableName string) (*gorm.DB, error) {
	if keywordMatcher == nil {
		return db, nil
	}
	if keywordMatcher.Value == Empty || cast.ToString(keywordMatcher.Value) == "" {
		return db, nil
	}

	// 包装LIKE查询的值
	likeValue := func(v any) string {
		return "%" + cast.ToString(v) + "%"
	}

	databaseType := DatabaseType(db.Dialector.Name())
	var exprList = make([]clause.Expression, 0, len(keywordMatcher.Names))

	for _, name := range keywordMatcher.Names {
		var field string
		if keywordMatcher.CustomTable {
			field = CamelToSnake(name)
		} else {
			field = fmt.Sprintf("%s.%s", tableName, CamelToSnake(name))
		}

		switch databaseType {
		case DatabaseMysql:
			exprList = append(exprList, gorm.Expr(fmt.Sprintf("LOWER(%s) like LOWER(?)", field), likeValue(keywordMatcher.Value)))
		case DatabaseSqlite:
			exprList = append(exprList, gorm.Expr(fmt.Sprintf("%s like ?", field), likeValue(keywordMatcher.Value)))
		case DatabasePostgres, DatabasePostgresql:
			exprList = append(exprList, gorm.Expr(fmt.Sprintf("%s ILIKE ?", field), likeValue(keywordMatcher.Value)))
		default:
			return nil, fmt.Errorf("database type %s not supported for keyword search", databaseType)
		}
	}

	if len(exprList) > 0 {
		db = db.Clauses(clause.OrConditions{
			Exprs: exprList,
		})
	}

	return db, nil
}

// ApplyMatchersWithKeyword 应用匹配器和关键词匹配器到查询
func ApplyMatchersWithKeyword(db *gorm.DB, matchers []Matcher, keywordMatcher *KeywordMatcher, tableName string, wrapQueryFunc func(Matcher) string) (*gorm.DB, error) {
	// 先应用普通匹配器
	var err error
	db, err = ApplyMatchers(db, matchers, tableName, wrapQueryFunc)
	if err != nil {
		return nil, err
	}

	// 再应用关键词匹配器
	db, err = ApplyKeywordMatcher(db, keywordMatcher, tableName)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// ApplyMatchers 应用匹配器到查询（独立函数，可被其他地方复用）
func ApplyMatchers(db *gorm.DB, matchers []Matcher, tableName string, wrapQueryFunc func(Matcher) string) (*gorm.DB, error) {
	// 使用数据库类型
	databaseType := DatabaseType(db.Dialector.Name())
	for _, matcher := range matchers {
		if matcher.Value == Empty {
			matcher.Value = ""
		} else {
			if matcher.Name == "" || matcher.Value == "" {
				continue
			}
		}

		var queryField string
		if wrapQueryFunc != nil {
			queryField = wrapQueryFunc(matcher)
		} else {
			// 默认处理：如果有自定义表则不加前缀，否则使用表名前缀
			if matcher.CustomTable {
				queryField = matcher.SnakeName()
			} else if tableName != "" {
				queryField = fmt.Sprintf("%s.%s", tableName, matcher.SnakeName())
			} else {
				queryField = matcher.SnakeName()
			}
		}

		// 包装LIKE查询的值
		likeValue := func(v any) string {
			return "%" + cast.ToString(v) + "%"
		}

		switch matcher.Mode {
		case MatcherContains:
			db = db.Where(fmt.Sprintf("%s like ?", queryField), likeValue(matcher.Value))
		case MatcherContainsIgnoreCase:
			switch databaseType {
			case DatabaseMysql:
				// 转大写之后再比较
				db = db.Where(fmt.Sprintf("LOWER(%s) like LOWER(?)", queryField), likeValue(matcher.Value))
			case DatabaseSqlite:
				db = db.Where(fmt.Sprintf("%s like ?", queryField), likeValue(matcher.Value))
			case DatabasePostgres, DatabasePostgresql:
				db = db.Where(fmt.Sprintf("%s ILIKE ?", queryField), likeValue(matcher.Value))
			default:
				return nil, fmt.Errorf("database type %s not supported for case-insensitive search", databaseType)
			}
		case MatcherEqual:
			db = db.Where(fmt.Sprintf("%s = ?", queryField), matcher.Value)
		case MatcherIn:
			db = db.Where(fmt.Sprintf("%s in ?", queryField), matcher.Value)
		case MatcherNotContains:
			db = db.Where(fmt.Sprintf("%s not like ?", queryField), likeValue(matcher.Value))
		case MatcherNotContainsIgnoreCase:
			switch databaseType {
			case DatabaseMysql:
				// 转大写之后再比较
				db = db.Where(fmt.Sprintf("LOWER(%s) not like LOWER(?)", queryField), likeValue(matcher.Value))
			case DatabaseSqlite:
				db = db.Where(fmt.Sprintf("%s not like ?", queryField), likeValue(matcher.Value))
			case DatabasePostgres, DatabasePostgresql:
				db = db.Where(fmt.Sprintf("%s not ILIKE ?", queryField), likeValue(matcher.Value))
			default:
				return nil, fmt.Errorf("database type %s not supported for case-insensitive search", databaseType)
			}
		case MatcherNotEqual:
			db = db.Where(fmt.Sprintf("%s != ?", queryField), matcher.Value)
		case MatcherNotIn:
			db = db.Where(fmt.Sprintf("%s not in ?", queryField), matcher.Value)
		case MatcherTags:
			tags := strings.Split(cast.ToString(matcher.Value), ",")
			for _, tag := range tags {
				db = db.Where(datatypes.JSONArrayQuery(matcher.SnakeName()).Contains(tag))
			}
		}
	}
	return db, nil
}

// match 处理查询匹配器
func (r *BaseRepository[T, ID]) match(db *gorm.DB, matchers []Matcher) (*gorm.DB, error) {
	return ApplyMatchers(db, matchers, r.GetTableName(), r.wrapQuery)
}

// FindOne 查找单个实体
func (r *BaseRepository[T, ID]) FindOne(ctx context.Context, matchers []Matcher) (T, error) {
	var entity T
	db := r.GetDB(ctx)

	db, err := r.match(db, matchers)
	if err != nil {
		return entity, err
	}

	err = db.Table(r.GetTableName()).First(&entity).Error
	return entity, err
}

// Exists 检查是否存在符合条件的实体
func (r *BaseRepository[T, ID]) Exists(ctx context.Context, matchers []Matcher) (bool, error) {
	count, err := r.CountByMatchers(ctx, matchers)
	return count > 0, err
}

// CountByMatchers 根据条件统计数量
func (r *BaseRepository[T, ID]) CountByMatchers(ctx context.Context, matchers []Matcher) (int64, error) {
	db := r.GetDB(ctx)

	db, err := r.match(db, matchers)
	if err != nil {
		return 0, err
	}

	var count int64
	err = db.Table(r.GetTableName()).Count(&count).Error
	return count, err
}
