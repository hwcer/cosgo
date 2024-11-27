package schema

import (
	"errors"
	"fmt"
	"reflect"
)

// ErrUnsupportedDataType unsupported data type
var ErrUnsupportedDataType = errors.New("unsupported data type")

type Schema struct {
	err  error
	init chan struct{}

	options        *Options
	Name           string
	Table          string
	Embedded       []*Field //嵌入字段
	ModelType      reflect.Type
	Fields         map[string]*Field
	fieldsPrivate  []*Field          //不包含内嵌结构的字段,需要使用所有字段，请使用Schema.Range 或者遍历 FieldsByName FieldsByDBName
	fieldsDatabase map[string]*Field //数据库字段索引
}

func (schema *Schema) initialized() {
	close(schema.init)
	schema.init = nil
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

// Make Make a Slice
func (schema *Schema) Make() reflect.Value {
	slice := reflect.MakeSlice(reflect.SliceOf(reflect.PtrTo(schema.ModelType)), 0, 20)
	results := reflect.New(slice.Type())
	results.Elem().Set(slice)
	return results
}

// Range 遍历字段
func (schema *Schema) Range(cb func(*Field) bool) {
	for _, field := range schema.fieldsDatabase {
		if !cb(field) {
			return
		}
	}
}
func (schema *Schema) LookUpField(name string) *Field {
	if field, ok := schema.Fields[name]; ok {
		return field
	}
	if field, ok := schema.fieldsDatabase[name]; ok {
		return field
	}
	return nil
}

func (schema *Schema) GetValue(obj any, key string, keys ...any) (r any) {
	keys = append([]any{key}, keys...)
	vf := ValueOf(obj)
	var sch *Schema
	var field *Field
	l := len(keys)
	for i := 0; i < l; i++ {
		sk := ToString(keys[i])
		if field != nil {
			sch = field.Embedded
		} else {
			sch = schema
		}
		if sch == nil {
			return nil
		}
		field = sch.LookUpField(sk)
		if field == nil {
			return nil
		}
		vf = field.Get(vf)
		if !vf.IsValid() {
			return nil
		}
	}
	return vf.Interface()
}

func (schema *Schema) SetValue(obj any, val any, key string, keys ...any) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%v", e)
		}
	}()
	keys = append([]any{key}, keys...)
	vf := ValueOf(obj)
	var sch *Schema
	var field *Field
	l := len(keys)
	n := l - 1
	for i := 0; i < l; i++ {
		sk := ToString(keys[i])
		if field != nil {
			sch = field.Embedded
		} else {
			sch = schema
		}
		if sch == nil {
			return fmt.Errorf("field not object:%v", keys)
		}
		field = sch.LookUpField(sk)
		if field == nil {
			return fmt.Errorf("field not exist:%v", sk)
		}
		if i < n {
			vf = field.Get(vf)

			//if v := field.Get(vf); v.IsZero() {
			//	v = reflect.New(field.FieldType)
			//	if err = field.Set(vf, v.Interface()); err != nil {
			//		return
			//	}
			//	vf = v
			//} else {
			//	vf = v
			//}

		}
	}
	return field.Set(vf, val)

}

func (schema *Schema) ParseField(fieldStruct reflect.StructField) *Field {
	//var err error

	field := &Field{
		Name:              fieldStruct.Name,
		FieldType:         fieldStruct.Type,
		IndirectFieldType: fieldStruct.Type,
		StructField:       fieldStruct,
		Schema:            schema,
	}

	for field.IndirectFieldType.Kind() == reflect.Ptr {
		field.IndirectFieldType = field.IndirectFieldType.Elem()
	}

	fieldValue := reflect.New(field.IndirectFieldType)
	if dbName := fieldStruct.Tag.Get("bson"); dbName != "" && dbName != "inline" {
		field.DBName = dbName
	} else {
		field.DBName = schema.options.ColumnName(schema.Table, field.Name)
	}
	field.Index = field.StructField.Index
	kind := reflect.Indirect(fieldValue).Kind()
	switch kind {
	case reflect.Struct:
		field.Embedded, schema.err = GetOrParse(fieldValue.Interface(), schema.options)
	case reflect.Map, reflect.Slice, reflect.Array:
		//初始化子结构
	case reflect.Invalid, reflect.Uintptr, reflect.Chan, reflect.Func, reflect.Interface, reflect.UnsafePointer, reflect.Complex64, reflect.Complex128:
		schema.err = fmt.Errorf("invalid embedded struct for %s's field %s, should be struct, but got %v", field.Schema.Name, field.Name, field.FieldType)
	}

	//if fieldStruct.Anonymous {
	//
	//}
	return field
}
