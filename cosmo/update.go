package cosmo

import (
	"errors"
	"github.com/hwcer/cosgo/cosmo/clause"
	"github.com/hwcer/cosgo/cosmo/schema"
	"go.mongodb.org/mongo-driver/bson"
	"reflect"
	"strings"
)

// Merge 只更新Model,不会修改数据库
// db.Model(m).Merge(i)
//参数支持 Struct,map[string]interface{}
func (db *DB) Merge(i interface{}) error {
	tx := db.Statement.Parse()
	if tx.Error != nil {
		return tx.Error
	}
	update, err := tx.bson(i)
	if err != nil {
		return err
	}

	var ok bool
	var values bson.M
	if values, ok = update[clause.UpdateTypeSet].(bson.M); !ok || len(values) == 0 {
		return nil
	}
	for k, v := range values {
		if err = tx.Statement.SetColumn(k, v); err != nil {
			return err
		}
	}
	return nil
}

func (db *DB) bson(i interface{}) (bson.M, error) {
	reflectValue := reflect.Indirect(reflect.ValueOf(i))
	switch reflectValue.Kind() {
	case reflect.Map:
		return db.parseMap(i)
	case reflect.Struct:
		return db.parseStruct(reflectValue)
	default:
		return nil, errors.New("类型错误")
	}
}

func (db *DB) parseMap(dest interface{}) (update bson.M, err error) {
	var destMap bson.M
	switch dest.(type) {
	case map[string]interface{}:
		destMap = bson.M(dest.(map[string]interface{}))
	case bson.M:
		destMap = dest.(bson.M)
	default:
		err = errors.New("Update方法参数仅支持 struct 和 map[string]interface{}")
	}
	if err != nil {
		return
	}

	update = make(bson.M)
	setData := make(bson.M)

	for k, v := range destMap {
		if strings.HasPrefix(k, clause.QueryOperationPrefix) {
			update[k] = v
		} else {
			setData[k] = v
		}
	}
	if len(setData) > 0 {
		update[clause.UpdateTypeSet] = setData
	}

	return
}

func (db *DB) parseStruct(reflectValue reflect.Value) (update bson.M, err error) {
	data := bson.M{}
	var destSchema *schema.Schema
	destSchema, err = db.Schema.Parse(reflectValue.Interface())
	if err != nil {
		return
	}

	for _, field := range destSchema.Fields {
		v := reflectValue.FieldByIndex(field.StructField.Index)
		if v.IsValid() && !v.IsZero() {
			data[field.DBName] = v.Interface()
		}
	}

	delete(data, clause.MongoPrimaryName)
	//if v, ok := data[clause.MongoPrimaryName]; ok {
	//	db.Where(v)
	//}

	update = bson.M{}
	update[clause.UpdateTypeSet] = data
	return
}
