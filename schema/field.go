package schema

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Field 表示结构体的一个字段，包含字段的元数据和操作方法
type Field struct {
	// 核心字段：字段标识和类型信息
	Name              string              // 字段名称
	FieldType         reflect.Type        // 字段类型
	IndirectFieldType reflect.Type        // 间接字段类型（去除指针）
	StructField       reflect.StructField // 原始结构体字段

	// 关联信息：与 Schema 和嵌入字段的关系
	Schema   *Schema // 所在的父 Schema
	Embedded *Schema // 嵌入子对象的 Schema

	// 访问信息：字段访问路径和缓存
	Index      []int                                  // 字段索引路径
	getValueFn func(reflect.Value) reflect.Value      // 缓存获取值的函数
	setValueFn func(reflect.Value, interface{}) error // 缓存设置值的函数

	// 命名缓存：避免重复计算
	dbName string // 缓存数据库字段名
	jsName string // 缓存 JSON 字段名
}

func (field *Field) JSName() string {
	if field.jsName == "" {
		field.jsName = field.GetName("json")
	}
	return field.jsName
}
func (field *Field) DBName() string {
	if field.dbName == "" {
		field.dbName = field.GetName("bson", "json")
	}
	return field.dbName
}

// GetName json,form,....
// 按照ks 指定的顺寻找以名字，所有TAG 都无法匹配时返回字段名
func (field *Field) GetName(ks ...string) string {
	for _, k := range ks {
		if v := field.GetTagName(k); v != "" {
			return v
		}
	}
	return field.Name
}

func (field *Field) GetTagName(k string) string {
	name := field.StructField.Tag.Get(k)
	if i := strings.Index(name, ","); i >= 0 {
		name = name[0:i]
	}
	if name == "-" {
		return ""
	}
	return name
}

func (field *Field) GetEmbeddedFields() (r []*Field) {
	if !field.StructField.Anonymous || field.Embedded == nil {
		return
	}
	for _, ef := range field.Embedded.Fields {
		i := *ef
		v := &i
		if field.FieldType.Kind() == reflect.Struct {
			v.Index = append([]int{field.StructField.Index[0]}, v.Index...)
		} else {
			v.Index = append([]int{-field.StructField.Index[0] - 1}, v.Index...)
		}
		r = append(r, v)
	}
	return
}

func (field *Field) Get(value reflect.Value) reflect.Value {
	return field.ReflectValueOf(value)
}

func (field *Field) ReflectValueOf(value reflect.Value) reflect.Value {
	// 如果没有缓存的获取函数，编译一个
	if field.getValueFn == nil {
		field.compileGetValueFn()
	}
	// 使用缓存的函数获取值
	return field.getValueFn(value)
}

// compileGetValueFn 编译获取值的函数，缓存反射访问路径
func (field *Field) compileGetValueFn() {
	switch {
	case len(field.Index) == 1:
		// 简单字段：直接访问
		idx := field.Index[0]
		field.getValueFn = func(value reflect.Value) reflect.Value {
			return reflect.Indirect(value).Field(idx)
		}
	case len(field.Index) == 2 && field.Index[0] >= 0 && field.FieldType.Kind() != reflect.Ptr:
		// 两层简单字段：直接访问
		idx1, idx2 := field.Index[0], field.Index[1]
		field.getValueFn = func(value reflect.Value) reflect.Value {
			return reflect.Indirect(value).Field(idx1).Field(idx2)
		}
	default:
		// 复杂字段：预编译索引路径
		indices := make([]int, len(field.Index))
		copy(indices, field.Index)
		field.getValueFn = func(value reflect.Value) reflect.Value {
			v := reflect.Indirect(value)
			for idx, fieldIdx := range indices {
				if fieldIdx >= 0 {
					v = v.Field(fieldIdx)
				} else {
					v = v.Field(-fieldIdx - 1)
				}
				if v.Kind() == reflect.Ptr {
					if v.Type().Elem().Kind() == reflect.Struct {
						if v.IsNil() {
							v.Set(reflect.New(v.Type().Elem()))
						}
					}
					if idx < len(indices)-1 {
						v = v.Elem()
					}
				}
			}
			return v
		}
	}
}

