package orz

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/go-orz/orz/config"
	"github.com/spf13/cast"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type SortDirection string

const (
	ASC  SortDirection = "asc"  // 正序
	DESC SortDirection = "desc" // 倒序
)

func NewSort(direction SortDirection, property string, allowedProperties ...string) Sort {
	if direction == "" {
		direction = DESC
	}
	return Sort{
		direction:         direction,
		property:          property,
		allowedProperties: allowedProperties,
	}
}

type Sort struct {
	direction         SortDirection // 排序方式
	property          string        // 排序字段
	allowedProperties []string      // 允许的排序字段
}

func (r Sort) Property() string {
	return CamelToSnake(r.property)
}

func (r Sort) Direction() SortDirection {
	switch r.direction {
	case "descend", DESC:
		return DESC
	case "ascend", ASC:
		return ASC
	default:
		return DESC
	}
}

func (r Sort) IsInvalid() bool {
	if r.property == "" {
		return false
	}
	return !slices.Contains(r.allowedProperties, r.property)
}

func NewPageRequest(pageIndex, pageSize int, sort Sort, matchers []Matcher) PageRequest {
	if pageIndex < 1 {
		pageIndex = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	return PageRequest{
		PageIndex: pageIndex,
		PageSize:  pageSize,
		Sort:      sort,
		Matchers:  matchers,
	}
}

type PageRequest struct {
	PageIndex int
	PageSize  int
	Sort      Sort
	Matchers  []Matcher
	Modifier  func(db *gorm.DB) *gorm.DB
	result    any
	result1   bool
}

func (r *PageRequest) Result(result any) {
	r.result = result
	r.result1 = true
}

func (r *PageRequest) Offset() int {
	return (r.PageIndex - 1) * r.PageSize
}

func (r *PageRequest) Limit() int {
	return r.PageSize
}

type PageResult[T any] struct {
	Items []T   `json:"items"`
	Total int64 `json:"total"`
}

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

	MatcherTags MatcherMode = "tags"
)

var Empty = struct{}{}

func NewMatcher(name string, value any, mode MatcherMode) Matcher {
	return Matcher{
		Name:  name,
		Value: value,
		Mode:  mode,
	}
}

type Matcher struct {
	Name        string
	Value       any
	Mode        MatcherMode
	CustomTable bool // 自定义表
}

func (r *Matcher) SnakeName() string {
	return CamelToSnake(r.Name)
}

func (r *Repo[T, ID]) wrap(property string) string {
	return fmt.Sprintf("%s.%s", r.GetTableName(), property)
}

func (r *Repo[T, ID]) wrapQuery(m Matcher) string {
	if m.CustomTable {
		return m.SnakeName()
	}
	return fmt.Sprintf("%s.%s", r.GetTableName(), m.SnakeName())
}

func (r *Repo[T, ID]) likeValue(v any) string {
	return "%" + cast.ToString(v) + "%"
}

func (r *Repo[T, ID]) Page(ctx context.Context, pageRequest *PageRequest) (page *PageResult[T], err error) {
	if pageRequest.Sort.IsInvalid() {
		return nil, errors.New(fmt.Sprintf("sort property %s is not invalid", pageRequest.Sort.Property()))
	}

	page = &PageResult[T]{}
	db := r.GetDB(ctx)
	if pageRequest.Modifier != nil {
		db = pageRequest.Modifier(db)
	}

	db, err = r.match(db, pageRequest.Matchers)
	if err != nil {
		return nil, err
	}

	sort := pageRequest.Sort
	if sort.Property() != "" {
		db = db.Order(fmt.Sprintf(`%s %s`, r.wrap(sort.Property()), sort.Direction()))
	}

	err = db.Table(r.GetTableName()).Count(&(page.Total)).Error
	if err != nil {
		return page, err
	}

	db = db.Table(r.GetTableName()).Offset(pageRequest.Offset()).Limit(pageRequest.Limit())
	if pageRequest.result1 {
		err = db.Find(pageRequest.result).Error
	} else {
		err = db.Find(&(page.Items)).Error
	}
	return
}

func (r *Repo[T, ID]) Find(ctx context.Context, matchers []Matcher, sort Sort) (items []T, err error) {
	db := r.GetDB(ctx)

	db, err = r.match(db, matchers)
	if err != nil {
		return nil, err
	}

	if sort.Property() != "" {
		db = db.Order(fmt.Sprintf(`%s %s`, r.wrap(sort.Property()), sort.Direction()))
	}

	err = db.Table(r.GetTableName()).Find(&items).Error
	return items, err
}

func (r *Repo[T, ID]) match(db *gorm.DB, matchers []Matcher) (*gorm.DB, error) {
	databaseType := config.Conf().Database.Type
	for _, matcher := range matchers {
		if matcher.Value == Empty {
			matcher.Value = ""
		} else {
			if matcher.Name == "" || matcher.Value == "" {
				continue
			}
		}

		switch matcher.Mode {
		case MatcherContains:
			db = db.Where(fmt.Sprintf("%s like ?", r.wrapQuery(matcher)), r.likeValue(matcher.Value))
		case MatcherContainsIgnoreCase:
			switch databaseType {
			case config.DatabaseMysql:
				// 转大写之后再比较
				db = db.Where(fmt.Sprintf("LOWER(%s) like LOWER(?)", r.wrapQuery(matcher)), r.likeValue(matcher.Value))
			case config.DatabaseSqlite:
				db = db.Where(fmt.Sprintf("%s like ?", r.wrapQuery(matcher)), r.likeValue(matcher.Value))
			case config.DatabasePostgres:
				db = db.Where(fmt.Sprintf("%s ILIKE ?", r.wrapQuery(matcher)), r.likeValue(matcher.Value))
			case config.DatabaseClickhouse:
				return nil, errors.New("not supported yet")
			}
		case MatcherEqual:
			db = db.Where(fmt.Sprintf("%s = ?", r.wrapQuery(matcher)), matcher.Value)
		case MatcherIn:
			db = db.Where(fmt.Sprintf("%s in ?", r.wrapQuery(matcher)), matcher.Value)
		case MatcherNotContains:
			db = db.Where(fmt.Sprintf("%s not like ?", r.wrapQuery(matcher)), r.likeValue(matcher.Value))
		case MatcherNotContainsIgnoreCase:
			switch databaseType {
			case config.DatabaseMysql:
				// 转大写之后再比较
				db = db.Where(fmt.Sprintf("LOWER(%s) not like LOWER(?)", r.wrapQuery(matcher)), r.likeValue(matcher.Value))
			case config.DatabaseSqlite:
				db = db.Where(fmt.Sprintf("%s not like ?", r.wrapQuery(matcher)), r.likeValue(matcher.Value))
			case config.DatabasePostgres:
				db = db.Where(fmt.Sprintf("%s not ILIKE ?", r.wrapQuery(matcher)), r.likeValue(matcher.Value))
			case config.DatabaseClickhouse:
				return nil, errors.New("not supported yet")
			}
		case MatcherNotEqual:
			db = db.Where(fmt.Sprintf("%s != ?", r.wrapQuery(matcher)), matcher.Value)
		case MatcherNotIn:
			db = db.Where(fmt.Sprintf("%s not in ?", r.wrapQuery(matcher)), matcher.Value)
		case MatcherTags:
			tags := strings.Split(cast.ToString(matcher.Value), ",")
			for _, tag := range tags {
				db = db.Where(datatypes.JSONArrayQuery(matcher.SnakeName()).Contains(tag))
			}
		}
	}
	return db, nil
}
