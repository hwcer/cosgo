package schema

import (
	"errors"
	"fmt"
	"reflect"
)

// ErrUnsupportedDataType unsupported data type
var ErrUnsupportedDataType = errors.New("unsupported data type")

type Schema struct {
	err            error
	options        *Options
	initialized    chan struct{}
	Name           string
	Table          string
	Fields         []*Field
	ModelType      reflect.Type
	FieldsByName   map[string]*Field
	FieldsByDBName map[string]*Field
}

func (schema *Schema) String() string {
	if schema.ModelType.Name() == "" {
		return fmt.Sprintf("%s(%s)", schema.Name, schema.Table)
	}
	return fmt.Sprintf("%s.%s", schema.ModelType.PkgPath(), schema.ModelType.Name())
}

func (schema *Schema) New() reflect.Value {
	results := reflect.New(schema.ModelType)
	return results
}

func (schema *Schema) MakeSlice() reflect.Value {
	slice := reflect.MakeSlice(reflect.SliceOf(reflect.PtrTo(schema.ModelType)), 0, 20)
	results := reflect.New(slice.Type())
	results.Elem().Set(slice)
	return results
}

func (schema *Schema) LookUpField(name string) *Field {
	if field, ok := schema.FieldsByDBName[name]; ok {
		return field
	}
	if field, ok := schema.FieldsByName[name]; ok {
		return field
	}
	return nil
}

// FieldDBName 查询对象字段对应的DBName
func (schema *Schema) FieldDBName(name string) string {
	if field := schema.LookUpField(name); field != nil {
		return field.DBName
	}
	return ""
}

func (schema *Schema) GetValue(i interface{}, key string) interface{} {
	field := schema.LookUpField(key)
	if field == nil {
		return nil
	}
	vf := reflect.Indirect(ValueOf(i))
	val := vf.FieldByIndex(field.Index)

	if val.IsValid() && !val.IsZero() {
		return val.Interface()
	}
	return nil
}

func (schema *Schema) SetValue(i interface{}, key string, val any) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%v", e)
		}
	}()
	field := schema.LookUpField(key)
	if field == nil {
		return fmt.Errorf("field not exist:%v", key)
	}
	reflectValue := reflect.ValueOf(i)
	return field.Set(reflectValue, val)
}