// Set valuer, setter when parse struct
func (field *Field) Set(value reflect.Value, v interface{}) error {
	// 如果没有缓存的设置函数，编译一个
	if field.setValueFn == nil {
		field.compileSetValueFn()
	}
	// 使用缓存的函数设置值
	return field.setValueFn(value, v)
}

// compileSetValueFn 编译设置值的函数，缓存类型判断和反射操作
func (field *Field) compileSetValueFn() {
	switch field.FieldType.Kind() {
	case reflect.Bool:
		field.setValueFn = field.SetBool
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		field.setValueFn = field.SetInt
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		field.setValueFn = field.SetUint
	case reflect.Float32, reflect.Float64:
		field.setValueFn = field.SetFloat
	case reflect.String:
		field.setValueFn = field.SetString
	default:
		field.setValueFn = func(value reflect.Value, v interface{}) error {
			return field.fallbackSetter(value, v, field.Set)
		}
	}
}

func (field *Field) fallbackSetter(value reflect.Value, v interface{}, setter func(reflect.Value, interface{}) error) (err error) {
	if v == nil {
		field.ReflectValueOf(value).Set(reflect.New(field.FieldType).Elem())
		return nil
	}
	reflectV := reflect.ValueOf(v)
	// Optimal value type acquisition for v
	reflectValType := reflectV.Type()

	if reflectValType.AssignableTo(field.FieldType) {
		field.ReflectValueOf(value).Set(reflectV)
		return
	} else if reflectValType.ConvertibleTo(field.FieldType) {
		field.ReflectValueOf(value).Set(reflectV.Convert(field.FieldType))
		return
	} else if field.FieldType.Kind() == reflect.Ptr {
		fieldValue := field.ReflectValueOf(value)
		fieldType := field.FieldType.Elem()

		if reflectValType.AssignableTo(fieldType) {
			if !fieldValue.IsValid() {
				fieldValue = reflect.New(fieldType)
			} else if fieldValue.IsNil() {
				fieldValue.Set(reflect.New(fieldType))
			}
			fieldValue.Elem().Set(reflectV)
			return
		} else if reflectValType.ConvertibleTo(fieldType) {
			if fieldValue.IsNil() {
				fieldValue.Set(reflect.New(fieldType))
			}

			fieldValue.Elem().Set(reflectV.Convert(fieldType))
			return
		}
	}

	if reflectV.Kind() == reflect.Ptr {
		if reflectV.IsNil() {
			field.ReflectValueOf(value).Set(reflect.New(field.FieldType).Elem())
		} else {
			err = setter(value, reflectV.Elem().Interface())
		}
	} else if valuer, ok := v.(driver.Valuer); ok {
		if v, err = valuer.Value(); err == nil {
			err = setter(value, v)
		}
	} else {
		return fmt.Errorf("failed to set value %+v to field %s", v, field.Name)
	}

	return
}

func (field *Field) SetBool(value reflect.Value, v interface{}) error {
	switch data := v.(type) {
	case bool:
		field.ReflectValueOf(value).SetBool(data)
	case *bool:
		if data != nil {
			field.ReflectValueOf(value).SetBool(*data)
		} else {
			field.ReflectValueOf(value).SetBool(false)
		}
	case int64:
		if data > 0 {
			field.ReflectValueOf(value).SetBool(true)
		} else {
			field.ReflectValueOf(value).SetBool(false)
		}
	case string:
		b, _ := strconv.ParseBool(data)
		field.ReflectValueOf(value).SetBool(b)
	default:
		return field.fallbackSetter(value, v, field.Set)
	}
	return nil
}
func (field *Field) SetInt(value reflect.Value, v interface{}) error {
	switch data := v.(type) {
	case int64:
		field.ReflectValueOf(value).SetInt(data)
	case int:
		field.ReflectValueOf(value).SetInt(int64(data))
	case int8:
		field.ReflectValueOf(value).SetInt(int64(data))
	case int16:
		field.ReflectValueOf(value).SetInt(int64(data))
	case int32:
		field.ReflectValueOf(value).SetInt(int64(data))
	case uint:
		field.ReflectValueOf(value).SetInt(int64(data))
	case uint8:
		field.ReflectValueOf(value).SetInt(int64(data))
	case uint16:
		field.ReflectValueOf(value).SetInt(int64(data))
	case uint32:
		field.ReflectValueOf(value).SetInt(int64(data))
	case uint64:
		field.ReflectValueOf(value).SetInt(int64(data))
	case float32:
		field.ReflectValueOf(value).SetInt(int64(data))
	case float64:
		field.ReflectValueOf(value).SetInt(int64(data))
	case []byte:
		return field.Set(value, string(data))
	case string:
		if data == "" {
			data = "0"
		}
		if i, err := strconv.ParseInt(data, 0, 64); err == nil {
			field.ReflectValueOf(value).SetInt(i)
		} else {
			return err
		}
	default:
		return field.fallbackSetter(value, v, field.Set)
	}
	return nil
}

