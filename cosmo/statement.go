package cosmo

import (
	"context"
	"errors"
	"github.com/hwcer/cosgo/cosmo/clause"
	"github.com/hwcer/cosgo/cosmo/schema"
	"go.mongodb.org/mongo-driver/bson"
	"reflect"
)

func NewStatement(db *DB) *Statement {
	return &Statement{
		DB:       db,
		Context:  context.Background(),
		Clause:   clause.New(),
		Paging:   &Paging{},
		settings: map[string]interface{}{},
	}
}

// Statement statement
type Statement struct {
	*DB
	Dest         interface{}
	Table        string
	Model        interface{}
	ReflectValue reflect.Value
	ReflectModel reflect.Value
	Omits        []string // omit columns
	Selects      []string // selected columns
	Schema       *schema.Schema
	Context      context.Context
	Clause       *clause.Query
	Paging       *Paging
	settings     map[string]interface{}
	multiple     bool
}

func (stmt *Statement) Parse() (tx *DB) {
	tx = stmt.DB
	if tx.Error != nil {
		return
	}
	// assign Dest values
	if stmt.Dest != nil {
		stmt.ReflectValue = reflect.ValueOf(stmt.Dest)
		for stmt.ReflectValue.Kind() == reflect.Ptr {
			if stmt.ReflectValue.IsNil() && stmt.ReflectValue.CanAddr() {
				stmt.ReflectValue.Set(reflect.New(stmt.ReflectValue.Type().Elem()))
			}
			stmt.ReflectValue = stmt.ReflectValue.Elem()
		}
		if !stmt.ReflectValue.IsValid() {
			tx.Errorf(ErrInvalidValue)
			return
		}
	}
	if stmt.Model != nil {
		stmt.ReflectModel = reflect.ValueOf(stmt.Model)
		if !stmt.ReflectModel.IsValid() || reflect.Indirect(stmt.ReflectModel).Kind() != reflect.Struct {
			tx.Errorf(ErrInvalidValue)
			return
		}
	}
	//fmt.Printf("Execute:%v,%+v\n", stmt.ReflectValue.Kind(), stmt.ReflectValue.Interface())
	var err error
	var model interface{}
	if stmt.Model != nil {
		model = stmt.Model
	} else {
		model = stmt.Dest
	}
	if stmt.Schema, err = tx.Schema.ParseWithSpecialTableName(model, ""); err != nil && !errors.Is(err, schema.ErrUnsupportedDataType) {
		tx.Errorf(ErrInvalidValue)
		return
	} else if stmt.Schema != nil && stmt.Table == "" {
		stmt.Table = stmt.Schema.Table
	}

	if stmt.Table == "" {
		tx.Errorf(errors.New("Table not set, please set it like: db.Model(&user) or db.Table(\"users\")"))
	}
	return
}

// SetColumn set column's value
//   stmt.SetColumn("Name", "jinzhu") // Hooks Method
func (stmt *Statement) SetColumn(name string, value interface{}) error {
	if stmt.Model == nil || stmt.Schema == nil || stmt.ReflectModel.Kind() != reflect.Ptr {
		return nil
	}
	if stmt.ReflectModel.Kind() != reflect.Ptr {
		return errors.New("非指针类型无法赋值")
	}
	field := stmt.Schema.LookUpField(name)
	if field == nil {
		return ErrInvalidField
	}
	reflectModel := reflect.Indirect(stmt.ReflectModel)
	return field.Set(reflectModel, value)
}

//projection 不能同时使用Select和Omit 优先Select生效
//可以使用model属性名或者数据库字段名
func (stmt *Statement) projection() (projection bson.M, order bson.D, err error) {
	fields := make(bson.M)
	for _, k := range stmt.Selects {
		fields[k] = 1
	}
	for _, k := range stmt.Omits {
		fields[k] = 0
	}
	if stmt.Schema == nil {
		projection = fields
		order = stmt.Paging.order
		return
	}
	projection = make(bson.M)
	for k, v := range fields {
		if field := stmt.Schema.LookUpField(k); field != nil {
			projection[field.DBName] = v
		} else {
			projection[k] = v
		}
	}

	for _, e := range stmt.Paging.order {
		if field := stmt.Schema.LookUpField(e.Key); field != nil {
			e.Key = field.DBName
		}
		order = append(order, e)
	}

	return
}
