package orz

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"reflect"

	"gorm.io/gorm/clause"
)

type Repo[T any, ID comparable] struct {
	tableName string
}

func (r *Repo[T, ID]) GetDB(ctx context.Context) *gorm.DB {
	val, ok := ctx.Value(_DB).(*gorm.DB)
	if !ok {
		return MustGetDB()
	}
	return val
}

func (r *Repo[T, ID]) GetTableName() string {
	if r.tableName == "" {
		var t T
		typeOf := reflect.TypeOf(t)
		s := reflect.New(typeOf)
		tableName := s.MethodByName("TableName")
		ret := tableName.Call(nil)
		if len(ret) != 0 {
			r.tableName = ret[0].String()
		}
	}
	return r.tableName
}

func (r *Repo[T, ID]) Create(ctx context.Context, m *T) (err error) {
	db := r.GetDB(ctx)
	err = db.Create(m).Error
	return
}

func (r *Repo[T, ID]) CreateOrUpdate(ctx context.Context, m *T) (err error) {
	db := r.GetDB(ctx)
	db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&m)
	return
}

func (r *Repo[T, ID]) CreateInBatches(ctx context.Context, m []T, batchSize int) (err error) {
	db := r.GetDB(ctx)
	err = db.CreateInBatches(m, batchSize).Error
	return
}

func (r *Repo[T, ID]) ExistsById(ctx context.Context, id ID) (exists bool, err error) {
	db := r.GetDB(ctx)
	var count int64
	err = db.Table(r.GetTableName()).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

func (r *Repo[T, ID]) FindById(ctx context.Context, id ID) (m T, err error) {
	db := r.GetDB(ctx)
	err = db.Table(r.GetTableName()).Where("id = ?", id).First(&m).Error
	return m, err
}

func (r *Repo[T, ID]) FindByIdExists(ctx context.Context, id ID) (m T, exists bool, err error) {
	m, err = r.FindById(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return m, false, nil
	}
	return m, true, err
}

func (r *Repo[T, ID]) FindByIdForUpdate(ctx context.Context, id ID) (m T, exists bool, err error) {
	db := r.GetDB(ctx)
	err = db.Clauses(clause.Locking{Strength: "UPDATE"}).Table(r.GetTableName()).Where("id = ?", id).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return m, false, nil
	}
	return m, true, err
}

func (r *Repo[T, ID]) FindByIdForShare(ctx context.Context, id ID) (m T, exists bool, err error) {
	db := r.GetDB(ctx)
	err = db.Clauses(clause.Locking{Strength: "SHARE", Table: clause.Table{Name: clause.CurrentTable}}).Table(r.GetTableName()).Where("id = ?", id).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return m, false, nil
	}
	return m, true, err
}

func (r *Repo[T, ID]) FindByIdForUpdateNowait(ctx context.Context, id ID) (m T, exists bool, err error) {
	db := r.GetDB(ctx)
	err = db.Clauses(clause.Locking{Strength: "UPDATE", Options: "NOWAIT"}).Table(r.GetTableName()).Where("id = ?", id).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return m, false, nil
	}
	return m, true, err
}

func (r *Repo[T, ID]) FindByIdIn(c context.Context, ids []ID) (o []T, err error) {
	err = r.GetDB(c).Where("id in ?", ids).Find(&o).Error
	return
}

func (r *Repo[T, ID]) FindAll(ctx context.Context) (m []T, err error) {
	db := r.GetDB(ctx)
	err = db.Table(r.GetTableName()).Find(&m).Error
	return
}

func (r *Repo[T, ID]) DeleteById(ctx context.Context, id ID) (err error) {
	db := r.GetDB(ctx)
	err = db.Table(r.GetTableName()).Where("id = ?", id).Delete(nil).Error
	return
}

func (r *Repo[T, ID]) DeleteByIdIn(c context.Context, ids []ID) error {
	return r.GetDB(c).Table(r.GetTableName()).Where("id in ?", ids).Delete(nil).Error
}

// UpdateById 根据主键更新数据，0，nil， false 空值等不会更新
func (r *Repo[T, ID]) UpdateById(ctx context.Context, m *T) (err error) {
	db := r.GetDB(ctx)
	err = db.Table(r.GetTableName()).Updates(m).Error
	return
}

func (r *Repo[T, ID]) UpdateColumnsById(ctx context.Context, id ID, m Map) (err error) {
	db := r.GetDB(ctx)
	var values = make(map[string]interface{}, len(m))
	for k, v := range m {
		values[k] = v
	}
	err = db.Table(r.GetTableName()).Where("id", id).UpdateColumns(values).Error
	return
}

// Save 保存所有的字段，即使字段是零值
func (r *Repo[T, ID]) Save(ctx context.Context, m *T) (err error) {
	db := r.GetDB(ctx)
	err = db.Table(r.GetTableName()).Save(m).Error
	return
}

func (r *Repo[T, ID]) Count(ctx context.Context) (total int64, err error) {
	db := r.GetDB(ctx)
	err = db.Table(r.GetTableName()).Count(&total).Error
	return
}