func (field *Field) SetUint(value reflect.Value, v interface{}) error {
	switch data := v.(type) {
	case uint64:
		field.ReflectValueOf(value).SetUint(data)
	case uint:
		field.ReflectValueOf(value).SetUint(uint64(data))
	case uint8:
		field.ReflectValueOf(value).SetUint(uint64(data))
	case uint16:
		field.ReflectValueOf(value).SetUint(uint64(data))
	case uint32:
		field.ReflectValueOf(value).SetUint(uint64(data))
	case int64:
		field.ReflectValueOf(value).SetUint(uint64(data))
	case int:
		field.ReflectValueOf(value).SetUint(uint64(data))
	case int8:
		field.ReflectValueOf(value).SetUint(uint64(data))
	case int16:
		field.ReflectValueOf(value).SetUint(uint64(data))
	case int32:
		field.ReflectValueOf(value).SetUint(uint64(data))
	case float32:
		field.ReflectValueOf(value).SetUint(uint64(data))
	case float64:
		field.ReflectValueOf(value).SetUint(uint64(data))
	case []byte:
		return field.Set(value, string(data))
	case string:
		if data == "" {
			data = "0"
		}
		if i, err := strconv.ParseUint(data, 0, 64); err == nil {
			field.ReflectValueOf(value).SetUint(i)
		} else {
			return err
		}
	default:
		return field.fallbackSetter(value, v, field.Set)
	}
	return nil
}

func (field *Field) SetFloat(value reflect.Value, v interface{}) error {
	switch data := v.(type) {
	case float64:
		field.ReflectValueOf(value).SetFloat(data)
	case float32:
		field.ReflectValueOf(value).SetFloat(float64(data))
	case int64:
		field.ReflectValueOf(value).SetFloat(float64(data))
	case int:
		field.ReflectValueOf(value).SetFloat(float64(data))
	case int8:
		field.ReflectValueOf(value).SetFloat(float64(data))
	case int16:
		field.ReflectValueOf(value).SetFloat(float64(data))
	case int32:
		field.ReflectValueOf(value).SetFloat(float64(data))
	case uint:
		field.ReflectValueOf(value).SetFloat(float64(data))
	case uint8:
		field.ReflectValueOf(value).SetFloat(float64(data))
	case uint16:
		field.ReflectValueOf(value).SetFloat(float64(data))
	case uint32:
		field.ReflectValueOf(value).SetFloat(float64(data))
	case uint64:
		field.ReflectValueOf(value).SetFloat(float64(data))
	case []byte:
		return field.Set(value, string(data))
	case string:
		if data == "" {
			data = "0"
		}
		if i, err := strconv.ParseFloat(data, 64); err == nil {
			field.ReflectValueOf(value).SetFloat(i)
		} else {
			return err
		}
	default:
		return field.fallbackSetter(value, v, field.Set)
	}
	return nil
}

func (field *Field) SetString(value reflect.Value, v interface{}) error {
	switch data := v.(type) {
	case string:
		field.ReflectValueOf(value).SetString(data)
	case []byte:
		field.ReflectValueOf(value).SetString(string(data))
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		field.ReflectValueOf(value).SetString(ToString(data))
	case float64, float32:
		field.ReflectValueOf(value).SetString(fmt.Sprintf("%v", data))
	default:
		return field.fallbackSetter(value, v, field.Set)
	}
	return nil
}
